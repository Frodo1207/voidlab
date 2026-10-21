package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/config"
	"voidlabai/apps/api/internal/repository"
	"voidlabai/apps/api/internal/service"
	"voidlabai/apps/api/internal/storage"
)

func TestOpsAgentCanIssueBuilderNetworkInviteWithoutMemberTokenAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := storage.OpenSQLite(filepath.Join(t.TempDir(), "network-invite-agent.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	router := NewRouter(config.Config{UploadsDir: t.TempDir()}, db)

	adminLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(
		`{"username":"admin","password":"admin"}`,
	))
	adminLogin.Header.Set("Content-Type", "application/json")
	adminLoginResponse := httptest.NewRecorder()
	router.ServeHTTP(adminLoginResponse, adminLogin)
	if adminLoginResponse.Code != http.StatusOK {
		t.Fatalf("admin login: status=%d body=%s", adminLoginResponse.Code, adminLoginResponse.Body.String())
	}
	var loginEnvelope struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminLoginResponse.Body.Bytes(), &loginEnvelope); err != nil {
		t.Fatalf("decode admin login: %v", err)
	}
	if loginEnvelope.Data.Token == "" {
		t.Fatal("admin session token missing")
	}

	adminTokenService := service.NewAgentTokenService(repository.NewAgentTokenRepository(db))
	agentRecord, agentToken, err := adminTokenService.Create(service.AgentTokenCreateInput{
		Name:   "BuilderNetwork invitation operator",
		Scopes: []string{"network_invitations:write"},
	}, 1)
	if err != nil {
		t.Fatalf("create ops agent token: %v", err)
	}
	if len(agentRecord.Scopes) != 1 || agentRecord.Scopes[0] != "network_invitations:write" {
		t.Fatalf("unexpected ops token scopes: %+v", agentRecord.Scopes)
	}

	createInvite := httptest.NewRequest(http.MethodPost, "/api/v1/network/invitations", strings.NewReader(
		`{"max_uses":1}`,
	))
	createInvite.Header.Set("Authorization", "Bearer "+agentToken)
	createInvite.Header.Set("Content-Type", "application/json")
	createInviteResponse := httptest.NewRecorder()
	router.ServeHTTP(createInviteResponse, createInvite)
	if createInviteResponse.Code != http.StatusCreated {
		t.Fatalf("ops agent should create invite: status=%d body=%s", createInviteResponse.Code, createInviteResponse.Body.String())
	}
	var inviteEnvelope struct {
		Data struct {
			Code string `json:"code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createInviteResponse.Body.Bytes(), &inviteEnvelope); err != nil {
		t.Fatalf("decode invite response: %v", err)
	}
	if inviteEnvelope.Data.Code == "" {
		t.Fatal("created invite code missing")
	}

	adminOnlyRequest := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	adminOnlyRequest.Header.Set("Authorization", "Bearer "+agentToken)
	adminOnlyResponse := httptest.NewRecorder()
	router.ServeHTTP(adminOnlyResponse, adminOnlyRequest)
	if adminOnlyResponse.Code != http.StatusForbidden {
		t.Fatalf("invitation scope must not grant user admin access: status=%d body=%s", adminOnlyResponse.Code, adminOnlyResponse.Body.String())
	}

	registerRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(
		`{"email":"invitee@example.com","password":"secure-password","invite_code":"`+inviteEnvelope.Data.Code+`","profile":{"display_name":"Invitee","primary_role":"Engineer"}}`,
	))
	registerRequest.Header.Set("Content-Type", "application/json")
	registerResponse := httptest.NewRecorder()
	router.ServeHTTP(registerResponse, registerRequest)
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("invitee should register with issued code: status=%d body=%s", registerResponse.Code, registerResponse.Body.String())
	}

	auditService := service.NewAuditLogService(repository.NewAuditLogRepository(db))
	logs, err := auditService.List(10)
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	foundInviteAudit := false
	for _, entry := range logs {
		if entry.EntityType != "network_invite_code" || entry.Action != "create" {
			continue
		}
		foundInviteAudit = true
		if entry.ActorType != "agent" || entry.AgentTokenID == nil {
			t.Fatalf("invite audit should identify the ops agent: %+v", entry)
		}
		if strings.Contains(entry.DetailJSON, inviteEnvelope.Data.Code) {
			t.Fatalf("invite audit leaked plaintext code: %s", entry.DetailJSON)
		}
	}
	if !foundInviteAudit {
		t.Fatalf("invite creation was not audited: %+v", logs)
	}
}
