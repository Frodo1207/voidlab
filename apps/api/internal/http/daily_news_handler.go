package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"voidlabai/apps/api/internal/dailynews"
	"voidlabai/apps/api/internal/service"
)

type DailyNewsHandler struct {
	news  *dailynews.Service
	audit *service.AuditLogService
}

func (h *DailyNewsHandler) response(ctx *gin.Context, task dailynews.Task) gin.H {
	ids := make([]int64, 0, len(task.Articles))
	for _, article := range task.Articles {
		ids = append(ids, article.ID)
	}
	return gin.H{
		"date": task.Date, "status": task.Status, "send_status": task.SendStatus,
		"items": task.Snapshot.Items, "article_ids": ids, "message": task.Message, "copy_text": task.CopyText,
		"message_id": task.MessageID, "last_error": task.LastError,
	}
}

func (h *DailyNewsHandler) fail(ctx *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "每日资讯操作失败，请查看服务日志或稍后重试"
	switch {
	case errors.Is(err, dailynews.ErrNotFound):
		status, message = http.StatusNotFound, "未找到每日资讯任务"
	case errors.Is(err, dailynews.ErrInvalidInput):
		status, message = http.StatusBadRequest, "请求参数不合法"
	case errors.Is(err, dailynews.ErrInvalidState), errors.Is(err, dailynews.ErrSendInProgress):
		status, message = http.StatusConflict, "任务状态不允许此操作，或发送尚未结束"
	}
	Fail(ctx, status, 5601, message)
}

func (h *DailyNewsHandler) List(ctx *gin.Context) {
	tasks, err := h.news.List(ctx.Request.Context(), 30, 0)
	if err != nil {
		h.fail(ctx, err)
		return
	}
	result := make([]gin.H, 0, len(tasks))
	for _, task := range tasks {
		full, err := h.news.Get(ctx.Request.Context(), task.ID)
		if err != nil {
			h.fail(ctx, err)
			return
		}
		result = append(result, h.response(ctx, full))
	}
	OK(ctx, result)
}

func (h *DailyNewsHandler) Collect(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}
	var input struct {
		Date string `json:"date"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		Fail(ctx, http.StatusBadRequest, 4601, "invalid request body")
		return
	}
	today := time.Now().In(time.FixedZone("Asia/Shanghai", 8*60*60)).Format("2006-01-02")
	if input.Date == "" {
		input.Date = today
	}
	// The upstream window is rolling 24h, not a historical date query.
	if input.Date != today {
		Fail(ctx, http.StatusBadRequest, 4602, "只能采集今天的滚动24小时资讯")
		return
	}
	task, err := h.news.Generate(ctx.Request.Context(), input.Date)
	if err != nil {
		h.fail(ctx, err)
		return
	}
	OK(ctx, h.response(ctx, task))
	recordAuditLog(ctx, h.audit, operator, "collect", "daily_news", toInt64Pointer(task.ID), task.Date, gin.H{"status": task.Status})
}

func (h *DailyNewsHandler) AppendItems(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}
	task, err := h.news.GetByDate(ctx.Request.Context(), ctx.Param("date"))
	if err != nil {
		h.fail(ctx, err)
		return
	}
	var input struct {
		Items []dailynews.Item `json:"items"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		Fail(ctx, http.StatusBadRequest, 4601, "invalid request body")
		return
	}
	task, err = h.news.AppendItems(ctx.Request.Context(), task.ID, input.Items)
	if err != nil {
		h.fail(ctx, err)
		return
	}
	OK(ctx, h.response(ctx, task))
	recordAuditLog(ctx, h.audit, operator, "append_items", "daily_news", toInt64Pointer(task.ID), task.Date, gin.H{
		"status": task.Status, "appended": len(input.Items), "article_count": len(task.Articles),
	})
}

func (h *DailyNewsHandler) Reset(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}
	task, err := h.news.GetByDate(ctx.Request.Context(), ctx.Param("date"))
	if err != nil {
		h.fail(ctx, err)
		return
	}
	articleIDs, err := h.news.Reset(ctx.Request.Context(), task.ID)
	if err != nil {
		h.fail(ctx, err)
		return
	}
	OK(ctx, gin.H{
		"date":                  task.Date,
		"status":                "reset",
		"deleted_article_ids":   articleIDs,
		"deleted_article_count": len(articleIDs),
	})
	recordAuditLog(ctx, h.audit, operator, "reset", "daily_news", toInt64Pointer(task.ID), task.Date, gin.H{
		"deleted_article_count": len(articleIDs), "previous_status": task.Status, "previous_send_status": task.SendStatus,
	})
}

func (h *DailyNewsHandler) Action(action string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		operator, ok := requireCurrentUser(ctx)
		if !ok {
			return
		}
		task, err := h.news.GetByDate(ctx.Request.Context(), ctx.Param("date"))
		if err != nil {
			h.fail(ctx, err)
			return
		}
		switch action {
		case "approve":
			task, err = h.news.Review(ctx.Request.Context(), task.ID, operator.ID, true, "")
		case "publish":
			task, err = h.news.Publish(ctx.Request.Context(), task.ID, operator.ID)
		case "send":
			task, err = h.news.Send(ctx.Request.Context(), task.ID)
		}
		if err != nil {
			h.fail(ctx, err)
			return
		}
		OK(ctx, h.response(ctx, task))
		recordAuditLog(ctx, h.audit, operator, action, "daily_news", toInt64Pointer(task.ID), task.Date, gin.H{"status": task.Status, "send_status": task.SendStatus})
	}
}
