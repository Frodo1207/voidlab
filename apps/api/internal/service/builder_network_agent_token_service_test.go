package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
	"voidlabai/apps/api/internal/storage"
)

func newNetworkAgentTokenFixture(t *testing.T) (
	*BuilderNetworkService,
	*BuilderNetworkAgentTokenService,
	*sql.DB,
	domain.NetworkSession,
	domain.NetworkSession,
) {
	t.Helper()
	db, err := storage.OpenSQLite(filepath.Join(t.TempDir(), "network-agent.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	networkRepo := repository.NewBuilderNetworkRepository(db)
	networkService := NewBuilderNetworkService(networkRepo)
	tokenService := NewBuilderNetworkAgentTokenService(networkRepo, repository.NewBuilderNetworkAgentTokenRepository(db))
	invite, err := networkService.CreateInvite(2, "", 1)
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}
	owner, err := networkService.Register("agent-owner@example.com", "secure-password-owner", invite, domain.NetworkProfile{
		DisplayName: "Agent Owner", PrimaryRole: "Engineer",
	})
	if err != nil {
		t.Fatalf("register owner: %v", err)
	}
	other, err := networkService.Register("agent-other@example.com", "secure-password-other", invite, domain.NetworkProfile{
		DisplayName: "Other User", PrimaryRole: "Designer",
	})
	if err != nil {
		t.Fatalf("register other user: %v", err)
	}
	return networkService, tokenService, db, owner, other
}

func TestBuilderNetworkAgentTokenLifecycleAndOwnership(t *testing.T) {
	_, tokenService, db, owner, other := newNetworkAgentTokenFixture(t)
	created, err := tokenService.Create(owner.Account.ID, NetworkAgentTokenCreateInput{
		Name: "Local assistant",
		Scopes: []string{
			"network_profiles:read",
			"network_works:write",
		},
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	if !strings.HasPrefix(created.Token, "vla_network_") || created.Record.ID == 0 {
		t.Fatalf("unexpected token create result: %+v", created)
	}
	var storedHash string
	if err := db.QueryRow(`SELECT token_hash FROM network_agent_tokens WHERE id = ?`, created.Record.ID).Scan(&storedHash); err != nil {
		t.Fatalf("read token hash: %v", err)
	}
	if storedHash == created.Token || storedHash != hashToken(created.Token) {
		t.Fatal("token must be stored as a hash")
	}

	resolved, err := tokenService.ResolveToken(created.Token)
	if err != nil {
		t.Fatalf("resolve token: %v", err)
	}
	if resolved.Account.ID != owner.Account.ID || resolved.Token.ProfileID != owner.Profile.ID {
		t.Fatalf("token resolved to wrong owner: %+v", resolved)
	}

	listed, err := tokenService.List(owner.Account.ID)
	if err != nil || len(listed) != 1 {
		t.Fatalf("list owner tokens: count=%d err=%v", len(listed), err)
	}
	encoded, err := json.Marshal(listed)
	if err != nil {
		t.Fatalf("marshal token list: %v", err)
	}
	if strings.Contains(string(encoded), created.Token) {
		t.Fatal("token list must not return token plaintext")
	}
	otherTokens, err := tokenService.List(other.Account.ID)
	if err != nil || len(otherTokens) != 0 {
		t.Fatalf("other account should not see token: tokens=%+v err=%v", otherTokens, err)
	}
	if _, err := tokenService.UpdateStatus(other.Account.ID, created.Record.ID, false); !errors.Is(err, repository.ErrNetworkNotFound) {
		t.Fatalf("other account should not disable token, got %v", err)
	}

	if _, err := tokenService.UpdateStatus(owner.Account.ID, created.Record.ID, false); err != nil {
		t.Fatalf("disable token: %v", err)
	}
	if _, err := tokenService.ResolveToken(created.Token); !errors.Is(err, ErrNetworkUnauthorized) {
		t.Fatalf("disabled token should be rejected, got %v", err)
	}
}

func TestNormalizeNetworkAgentScopes(t *testing.T) {
	scopes, err := normalizeNetworkAgentScopes([]string{
		" network_works:read ",
		"network_works:read",
		"network_projects:read",
		"network_contact_exchange:write",
	})
	if err != nil || len(scopes) != 3 {
		t.Fatalf("expected normalized unique scopes, scopes=%v err=%v", scopes, err)
	}
	if _, err := normalizeNetworkAgentScopes([]string{"admin:*"}); !errors.Is(err, ErrNetworkValidation) {
		t.Fatalf("unknown scope should be rejected, got %v", err)
	}
	if _, err := normalizeNetworkAgentScopes([]string{" "}); !errors.Is(err, ErrNetworkValidation) {
		t.Fatalf("empty scope list should be rejected, got %v", err)
	}
}

func TestBuilderNetworkContactExchangeCanBeRevokedByEitherParty(t *testing.T) {
	networkService, _, _, owner, applicant := newNetworkAgentTokenFixture(t)
	if err := networkService.UpdateContacts(owner.Account.ID, domain.NetworkContactDetails{WeChat: "owner-private"}, false, true, false); err != nil {
		t.Fatalf("save owner contacts: %v", err)
	}
	if err := networkService.UpdateContacts(applicant.Account.ID, domain.NetworkContactDetails{WeChat: "applicant-private"}, false, true, false); err != nil {
		t.Fatalf("save applicant contacts: %v", err)
	}
	opportunity, err := networkService.CreateOpportunity(owner.Account.ID, domain.NetworkOpportunity{
		Title: "Collaborate", Summary: "Build together", OpportunityType: "collaboration",
		LookingFor: []string{"Engineer"},
	})
	if err != nil {
		t.Fatalf("create opportunity: %v", err)
	}
	request, err := networkService.CreateContactRequest(applicant.Account.ID, domain.NetworkContactRequest{
		TargetType: "opportunity", TargetID: opportunity.ID, Message: "I can help",
		ShareWeChat: true,
	})
	if err != nil {
		t.Fatalf("create contact request: %v", err)
	}
	if err := networkService.AcceptRequest(request.ID, owner.Account.ID); err != nil {
		t.Fatalf("accept request: %v", err)
	}
	if _, err := networkService.ContactExchange(request.ID, applicant.Account.ID); err != nil {
		t.Fatalf("exchange should be readable before revocation: %v", err)
	}
	if err := networkService.RevokeContactExchange(request.ID, owner.Account.ID); err != nil {
		t.Fatalf("recipient revoke: %v", err)
	}
	for _, accountID := range []int64{owner.Account.ID, applicant.Account.ID} {
		if _, err := networkService.ContactExchange(request.ID, accountID); !errors.Is(err, repository.ErrNetworkNotFound) {
			t.Fatalf("exchange should be hidden after revocation for account %d, got %v", accountID, err)
		}
	}
	if err := networkService.RevokeContactExchange(request.ID, applicant.Account.ID); !errors.Is(err, repository.ErrNetworkNotFound) {
		t.Fatalf("already revoked exchange should not be revoked again, got %v", err)
	}
}
