package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	selectedItemsURL = "https://aihot.virxact.com/api/v1/items?mode=selected&window=24h&limit=8"
	allItemsURL      = "https://aihot.virxact.com/api/v1/items?mode=all&window=24h&limit=16"
)

type runnerConfig struct {
	Enabled      bool
	APIBaseURL   string
	AgentToken   string
	RunTime      string
	TargetCount  int
	CoverCommand string
}

type apiEnvelope[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type item struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Source  struct {
		Name string `json:"name"`
	} `json:"source"`
	Links struct {
		Original string `json:"original"`
	} `json:"links"`
}

type dailyTask struct {
	Date       string  `json:"date"`
	Status     string  `json:"status"`
	SendStatus string  `json:"send_status"`
	Items      []item  `json:"items"`
	ArticleIDs []int64 `json:"article_ids"`
	Message    string  `json:"message"`
	CopyText   string  `json:"copy_text"`
	MessageID  string  `json:"message_id"`
	LastError  string  `json:"last_error"`
}

type article struct {
	ID         int64    `json:"id"`
	Title      string   `json:"title"`
	Slug       string   `json:"slug"`
	Summary    string   `json:"summary"`
	Category   string   `json:"category"`
	Audience   string   `json:"audience"`
	Tags       []string `json:"tags"`
	CoverURL   string   `json:"cover_url"`
	Content    string   `json:"content"`
	SourceName string   `json:"source_name"`
	SourceURL  string   `json:"source_url"`
	Featured   bool     `json:"featured"`
	Status     string   `json:"status"`
}

type mediaGallery struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	IsDefault bool   `json:"is_default"`
}

type mediaGalleryItem struct {
	ID          int64  `json:"id"`
	FileName    string `json:"file_name"`
	ObjectURL   string `json:"object_url"`
	ContentType string `json:"content_type"`
}

type mediaAsset struct {
	ID        int64  `json:"id"`
	ObjectURL string `json:"object_url"`
}

type snapshotResponse struct {
	SchemaVersion int    `json:"schemaVersion"`
	Items         []item `json:"items"`
}

