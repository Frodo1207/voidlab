package http

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	stdhttp "net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/service"
)

type KnowledgeHandler struct {
	knowledgeService      *service.KnowledgeService
	knowledgeAssetService *service.KnowledgeAssetService
	auditLogService       *service.AuditLogService
}

func NewKnowledgeHandler(
	knowledgeService *service.KnowledgeService,
	knowledgeAssetService *service.KnowledgeAssetService,
	auditLogService *service.AuditLogService,
) *KnowledgeHandler {
	return &KnowledgeHandler{
		knowledgeService:      knowledgeService,
		knowledgeAssetService: knowledgeAssetService,
		auditLogService:       auditLogService,
	}
}

func (h *KnowledgeHandler) PublicListSpaces(ctx *gin.Context) {
	spaces, err := h.knowledgeService.ListPublishedSpaces()
	if err != nil {
		logKnowledgePublicError(ctx, "list_spaces", err, gin.H{})
		if isTransientKnowledgeStorageError(err) {
			Fail(ctx, stdhttp.StatusServiceUnavailable, 5601, "knowledge storage temporarily busy")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5601, "failed to list knowledge spaces")
		return
	}
	OK(ctx, spaces)
}

func (h *KnowledgeHandler) PublicSpaceTOC(ctx *gin.Context) {
	slug := strings.TrimSpace(ctx.Param("slug"))
	space, entries, err := h.getPublishedSpaceAndEntries(slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 5641, "knowledge space not found")
			return
		}
		logKnowledgePublicError(ctx, "get_space_toc", err, gin.H{
			"slug": slug,
		})
		if isTransientKnowledgeStorageError(err) {
			Fail(ctx, stdhttp.StatusServiceUnavailable, 5602, "knowledge storage temporarily busy")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5602, "failed to get knowledge space")
		return
	}

	records, _ := entries.([]domain.KnowledgeEntry)
	payload := make([]gin.H, 0, len(records))
	for _, entry := range records {
		payload = append(payload, gin.H{
			"id":                     entry.ID,
			"space_id":               entry.SpaceID,
			"space_slug":             entry.SpaceSlug,
			"title":                  entry.Title,
			"slug":                   entry.Slug,
			"section_name":           entry.SectionName,
			"sort_order":             entry.SortOrder,
			"estimated_read_minutes": entry.EstimatedReadMinutes,
			"public_summary":         entry.PublicSummary,
			"cover_url":              entry.CoverURL,
			"is_preview":             entry.IsPreview,
			"status":                 entry.Status,
			"updated_at":             entry.UpdatedAt,
		})
	}

	OK(ctx, gin.H{
		"space":   space,
		"entries": payload,
	})
}

func (h *KnowledgeHandler) PublicVerifyToken(ctx *gin.Context) {
	space, err := h.knowledgeService.GetPublishedSpaceBySlug(ctx.Param("slug"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 5641, "knowledge space not found")
			return
		}
		logKnowledgePublicError(ctx, "verify_token_get_space", err, gin.H{
			"slug": ctx.Param("slug"),
		})
		if isTransientKnowledgeStorageError(err) {
			Fail(ctx, stdhttp.StatusServiceUnavailable, 5602, "knowledge storage temporarily busy")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5602, "failed to get knowledge space")
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5401, "invalid request body")
		return
	}

	record, grant, verifyErr := h.knowledgeService.VerifySpaceToken(space, req.Token, ctx.ClientIP(), ctx.Request.UserAgent())
	if verifyErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5404, verifyErr.Error())
		return
	}

	OK(ctx, gin.H{
		"grant": grant,
		"space": gin.H{
			"id":   space.ID,
			"slug": space.Slug,
		},
		"access": gin.H{
			"access_level": record.AccessLevel,
			"scope_type":   record.ScopeType,
			"space_ids":    record.SpaceIDs,
		},
	})
}

