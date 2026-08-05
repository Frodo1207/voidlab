package http

import (
	stdhttp "net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/service"
)

type AuditLogHandler struct {
	auditLogService *service.AuditLogService
}

func NewAuditLogHandler(auditLogService *service.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{auditLogService: auditLogService}
}

func (h *AuditLogHandler) List(ctx *gin.Context) {
	limit := 100
	if rawLimit := ctx.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			Fail(ctx, stdhttp.StatusBadRequest, 4601, "invalid limit")
			return
		}
		limit = parsedLimit
	}

	records, err := h.auditLogService.List(limit)
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5601, "failed to list audit logs")
		return
	}

	payload := make([]gin.H, 0, len(records))
	for _, record := range records {
		payload = append(payload, gin.H{
			"id":             record.ID,
			"actor_type":     record.ActorType,
			"actor_id":       record.ActorID,
			"actor_username": record.ActorUsername,
			"actor_role":     record.ActorRole,
			"agent_token_id": record.AgentTokenID,
			"action":         record.Action,
			"entity_type":    record.EntityType,
			"entity_id":      record.EntityID,
			"entity_label":   record.EntityLabel,
			"detail":         parseSiteConfigValue(record.DetailJSON),
			"created_at":     record.CreatedAt,
		})
	}

	OK(ctx, payload)
}
