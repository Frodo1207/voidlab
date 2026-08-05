package http

import (
	stdhttp "net/http"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/http/middleware"
	"voidlabai/apps/api/internal/service"
)

func requireCurrentUser(ctx *gin.Context) (domain.User, bool) {
	user, ok := middleware.CurrentUser(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return domain.User{}, false
	}

	return user, true
}

func recordAuditLog(
	ctx *gin.Context,
	auditLogService *service.AuditLogService,
	actor domain.User,
	action string,
	entityType string,
	entityID *int64,
	entityLabel string,
	detail any,
) {
	_ = actor

	if auditLogService == nil {
		return
	}

	currentActor, ok := middleware.CurrentActor(ctx)
	if !ok {
		return
	}

	_ = auditLogService.Record(service.AuditLogInput{
		ActorType:     currentActor.Type,
		ActorID:       currentActor.ID,
		ActorUsername: currentActor.Username,
		ActorRole:     currentActor.Role,
		AgentTokenID:  currentActor.AgentTokenID,
		Action:        action,
		EntityType:    entityType,
		EntityID:      entityID,
		EntityLabel:   entityLabel,
		Detail:        detail,
	})
}

func toInt64Pointer(value int64) *int64 {
	return &value
}
