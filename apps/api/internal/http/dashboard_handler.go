package http

import (
	stdhttp "net/http"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/service"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

func (h *DashboardHandler) Stats(ctx *gin.Context) {
	stats, err := h.dashboardService.Stats()
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5701, "failed to load dashboard stats")
		return
	}

	OK(ctx, stats)
}
