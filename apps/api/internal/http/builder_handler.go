package http

import (
	"database/sql"
	"errors"
	stdhttp "net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/service"
)

type BuilderHandler struct {
	builderService  *service.BuilderService
	auditLogService *service.AuditLogService
}

func NewBuilderHandler(builderService *service.BuilderService, auditLogService *service.AuditLogService) *BuilderHandler {
	return &BuilderHandler{
		builderService:  builderService,
		auditLogService: auditLogService,
	}
}

func (h *BuilderHandler) List(ctx *gin.Context) {
	builders, err := h.builderService.List()
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5201, "failed to list builders")
		return
	}

	OK(ctx, builders)
}

func (h *BuilderHandler) PublicList(ctx *gin.Context) {
	builders, err := h.builderService.ListPublished()
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5206, "failed to list published builders")
		return
	}

	OK(ctx, builders)
}

func (h *BuilderHandler) Detail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4203, "invalid builder id")
		return
	}

	builder, getErr := h.builderService.GetByID(id)
	if getErr != nil {
		if errors.Is(getErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 4241, "builder not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5202, "failed to get builder")
		return
	}

	OK(ctx, builder)
}

func (h *BuilderHandler) PublicDetail(ctx *gin.Context) {
	builder, getErr := h.builderService.GetPublishedBySlug(ctx.Param("slug"))
	if getErr != nil {
		if errors.Is(getErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 4242, "builder not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5207, "failed to get published builder")
		return
	}

	OK(ctx, builder)
}

func (h *BuilderHandler) Create(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	var req service.BuilderInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4201, "invalid request body")
		return
	}

	id, err := h.builderService.Create(req, operator.ID)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4204, err.Error())
		return
	}

	builder, getErr := h.builderService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5203, "created but failed to fetch builder")
		return
	}

	ctx.JSON(stdhttp.StatusCreated, gin.H{
		"code":    0,
		"message": "ok",
		"data":    builder,
	})

	recordAuditLog(ctx, h.auditLogService, operator, "create", "builder", toInt64Pointer(builder.ID), builder.Name, gin.H{
		"status": builder.Status,
		"slug":   builder.Slug,
	})
}

func (h *BuilderHandler) Update(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4203, "invalid builder id")
		return
	}

	var req service.BuilderInput
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4201, "invalid request body")
		return
	}

	if updateErr := h.builderService.Update(id, req, operator.ID); updateErr != nil {
		if updateErr.Error() == "builder not found" {
			Fail(ctx, stdhttp.StatusNotFound, 4241, "builder not found")
			return
		}
		Fail(ctx, stdhttp.StatusBadRequest, 4205, updateErr.Error())
		return
	}

	builder, getErr := h.builderService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5204, "updated but failed to fetch builder")
		return
	}

	OK(ctx, builder)

	recordAuditLog(ctx, h.auditLogService, operator, "update", "builder", toInt64Pointer(builder.ID), builder.Name, gin.H{
		"status": builder.Status,
		"slug":   builder.Slug,
	})
}

func (h *BuilderHandler) UpdateStatus(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4203, "invalid builder id")
		return
	}

	var req service.ContentStatusInput
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4201, "invalid request body")
		return
	}

	if updateErr := h.builderService.UpdateStatus(id, req, operator.ID); updateErr != nil {
		if updateErr.Error() == "builder not found" {
			Fail(ctx, stdhttp.StatusNotFound, 4241, "builder not found")
			return
		}
		Fail(ctx, stdhttp.StatusBadRequest, 4206, updateErr.Error())
		return
	}

	builder, getErr := h.builderService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5208, "updated but failed to fetch builder")
		return
	}

	OK(ctx, builder)

	recordAuditLog(ctx, h.auditLogService, operator, "update_status", "builder", toInt64Pointer(builder.ID), builder.Name, gin.H{
		"status": builder.Status,
	})
}

func (h *BuilderHandler) Delete(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4203, "invalid builder id")
		return
	}

	builder, getErr := h.builderService.GetByID(id)
	if getErr != nil {
		if errors.Is(getErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 4241, "builder not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5202, "failed to get builder")
		return
	}

	if deleteErr := h.builderService.Delete(id); deleteErr != nil {
		if deleteErr.Error() == "builder not found" {
			Fail(ctx, stdhttp.StatusNotFound, 4241, "builder not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5205, "failed to delete builder")
		return
	}

	OK(ctx, gin.H{"deleted": true, "id": id})

	recordAuditLog(ctx, h.auditLogService, operator, "delete", "builder", toInt64Pointer(id), builder.Name, gin.H{
		"slug": builder.Slug,
	})
}
