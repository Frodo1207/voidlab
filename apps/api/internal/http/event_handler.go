package http

import (
	"database/sql"
	"errors"
	stdhttp "net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/service"
)

type EventHandler struct {
	eventService    *service.EventService
	auditLogService *service.AuditLogService
}

func NewEventHandler(eventService *service.EventService, auditLogService *service.AuditLogService) *EventHandler {
	return &EventHandler{
		eventService:    eventService,
		auditLogService: auditLogService,
	}
}

func (h *EventHandler) List(ctx *gin.Context) {
	events, err := h.eventService.List()
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5101, "failed to list events")
		return
	}

	OK(ctx, events)
}

func (h *EventHandler) PublicList(ctx *gin.Context) {
	events, err := h.eventService.ListPublished()
	if err != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5106, "failed to list published events")
		return
	}

	OK(ctx, events)
}

func (h *EventHandler) Detail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4103, "invalid event id")
		return
	}

	event, getErr := h.eventService.GetByID(id)
	if getErr != nil {
		if errors.Is(getErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 4141, "event not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5102, "failed to get event")
		return
	}

	OK(ctx, event)
}

func (h *EventHandler) PublicDetail(ctx *gin.Context) {
	event, err := h.eventService.GetPublishedBySlug(ctx.Param("slug"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 4141, "event not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5107, "failed to get published event")
		return
	}

	OK(ctx, event)
}

func (h *EventHandler) Create(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	var req service.EventInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4101, "invalid request body")
		return
	}

	id, err := h.eventService.Create(req, operator.ID)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4104, err.Error())
		return
	}

	event, getErr := h.eventService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5103, "created but failed to fetch event")
		return
	}

	ctx.JSON(stdhttp.StatusCreated, gin.H{
		"code":    0,
		"message": "ok",
		"data":    event,
	})

	recordAuditLog(ctx, h.auditLogService, operator, "create", "event", toInt64Pointer(event.ID), event.Title, gin.H{
		"status":     event.Status,
		"slug":       event.Slug,
		"event_time": event.EventTime,
	})
}

func (h *EventHandler) Update(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4103, "invalid event id")
		return
	}

	var req service.EventInput
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4101, "invalid request body")
		return
	}

	if updateErr := h.eventService.Update(id, req, operator.ID); updateErr != nil {
		if updateErr.Error() == "event not found" {
			Fail(ctx, stdhttp.StatusNotFound, 4141, "event not found")
			return
		}
		Fail(ctx, stdhttp.StatusBadRequest, 4105, updateErr.Error())
		return
	}

	event, getErr := h.eventService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5104, "updated but failed to fetch event")
		return
	}

	OK(ctx, event)

	recordAuditLog(ctx, h.auditLogService, operator, "update", "event", toInt64Pointer(event.ID), event.Title, gin.H{
		"status":     event.Status,
		"slug":       event.Slug,
		"event_time": event.EventTime,
	})
}

func (h *EventHandler) UpdateStatus(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4103, "invalid event id")
		return
	}

	var req service.ContentStatusInput
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4101, "invalid request body")
		return
	}

	if updateErr := h.eventService.UpdateStatus(id, req, operator.ID); updateErr != nil {
		if updateErr.Error() == "event not found" {
			Fail(ctx, stdhttp.StatusNotFound, 4141, "event not found")
			return
		}
		Fail(ctx, stdhttp.StatusBadRequest, 4106, updateErr.Error())
		return
	}

	event, getErr := h.eventService.GetByID(id)
	if getErr != nil {
		Fail(ctx, stdhttp.StatusInternalServerError, 5108, "updated but failed to fetch event")
		return
	}

	OK(ctx, event)

	recordAuditLog(ctx, h.auditLogService, operator, "update_status", "event", toInt64Pointer(event.ID), event.Title, gin.H{
		"status": event.Status,
	})
}

func (h *EventHandler) Delete(ctx *gin.Context) {
	operator, ok := requireCurrentUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4103, "invalid event id")
		return
	}

	event, getErr := h.eventService.GetByID(id)
	if getErr != nil {
		if errors.Is(getErr, sql.ErrNoRows) {
			Fail(ctx, stdhttp.StatusNotFound, 4141, "event not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5102, "failed to get event")
		return
	}

	if deleteErr := h.eventService.Delete(id); deleteErr != nil {
		if deleteErr.Error() == "event not found" {
			Fail(ctx, stdhttp.StatusNotFound, 4141, "event not found")
			return
		}
		Fail(ctx, stdhttp.StatusInternalServerError, 5105, "failed to delete event")
		return
	}

	OK(ctx, gin.H{"deleted": true, "id": id})

	recordAuditLog(ctx, h.auditLogService, operator, "delete", "event", toInt64Pointer(id), event.Title, gin.H{
		"slug": event.Slug,
	})
}
