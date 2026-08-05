package http

import (
	stdhttp "net/http"
	"voidlabai/apps/api/internal/service"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	mediaService    *service.MediaService
	auditLogService *service.AuditLogService
}

func NewMediaHandler(mediaService *service.MediaService, auditLogService *service.AuditLogService) *MediaHandler {
	return &MediaHandler{
		mediaService:    mediaService,
		auditLogService: auditLogService,
	}
}

func (h *MediaHandler) List(ctx *gin.Context) {
	assets, err := h.mediaService.List()
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5301, "failed to list media assets")
		return
	}

	OK(ctx, assets)
}

func (h *MediaHandler) Upload(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4301, "file is required")
		return
	}

	record, uploadErr := h.mediaService.Upload(fileHeader, operator.ID)
	if uploadErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5302, uploadErr.Error())
		return
	}

	ctx.JSON(stdhttp.StatusCreated, gin.H{
		"code":    0,
		"message": "ok",
		"data":    record,
	})

	recordAuditLog(ctx, h.auditLogService, operator, "upload", "media_asset", toInt64Pointer(record.ID), record.FileName, gin.H{
		"content_type": record.ContentType,
		"object_url":   record.ObjectURL,
	})
}
