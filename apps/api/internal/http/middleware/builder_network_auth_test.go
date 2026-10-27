package middleware

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
	"voidlabai/apps/api/internal/service"
	"voidlabai/apps/api/internal/storage"
)

func TestNetworkAgentScopeAndSessionGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := storage.OpenSQLite(filepath.Join(t.TempDir(), "network-auth.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	networkRepo := repository.NewBuilderNetworkRepository(db)
	networkService := service.NewBuilderNetworkService(networkRepo)
	tokenService := service.NewBuilderNetworkAgentTokenService(
		networkRepo,
		repository.NewBuilderNetworkAgentTokenRepository(db),
	)
	invite, err := networkService.CreateInvite(1, "", 1)
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}
	session, err := networkService.Register("scope-user@example.com", "secure-password", invite, domain.NetworkProfile{
		DisplayName: "Scope User", PrimaryRole: "Engineer",
	})
	if err != nil {
		t.Fatalf("register user: %v", err)
	}
	agentToken, err := tokenService.Create(session.Account.ID, service.NetworkAgentTokenCreateInput{
		Name:   "Read-only agent",
		Scopes: []string{"network_profiles:read"},
	})
	if err != nil {
		t.Fatalf("create agent token: %v", err)
	}
	projectToken, err := tokenService.Create(session.Account.ID, service.NetworkAgentTokenCreateInput{
		Name:   "Project reader",
		Scopes: []string{"network_projects:read"},
	})
	if err != nil {
		t.Fatalf("create project token: %v", err)
	}

	engine := gin.New()
	network := engine.Group("/network")
	network.Use(RequireNetworkAuth(networkService, tokenService))
	network.GET("/read", RequireNetworkScopes("network_profiles:read"), func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})
	network.GET("/projects", RequireNetworkScopes("network_projects:read"), func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})
	network.POST("/write", RequireNetworkScopes("network_works:write"), func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})
	network.POST("/sensitive", RequireNetworkSession(), func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/network/read", nil)
	request.Header.Set("Authorization", "Bearer "+agentToken.Token)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("read scope should be permitted, status=%d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/network/projects", nil)
	request.Header.Set("Authorization", "Bearer "+agentToken.Token)
	response = httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("profile scope must not grant Project access, status=%d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/network/projects", nil)
	request.Header.Set("Authorization", "Bearer "+projectToken.Token)
	response = httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("Project scope should permit Project reads, status=%d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/network/write", nil)
	request.Header.Set("Authorization", "Bearer "+agentToken.Token)
	response = httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("missing write scope should be rejected, status=%d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/network/sensitive", nil)
	request.Header.Set("Authorization", "Bearer "+agentToken.Token)
	response = httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("agent should not perform session-only action, status=%d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/network/sensitive", nil)
	request.Header.Set("Authorization", "Bearer "+session.Token)
	response = httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("member session should perform session-only action, status=%d body=%s", response.Code, response.Body.String())
	}
}
