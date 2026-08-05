package http

import (
	"database/sql"
	"errors"
	stdhttp "net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/service"
)

type LeadHandler struct {
	leadService     *service.LeadService
	auditLogService *service.AuditLogService
}

func NewLeadHandler(leadService *service.LeadService, auditLogService *service.AuditLogService) *LeadHandler {
	return &LeadHandler{
		leadService:     leadService,
		auditLogService: auditLogService,
	}
}

func (h *LeadHandler) List(ctx *gin.Context) {
	leads, err := h.leadService.List()
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5301, "failed to list leads")
		return
	}

	OK(ctx, leads)
}

func (h *LeadHandler) Detail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4303, "invalid lead id")
		return
	}

	lead, getErr := h.leadService.GetByID(id)
	if getErr != nil {
		if errors.Is(getErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 4341, "lead not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5302, "failed to get lead")
		return
	}

	OK(ctx, lead)
}

func (h *LeadHandler) Create(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	var req service.LeadInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4301, "invalid request body")
		return
	}

	id, err := h.leadService.Create(req)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4304, err.Error())
		return
	}

	lead, getErr := h.leadService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5303, "created but failed to fetch lead")
		return
	}

	ctx.JSON(stdhttp.StatusCreated, gin.H{
		"code":    0,
		"message": "ok",
		"data":    lead,
	})

	recordAuditLog(ctx, h.auditLogService, operator, "create", "lead", toInt64Pointer(lead.ID), lead.Name, gin.H{
		"source_type": lead.SourceType,
		"status":      lead.Status,
	})
}

func (h *LeadHandler) UpdateStatus(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4303, "invalid lead id")
		return
	}

	var req service.LeadStatusInput
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4301, "invalid request body")
		return
	}

	if updateErr := h.leadService.UpdateStatus(id, req); updateErr != nil {
		if updateErr.Error() == "lead not found" {
			Fail(ctx, stdhttp.StatusNotFound, 4341, "lead not found")
			return
		}
		Fail(ctx, stdhttp.StatusBadRequest, 4305, updateErr.Error())
		return
	}

	lead, getErr := h.leadService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5304, "updated but failed to fetch lead")
		return
	}

	OK(ctx, lead)

	recordAuditLog(ctx, h.auditLogService, operator, "update_status", "lead", toInt64Pointer(lead.ID), lead.Name, gin.H{
		"status": lead.Status,
	})
}

func (h *LeadHandler) AddLog(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4303, "invalid lead id")
		return
	}

	var req service.LeadLogInput
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4301, "invalid request body")
		return
	}

	if addErr := h.leadService.AddLog(id, req, operator.ID); addErr != nil {
		if addErr.Error() == "lead not found" {
			Fail(ctx, stdhttp.StatusNotFound, 4341, "lead not found")
			return
		}
		Fail(ctx, stdhttp.StatusBadRequest, 4306, addErr.Error())
		return
	}

	lead, getErr := h.leadService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5305, "logged but failed to fetch lead")
		return
	}

	OK(ctx, lead)

	recordAuditLog(ctx, h.auditLogService, operator, "add_log", "lead", toInt64Pointer(lead.ID), lead.Name, gin.H{
		"action":  req.Action,
		"content": req.Content,
	})
}
