package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	stdhttp "net/http"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/service"
)

type SiteConfigHandler struct {
	siteConfigService *service.SiteConfigService
	auditLogService   *service.AuditLogService
}

func NewSiteConfigHandler(siteConfigService *service.SiteConfigService, auditLogService *service.AuditLogService) *SiteConfigHandler {
	return &SiteConfigHandler{
		siteConfigService: siteConfigService,
		auditLogService:   auditLogService,
	}
}

func (h *SiteConfigHandler) PublicList(ctx *gin.Context) {
	records, err := h.siteConfigService.List()
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5501, "failed to list site configs")
		return
	}

	OK(ctx, buildSiteConfigMap(records))
}

func (h *SiteConfigHandler) List(ctx *gin.Context) {
	records, err := h.siteConfigService.List()
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5501, "failed to list site configs")
		return
	}

	OK(ctx, buildSiteConfigRecords(records))
}

func (h *SiteConfigHandler) Update(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	key := ctx.Param("key")

	var req service.SiteConfigUpsertInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4501, "invalid request body")
		return
	}

	if err := h.siteConfigService.Upsert(key, req, operator.ID); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4502, err.Error())
		return
	}

	record, err := h.siteConfigService.GetByKey(key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 4541, "site config not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5502, "updated but failed to fetch site config")
		return
	}

	OK(ctx, buildSiteConfigRecord(record))

	recordAuditLog(ctx, h.auditLogService, operator, "update", "site_config", nil, key, gin.H{
		"config_key": key,
		"updated_by": operator.Username,
	})
}

func buildSiteConfigMap(records []domain.SiteConfig) gin.H {
	payload := gin.H{}
	for _, record := range records {
		payload[record.ConfigKey] = parseSiteConfigValue(record.ConfigValueJSON)
	}

	return payload
}

func buildSiteConfigRecords(records []domain.SiteConfig) []gin.H {
	payload := make([]gin.H, 0, len(records))
	for _, record := range records {
		payload = append(payload, buildSiteConfigRecord(record))
	}

	return payload
}

func buildSiteConfigRecord(record domain.SiteConfig) gin.H {
	return gin.H{
		"id":           record.ID,
		"config_key":   record.ConfigKey,
		"config_value": parseSiteConfigValue(record.ConfigValueJSON),
		"updated_by":   record.UpdatedBy,
		"updated_at":   record.UpdatedAt,
	}
}

func parseSiteConfigValue(raw string) any {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return raw
	}

	return value
}
