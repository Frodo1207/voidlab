package service

import (
	"errors"
	"path/filepath"
	"testing"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
	"voidlabai/apps/api/internal/storage"
)

func newBuilderNetworkFixture(t *testing.T) *BuilderNetworkService {
	t.Helper()
	db, err := storage.OpenSQLite(filepath.Join(t.TempDir(), "network.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return NewBuilderNetworkService(repository.NewBuilderNetworkRepository(db))
}

func TestBuilderNetworkInviteLimitAndRegistration(t *testing.T) {
	svc := newBuilderNetworkFixture(t)
	code, err := svc.CreateInvite(1, "", 1)
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}
	valid, err := svc.VerifyInvite(code)
	if err != nil || !valid {
		t.Fatalf("verify invite: valid=%v err=%v", valid, err)
	}

	profile := domain.NetworkProfile{DisplayName: "Lin", PrimaryRole: "Product Engineer"}
	session, err := svc.Register("lin@example.com", "secure-password-123", code, profile)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if session.Token == "" || session.Account.ID == 0 || session.Profile.ID == 0 {
		t.Fatalf("register returned incomplete session: %+v", session)
	}

	if _, err := svc.Register("another@example.com", "secure-password-456", code, domain.NetworkProfile{
		DisplayName: "Another", PrimaryRole: "Designer",
	}); !errors.Is(err, repository.ErrNetworkInviteInvalid) {
		t.Fatalf("expected exhausted invite error, got %v", err)
	}
	if _, err := svc.ResolveSession(session.Token); err != nil {
		t.Fatalf("resolve session: %v", err)
	}
	if err := svc.Logout(session.Token); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := svc.ResolveSession(session.Token); err == nil {
		t.Fatal("expected logged out session to be rejected")
	}
}

func TestBuilderNetworkAdminCanListAndDisableInvites(t *testing.T) {
	svc := newBuilderNetworkFixture(t)
	code, err := svc.CreateInvite(3, "", 7)
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}
	invites, err := svc.ListInvites()
	if err != nil || len(invites) != 1 {
		t.Fatalf("list invites: count=%d err=%v", len(invites), err)
	}
	invite := invites[0]
	if invite.MaxUses != 3 || invite.UsedCount != 0 || invite.Status != "active" || invite.CreatedBy != 7 {
		t.Fatalf("unexpected invite metadata: %+v", invite)
	}
	if err := svc.UpdateInviteStatus(invite.ID, "disabled"); err != nil {
		t.Fatalf("disable invite: %v", err)
	}
	valid, err := svc.VerifyInvite(code)
	if err != nil || valid {
		t.Fatalf("disabled invite should not verify: valid=%v err=%v", valid, err)
	}
	invites, err = svc.ListInvites()
	if err != nil || len(invites) != 1 || invites[0].Status != "disabled" {
		t.Fatalf("list disabled invite: invites=%+v err=%v", invites, err)
	}
}