type runSummary struct {
	Date           string   `json:"date"`
	Collect        string   `json:"collect"`
	TopUp          string   `json:"top_up"`
	Approve        string   `json:"approve"`
	Publish        string   `json:"publish"`
	GenerateCovers string   `json:"generate_covers"`
	UploadAssets   string   `json:"upload_assets"`
	BindCovers     string   `json:"bind_covers"`
	BuildDigest    string   `json:"build_digest"`
	SendFeishu     string   `json:"send_feishu"`
	Failures       []string `json:"failures,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
	Complete       bool     `json:"complete"`
}

type coverRunSummary struct {
	Generate string
	Upload   string
	Bind     string
	Failures []string
	Warnings []string
	Complete bool
}

type apiClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	c := &apiClient{
		baseURL: strings.TrimRight(cfg.APIBaseURL, "/"),
		token:   strings.TrimSpace(cfg.AgentToken),
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	beijing := time.FixedZone("Asia/Shanghai", 8*60*60)
	lastSuccessDate := ""
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		now := time.Now().In(beijing)
		if cfg.Enabled && now.Format("15:04") >= cfg.RunTime && lastSuccessDate != now.Format("2006-01-02") {
			summary := runOnce(ctx, c, cfg, now.Format("2006-01-02"))
			printSummary(summary)
			if summary.Complete {
				lastSuccessDate = summary.Date
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func loadConfig() (runnerConfig, error) {
	target := 5
	if raw := strings.TrimSpace(os.Getenv("DAILY_NEWS_TARGET_COUNT")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 8 {
			return runnerConfig{}, errors.New("DAILY_NEWS_TARGET_COUNT must be 1..8")
		}
		target = value
	}
	runTime := strings.TrimSpace(os.Getenv("DAILY_NEWS_RUNNER_TIME"))
	if runTime == "" {
		runTime = "10:30"
	}
	if parsed, err := time.Parse("15:04", runTime); err != nil || parsed.Format("15:04") != runTime {
		return runnerConfig{}, errors.New("DAILY_NEWS_RUNNER_TIME must be HH:MM")
	}
	baseURL := strings.TrimSpace(os.Getenv("VOIDLAB_API_BASE_URL"))
	if baseURL == "" {
		baseURL = "http://api:8080"
	}
	token := strings.TrimSpace(os.Getenv("VOIDLAB_AGENT_TOKEN"))
	if token == "" {
		log.Printf("daily-news runner: VOIDLAB_AGENT_TOKEN is empty; runner will remain idle")
	}
	return runnerConfig{
		Enabled:      strings.EqualFold(strings.TrimSpace(os.Getenv("DAILY_NEWS_RUNNER_ENABLED")), "true"),
		APIBaseURL:   baseURL,
		AgentToken:   token,
		RunTime:      runTime,
		TargetCount:  target,
		CoverCommand: strings.TrimSpace(os.Getenv("DAILY_NEWS_COVER_COMMAND")),
	}, nil
}

func runOnce(ctx context.Context, c *apiClient, cfg runnerConfig, date string) runSummary {
	summary := runSummary{
		Date:           date,
		Collect:        "pending",
		TopUp:          "skip",
		Approve:        "skip",
		Publish:        "skip",
		GenerateCovers: "skip",
		UploadAssets:   "skip",
		BindCovers:     "skip",
		BuildDigest:    "skip",
		SendFeishu:     "skip",
	}
	if strings.TrimSpace(cfg.AgentToken) == "" {
		summary.Failures = append(summary.Failures, "missing VOIDLAB_AGENT_TOKEN")
		return summary
	}

	if err := c.health(ctx); err != nil {
		summary.Failures = append(summary.Failures, "healthz failed: "+err.Error())
		return summary
	}
	if err := c.authMe(ctx); err != nil {
		summary.Failures = append(summary.Failures, "auth handshake failed: "+err.Error())
		return summary
	}

	task, err := c.collectTask(ctx, date)
	if err != nil {
		summary.Collect = "failed"
		summary.Failures = append(summary.Failures, "collect failed: "+err.Error())
		return summary
	}
	summary.Collect = fmt.Sprintf("ok (%d items)", len(task.Items))

	if len(task.Items) < cfg.TargetCount {
		added, topErr := c.topUpTask(ctx, task, cfg.TargetCount)
		if topErr != nil {
			summary.TopUp = "failed"
			summary.Failures = append(summary.Failures, "top-up failed: "+topErr.Error())
		} else {
			task, _ = c.collectTask(ctx, date)
			summary.TopUp = fmt.Sprintf("ok (+%d, total %d)", added, len(task.Items))
		}
	} else {
		summary.TopUp = "skip"
	}

	switch task.Status {
	case "pending":
		task, err = c.taskAction(ctx, date, "approve")
		if err != nil {
			summary.Approve = "failed"
			summary.Failures = append(summary.Failures, "approve failed: "+err.Error())
			return summary
		}
		summary.Approve = "ok"
	case "approved":
		summary.Approve = "already-approved"
	case "published":
		summary.Approve = "already-approved"
	case "rejected":
		summary.Approve = "blocked (rejected)"
		summary.Failures = append(summary.Failures, "approve blocked: task is rejected")
		return summary
	default:
		summary.Approve = "blocked (" + task.Status + ")"
		summary.Failures = append(summary.Failures, "approve blocked: unexpected task status "+task.Status)
		return summary
	}

	if task.Status == "approved" {
		task, err = c.taskAction(ctx, date, "publish")
		if err != nil {
			summary.Publish = "failed"
			summary.Failures = append(summary.Failures, "publish failed: "+err.Error())
			return summary
		}
		summary.Publish = "ok"
	} else if task.Status == "published" {
		summary.Publish = "already-published"
	}

	coverSummary := c.ensureCovers(ctx, cfg, task.ArticleIDs)
	summary.GenerateCovers = coverSummary.Generate
	summary.UploadAssets = coverSummary.Upload
	summary.BindCovers = coverSummary.Bind
	summary.Failures = append(summary.Failures, coverSummary.Failures...)
	summary.Warnings = append(summary.Warnings, coverSummary.Warnings...)

	digestStatus, digestFailures := ensureDigest(task)
	summary.BuildDigest = digestStatus
	summary.Failures = append(summary.Failures, digestFailures...)

	if task.SendStatus != "sent" {
		task, err = c.taskAction(ctx, date, "send")
		if err != nil {
			summary.SendFeishu = "failed"
			summary.Failures = append(summary.Failures, "send failed: "+err.Error())
		} else {
			summary.SendFeishu = "ok"
		}
	} else {
		summary.SendFeishu = "already-sent"
	}

	summary.Complete = len(summary.Failures) == 0 &&
		hasStatus(summary.Collect, "ok") &&
		(summary.Publish == "ok" || summary.Publish == "already-published") &&
		coverSummary.Complete &&
		summary.BuildDigest == "ok" &&
		(summary.SendFeishu == "ok" || summary.SendFeishu == "already-sent")
	return summary
}

func printSummary(summary runSummary) {
	data, err := json.Marshal(summary)
	if err != nil {
		log.Printf("daily-news runner summary marshal failed: %v", err)
		return
	}
	log.Printf("daily-news runner: %s", data)
}

func (c *apiClient) health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/healthz", nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

func (c *apiClient) authMe(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/auth/me", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

func (c *apiClient) collectTask(ctx context.Context, date string) (dailyTask, error) {
	payload := map[string]string{"date": date}
	var task dailyTask
	err := c.doJSON(ctx, http.MethodPost, "/api/v1/daily-news/collect", payload, &task)
	return task, err
}

func (c *apiClient) taskAction(ctx context.Context, date, action string) (dailyTask, error) {
	var task dailyTask
	err := c.doJSON(ctx, http.MethodPost, "/api/v1/daily-news/"+date+"/"+action, nil, &task)
	return task, err
}

func (c *apiClient) topUpTask(ctx context.Context, task dailyTask, target int) (int, error) {
	snapshot, err := fetchItems(ctx, allItemsURL)
	if err != nil {
		return 0, err
	}
	existingIDs := make(map[string]struct{}, len(task.Items))
	existingURLs := make(map[string]struct{}, len(task.Items))
	for _, existing := range task.Items {
		existingIDs[existing.ID] = struct{}{}
		existingURLs[existing.Links.Original] = struct{}{}
	}
	additions := make([]item, 0, target-len(task.Items))
	for _, candidate := range snapshot.Items {
		if len(task.Items)+len(additions) >= target {
			break
		}
		if _, ok := existingIDs[candidate.ID]; ok {
			continue
		}
		if _, ok := existingURLs[candidate.Links.Original]; ok {
			continue
		}
		additions = append(additions, candidate)
		existingIDs[candidate.ID] = struct{}{}
		existingURLs[candidate.Links.Original] = struct{}{}
	}
	if len(additions) == 0 {
		return 0, nil
	}
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/daily-news/"+task.Date+"/items", map[string]any{"items": additions}, &task); err != nil {
		return 0, err
	}
	return len(additions), nil
}

func fetchItems(ctx context.Context, endpoint string) (snapshotResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return snapshotResponse{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "voidlab-dailynews-runner/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return snapshotResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return snapshotResponse{}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return snapshotResponse{}, err
	}
	var snapshot snapshotResponse
	if err := json.Unmarshal(body, &snapshot); err != nil {
		return snapshotResponse{}, err
	}
	return snapshot, nil
}

func (c *apiClient) ensureCovers(ctx context.Context, cfg runnerConfig, articleIDs []int64) coverRunSummary {
	if len(articleIDs) == 0 {
		return coverRunSummary{
			Generate: "skip",
			Upload:   "skip",
			Bind:     "skip",
			Complete: true,
		}
	}
	failures := make([]string, 0)
	warnings := make([]string, 0)
	alreadyCoveredCount := 0
	boundCount := 0
	reusedCount := 0
	generatedCount := 0
	generateFailed := 0
	uploadedCount := 0
	uploadFailed := 0
	bindFailed := 0
	missingCount := 0
	var galleryItems []mediaGalleryItem
	galleryLoaded := false
	loadGallery := func() ([]mediaGalleryItem, error) {
		if galleryLoaded {
			return galleryItems, nil
		}
		_, items, err := c.loadBaseGallery(ctx)
		if err != nil {
			return nil, err
		}
		galleryItems = items
		galleryLoaded = true
		return galleryItems, nil
	}
	for _, articleID := range articleIDs {
		record, err := c.getArticle(ctx, articleID)
		if err != nil {
			bindFailed++
			failures = append(failures, fmt.Sprintf("bind cover failed for article %d: read article failed: %v", articleID, err))
			continue
		}
		if strings.TrimSpace(record.CoverURL) != "" {
			alreadyCoveredCount++
			continue
		}
		missingCount++
		coverURL := ""
		if cfg.CoverCommand != "" {
			outputPath, cleanup, genErr := generateCoverFile(ctx, cfg.CoverCommand, record)
			if genErr != nil {
				generateFailed++
				warnings = append(warnings, fmt.Sprintf("generate cover failed for %s: %v", record.Slug, genErr))
			} else {
				generatedCount++
				asset, uploadErr := c.uploadMedia(ctx, outputPath)
				cleanup()
				if uploadErr != nil {
					uploadFailed++
					warnings = append(warnings, fmt.Sprintf("upload asset failed for %s: %v", record.Slug, uploadErr))
				} else {
					uploadedCount++
					coverURL = asset.ObjectURL
				}
			}
		}
		if coverURL == "" {
			items, err := loadGallery()
			if err != nil {
				bindFailed++
				failures = append(failures, fmt.Sprintf("bind cover failed for %s: load base-gallery failed: %v", record.Slug, err))
				continue
			}
			match := pickGalleryCover(items, record.Slug)
			if strings.TrimSpace(match.ObjectURL) == "" {
				bindFailed++
				failures = append(failures, fmt.Sprintf("bind cover failed for %s: no base-gallery asset", record.Slug))
				continue
			}
			coverURL = match.ObjectURL
			reusedCount++
		}
		record.CoverURL = coverURL
		if err := c.updateArticle(ctx, record); err != nil {
			bindFailed++
			failures = append(failures, fmt.Sprintf("bind cover failed for %s: %v", record.Slug, err))
			continue
		}
		boundCount++
	}
	generateStatus := "skip"
	switch {
	case missingCount == 0:
		generateStatus = "already-complete"
	case cfg.CoverCommand == "":
		generateStatus = fmt.Sprintf("skip (no command, %d gallery fallback)", reusedCount)
	case generateFailed == 0:
		generateStatus = fmt.Sprintf("ok (%d generated)", generatedCount)
	case generatedCount == 0:
		generateStatus = fmt.Sprintf("failed (%d failed)", generateFailed)
	default:
		generateStatus = fmt.Sprintf("fallback (%d generated, %d failed)", generatedCount, generateFailed)
	}
	uploadStatus := "skip"
	switch {
	case missingCount == 0:
		uploadStatus = "already-complete"
	case generatedCount == 0 && uploadFailed == 0:
		if reusedCount > 0 {
			uploadStatus = fmt.Sprintf("skip (%d gallery reused)", reusedCount)
		}
	case uploadFailed == 0:
		uploadStatus = fmt.Sprintf("ok (%d uploaded)", uploadedCount)
	case uploadedCount == 0:
		uploadStatus = fmt.Sprintf("failed (%d failed)", uploadFailed)
	default:
		uploadStatus = fmt.Sprintf("fallback (%d uploaded, %d failed)", uploadedCount, uploadFailed)
	}
	bindStatus := "skip"
	switch {
	case missingCount == 0:
		bindStatus = "already-complete"
	case bindFailed == 0:
		bindStatus = fmt.Sprintf("ok (%d bound, %d gallery reused, %d already had covers)", boundCount, reusedCount, alreadyCoveredCount)
	case boundCount == 0:
		bindStatus = fmt.Sprintf("failed (%d failed)", bindFailed)
	default:
		bindStatus = fmt.Sprintf("partial (%d bound, %d failed)", boundCount, bindFailed)
	}
	return coverRunSummary{
		Generate: generateStatus,
		Upload:   uploadStatus,
		Bind:     bindStatus,
		Failures: failures,
		Warnings: warnings,
		Complete: bindFailed == 0,
	}
}

func generateCoverFile(ctx context.Context, command string, record article) (string, func(), error) {
	tempDir, err := os.MkdirTemp("", "voidlab-dailynews-cover")
	if err != nil {
		return "", nil, err
	}
	outputPath := filepath.Join(tempDir, record.Slug+".jpg")
	cmd := exec.CommandContext(ctx, "/bin/sh", "-lc", command)
	cmd.Env = append(os.Environ(),
		"VOIDLAB_DAILY_NEWS_TITLE="+record.Title,
		"VOIDLAB_DAILY_NEWS_SUMMARY="+record.Summary,
		"VOIDLAB_DAILY_NEWS_SLUG="+record.Slug,
		"VOIDLAB_DAILY_NEWS_OUTPUT="+outputPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tempDir)
		return "", nil, fmt.Errorf("%v: %s", err, strings.TrimSpace(string(output)))
	}
	if _, err := os.Stat(outputPath); err != nil {
		os.RemoveAll(tempDir)
		return "", nil, fmt.Errorf("generator did not create %s", outputPath)
	}
	return outputPath, func() { _ = os.RemoveAll(tempDir) }, nil
}

func ensureDigest(task dailyTask) (string, []string) {
	if task.Status != "published" {
		return "skip", nil
	}
	failures := make([]string, 0, 3)
	if strings.TrimSpace(task.Message) == "" {
		failures = append(failures, "build digest failed: reading digest is empty")
	}
	if strings.TrimSpace(task.CopyText) == "" {
		failures = append(failures, "build digest failed: copy digest is empty")
	}
	if len(task.ArticleIDs) > 0 && !strings.Contains(task.CopyText, "/insights/") {
		failures = append(failures, "build digest failed: copy digest does not contain VOIDLAB insights links")
	}
	if mentionsAIHOT(task.Message) || mentionsAIHOT(task.CopyText) {
		failures = append(failures, "build digest failed: digest still contains forbidden AI HOT branding")
	}
	if len(failures) > 0 {
		return "failed", failures
	}
	return "ok", nil
}

func mentionsAIHOT(text string) bool {
	normalized := strings.ToLower(strings.TrimSpace(text))
	return strings.Contains(normalized, "ai hot") || strings.Contains(normalized, "aihot")
}

func hasStatus(status, prefix string) bool {
	return strings.HasPrefix(status, prefix)
}

func pickGalleryCover(items []mediaGalleryItem, slug string) mediaGalleryItem {
	pool := make([]mediaGalleryItem, 0, len(items))
	for _, item := range items {
		raw := strings.ToLower(strings.TrimSpace(item.ObjectURL + " " + item.FileName))
		if raw == "" || strings.Contains(raw, ".svg") {
			continue
		}
		pool = append(pool, item)
		if strings.Contains(raw, strings.ToLower(slug)) {
			return item
		}
	}
	if len(pool) == 0 {
		return mediaGalleryItem{}
	}
	sum := sha256.Sum256([]byte(slug))
	index := int(sum[0]) % len(pool)
	return pool[index]
}

func (c *apiClient) loadBaseGallery(ctx context.Context) (mediaGallery, []mediaGalleryItem, error) {
	var galleries []mediaGallery
	if err := c.doJSON(ctx, http.MethodGet, "/api/v1/media/galleries", nil, &galleries); err != nil {
		return mediaGallery{}, nil, err
	}
	var base mediaGallery
	for _, gallery := range galleries {
		if gallery.IsDefault || gallery.Slug == "base-gallery" {
			base = gallery
			break
		}
	}
	if base.ID == 0 {
		return mediaGallery{}, nil, errors.New("base-gallery not found")
	}
	var items []mediaGalleryItem
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/media/galleries/%d/items", base.ID), nil, &items); err != nil {
		return mediaGallery{}, nil, err
	}
	return base, items, nil
}

func (c *apiClient) getArticle(ctx context.Context, id int64) (article, error) {
	var record article
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/articles/%d", id), nil, &record)
	return record, err
}

func (c *apiClient) updateArticle(ctx context.Context, record article) error {
	payload := map[string]any{
		"title":       record.Title,
		"slug":        record.Slug,
		"summary":     record.Summary,
		"category":    record.Category,
		"audience":    record.Audience,
		"tags":        record.Tags,
		"cover_url":   record.CoverURL,
		"content":     record.Content,
		"source_name": record.SourceName,
		"source_url":  record.SourceURL,
		"featured":    record.Featured,
		"status":      record.Status,
	}
	var updated article
	return c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/articles/%d", record.ID), payload, &updated)
}

func (c *apiClient) uploadMedia(ctx context.Context, path string) (mediaAsset, error) {
	file, err := os.Open(path)
	if err != nil {
		return mediaAsset{}, err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return mediaAsset{}, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return mediaAsset{}, err
	}
	if err := writer.Close(); err != nil {
		return mediaAsset{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/media/upload", &body)
	if err != nil {
		return mediaAsset{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := c.client.Do(req)
	if err != nil {
		return mediaAsset{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return mediaAsset{}, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return mediaAsset{}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	var envelope apiEnvelope[mediaAsset]
	if err := json.Unmarshal(data, &envelope); err != nil {
		return mediaAsset{}, err
	}
	if envelope.Code != 0 {
		return mediaAsset{}, fmt.Errorf("api code %d: %s", envelope.Code, envelope.Message)
	}
	return envelope.Data, nil
}

func (c *apiClient) doJSON(ctx context.Context, method, path string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	if out == nil {
		return nil
	}
	var envelope apiEnvelope[json.RawMessage]
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	if envelope.Code != 0 {
		return fmt.Errorf("api code %d: %s", envelope.Code, envelope.Message)
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil
	}
	return json.Unmarshal(envelope.Data, out)
}
