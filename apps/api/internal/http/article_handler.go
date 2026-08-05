package http

import (
	"database/sql"
	"errors"
	stdhttp "net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/service"
)

type ArticleHandler struct {
	articleService  *service.ArticleService
	auditLogService *service.AuditLogService
}

func NewArticleHandler(articleService *service.ArticleService, auditLogService *service.AuditLogService) *ArticleHandler {
	return &ArticleHandler{
		articleService:  articleService,
		auditLogService: auditLogService,
	}
}

func (h *ArticleHandler) List(ctx *gin.Context) {
	articles, err := h.articleService.List()
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5001, "failed to list articles")
		return
	}

	OK(ctx, articles)
}

func (h *ArticleHandler) PublicList(ctx *gin.Context) {
	articles, err := h.articleService.ListPublished()
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5006, "failed to list published articles")
		return
	}

	OK(ctx, articles)
}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4003, "invalid article id")
		return
	}

	article, getErr := h.articleService.GetByID(id)
	if getErr != nil {
		if errors.Is(getErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 4041, "article not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5002, "failed to get article")
		return
	}

	OK(ctx, article)
}

func (h *ArticleHandler) PublicDetail(ctx *gin.Context) {
	article, getErr := h.articleService.GetPublishedBySlug(ctx.Param("slug"))
	if getErr != nil {
		if errors.Is(getErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 4042, "article not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5007, "failed to get published article")
		return
	}

	OK(ctx, article)
}

func (h *ArticleHandler) Create(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	var req service.ArticleInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4001, "invalid request body")
		return
	}

	id, err := h.articleService.Create(req, operator.ID)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4004, err.Error())
		return
	}

	article, getErr := h.articleService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5003, "created but failed to fetch article")
		return
	}

	ctx.JSON(stdhttp.StatusCreated, gin.H{
		"code":    0,
		"message": "ok",
		"data":    article,
	})

	recordAuditLog(ctx, h.auditLogService, operator, "create", "article", toInt64Pointer(article.ID), article.Title, gin.H{
		"status": article.Status,
		"slug":   article.Slug,
	})
}

func (h *ArticleHandler) Update(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4003, "invalid article id")
		return
	}

	var req service.ArticleInput
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4001, "invalid request body")
		return
	}

	if updateErr := h.articleService.Update(id, req, operator.ID); updateErr != nil {
		if updateErr.Error() == "article not found" {
			Fail(ctx, stdhttp.StatusNotFound, 4041, "article not found")
			return
		}
		Fail(ctx, stdhttp.StatusBadRequest, 4005, updateErr.Error())
		return
	}

	article, getErr := h.articleService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5005, "updated but failed to fetch article")
		return
	}

	OK(ctx, article)

	recordAuditLog(ctx, h.auditLogService, operator, "update", "article", toInt64Pointer(article.ID), article.Title, gin.H{
		"status": article.Status,
		"slug":   article.Slug,
	})
}

func (h *ArticleHandler) UpdateStatus(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4003, "invalid article id")
		return
	}

	var req service.ContentStatusInput
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4001, "invalid request body")
		return
	}

	if updateErr := h.articleService.UpdateStatus(id, req, operator.ID); updateErr != nil {
		if updateErr.Error() == "article not found" {
			Fail(ctx, stdhttp.StatusNotFound, 4041, "article not found")
			return
		}
		Fail(ctx, stdhttp.StatusBadRequest, 4006, updateErr.Error())
		return
	}

	article, getErr := h.articleService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5008, "updated but failed to fetch article")
		return
	}

	OK(ctx, article)

	recordAuditLog(ctx, h.auditLogService, operator, "update_status", "article", toInt64Pointer(article.ID), article.Title, gin.H{
		"status": article.Status,
	})
}

func (h *ArticleHandler) Delete(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4003, "invalid article id")
		return
	}

	article, getErr := h.articleService.GetByID(id)
	if getErr != nil {
		if errors.Is(getErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 4041, "article not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5002, "failed to get article")
		return
	}

	if deleteErr := h.articleService.Delete(id); deleteErr != nil {
		if deleteErr.Error() == "article not found" {
			Fail(ctx, stdhttp.StatusNotFound, 4041, "article not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5004, "failed to delete article")
		return
	}

	OK(ctx, gin.H{"deleted": true, "id": id})

	recordAuditLog(ctx, h.auditLogService, operator, "delete", "article", toInt64Pointer(id), article.Title, gin.H{
		"slug": article.Slug,
	})
}