func (h *KnowledgeHandler) PublicEntryDetail(ctx *gin.Context) {
	slug := strings.TrimSpace(ctx.Param("slug"))
	entrySlug := strings.TrimSpace(ctx.Param("entrySlug"))
	space, err := h.knowledgeService.GetPublishedSpaceBySlug(slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 5641, "knowledge space not found")
			return
		}
		logKnowledgePublicError(ctx, "get_entry_space", err, gin.H{
			"slug":       slug,
			"entry_slug": entrySlug,
		})
		if isTransientKnowledgeStorageError(err) {
			Fail(ctx, stdhttp.StatusServiceUnavailable, 5602, "knowledge storage temporarily busy")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5602, "failed to get knowledge space")
		return
	}

	entry, entryErr := h.knowledgeService.GetPublishedEntryBySpaceAndSlug(space.ID, entrySlug)
	if entryErr != nil {
		if errors.Is(entryErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 5642, "knowledge entry not found")
			return
		}
		logKnowledgePublicError(ctx, "get_entry_detail", entryErr, gin.H{
			"slug":       slug,
			"entry_slug": entrySlug,
			"space_id":   space.ID,
		})
		if isTransientKnowledgeStorageError(entryErr) {
			Fail(ctx, stdhttp.StatusServiceUnavailable, 5603, "knowledge storage temporarily busy")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5603, "failed to get knowledge entry")
		return
	}

	grant := strings.TrimSpace(ctx.GetHeader("X-Knowledge-Grant"))
	if grant == "" {
		grant = strings.TrimSpace(ctx.Query("grant"))
	}

	allowed, allowErr := h.knowledgeService.CanReadEntry(space, entry, grant, ctx.ClientIP(), ctx.Request.UserAgent())
	if allowErr != nil {
		logKnowledgePublicError(ctx, "verify_entry_access", allowErr, gin.H{
			"slug":       slug,
			"entry_slug": entrySlug,
			"space_id":   space.ID,
			"entry_id":   entry.ID,
		})
		if isTransientKnowledgeStorageError(allowErr) {
			Fail(ctx, stdhttp.StatusServiceUnavailable, 5604, "knowledge storage temporarily busy")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5604, "failed to verify knowledge access")
		return
	}
	if !allowed {
		Fail(ctx, stdhttp.StatusForbidden, 5430, "knowledge entry is locked")
		return
	}

	OK(ctx, gin.H{
		"space": space,
		"entry": entry,
	})
}

func (h *KnowledgeHandler) PublicAsset(ctx *gin.Context) {
	slug := strings.TrimSpace(ctx.Param("slug"))
	space, err := h.knowledgeService.GetPublishedSpaceBySlug(slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 5641, "knowledge space not found")
			return
		}
		logKnowledgePublicError(ctx, "get_asset_space", err, gin.H{
			"slug": slug,
		})
		if isTransientKnowledgeStorageError(err) {
			Fail(ctx, stdhttp.StatusServiceUnavailable, 5602, "knowledge storage temporarily busy")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5602, "failed to get knowledge space")
		return
	}

	assetID, err := strconv.ParseInt(ctx.Param("assetID"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5409, "invalid knowledge asset id")
		return
	}

	grant := strings.TrimSpace(ctx.Query("grant"))
	if grant == "" {
		grant = strings.TrimSpace(ctx.GetHeader("X-Knowledge-Grant"))
	}

	allowed, allowErr := h.knowledgeService.CanReadSpaceContent(space, grant)
	if allowErr != nil {
		logKnowledgePublicError(ctx, "verify_asset_access", allowErr, gin.H{
			"slug":     slug,
			"space_id": space.ID,
			"asset_id": assetID,
		})
		if isTransientKnowledgeStorageError(allowErr) {
			Fail(ctx, stdhttp.StatusServiceUnavailable, 5604, "knowledge storage temporarily busy")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5604, "failed to verify knowledge access")
		return
	}
	if !allowed {
		Fail(ctx, stdhttp.StatusForbidden, 5431, "knowledge asset is locked")
		return
	}

	asset, path, assetErr := h.knowledgeAssetService.GetStorageBySpaceAndID(space.ID, assetID)
	if assetErr != nil {
		if errors.Is(assetErr, sql.ErrNoRows) || errors.Is(assetErr, os.ErrNotExist) {
			Fail(ctx, stdhttp.StatusNotFound, 5644, "knowledge asset not found")
			return
		}
		logKnowledgePublicError(ctx, "get_asset_storage", assetErr, gin.H{
			"slug":     slug,
			"space_id": space.ID,
			"asset_id": assetID,
		})
		if isTransientKnowledgeStorageError(assetErr) {
			Fail(ctx, stdhttp.StatusServiceUnavailable, 5615, "knowledge storage temporarily busy")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5615, "failed to get knowledge asset")
		return
	}

	ctx.Header("Content-Type", asset.ContentType)
	ctx.File(path)
}

