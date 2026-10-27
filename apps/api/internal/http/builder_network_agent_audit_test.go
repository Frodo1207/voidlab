package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/config"
	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
	"voidlabai/apps/api/internal/service"
	"voidlabai/apps/api/internal/storage"
)

func TestProjectRoutesRequireProjectScopesAndReturnOwnedProject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := storage.OpenSQLite(filepath.Join(t.TempDir(), "network-project-api.sqlite"))
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
	owner, err := networkService.Register("project-api-owner@example.com", "secure-password", invite, domain.NetworkProfile{
		DisplayName: "Project API Owner", PrimaryRole: "Founder",
	})
	if err != nil {
		t.Fatalf("register owner: %v", err)
	}
	projectToken, err := tokenService.Create(owner.Account.ID, service.NetworkAgentTokenCreateInput{
		Name:   "Project manager",
		Scopes: []string{"network_projects:read", "network_projects:write"},
	})
	if err != nil {
		t.Fatalf("create project token: %v", err)
	}
	legacyToken, err := tokenService.Create(owner.Account.ID, service.NetworkAgentTokenCreateInput{
		Name:   "Legacy opportunity reader",
		Scopes: []string{"network_opportunities:read"},
	})
	if err != nil {
		t.Fatalf("create legacy token: %v", err)
	}

	router := NewRouter(config.Config{UploadsDir: t.TempDir()}, db)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/network/projects", strings.NewReader(
		`{"title":"API Project","summary":"Created through the scoped Project API","status":"published","project_plan":{"milestones":[],"sections":[]},"members":[]}`,
	))
	request.Header.Set("Authorization", "Bearer "+projectToken.Token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("create Project: status=%d body=%s", response.Code, response.Body.String())
	}
	var created struct {
		Data domain.NetworkProject `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil || created.Data.ID == 0 {
		t.Fatalf("decode created Project: data=%+v err=%v", created.Data, err)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/network/projects", nil)
	request.Header.Set("Authorization", "Bearer "+legacyToken.Token)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("legacy Opportunity scope must not grant Project access: status=%d body=%s", response.Code, response.Body.String())
	}

	projectPath := "/api/v1/network/projects/" + strconv.FormatInt(created.Data.ID, 10)
	request = httptest.NewRequest(http.MethodGet, projectPath, nil)
	request.Header.Set("Authorization", "Bearer "+projectToken.Token)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "API Project") {
		t.Fatalf("read owned Project: status=%d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/public/network/projects/"+created.Data.Slug, nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "API Project") {
		t.Fatalf("read public Project: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAgentContactExchangeReadIsAuditedWithoutPrivateValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := storage.OpenSQLite(filepath.Join(t.TempDir(), "network-audit.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	networkRepo := repository.NewBuilderNetworkRepository(db)
	networkService := service.NewBuilderNetworkService(networkRepo)
	tokenRepo := repository.NewBuilderNetworkAgentTokenRepository(db)
	tokenService := service.NewBuilderNetworkAgentTokenService(networkRepo, tokenRepo)
	auditService := service.NewAuditLogService(repository.NewAuditLogRepository(db))
	invite, err := networkService.CreateInvite(2, "", 1)
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}
	owner, err := networkService.Register("audit-owner@example.com", "secure-password-owner", invite, domain.NetworkProfile{
		DisplayName: "Audit Owner", PrimaryRole: "Founder",
	})
	if err != nil {
		t.Fatalf("register owner: %v", err)
	}
	applicant, err := networkService.Register("audit-applicant@example.com", "secure-password-applicant", invite, domain.NetworkProfile{
		DisplayName: "Audit Applicant", PrimaryRole: "Engineer",
	})
	if err != nil {
		t.Fatalf("register applicant: %v", err)
	}
	if err := networkService.UpdateContacts(owner.Account.ID, domain.NetworkContactDetails{WeChat: "owner-private-value"}, false, true, false); err != nil {
		t.Fatalf("save owner contacts: %v", err)
	}
	if err := networkService.UpdateContacts(applicant.Account.ID, domain.NetworkContactDetails{WeChat: "applicant-private-value"}, false, true, false); err != nil {
		t.Fatalf("save applicant contacts: %v", err)
	}
	opportunity, err := networkService.CreateOpportunity(owner.Account.ID, domain.NetworkOpportunity{
		Title: "Audit test collaboration", Summary: "Build a prototype", OpportunityType: "collaboration",
		LookingFor: []string{"Engineer"},
	})
	if err != nil {
		t.Fatalf("create opportunity: %v", err)
	}
	contactRequest, err := networkService.CreateContactRequest(applicant.Account.ID, domain.NetworkContactRequest{
		TargetType: "opportunity", TargetID: opportunity.ID, Message: "private-request-message",
		ShareWeChat: true,
	})
	if err != nil {
		t.Fatalf("create contact request: %v", err)
	}
	if err := networkService.AcceptRequest(contactRequest.ID, owner.Account.ID); err != nil {
		t.Fatalf("accept request: %v", err)
	}
	agentToken, err := tokenService.Create(applicant.Account.ID, service.NetworkAgentTokenCreateInput{
		Name: "Exchange reader", Scopes: []string{"network_contact_exchange:read"},
	})
	if err != nil {
		t.Fatalf("create agent token: %v", err)
	}
	profileReadToken, err := tokenService.Create(applicant.Account.ID, service.NetworkAgentTokenCreateInput{
		Name: "Profile reader", Scopes: []string{"network_profiles:read"},
	})
	if err != nil {
		t.Fatalf("create profile read token: %v", err)
	}
	profileWriteToken, err := tokenService.Create(applicant.Account.ID, service.NetworkAgentTokenCreateInput{
		Name: "Profile writer", Scopes: []string{"network_profiles:write"},
	})
	if err != nil {
		t.Fatalf("create profile write token: %v", err)
	}

	router := NewRouter(config.Config{UploadsDir: t.TempDir()}, db)
	exchangePath := "/api/v1/network/contact-requests/" + intString(contactRequest.ID) + "/contacts"
	request := httptest.NewRequest(http.MethodGet, exchangePath, nil)
	request.Header.Set("Authorization", "Bearer "+agentToken.Token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("read exchange: status=%d body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "applicant-private-value") {
		t.Fatalf("exchange response did not include authorized contact: %s", response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/network/me/contacts", nil)
	request.Header.Set("Authorization", "Bearer "+profileReadToken.Token)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("profile read scope must not expose private contacts, status=%d body=%s", response.Code, response.Body.String())
	}
	request = httptest.NewRequest(http.MethodGet, "/api/v1/network/me/contacts", nil)
	request.Header.Set("Authorization", "Bearer "+agentToken.Token)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "applicant-private-value") {
		t.Fatalf("contact exchange read scope should read own contact details, status=%d body=%s", response.Code, response.Body.String())
	}
	request = httptest.NewRequest(http.MethodPut, "/api/v1/network/me/contacts", strings.NewReader(`{"contacts":{"wechat":"changed"}}`))
	request.Header.Set("Authorization", "Bearer "+profileWriteToken.Token)
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("profile write scope must not update private contacts, status=%d body=%s", response.Code, response.Body.String())
	}

	revokePath := "/api/v1/network/contact-exchanges/" + intString(contactRequest.ID) + "/revoke"
	request = httptest.NewRequest(http.MethodPost, revokePath, nil)
	request.Header.Set("Authorization", "Bearer "+agentToken.Token)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("read-only agent should not revoke exchange, status=%d body=%s", response.Code, response.Body.String())
	}

	revokeToken, err := tokenService.Create(applicant.Account.ID, service.NetworkAgentTokenCreateInput{
		Name: "Exchange revoker", Scopes: []string{"network_contact_exchange:write"},
	})
	if err != nil {
		t.Fatalf("create revoke token: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, revokePath, nil)
	request.Header.Set("Authorization", "Bearer "+revokeToken.Token)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("revoke exchange: status=%d body=%s", response.Code, response.Body.String())
	}
	request = httptest.NewRequest(http.MethodGet, exchangePath, nil)
	request.Header.Set("Authorization", "Bearer "+agentToken.Token)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("revoked exchange should no longer be readable, status=%d body=%s", response.Code, response.Body.String())
	}

	logs, err := auditService.List(10)
	if err != nil || len(logs) != 3 {
		t.Fatalf("read audit logs: logs=%+v err=%v", logs, err)
	}
	seenActions := map[string]bool{}
	for _, log := range logs {
		if log.ActorType != "agent" || log.AgentTokenID == nil {
			t.Fatalf("unexpected audit actor/event: %+v", log)
		}
		if log.EntityType != "network_contact_exchange" && log.EntityType != "network_private_contacts" {
			t.Fatalf("unexpected audited entity: %+v", log)
		}
		if log.Action != "read" && log.Action != "revoke" {
			t.Fatalf("unexpected contact exchange audit action: %+v", log)
		}
		seenActions[log.Action+":"+log.EntityType] = true
		if strings.Contains(log.DetailJSON, agentToken.Token) ||
			strings.Contains(log.DetailJSON, revokeToken.Token) ||
			strings.Contains(log.DetailJSON, profileReadToken.Token) ||
			strings.Contains(log.DetailJSON, profileWriteToken.Token) ||
			strings.Contains(log.DetailJSON, "applicant-private-value") ||
			strings.Contains(log.DetailJSON, "owner-private-value") ||
			strings.Contains(log.DetailJSON, "private-request-message") {
			t.Fatalf("audit log contains secret data: %s", log.DetailJSON)
		}
	}
	for _, expected := range []string{
		"read:network_contact_exchange",
		"read:network_private_contacts",
		"revoke:network_contact_exchange",
	} {
		if !seenActions[expected] {
			t.Fatalf("missing audit event %q: %+v", expected, logs)
		}
	}
}

func intString(value int64) string {
	return strconv.FormatInt(value, 10)
}
