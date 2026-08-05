package http

import (
	"errors"
	stdhttp "net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/service"
)

type UserHandler struct {
	userService     *service.UserService
	auditLogService *service.AuditLogService
}

func NewUserHandler(userService *service.UserService, auditLogService *service.AuditLogService) *UserHandler {
	return &UserHandler{
		userService:     userService,
		auditLogService: auditLogService,
	}
}

func (h *UserHandler) List(ctx *gin.Context) {
	users, err := h.userService.List()
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5801, "failed to list users")
		return
	}

	OK(ctx, users)
}

func (h *UserHandler) Create(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	var req service.UserCreateInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4801, "invalid request body")
		return
	}

	id, err := h.userService.Create(req)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4802, err.Error())
		return
	}

	user, getErr := h.userService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5802, "created but failed to fetch user")
		return
	}

	OK(ctx, user)
	recordAuditLog(ctx, h.auditLogService, operator, "create", "user", toInt64Pointer(user.ID), user.Username, gin.H{
		"role":      user.Role,
		"is_active": user.IsActive,
	})
}

func (h *UserHandler) UpdateRole(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4803, "invalid user id")
		return
	}

	var req service.UserRoleInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4804, "invalid request body")
		return
	}

	if err := h.userService.UpdateRole(id, req, operator.ID); err != nil {
		handleUserMutationError(ctx, err)
		return
	}

	user, getErr := h.userService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5803, "updated but failed to fetch user")
		return
	}

	OK(ctx, user)
	recordAuditLog(ctx, h.auditLogService, operator, "update_role", "user", toInt64Pointer(user.ID), user.Username, gin.H{
		"role": user.Role,
	})
}

func (h *UserHandler) UpdateStatus(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4803, "invalid user id")
		return
	}

	var req service.UserStatusInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4805, "invalid request body")
		return
	}

	if err := h.userService.UpdateStatus(id, req, operator.ID); err != nil {
		handleUserMutationError(ctx, err)
		return
	}

	user, getErr := h.userService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5804, "updated but failed to fetch user")
		return
	}

	OK(ctx, user)
	recordAuditLog(ctx, h.auditLogService, operator, "update_status", "user", toInt64Pointer(user.ID), user.Username, gin.H{
		"is_active": user.IsActive,
	})
}

func (h *UserHandler) ResetPassword(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4803, "invalid user id")
		return
	}

	var req service.UserPasswordInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4806, "invalid request body")
		return
	}

	if err := h.userService.ResetPassword(id, req); err != nil {
		handleUserMutationError(ctx, err)
		return
	}

	user, getErr := h.userService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5805, "updated but failed to fetch user")
		return
	}

	OK(ctx, gin.H{"updated": true, "id": user.ID})
	recordAuditLog(ctx, h.auditLogService, operator, "reset_password", "user", toInt64Pointer(user.ID), user.Username, gin.H{})
}

func handleUserMutationError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, stdhttp.ErrBodyNotAllowed):
		Fail(ctx, stdhttp.StatusBadRequest, 4807, "invalid request")
	case err.Error() == "user not found":
		Fail(ctx, stdhttp.StatusNotFound, 4841, "user not found")
	default:
		Fail(ctx, stdhttp.StatusBadRequest, 4808, err.Error())
	}
}