func (h *KnowledgeHandler) ListSpaces(ctx *gin.Context) {
	spaces, err := h.knowledgeService.ListSpaces()
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5605, "failed to list knowledge spaces")
		return
	}
	OK(ctx, spaces)
}

func (h *KnowledgeHandler) CreateSpace(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	var req service.KnowledgeSpaceInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5401, "invalid request body")
		return
	}

	id, createErr := h.knowledgeService.CreateSpace(req, operator.ID)
	if createErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5405, createErr.Error())
		return
	}

	space, getErr := h.knowledgeService.GetSpaceByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5606, "created but failed to fetch knowledge space")
		return
	}

	ctx.JSON(stdhttp.StatusCreated, gin.H{
		"code":    0,
		"message": "ok",
		"data":    space,
	})

	recordAuditLog(ctx, h.auditLogService, operator, "create", "knowledge_space", toInt64Pointer(space.ID), space.Title, gin.H{
		"slug":            space.Slug,
		"visibility_mode": space.VisibilityMode,
		"status":          space.Status,
	})
}

func (h *KnowledgeHandler) UpdateSpace(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5402, "invalid knowledge space id")
		return
	}

	var req service.KnowledgeSpaceInput
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5401, "invalid request body")
		return
	}

	if updateErr := h.knowledgeService.UpdateSpace(id, req, operator.ID); updateErr != nil {
		if updateErr.Error() == "knowledge space not found" {
			Fail(ctx, stdhttp.StatusNotFound, 5641, "knowledge space not found")
			return
		}
		Fail(ctx, stdhttp.StatusBadRequest, 5406, updateErr.Error())
		return
	}

	space, getErr := h.knowledgeService.GetSpaceByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5607, "updated but failed to fetch knowledge space")
		return
	}

	OK(ctx, space)

	recordAuditLog(ctx, h.auditLogService, operator, "update", "knowledge_space", toInt64Pointer(space.ID), space.Title, gin.H{
		"slug":            space.Slug,
		"visibility_mode": space.VisibilityMode,
		"status":          space.Status,
	})
}

func (h *KnowledgeHandler) DeleteSpace(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5402, "invalid knowledge space id")
		return
	}

	space, getErr := h.knowledgeService.GetSpaceByID(id)
	if getErr != nil {
		if errors.Is(getErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 5641, "knowledge space not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5602, "failed to get knowledge space")
		return
	}

	if deleteErr := h.knowledgeService.DeleteSpace(id); deleteErr != nil {
		if deleteErr.Error() == "knowledge space not found" {
			Fail(ctx, stdhttp.StatusNotFound, 5641, "knowledge space not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5608, "failed to delete knowledge space")
		return
	}

	OK(ctx, gin.H{"deleted": true, "id": id})

	recordAuditLog(ctx, h.auditLogService, operator, "delete", "knowledge_space", toInt64Pointer(space.ID), space.Title, gin.H{
		"slug": space.Slug,
	})
}

func (h *KnowledgeHandler) ListEntries(ctx *gin.Context) {
	var spaceID *int64
	if raw := strings.TrimSpace(ctx.Query("space_id")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			Fail(ctx, stdhttp.StatusBadRequest, 5403, "invalid knowledge space id")
			return
		}
		spaceID = &parsed
	}

	entries, err := h.knowledgeService.ListEntries(spaceID)
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5609, "failed to list knowledge entries")
		return
	}
	OK(ctx, entries)
}

func (h *KnowledgeHandler) CreateEntry(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	var req service.KnowledgeEntryInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5401, "invalid request body")
		return
	}

	id, createErr := h.knowledgeService.CreateEntry(req, operator.ID)
	if createErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5410, createErr.Error())
		return
	}

	entry, getErr := h.knowledgeService.GetEntryByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5610, "created but failed to fetch knowledge entry")
		return
	}

	ctx.JSON(stdhttp.StatusCreated, gin.H{
		"code":    0,
		"message": "ok",
		"data":    entry,
	})

	recordAuditLog(ctx, h.auditLogService, operator, "create", "knowledge_entry", toInt64Pointer(entry.ID), entry.Title, gin.H{
		"space_id":   entry.SpaceID,
		"slug":       entry.Slug,
		"section":    entry.SectionName,
		"status":     entry.Status,
		"is_preview": entry.IsPreview,
	})
}