func TestBuilderNetworkContactExchangeRequiresAcceptanceAndConsent(t *testing.T) {
	svc := newBuilderNetworkFixture(t)
	code, err := svc.CreateInvite(2, "", 1)
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}
	owner, err := svc.Register("owner@example.com", "secure-password-owner", code, domain.NetworkProfile{
		DisplayName: "Project Owner", PrimaryRole: "Founder",
	})
	if err != nil {
		t.Fatalf("register owner: %v", err)
	}
	applicant, err := svc.Register("builder@example.com", "secure-password-builder", code, domain.NetworkProfile{
		DisplayName: "Builder", PrimaryRole: "Engineer",
	})
	if err != nil {
		t.Fatalf("register applicant: %v", err)
	}

	if err := svc.UpdateContacts(owner.Account.ID, domain.NetworkContactDetails{WeChat: "owner-wx"}, false, true, false); err != nil {
		t.Fatalf("save owner contact: %v", err)
	}
	if err := svc.UpdateContacts(applicant.Account.ID, domain.NetworkContactDetails{WeChat: "builder-wx"}, false, true, false); err != nil {
		t.Fatalf("save applicant contact: %v", err)
	}

	opportunity, err := svc.CreateOpportunity(owner.Account.ID, domain.NetworkOpportunity{
		Title: "Build a prototype", Summary: "Looking for a technical collaborator.",
		OpportunityType: "collaboration", LookingFor: []string{"Engineer"},
	})
	if err != nil {
		t.Fatalf("create opportunity: %v", err)
	}
	request, err := svc.CreateContactRequest(applicant.Account.ID, domain.NetworkContactRequest{
		TargetType: "opportunity", TargetID: opportunity.ID,
		Message:     "I have built a similar product and can help deliver the prototype.",
		ShareWeChat: true,
	})
	if err != nil {
		t.Fatalf("create contact request: %v", err)
	}

	if _, err := svc.ContactExchange(request.ID, applicant.Account.ID); !errors.Is(err, repository.ErrNetworkNotFound) {
		t.Fatalf("expected exchange to remain hidden before acceptance, got %v", err)
	}
	incoming, err := svc.ContactRequests(owner.Account.ID, true)
	if err != nil || len(incoming) != 1 || incoming[0].Message == "" {
		t.Fatalf("list incoming requests: count=%d err=%v", len(incoming), err)
	}
	if err := svc.AcceptRequest(request.ID, owner.Account.ID); err != nil {
		t.Fatalf("accept request: %v", err)
	}
	exchange, err := svc.ContactExchange(request.ID, applicant.Account.ID)
	if err != nil {
		t.Fatalf("read exchange as applicant: %v", err)
	}
	if exchange.Applicant.WeChat != "builder-wx" || exchange.Recipient.WeChat != "owner-wx" {
		t.Fatalf("unexpected exchanged contacts: %+v", exchange)
	}
	if err := svc.UpdateContactRequestProgress(request.ID, owner.Account.ID, "discussing"); err != nil {
		t.Fatalf("update cooperation progress: %v", err)
	}
	outgoing, err := svc.ContactRequests(applicant.Account.ID, false)
	if err != nil || len(outgoing) != 1 || outgoing[0].ProgressStatus != "discussing" {
		t.Fatalf("progress should be visible to both participants: requests=%+v err=%v", outgoing, err)
	}
	if err := svc.UpdateContactRequestProgress(request.ID, applicant.Account.ID, "invalid"); !errors.Is(err, ErrNetworkValidation) {
		t.Fatalf("invalid progress should be rejected, got %v", err)
	}
}

func TestBuilderNetworkExchangeHonorsPerRequestAndGlobalConsent(t *testing.T) {
	svc := newBuilderNetworkFixture(t)
	code, err := svc.CreateInvite(2, "", 1)
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}
	owner, err := svc.Register("consent-owner@example.com", "secure-password-owner", code, domain.NetworkProfile{
		DisplayName: "Consent Owner", PrimaryRole: "Founder",
	})
	if err != nil {
		t.Fatalf("register owner: %v", err)
	}
	applicant, err := svc.Register("consent-applicant@example.com", "secure-password-applicant", code, domain.NetworkProfile{
		DisplayName: "Consent Applicant", PrimaryRole: "Engineer",
	})
	if err != nil {
		t.Fatalf("register applicant: %v", err)
	}
	if err := svc.UpdateContacts(owner.Account.ID, domain.NetworkContactDetails{
		Email: "owner@example.com", WeChat: "owner-wx",
	}, true, true, false); err != nil {
		t.Fatalf("save owner contacts: %v", err)
	}
	if err := svc.UpdateContacts(applicant.Account.ID, domain.NetworkContactDetails{
		Email: "applicant@example.com", WeChat: "applicant-wx", Phone: "123456",
	}, true, true, true); err != nil {
		t.Fatalf("save applicant contacts: %v", err)
	}
	opportunity, err := svc.CreateOpportunity(owner.Account.ID, domain.NetworkOpportunity{
		Title: "Consent test", Summary: "Test explicit field consent",
		OpportunityType: "collaboration", LookingFor: []string{"Engineer"},
	})
	if err != nil {
		t.Fatalf("create opportunity: %v", err)
	}
	request, err := svc.CreateContactRequest(applicant.Account.ID, domain.NetworkContactRequest{
		TargetType: "opportunity", TargetID: opportunity.ID, Message: "Let's collaborate",
		ShareEmail: true,
	})
	if err != nil {
		t.Fatalf("create contact request: %v", err)
	}
	if err := svc.AcceptRequest(request.ID, owner.Account.ID); err != nil {
		t.Fatalf("accept request: %v", err)
	}
	exchange, err := svc.ContactExchange(request.ID, applicant.Account.ID)
	if err != nil {
		t.Fatalf("read contact exchange: %v", err)
	}
	if exchange.Applicant.Email != "applicant@example.com" || exchange.Applicant.WeChat != "" || exchange.Applicant.Phone != "" {
		t.Fatalf("applicant contacts exceeded per-request consent: %+v", exchange.Applicant)
	}
	if exchange.Recipient.Email != "owner@example.com" || exchange.Recipient.WeChat != "owner-wx" {
		t.Fatalf("recipient contacts should follow recipient sharing preferences: %+v", exchange.Recipient)
	}
	if err := svc.UpdateContacts(owner.Account.ID, domain.NetworkContactDetails{
		Email: "owner@example.com", WeChat: "owner-wx",
	}, true, false, false); err != nil {
		t.Fatalf("revoke recipient WeChat preference: %v", err)
	}
	exchange, err = svc.ContactExchange(request.ID, applicant.Account.ID)
	if err != nil {
		t.Fatalf("read exchange after recipient preference update: %v", err)
	}
	if exchange.Applicant.Email != "applicant@example.com" || exchange.Recipient.Email != "owner@example.com" || exchange.Recipient.WeChat != "" {
		t.Fatalf("recipient revocation must not affect applicant consent: applicant=%+v recipient=%+v", exchange.Applicant, exchange.Recipient)
	}
	if err := svc.UpdateContacts(applicant.Account.ID, domain.NetworkContactDetails{
		Email: "applicant@example.com", WeChat: "applicant-wx", Phone: "123456",
	}, false, true, false); err != nil {
		t.Fatalf("revoke applicant email preference: %v", err)
	}
	if err := svc.UpdateContacts(applicant.Account.ID, domain.NetworkContactDetails{
		Email: "applicant@example.com", WeChat: "applicant-wx", Phone: "123456",
	}, true, true, false); err != nil {
		t.Fatalf("restore applicant email preference: %v", err)
	}
	exchange, err = svc.ContactExchange(request.ID, applicant.Account.ID)
	if err != nil {
		t.Fatalf("read exchange after preference update: %v", err)
	}
	if exchange.Applicant.Email != "" {
		t.Fatalf("revoked field consent must not be restored by re-enabling global preference: %+v", exchange.Applicant)
	}
	if exchange.Recipient.Email != "owner@example.com" {
		t.Fatalf("applicant revocation must not affect recipient consent: %+v", exchange.Recipient)
	}

	_, err = svc.CreateContactRequest(applicant.Account.ID, domain.NetworkContactRequest{
		TargetType: "opportunity", TargetID: opportunity.ID, Message: "Try another contact field",
		SharePhone: true,
	})
	if !errors.Is(err, ErrNetworkValidation) {
		t.Fatalf("field disabled in global preferences should not be authorized, got %v", err)
	}
}

