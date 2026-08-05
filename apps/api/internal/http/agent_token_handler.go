package http

import (
	stdhttp "net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/service"
)

type AgentTokenHandler struct {
	agentTokenService *service.AgentTokenService
	auditLogService   *service.AuditLogService
}

func NewAgentTokenHandler(agentTokenService *service.AgentTokenService, auditLogService *service.AuditLogService) *AgentTokenHandler {
	return &AgentTokenHandler{
		agentTokenService: agentTokenService,
		auditLogService:   auditLogService,
	}
}

func (h *AgentTokenHandler) List(ctx *gin.Context) {
	records, err := h.agentTokenService.List()
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5901, "failed to list agent tokens")
		return
	}

	OK(ctx, records)
}

func (h *AgentTokenHandler) Create(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	var req service.AgentTokenCreateInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4901, "invalid request body")
		return
	}

	record, plainToken, err := h.agentTokenService.Create(req, operator.ID)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4902, err.Error())
		return
	}

	OK(ctx, gin.H{
		"record": record,
		"token":  plainToken,
	})

	recordAuditLog(ctx, h.auditLogService, operator, "create", "agent_token", toInt64Pointer(record.ID), record.Name, gin.H{
		"scopes": record.Scopes,
	})
}

func (h *AgentTokenHandler) UpdateStatus(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4903, "invalid agent token id")
		return
	}

	var req service.AgentTokenStatusInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4904, "invalid request body")
		return
	}

	if err := h.agentTokenService.UpdateStatus(id, req); err != nil {
		if err.Error() == "agent token not found" {
			Fail(ctx, stdhttp.StatusNotFound, 4941, "agent token not found")
			return
		}
		Fail(ctx, stdhttp.StatusBadRequest, 4905, err.Error())
		return
	}

	record, getErr := h.agentTokenService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5902, "updated but failed to fetch agent token")
		return
	}

	OK(ctx, record)
	recordAuditLog(ctx, h.auditLogService, operator, "update_status", "agent_token", toInt64Pointer(record.ID), record.Name, gin.H{
		"is_active": record.IsActive,
	})
}
