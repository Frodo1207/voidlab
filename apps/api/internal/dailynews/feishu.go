package dailynews

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	feishuTokenURL     = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"
	feishuMessagesURL  = "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=chat_id"
	feishuTimeout      = 15 * time.Second
	feishuResponseSize = 1 << 20
)

// Sender sends text using a caller-supplied idempotency key and returns a message ID.
// Retries of the same logical message must reuse the same uuid.
type Sender interface {
	Send(ctx context.Context, text, uuid string) (string, error)
}

// FeishuSender uses FEISHU_APP_ID, FEISHU_APP_SECRET and FEISHU_CHAT_ID.
// Its zero value is ready to use and may be shared between goroutines.
type FeishuSender struct{}

var _ Sender = (*FeishuSender)(nil)

func (s *FeishuSender) Send(ctx context.Context, text, uuid string) (string, error) {
	if ctx == nil {
		return "", errors.New("feishu: context is required")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if strings.TrimSpace(text) == "" || !utf8.ValidString(text) {
		return "", errors.New("feishu: valid non-empty text is required")
	}
	// Feishu accepts at most 50 characters; a byte limit also accepts standard UUIDs.
	if strings.TrimSpace(uuid) == "" || len(uuid) > 50 || !utf8.ValidString(uuid) {
		return "", errors.New("feishu: uuid must contain 1 to 50 bytes")
	}

	appID := strings.TrimSpace(os.Getenv("FEISHU_APP_ID"))
	appSecret := strings.TrimSpace(os.Getenv("FEISHU_APP_SECRET"))
	chatID := strings.TrimSpace(os.Getenv("FEISHU_CHAT_ID"))
	if appID == "" || appSecret == "" || chatID == "" {
		return "", errors.New("feishu: FEISHU_APP_ID, FEISHU_APP_SECRET and FEISHU_CHAT_ID are required")
	}

	// One deadline covers both authentication and message delivery.
	ctx, cancel := context.WithTimeout(ctx, feishuTimeout)
	defer cancel()
	client := &http.Client{
		Timeout: feishuTimeout,
		// Never forward credentials or bearer tokens to a redirect destination.
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	var auth struct {
		Code  *int   `json:"code"`
		Token string `json:"tenant_access_token"`
	}
	err := feishuPost(ctx, client, feishuTokenURL, "", struct {
		AppID     string `json:"app_id"`
		AppSecret string `json:"app_secret"`
	}{appID, appSecret}, &auth)
	if err != nil {
		return "", err
	}
	if err := feishuCheckCode(auth.Code); err != nil {
		return "", err
	}
	if auth.Token == "" || strings.ContainsAny(auth.Token, " \t\r\n") {
		return "", errors.New("feishu: invalid authentication response")
	}

	// The API expects content to be a JSON-encoded string, not a nested object.
	content, err := json.Marshal(struct {
		Text string `json:"text"`
	}{text})
	if err != nil {
		return "", errors.New("feishu: cannot encode message")
	}
	var result struct {
		Code *int `json:"code"`
		Data struct {
			MessageID string `json:"message_id"`
		} `json:"data"`
	}
	err = feishuPost(ctx, client, feishuMessagesURL, auth.Token, struct {
		ReceiveID string `json:"receive_id"`
		MsgType   string `json:"msg_type"`
		Content   string `json:"content"`
		UUID      string `json:"uuid"`
	}{chatID, "text", string(content), uuid}, &result)
	if err != nil {
		return "", err
	}
	if err := feishuCheckCode(result.Code); err != nil {
		return "", err
	}
	if result.Data.MessageID == "" {
		return "", errors.New("feishu: missing message ID")
	}
	return result.Data.MessageID, nil
}

func feishuPost(ctx context.Context, client *http.Client, endpoint, token string, payload, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return errors.New("feishu: cannot encode request")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return errors.New("feishu: cannot create request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		// Network errors can contain URLs or proxy credentials. Do not wrap them.
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return errors.New("feishu: request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("feishu: HTTP status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, feishuResponseSize+1))
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return errors.New("feishu: cannot read response")
	}
	if len(data) > feishuResponseSize {
		return errors.New("feishu: response exceeds size limit")
	}
	// Neither decoding errors nor remote error messages are safe to expose.
	if err := json.Unmarshal(data, result); err != nil {
		return errors.New("feishu: invalid JSON response")
	}
	return nil
}

func feishuCheckCode(code *int) error {
	if code == nil {
		return errors.New("feishu: missing API response code")
	}
	if *code != 0 {
		return fmt.Errorf("feishu: API error code %d", *code)
	}
	return nil
}
