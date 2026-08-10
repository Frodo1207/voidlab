package http

import (
	"database/sql"
	"errors"
	"fmt"
	stdhttp "net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/service"
)

type IntakeHandler struct {
	leadService    *service.LeadService
	eventService   *service.EventService
	builderService *service.BuilderService
}

type ContactSubmitRequest struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Message string `json:"message"`
}

type EventRSVPRequest struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Message string `json:"message"`
}

type BuilderInquiryRequest struct {
	BuilderSlug string `json:"builder_slug"`
	BuilderName string `json:"builder_name"`
	Name        string `json:"name"`
	Contact     string `json:"contact"`
	Message     string `json:"message"`
}

var nonDigitPattern = regexp.MustCompile(`\D`)
var labeledPhonePattern = regexp.MustCompile(`手机\s*[:：]\s*([^\n/]+)`)

func NewIntakeHandler(
	leadService *service.LeadService,
	eventService *service.EventService,
	builderService *service.BuilderService,
) *IntakeHandler {
	return &IntakeHandler{
		leadService:    leadService,
		eventService:   eventService,
		builderService: builderService,
	}
}

func (h *IntakeHandler) ContactSubmit(ctx *gin.Context) {
	var req ContactSubmitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4401, "invalid request body")
		return
	}

	if _, err := h.leadService.Create(service.LeadInput{
		SourceType: "contact",
		Name:       strings.TrimSpace(req.Name),
		Contact:    strings.TrimSpace(req.Contact),
		Message:    strings.TrimSpace(req.Message),
	}); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4402, err.Error())
		return
	}

	OK(ctx, gin.H{"submitted": true})
}

func (h *IntakeHandler) EventRSVP(ctx *gin.Context) {
	eventID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4403, "invalid event id")
		return
	}

	event, getErr := h.eventService.GetByID(eventID)
	if getErr != nil {
		if errors.Is(getErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 4441, "event not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5401, "failed to get event")
		return
	}

	signupStatus := service.DeriveEventSignupStatus(event)
	if signupStatus != service.EventSignupStatusOpen && signupStatus != service.EventSignupStatusExternal {
		message := "当前活动暂不可报名"
		switch signupStatus {
		case service.EventSignupStatusNotStarted:
			message = "报名尚未开始"
		case service.EventSignupStatusClosed:
			message = "报名已关闭"
		case service.EventSignupStatusFull:
			message = "报名人数已满"
		case service.EventSignupStatusLive:
			message = "活动进行中，当前不接受报名"
		case service.EventSignupStatusEnded:
			message = "活动已结束"
		}
		Fail(ctx, stdhttp.StatusBadRequest, 4405, message)
		return
	}

	if signupStatus == service.EventSignupStatusExternal {
		Fail(ctx, stdhttp.StatusBadRequest, 4406, "当前活动需跳转外部报名链接")
		return
	}

	var req EventRSVPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4401, "invalid request body")
		return
	}

	if _, ok := extractNormalizedEventPhone(req.Contact); !ok {
		Fail(ctx, stdhttp.StatusBadRequest, 4408, "请输入有效的 11 位手机号")
		return
	}

	if _, err := h.leadService.Create(service.LeadInput{
		SourceType: "event",
		SourceID:   &eventID,
		Name:       strings.TrimSpace(req.Name),
		Contact:    strings.TrimSpace(req.Contact),
		Status:     "applied",
		Message:    formatEventLeadMessage(event.Title, req.Message),
	}); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "dedupe") {
			Fail(ctx, stdhttp.StatusBadRequest, 4407, "你已经报过这个活动了，我们这边会直接基于已有记录继续跟进")
			return
		}
		Fail(ctx, stdhttp.StatusBadRequest, 4402, err.Error())
		return
	}

	OK(ctx, gin.H{"submitted": true})
}

func (h *IntakeHandler) BuilderInquiry(ctx *gin.Context) {
	var req BuilderInquiryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4401, "invalid request body")
		return
	}

	sourceID, message, err := h.resolveBuilderInquiry(req)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4404, err.Error())
		return
	}

	if _, err := h.leadService.Create(service.LeadInput{
		SourceType: "builder",
		SourceID:   sourceID,
		Name:       strings.TrimSpace(req.Name),
		Contact:    strings.TrimSpace(req.Contact),
		Message:    message,
	}); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4402, err.Error())
		return
	}

	OK(ctx, gin.H{"submitted": true})
}

func (h *IntakeHandler) resolveBuilderInquiry(req BuilderInquiryRequest) (*int64, string, error) {
	builderSlug := strings.TrimSpace(req.BuilderSlug)
	builderName := strings.TrimSpace(req.BuilderName)
	message := strings.TrimSpace(req.Message)

	if builderSlug == "" && builderName == "" {
		return nil, "", errors.New("builder identity is required")
	}

	if builderSlug != "" {
		builders, err := h.builderService.List()
		if err != nil {
			return nil, "", fmt.Errorf("failed to load builders")
		}

		for _, builder := range builders {
			if builder.Slug == builderSlug {
				return &builder.ID, formatBuilderLeadMessage(builder.Name, builder.Slug, message), nil
			}
		}
	}

	if builderName == "" {
		builderName = builderSlug
	}

	return nil, formatBuilderLeadMessage(builderName, builderSlug, message), nil
}

func formatEventLeadMessage(eventTitle string, message string) string {
	parts := []string{}
	if eventTitle != "" {
		parts = append(parts, fmt.Sprintf("活动：%s", eventTitle))
	}
	if strings.TrimSpace(message) != "" {
		parts = append(parts, fmt.Sprintf("报名说明：%s", strings.TrimSpace(message)))
	}
	if len(parts) == 0 {
		return "用户提交了活动报名/预约意向。"
	}
	return strings.Join(parts, "\n")
}

func formatBuilderLeadMessage(builderName string, builderSlug string, message string) string {
	parts := []string{}
	if builderName != "" {
		parts = append(parts, fmt.Sprintf("目标 Builder：%s", builderName))
	}
	if builderSlug != "" {
		parts = append(parts, fmt.Sprintf("Builder Slug：%s", builderSlug))
	}
	if strings.TrimSpace(message) != "" {
		parts = append(parts, fmt.Sprintf("合作诉求：%s", strings.TrimSpace(message)))
	}
	if len(parts) == 0 {
		return "用户提交了 Builder 合作意向。"
	}
	return strings.Join(parts, "\n")
}

func extractNormalizedEventPhone(contact string) (string, bool) {
	trimmed := strings.TrimSpace(contact)
	if trimmed == "" {
		return "", false
	}

	candidates := []string{trimmed}
	if match := labeledPhonePattern.FindStringSubmatch(trimmed); len(match) == 2 {
		candidates = append([]string{match[1]}, candidates...)
	}

	for _, candidate := range candidates {
		normalized := nonDigitPattern.ReplaceAllString(candidate, "")
		if len(normalized) == 13 && strings.HasPrefix(normalized, "86") {
			normalized = normalized[2:]
		}
		if len(normalized) == 11 && strings.HasPrefix(normalized, "1") {
			secondDigit := normalized[1]
			if secondDigit >= '3' && secondDigit <= '9' {
				return normalized, true
			}
		}
	}

	return "", false
}