func (h *KnowledgeHandler) UpdateEntry(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5407, "invalid knowledge entry id")
		return
	}

	var req service.KnowledgeEntryInput
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5401, "invalid request body")
		return
	}

	if updateErr := h.knowledgeService.UpdateEntry(id, req, operator.ID); updateErr != nil {
		if updateErr.Error() == "knowledge entry not found" {
			Fail(ctx, stdhttp.StatusNotFound, 5642, "knowledge entry not found")
			return
		}
		Fail(ctx, stdhttp.StatusBadRequest, 5411, updateErr.Error())
		return
	}

	entry, getErr := h.knowledgeService.GetEntryByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5611, "updated but failed to fetch knowledge entry")
		return
	}

	OK(ctx, entry)

	recordAuditLog(ctx, h.auditLogService, operator, "update", "knowledge_entry", toInt64Pointer(entry.ID), entry.Title, gin.H{
		"space_id":   entry.SpaceID,
		"slug":       entry.Slug,
		"section":    entry.SectionName,
		"status":     entry.Status,
		"is_preview": entry.IsPreview,
	})
}

func (h *KnowledgeHandler) DeleteEntry(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5407, "invalid knowledge entry id")
		return
	}

	entry, getErr := h.knowledgeService.GetEntryByID(id)
	if getErr != nil {
		if errors.Is(getErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 5642, "knowledge entry not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5603, "failed to get knowledge entry")
		return
	}

	if deleteErr := h.knowledgeService.DeleteEntry(id); deleteErr != nil {
		if deleteErr.Error() == "knowledge entry not found" {
			Fail(ctx, stdhttp.StatusNotFound, 5642, "knowledge entry not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5612, "failed to delete knowledge entry")
		return
	}

	OK(ctx, gin.H{"deleted": true, "id": id})

	recordAuditLog(ctx, h.auditLogService, operator, "delete", "knowledge_entry", toInt64Pointer(entry.ID), entry.Title, gin.H{
		"space_id": entry.SpaceID,
		"slug":     entry.Slug,
	})
}

func (h *KnowledgeHandler) ImportEntryMarkdown(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4301, "file is required")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5414, "failed to open markdown file")
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5618, "failed to read markdown file")
		return
	}

	result, parseErr := h.knowledgeService.ParseMarkdownImport(fileHeader.Filename, content)
	if parseErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5415, parseErr.Error())
		return
	}

	OK(ctx, result)

	recordAuditLog(ctx, h.auditLogService, operator, "import_markdown", "knowledge_entry", nil, fileHeader.Filename, gin.H{
		"title":                  result.Title,
		"slug":                   result.Slug,
		"estimated_read_minutes": result.EstimatedReadMinutes,
		"status":                 result.Status,
	})
}

func (h *KnowledgeHandler) ListAccessTokens(ctx *gin.Context) {
	var spaceID *int64
	if raw := strings.TrimSpace(ctx.Query("space_id")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			Fail(ctx, stdhttp.StatusBadRequest, 5403, "invalid knowledge space id")
			return
		}
		spaceID = &parsed
	}

	tokens, err := h.knowledgeService.ListAccessTokens(spaceID)
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5613, "failed to list knowledge access tokens")
		return
	}
	OK(ctx, tokens)
}

func (h *KnowledgeHandler) ListAssets(ctx *gin.Context) {
	spaceID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5402, "invalid knowledge space id")
		return
	}

	assets, err := h.knowledgeAssetService.ListBySpace(spaceID)
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5616, "failed to list knowledge assets")
		return
	}

	OK(ctx, assets)
}

func (h *KnowledgeHandler) UploadAsset(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	spaceID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5402, "invalid knowledge space id")
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4301, "file is required")
		return
	}

	asset, uploadErr := h.knowledgeAssetService.Upload(spaceID, fileHeader, operator.ID)
	if uploadErr != nil {
		if uploadErr.Error() == "knowledge space not found" {
			Fail(ctx, stdhttp.StatusNotFound, 5641, "knowledge space not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5617, uploadErr.Error())
		return
	}

	space, _ := h.knowledgeService.GetSpaceByID(spaceID)
	markdownURL := "knowledge-asset://" + strconv.FormatInt(asset.ID, 10)
	publicURL := "/api/v1/public/knowledge/spaces/" + space.Slug + "/assets/" + strconv.FormatInt(asset.ID, 10)

	ctx.JSON(stdhttp.StatusCreated, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"asset":            asset,
			"markdown_url":     markdownURL,
			"markdown_snippet": "![" + asset.FileName + "](" + markdownURL + ")",
			"public_url":       publicURL,
		},
	})

	recordAuditLog(ctx, h.auditLogService, operator, "upload", "knowledge_asset", toInt64Pointer(asset.ID), asset.FileName, gin.H{
		"space_id":       asset.SpaceID,
		"media_asset_id": asset.MediaAssetID,
		"content_type":   asset.ContentType,
	})
}

