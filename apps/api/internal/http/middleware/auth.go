package middleware

import (
	stdhttp "net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/service"
)

const currentUserKey = "auth.current_user"
const currentActorKey = "auth.current_actor"

func RequireAuth(authService *service.AuthService, agentTokenService *service.AgentTokenService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, ok := parseBearerToken(ctx.GetHeader("Authorization"))
		if !ok {
			abortWithError(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
			return
		}

		user, err := authService.ResolveToken(token)
		if err == nil {
			ctx.Set(currentUserKey, user)
			ctx.Set(currentActorKey, domain.Actor{
				Type:     "human",
				ID:       user.ID,
				Username: user.Username,
				Role:     user.Role,
			})
			ctx.Next()
			return
		}

		agentToken, agentErr := agentTokenService.ResolveToken(token)
		if agentErr != nil {
			abortWithError(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
			return
		}

		agentTokenID := agentToken.ID
		actor := domain.Actor{
			Type:         "agent",
			ID:           0,
			Username:     agentToken.Name,
			Role:         "agent",
			AgentTokenID: &agentTokenID,
			Scopes:       agentToken.Scopes,
		}

		ctx.Set(currentUserKey, domain.User{
			ID:          0,
			Username:    agentToken.Name,
			Role:        "agent",
			DisplayName: agentToken.Name,
			IsActive:    true,
		})
		ctx.Set(currentActorKey, actor)
		ctx.Next()
	}
}

func RequireAnyRole(roles ...string) gin.HandlerFunc {
	return RequireRolesOrScopes(roles, nil)
}

func RequireRolesOrScopes(roles []string, scopes []string) gin.HandlerFunc {
	allowedRoles := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		normalizedRole := strings.TrimSpace(role)
		if normalizedRole == "" {
			continue
		}
		allowedRoles[normalizedRole] = struct{}{}
	}

	allowedScopes := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		normalizedScope := strings.TrimSpace(scope)
		if normalizedScope == "" {
			continue
		}
		allowedScopes[normalizedScope] = struct{}{}
	}

	return func(ctx *gin.Context) {
		actor, ok := CurrentActor(ctx)
		if !ok {
			abortWithError(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
			return
		}

		switch actor.Type {
		case "human":
			if len(allowedRoles) == 0 {
				ctx.Next()
				return
			}
			if _, ok := allowedRoles[actor.Role]; !ok {
				abortWithError(ctx, stdhttp.StatusForbidden, 4030, "forbidden")
				return
			}
		case "agent":
			if len(allowedScopes) == 0 {
				abortWithError(ctx, stdhttp.StatusForbidden, 4030, "forbidden")
				return
			}
			for _, scope := range actor.Scopes {
				if _, ok := allowedScopes[scope]; ok {
					ctx.Next()
					return
				}
			}
			abortWithError(ctx, stdhttp.StatusForbidden, 4030, "forbidden")
			return
		default:
			abortWithError(ctx, stdhttp.StatusForbidden, 4030, "forbidden")
			return
		}
		ctx.Next()
	}
}

func CurrentUser(ctx *gin.Context) (domain.User, bool) {
	value, exists := ctx.Get(currentUserKey)
	if !exists {
		return domain.User{}, false
	}

	user, ok := value.(domain.User)
	if !ok {
		return domain.User{}, false
	}

	return user, true
}

func CurrentActor(ctx *gin.Context) (domain.Actor, bool) {
	value, exists := ctx.Get(currentActorKey)
	if !exists {
		return domain.Actor{}, false
	}

	actor, ok := value.(domain.Actor)
	if !ok {
		return domain.Actor{}, false
	}

	return actor, true
}

func parseBearerToken(authHeader string) (string, bool) {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return "", false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	return token, token != ""
}

func abortWithError(ctx *gin.Context, statusCode int, code int, message string) {
	ctx.JSON(statusCode, gin.H{
		"code":    code,
		"message": message,
		"data":    nil,
	})
	ctx.Abort()
}
