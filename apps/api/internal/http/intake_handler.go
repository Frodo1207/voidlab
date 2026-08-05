package http

import (
	"database/sql"
	"errors"
	"fmt"
	stdhttp "net/http"
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

	var req EventRSVPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4401, "invalid request body")
		return
	}

	if _, err := h.leadService.Create(service.LeadInput{
		SourceType: "event",
		SourceID:   &eventID,
		Name:       strings.TrimSpace(req.Name),
		Contact:    strings.TrimSpace(req.Contact),
		Message:    formatEventLeadMessage(event.Title, req.Message),
	}); err != nil {
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