func (h *KnowledgeHandler) CreateAccessToken(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	var req service.KnowledgeAccessTokenCreateInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5401, "invalid request body")
		return
	}

	record, plainToken, createErr := h.knowledgeService.CreateAccessToken(req, operator.ID)
	if createErr != nil {
		if createErr.Error() == "knowledge space not found" {
			Fail(ctx, stdhttp.StatusNotFound, 5641, "knowledge space not found")
			return
		}
		Fail(ctx, stdhttp.StatusBadRequest, 5412, createErr.Error())
		return
	}

	OK(ctx, gin.H{
		"record": record,
		"token":  plainToken,
	})

	recordAuditLog(ctx, h.auditLogService, operator, "create", "knowledge_access_token", toInt64Pointer(record.ID), record.Name, gin.H{
		"space_ids":    record.SpaceIDs,
		"scope_type":   record.ScopeType,
		"access_level": record.AccessLevel,
		"expires_at":   record.ExpiresAt,
		"is_active":    record.IsActive,
	})
}

func (h *KnowledgeHandler) UpdateAccessTokenStatus(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5408, "invalid knowledge token id")
		return
	}

	var req service.KnowledgeAccessTokenStatusInput
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 5401, "invalid request body")
		return
	}

	if updateErr := h.knowledgeService.UpdateAccessTokenStatus(id, req); updateErr != nil {
		if updateErr.Error() == "knowledge access token not found" {
			Fail(ctx, stdhttp.StatusNotFound, 5643, "knowledge access token not found")
			return
		}
		Fail(ctx, stdhttp.StatusBadRequest, 5413, updateErr.Error())
		return
	}

	tokens, listErr := h.knowledgeService.ListAccessTokens(nil)
	if listErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5614, "updated but failed to fetch knowledge token")
		return
	}

	var target any = gin.H{"id": id, "is_active": req.IsActive}
	for _, token := range tokens {
		if token.ID == id {
			target = token
			break
		}
	}

	OK(ctx, target)

	recordAuditLog(ctx, h.auditLogService, operator, "update_status", "knowledge_access_token", toInt64Pointer(id), "knowledge access token", gin.H{
		"is_active": req.IsActive,
	})
}

func (h *KnowledgeHandler) getPublishedSpaceAndEntries(slug string) (any, any, error) {
	space, err := h.knowledgeService.GetPublishedSpaceBySlug(slug)
	if err != nil {
		return nil, nil, err
	}

	entries, entryErr := h.knowledgeService.ListPublishedEntriesBySpace(space.ID)
	if entryErr != nil {
		return nil, nil, entryErr
	}

	return space, entries, nil
}

func isTransientKnowledgeStorageError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "database is locked") ||
		strings.Contains(message, "database table is locked") ||
		strings.Contains(message, "database is busy") ||
		strings.Contains(message, "sql_busy") ||
		strings.Contains(message, "sql_locked")
}

func logKnowledgePublicError(ctx *gin.Context, operation string, err error, detail gin.H) {
	if err == nil {
		return
	}

	fields := []string{
		"component=knowledge_public",
		"operation=" + operation,
		"method=" + ctx.Request.Method,
		"path=" + ctx.FullPath(),
		"client_ip=" + ctx.ClientIP(),
		"error=" + strconv.Quote(err.Error()),
	}

	for key, value := range detail {
		fields = append(fields, key+"="+strconv.Quote(strings.TrimSpace(toDebugString(value))))
	}

	log.Printf("%s", strings.Join(fields, " "))
}

func toDebugString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case int64:
		return strconv.FormatInt(typed, 10)
	case int:
		return strconv.Itoa(typed)
	default:
		return fmt.Sprint(value)
	}
}
