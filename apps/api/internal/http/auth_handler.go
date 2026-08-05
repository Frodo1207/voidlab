package http

import (
	stdhttp "net/http"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/http/middleware"
	"voidlabai/apps/api/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4001, "invalid request body")
		return
	}

	session, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4002, err.Error())
		return
	}

	OK(ctx, session)
}

func (h *AuthHandler) Me(ctx *gin.Context) {
	user, ok := middleware.CurrentUser(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}

	OK(ctx, user)
}

func (h *AuthHandler) Logout(ctx *gin.Context) {
	OK(ctx, gin.H{"logged_out": true})
}