func TestBuilderNetworkOpportunityRequestRequiresPublicOwnerProfile(t *testing.T) {
	db, err := storage.OpenSQLite(filepath.Join(t.TempDir(), "network-hidden-opportunity.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := repository.NewBuilderNetworkRepository(db)
	svc := NewBuilderNetworkService(repo)
	code, err := svc.CreateInvite(2, "", 1)
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}
	owner, err := svc.Register("hidden-owner@example.com", "secure-password-owner", code, domain.NetworkProfile{
		DisplayName: "Hidden Owner", PrimaryRole: "Founder",
	})
	if err != nil {
		t.Fatalf("register owner: %v", err)
	}
	applicant, err := svc.Register("hidden-applicant@example.com", "secure-password-applicant", code, domain.NetworkProfile{
		DisplayName: "Applicant", PrimaryRole: "Engineer",
	})
	if err != nil {
		t.Fatalf("register applicant: %v", err)
	}
	opportunity, err := svc.CreateOpportunity(owner.Account.ID, domain.NetworkOpportunity{
		Title: "Hidden opportunity", Summary: "Should not accept applications",
		OpportunityType: "collaboration", LookingFor: []string{"Engineer"},
	})
	if err != nil {
		t.Fatalf("create opportunity: %v", err)
	}
	if _, err := db.Exec(`UPDATE network_profiles SET status = 'hidden' WHERE id = ?`, owner.Profile.ID); err != nil {
		t.Fatalf("hide opportunity owner: %v", err)
	}
	_, err = svc.CreateContactRequest(applicant.Account.ID, domain.NetworkContactRequest{
		TargetType: "opportunity", TargetID: opportunity.ID, Message: "I would like to apply",
	})
	if !errors.Is(err, repository.ErrNetworkNotFound) {
		t.Fatalf("hidden profile opportunity should reject applications, got %v", err)
	}
}

func TestBuilderNetworkPublicProfileDoesNotExposePrivateContact(t *testing.T) {
	svc := newBuilderNetworkFixture(t)
	code, err := svc.CreateInvite(1, "", 1)
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}
	session, err := svc.Register("private@example.com", "secure-password-789", code, domain.NetworkProfile{
		DisplayName: "Private User", PrimaryRole: "Designer",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := svc.UpdateContacts(session.Account.ID, domain.NetworkContactDetails{
		Email: "private@example.com", WeChat: "secret-wechat",
	}, false, true, false); err != nil {
		t.Fatalf("save contacts: %v", err)
	}
	profiles, err := svc.PublicProfiles()
	if err != nil || len(profiles) != 1 {
		t.Fatalf("list public profiles: count=%d err=%v", len(profiles), err)
	}
	if profiles[0].Email != "" || profiles[0].WeChat != "" || profiles[0].Phone != "" {
		t.Fatalf("public profile exposed private contact: %+v", profiles[0])
	}
}

func TestBuilderNetworkWorkMembersAndOpportunityLifecycle(t *testing.T) {
	svc := newBuilderNetworkFixture(t)
	code, err := svc.CreateInvite(2, "", 1)
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}
	owner, err := svc.Register("work-owner@example.com", "secure-password-owner", code, domain.NetworkProfile{
		DisplayName: "Work Owner", PrimaryRole: "Founder",
	})
	if err != nil {
		t.Fatalf("register owner: %v", err)
	}
	collaborator, err := svc.Register("work-collaborator@example.com", "secure-password-collab", code, domain.NetworkProfile{
		DisplayName: "Collaborator", PrimaryRole: "Engineer",
	})
	if err != nil {
		t.Fatalf("register collaborator: %v", err)
	}

	work, err := svc.CreateWork(owner.Account.ID, domain.NetworkWork{
		Title: "Shared project", Summary: "First version",
		Members: []domain.NetworkWorkMember{{
			ProfileID: collaborator.Profile.ID, Role: "Engineer",
			Contribution: "Built the API", SortOrder: 1,
		}},
	})
	if err != nil {
		t.Fatalf("create work: %v", err)
	}
	if len(work.Members) != 2 || !work.Members[0].IsOwner || work.Members[1].ProfileID != collaborator.Profile.ID {
		t.Fatalf("work members were not persisted: %+v", work.Members)
	}
	if _, err := svc.UpdateWork(collaborator.Account.ID, work.ID, domain.NetworkWork{
		Title: work.Title, Summary: "Unauthorized update", Status: "published",
	}); !errors.Is(err, repository.ErrNetworkNotFound) {
		t.Fatalf("expected non-owner work update to be rejected, got %v", err)
	}

	work.Summary = "Updated project"
	work.Status = "archived"
	work, err = svc.UpdateWork(owner.Account.ID, work.ID, work)
	if err != nil {
		t.Fatalf("archive work: %v", err)
	}
	if work.Status != "archived" || work.Summary != "Updated project" {
		t.Fatalf("work update was not persisted: %+v", work)
	}
	if _, err := svc.PublicWorkBySlug(work.Slug); !errors.Is(err, repository.ErrNetworkNotFound) {
		t.Fatalf("archived work should not be publicly readable, got %v", err)
	}
	work.Status = "published"
	if _, err := svc.UpdateWork(owner.Account.ID, work.ID, work); err != nil {
		t.Fatalf("republish work: %v", err)
	}
	publicWork, err := svc.PublicWorkBySlug(work.Slug)
	if err != nil || len(publicWork.Members) != 2 {
		t.Fatalf("published work detail missing members: work=%+v err=%v", publicWork, err)
	}

	opportunity, err := svc.CreateOpportunity(owner.Account.ID, domain.NetworkOpportunity{
		Title: "Find a designer", Summary: "Design a prototype",
		OpportunityType: "collaboration", LookingFor: []string{"Designer"},
		ExpiresAt: "2099-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("create opportunity: %v", err)
	}
	opportunity.Status = "paused"
	opportunity.Summary = "Temporarily paused"
	opportunity, err = svc.UpdateOpportunity(owner.Account.ID, opportunity.ID, opportunity)
	if err != nil {
		t.Fatalf("pause opportunity: %v", err)
	}
	if opportunity.Status != "paused" || opportunity.Summary != "Temporarily paused" {
		t.Fatalf("opportunity update was not persisted: %+v", opportunity)
	}
	if _, err := svc.PublicOpportunityBySlug(opportunity.Slug); !errors.Is(err, repository.ErrNetworkNotFound) {
		t.Fatalf("paused opportunity should not be publicly readable, got %v", err)
	}
	if _, err := svc.UpdateOpportunity(collaborator.Account.ID, opportunity.ID, opportunity); !errors.Is(err, repository.ErrNetworkNotFound) {
		t.Fatalf("expected non-owner opportunity update to be rejected, got %v", err)
	}
	opportunity.Status = "open"
	if _, err := svc.UpdateOpportunity(owner.Account.ID, opportunity.ID, opportunity); err != nil {
		t.Fatalf("reopen opportunity using returned expiry timestamp: %v", err)
	}
	if _, err := svc.PublicOpportunityBySlug(opportunity.Slug); err != nil {
		t.Fatalf("reopened opportunity should be publicly readable: %v", err)
	}
}

func TestBuilderNetworkDiscoveryFiltersAndPagination(t *testing.T) {
	svc := newBuilderNetworkFixture(t)
	code, err := svc.CreateInvite(2, "", 1)
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}
	engineer, err := svc.Register("discovery-engineer@example.com", "secure-password-engineer", code, domain.NetworkProfile{
		DisplayName: "Discovery Engineer", PrimaryRole: "Engineer", City: "Shanghai",
		Skills: []string{"agents", "go"}, Availability: "open",
	})
	if err != nil {
		t.Fatalf("register engineer: %v", err)
	}
	founder, err := svc.Register("discovery-founder@example.com", "secure-password-founder", code, domain.NetworkProfile{
		DisplayName: "Discovery Founder", PrimaryRole: "Founder", City: "Beijing",
		Skills: []string{"product"}, Availability: "limited",
	})
	if err != nil {
		t.Fatalf("register founder: %v", err)
	}
	if _, err := svc.CreateWork(engineer.Account.ID, domain.NetworkWork{
		Title: "Agent workspace", Summary: "Build useful agents", Tags: []string{"agents", "go"},
	}); err != nil {
		t.Fatalf("create engineer work: %v", err)
	}
	if _, err := svc.CreateWork(founder.Account.ID, domain.NetworkWork{
		Title: "Product studio", Summary: "Explore new products", Tags: []string{"product"},
	}); err != nil {
		t.Fatalf("create founder work: %v", err)
	}
	for _, input := range []domain.NetworkOpportunity{
		{Title: "Find a designer", Summary: "Design a product", OpportunityType: "collaboration", LookingFor: []string{"Designer"}, RemoteAllowed: true},
		{Title: "Find an engineer", Summary: "Build an API", OpportunityType: "collaboration", LookingFor: []string{"Engineer"}, RemoteAllowed: false},
	} {
		if _, err := svc.CreateOpportunity(founder.Account.ID, input); err != nil {
			t.Fatalf("create opportunity: %v", err)
		}
	}

	profiles, err := svc.BrowsePublicProfiles(domain.NetworkDiscoveryQuery{
		Role: "Engineer", City: "Shanghai", Availability: "open", Page: 1, PageSize: 1,
	})
	if err != nil || profiles.Total != 1 || len(profiles.Items) != 1 || profiles.Items[0].ID != engineer.Profile.ID {
		t.Fatalf("profile discovery filters failed: result=%+v err=%v", profiles, err)
	}
	works, err := svc.BrowsePublicWorks(domain.NetworkDiscoveryQuery{
		Query: "Agent", Skill: "agents", Page: 1, PageSize: 1,
	})
	if err != nil || works.Total != 1 || len(works.Items) != 1 || works.Items[0].ProfileID != engineer.Profile.ID {
		t.Fatalf("work discovery filters failed: result=%+v err=%v", works, err)
	}
	opportunities, err := svc.BrowsePublicOpportunities(domain.NetworkDiscoveryQuery{
		Type: "collaboration", Role: "Designer", Remote: boolPointer(true), Page: 1, PageSize: 1,
	})
	if err != nil || opportunities.Total != 1 || len(opportunities.Items) != 1 || opportunities.Items[0].LookingFor[0] != "Designer" {
		t.Fatalf("opportunity discovery filters failed: result=%+v err=%v", opportunities, err)
	}
	paged, err := svc.BrowsePublicOpportunities(domain.NetworkDiscoveryQuery{Page: 2, PageSize: 1})
	if err != nil || paged.Total != 2 || len(paged.Items) != 1 || paged.Page != 2 || paged.PageSize != 1 {
		t.Fatalf("opportunity pagination failed: result=%+v err=%v", paged, err)
	}
}

func boolPointer(value bool) *bool {
	return &value
}
