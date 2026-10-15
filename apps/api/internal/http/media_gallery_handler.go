package http

import (
	"database/sql"
	"errors"
	stdhttp "net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/service"
)

type MediaGalleryHandler struct {
	galleryService  *service.MediaGalleryService
	auditLogService *service.AuditLogService
}

func NewMediaGalleryHandler(galleryService *service.MediaGalleryService, auditLogService *service.AuditLogService) *MediaGalleryHandler {
	return &MediaGalleryHandler{
		galleryService:  galleryService,
		auditLogService: auditLogService,
	}
}

func (h *MediaGalleryHandler) List(ctx *gin.Context) {
	galleries, err := h.galleryService.List()
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5310, "failed to list media galleries")
		return
	}
	OK(ctx, galleries)
}

func (h *MediaGalleryHandler) Detail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4310, "invalid gallery id")
		return
	}

	gallery, getErr := h.galleryService.GetByID(id)
	if getErr != nil {
		if errors.Is(getErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 4311, "gallery not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5311, "failed to get gallery")
		return
	}

	OK(ctx, gallery)
}

func (h *MediaGalleryHandler) Create(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	var req service.MediaGalleryInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4312, "invalid request body")
		return
	}

	id, err := h.galleryService.Create(req, operator.ID)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4313, err.Error())
		return
	}

	gallery, getErr := h.galleryService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5312, "created but failed to fetch gallery")
		return
	}

	ctx.JSON(stdhttp.StatusCreated, gin.H{
		"code":    0,
		"message": "ok",
		"data":    gallery,
	})

	recordAuditLog(ctx, h.auditLogService, operator, "create", "media_gallery", toInt64Pointer(gallery.ID), gallery.Title, gin.H{
		"slug":       gallery.Slug,
		"is_default": gallery.IsDefault,
	})
}

func (h *MediaGalleryHandler) Update(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4310, "invalid gallery id")
		return
	}

	var req service.MediaGalleryInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4312, "invalid request body")
		return
	}

	if err := h.galleryService.Update(id, req, operator.ID); err != nil {
		if err.Error() == "gallery not found" {
			Fail(ctx, stdhttp.StatusNotFound, 4311, "gallery not found")
			return
		}
		Fail(ctx, stdhttp.StatusBadRequest, 4314, err.Error())
		return
	}

	gallery, getErr := h.galleryService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5313, "updated but failed to fetch gallery")
		return
	}

	OK(ctx, gallery)

	recordAuditLog(ctx, h.auditLogService, operator, "update", "media_gallery", toInt64Pointer(gallery.ID), gallery.Title, gin.H{
		"slug":       gallery.Slug,
		"is_default": gallery.IsDefault,
	})
}

func (h *MediaGalleryHandler) Delete(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4310, "invalid gallery id")
		return
	}

	gallery, getErr := h.galleryService.GetByID(id)
	if getErr != nil {
		if errors.Is(getErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 4311, "gallery not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5311, "failed to get gallery")
		return
	}

	if err := h.galleryService.Delete(id); err != nil {
		if err.Error() == "gallery not found" {
			Fail(ctx, stdhttp.StatusNotFound, 4311, "gallery not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5314, "failed to delete gallery")
		return
	}

	OK(ctx, gin.H{"deleted": true})

	recordAuditLog(ctx, h.auditLogService, operator, "delete", "media_gallery", toInt64Pointer(gallery.ID), gallery.Title, gin.H{
		"slug":       gallery.Slug,
		"is_default": gallery.IsDefault,
	})
}

func (h *MediaGalleryHandler) ListItems(ctx *gin.Context) {
	galleryID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4310, "invalid gallery id")
		return
	}

	items, listErr := h.galleryService.ListItems(galleryID)
	if listErr != nil {
		if errors.Is(listErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 4311, "gallery not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5315, "failed to list gallery items")
		return
	}

	OK(ctx, items)
}

func (h *MediaGalleryHandler) AddItems(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	galleryID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4310, "invalid gallery id")
		return
	}

	var req service.MediaGalleryItemsAppendInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4312, "invalid request body")
		return
	}

	if err := h.galleryService.AddItems(galleryID, req); err != nil {
		switch err.Error() {
		case "gallery not found":
			Fail(ctx, stdhttp.StatusNotFound, 4311, "gallery not found")
		default:
			Fail(ctx, stdhttp.StatusBadRequest, 4315, err.Error())
		}
		return
	}

	items, listErr := h.galleryService.ListItems(galleryID)
	if listErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5315, "items added but failed to list gallery items")
		return
	}

	OK(ctx, items)

	recordAuditLog(ctx, h.auditLogService, operator, "add_items", "media_gallery", toInt64Pointer(galleryID), strconv.FormatInt(galleryID, 10), gin.H{
		"media_asset_ids": req.MediaAssetIDs,
	})
}

func (h *MediaGalleryHandler) UpdateItem(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	galleryID, galleryErr := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if galleryErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4310, "invalid gallery id")
		return
	}

	itemID, itemErr := strconv.ParseInt(ctx.Param("itemID"), 10, 64)
	if itemErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4316, "invalid gallery item id")
		return
	}

	var req service.MediaGalleryItemUpdateInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4312, "invalid request body")
		return
	}

	if err := h.galleryService.UpdateItem(galleryID, itemID, req); err != nil {
		switch err.Error() {
		case "gallery not found":
			Fail(ctx, stdhttp.StatusNotFound, 4311, "gallery not found")
		case "gallery item not found":
			Fail(ctx, stdhttp.StatusNotFound, 4317, "gallery item not found")
		default:
			Fail(ctx, stdhttp.StatusBadRequest, 4318, err.Error())
		}
		return
	}

	item, getErr := h.galleryService.ListItems(galleryID)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5316, "updated but failed to list gallery items")
		return
	}

	OK(ctx, item)

	recordAuditLog(ctx, h.auditLogService, operator, "update_item", "media_gallery", toInt64Pointer(galleryID), strconv.FormatInt(galleryID, 10), gin.H{
		"item_id":    itemID,
		"sort_order": req.SortOrder,
		"is_active":  req.IsActive,
	})
}

func (h *MediaGalleryHandler) DeleteItem(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	galleryID, galleryErr := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if galleryErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4310, "invalid gallery id")
		return
	}

	itemID, itemErr := strconv.ParseInt(ctx.Param("itemID"), 10, 64)
	if itemErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4316, "invalid gallery item id")
		return
	}

	if err := h.galleryService.DeleteItem(galleryID, itemID); err != nil {
		if err.Error() == "gallery item not found" {
			Fail(ctx, stdhttp.StatusNotFound, 4317, "gallery item not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5317, "failed to delete gallery item")
		return
	}

	OK(ctx, gin.H{"deleted": true})

	recordAuditLog(ctx, h.auditLogService, operator, "delete_item", "media_gallery", toInt64Pointer(galleryID), strconv.FormatInt(galleryID, 10), gin.H{
		"item_id": itemID,
	})
}
