package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

var ErrNetworkUnauthorized = errors.New("invalid email or password")
var ErrNetworkValidation = errors.New("invalid builder network input")

type BuilderNetworkService struct {
	repo *repository.BuilderNetworkRepository
}

func NewBuilderNetworkService(repo *repository.BuilderNetworkRepository) *BuilderNetworkService {
	return &BuilderNetworkService{repo: repo}
}

func (s *BuilderNetworkService) CreateInvite(maxUses int, expiresAt string, createdBy int64) (string, error) {
	if maxUses < 1 || maxUses > 1000 {
		return "", fmt.Errorf("%w: max_uses must be between 1 and 1000", ErrNetworkValidation)
	}
	if expiresAt != "" {
		expiry, err := time.Parse(time.RFC3339, expiresAt)
		if err != nil {
			return "", fmt.Errorf("%w: expires_at must be RFC3339", ErrNetworkValidation)
		}
		expiresAt = expiry.UTC().Format("2006-01-02 15:04:05")
	}
	code, err := randomToken(9)
	if err != nil {
		return "", err
	}
	code = strings.ToUpper(code)
	if err := s.repo.CreateInvite(hashToken(code), maxUses, expiresAt, createdBy); err != nil {
		return "", err
	}
	return code, nil
}

func (s *BuilderNetworkService) ListInvites() ([]domain.NetworkInviteCode, error) {
	return s.repo.ListInvites()
}

func (s *BuilderNetworkService) UpdateInviteStatus(id int64, status string) error {
	if id <= 0 || (status != "active" && status != "disabled") {
		return fmt.Errorf("%w: status must be active or disabled", ErrNetworkValidation)
	}
	return s.repo.UpdateInviteStatus(id, status)
}

func (s *BuilderNetworkService) VerifyInvite(code string) (bool, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) < 8 || len(code) > 32 {
		return false, nil
	}
	return s.repo.VerifyInvite(hashToken(code))
}

func (s *BuilderNetworkService) Register(email, password, inviteCode string, profile domain.NetworkProfile) (domain.NetworkSession, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if _, err := mail.ParseAddress(email); err != nil || !strings.Contains(email, "@") {
		return domain.NetworkSession{}, fmt.Errorf("%w: valid email is required", ErrNetworkValidation)
	}
	if len(password) < 10 {
		return domain.NetworkSession{}, fmt.Errorf("%w: password must be at least 10 characters", ErrNetworkValidation)
	}
	profile.DisplayName = strings.TrimSpace(profile.DisplayName)
	profile.PrimaryRole = strings.TrimSpace(profile.PrimaryRole)
	if profile.DisplayName == "" || profile.PrimaryRole == "" {
		return domain.NetworkSession{}, fmt.Errorf("%w: display_name and primary_role are required", ErrNetworkValidation)
	}
	if profile.Availability == "" {
		profile.Availability = "open"
	}
	if !validAvailability(profile.Availability) {
		return domain.NetworkSession{}, fmt.Errorf("%w: availability must be open, limited, or unavailable", ErrNetworkValidation)
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.NetworkSession{}, err
	}
	account, _, err := s.repo.Register(email, string(passwordHash), hashToken(strings.ToUpper(strings.TrimSpace(inviteCode))), profile)
	if err != nil {
		return domain.NetworkSession{}, err
	}
	return s.newSession(account)
}

func (s *BuilderNetworkService) Login(email, password string) (domain.NetworkSession, error) {
	account, passwordHash, err := s.repo.FindAccountByEmail(email)
	if err != nil || account.Status != "active" || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
		return domain.NetworkSession{}, ErrNetworkUnauthorized
	}
	return s.newSession(account)
}

func (s *BuilderNetworkService) ResolveSession(token string) (domain.NetworkAccount, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.NetworkAccount{}, ErrNetworkUnauthorized
	}
	return s.repo.ResolveSession(hashToken(token))
}

func (s *BuilderNetworkService) Logout(token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrNetworkUnauthorized
	}
	return s.repo.DeleteSession(hashToken(strings.TrimSpace(token)))
}

func (s *BuilderNetworkService) Me(accountID int64) (domain.NetworkAccount, domain.NetworkProfile, error) {
	return s.repo.GetAccountProfile(accountID)
}

func (s *BuilderNetworkService) UpdateProfile(accountID int64, profile domain.NetworkProfile) (domain.NetworkProfile, error) {
	profile.DisplayName = strings.TrimSpace(profile.DisplayName)
	profile.PrimaryRole = strings.TrimSpace(profile.PrimaryRole)
	if profile.DisplayName == "" || profile.PrimaryRole == "" {
		return domain.NetworkProfile{}, fmt.Errorf("%w: display_name and primary_role are required", ErrNetworkValidation)
	}
	if len([]rune(profile.DisplayName)) > 80 || len([]rune(profile.PrimaryRole)) > 80 ||
		len([]rune(profile.Headline)) > 180 || len([]rune(profile.Bio)) > 2000 ||
		len([]rune(profile.City)) > 100 || len([]rune(profile.AvailableFrom)) > 80 {
		return domain.NetworkProfile{}, fmt.Errorf("%w: profile fields exceed the allowed length", ErrNetworkValidation)
	}
	if profile.Availability == "" {
		profile.Availability = "open"
	}
	if !validAvailability(profile.Availability) {
		return domain.NetworkProfile{}, fmt.Errorf("%w: invalid availability", ErrNetworkValidation)
	}
	return s.repo.UpdateProfile(profile, accountID)
}

func (s *BuilderNetworkService) UpdateContacts(accountID int64, details domain.NetworkContactDetails, shareEmail, shareWeChat, sharePhone bool) error {
	_, profile, err := s.repo.GetAccountProfile(accountID)
	if err != nil {
		return err
	}
	if shareEmail && strings.TrimSpace(details.Email) == "" ||
		shareWeChat && strings.TrimSpace(details.WeChat) == "" ||
		sharePhone && strings.TrimSpace(details.Phone) == "" {
		return fmt.Errorf("%w: a contact value is required before enabling its sharing", ErrNetworkValidation)
	}
	return s.repo.SaveContacts(profile.ID, strings.TrimSpace(details.Email), strings.TrimSpace(details.WeChat),
		strings.TrimSpace(details.Phone), shareEmail, shareWeChat, sharePhone)
}

func (s *BuilderNetworkService) Contacts(accountID int64) (domain.NetworkContactDetails, bool, bool, bool, error) {
	_, profile, err := s.repo.GetAccountProfile(accountID)
	if err != nil {
		return domain.NetworkContactDetails{}, false, false, false, err
	}
	return s.repo.GetContacts(profile.ID)
}

func (s *BuilderNetworkService) PublicProfiles() ([]domain.NetworkProfile, error) {
	return s.repo.ListPublicProfiles()
}

func (s *BuilderNetworkService) BrowsePublicProfiles(query domain.NetworkDiscoveryQuery) (domain.NetworkPage[domain.NetworkProfile], error) {
	return s.repo.BrowsePublicProfiles(query)
}

func (s *BuilderNetworkService) PublicProfileBySlug(slug string) (domain.NetworkProfile, error) {
	return s.repo.GetPublicProfileBySlug(strings.TrimSpace(slug))
}

func (s *BuilderNetworkService) CreateWork(accountID int64, work domain.NetworkWork) (domain.NetworkWork, error) {
	work.Title = strings.TrimSpace(work.Title)
	if work.Title == "" || len([]rune(work.Title)) > 100 || len([]rune(work.Summary)) > 300 || len([]rune(work.Body)) > 10000 {
		return domain.NetworkWork{}, fmt.Errorf("%w: title, summary, or body is invalid", ErrNetworkValidation)
	}
	if work.Status == "" {
		work.Status = "published"
	}
	if !validWorkStatus(work.Status) {
		return domain.NetworkWork{}, fmt.Errorf("%w: status must be draft, published, or archived", ErrNetworkValidation)
	}
	_, profile, err := s.repo.GetAccountProfile(accountID)
	if err != nil {
		return domain.NetworkWork{}, err
	}
	if err := validateWorkMembers(profile.ID, work.Members); err != nil {
		return domain.NetworkWork{}, err
	}
	work.Slug = slugWithSuffix(work.Title)
	return s.repo.CreateWork(profile.ID, work)
}

func (s *BuilderNetworkService) UpdateWork(accountID, workID int64, work domain.NetworkWork) (domain.NetworkWork, error) {
	if workID <= 0 {
		return domain.NetworkWork{}, fmt.Errorf("%w: invalid work id", ErrNetworkValidation)
	}
	work.Title = strings.TrimSpace(work.Title)
	if work.Title == "" || len([]rune(work.Title)) > 100 || len([]rune(work.Summary)) > 300 || len([]rune(work.Body)) > 10000 {
		return domain.NetworkWork{}, fmt.Errorf("%w: title, summary, or body is invalid", ErrNetworkValidation)
	}
	if !validWorkStatus(work.Status) {
		return domain.NetworkWork{}, fmt.Errorf("%w: status must be draft, published, or archived", ErrNetworkValidation)
	}
	_, profile, err := s.repo.GetAccountProfile(accountID)
	if err != nil {
		return domain.NetworkWork{}, err
	}
	if err := validateWorkMembers(profile.ID, work.Members); err != nil {
		return domain.NetworkWork{}, err
	}
	return s.repo.UpdateWork(profile.ID, workID, work)
}

func (s *BuilderNetworkService) PublicWorks() ([]domain.NetworkWork, error) {
	return s.repo.ListPublicWorks()
}

func (s *BuilderNetworkService) BrowsePublicWorks(query domain.NetworkDiscoveryQuery) (domain.NetworkPage[domain.NetworkWork], error) {
	return s.repo.BrowsePublicWorks(query)
}

func (s *BuilderNetworkService) PublicWorkBySlug(slug string) (domain.NetworkWork, error) {
	return s.repo.GetPublicWorkBySlug(strings.TrimSpace(slug))
}

func (s *BuilderNetworkService) MyWorks(accountID int64) ([]domain.NetworkWork, error) {
	_, profile, err := s.repo.GetAccountProfile(accountID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListWorksByProfile(profile.ID)
}

func (s *BuilderNetworkService) CreateOpportunity(accountID int64, opportunity domain.NetworkOpportunity) (domain.NetworkOpportunity, error) {
	opportunity.Title = strings.TrimSpace(opportunity.Title)
	opportunity.OpportunityType = strings.TrimSpace(opportunity.OpportunityType)
	if opportunity.Title == "" || opportunity.OpportunityType == "" || strings.TrimSpace(opportunity.Summary) == "" {
		return domain.NetworkOpportunity{}, fmt.Errorf("%w: title, opportunity_type, and summary are required", ErrNetworkValidation)
	}
	if len(opportunity.LookingFor) == 0 {
		return domain.NetworkOpportunity{}, fmt.Errorf("%w: at least one role is required", ErrNetworkValidation)
	}
	if len([]rune(opportunity.Title)) > 100 || len([]rune(opportunity.Summary)) > 3000 || len(opportunity.LookingFor) > 12 {
		return domain.NetworkOpportunity{}, fmt.Errorf("%w: opportunity fields exceed the allowed limits", ErrNetworkValidation)
	}
	if opportunity.Status == "" {
		opportunity.Status = "open"
	}
	if opportunity.Status != "open" && opportunity.Status != "draft" {
		return domain.NetworkOpportunity{}, fmt.Errorf("%w: new opportunities must be draft or open", ErrNetworkValidation)
	}
	if opportunity.ExpiresAt != "" {
		expiresAt, err := parseOpportunityExpiry(opportunity.ExpiresAt)
		if err != nil {
			return domain.NetworkOpportunity{}, fmt.Errorf("%w: expires_at must be RFC3339 or a stored UTC timestamp", ErrNetworkValidation)
		}
		opportunity.ExpiresAt = expiresAt
	}
	_, profile, err := s.repo.GetAccountProfile(accountID)
	if err != nil {
		return domain.NetworkOpportunity{}, err
	}
	opportunity.Slug = slugWithSuffix(opportunity.Title)
	return s.repo.CreateOpportunity(profile.ID, opportunity)
}

func (s *BuilderNetworkService) UpdateOpportunity(accountID, opportunityID int64, opportunity domain.NetworkOpportunity) (domain.NetworkOpportunity, error) {
	if opportunityID <= 0 {
		return domain.NetworkOpportunity{}, fmt.Errorf("%w: invalid opportunity id", ErrNetworkValidation)
	}
	opportunity.Title = strings.TrimSpace(opportunity.Title)
	opportunity.OpportunityType = strings.TrimSpace(opportunity.OpportunityType)
	if opportunity.Title == "" || opportunity.OpportunityType == "" || strings.TrimSpace(opportunity.Summary) == "" {
		return domain.NetworkOpportunity{}, fmt.Errorf("%w: title, opportunity_type, and summary are required", ErrNetworkValidation)
	}
	if len([]rune(opportunity.Title)) > 100 || len([]rune(opportunity.Summary)) > 3000 || len(opportunity.LookingFor) == 0 || len(opportunity.LookingFor) > 12 {
		return domain.NetworkOpportunity{}, fmt.Errorf("%w: opportunity fields exceed the allowed limits", ErrNetworkValidation)
	}
	if !validOpportunityStatus(opportunity.Status) {
		return domain.NetworkOpportunity{}, fmt.Errorf("%w: invalid opportunity status", ErrNetworkValidation)
	}
	if opportunity.ExpiresAt != "" {
		expiresAt, err := parseOpportunityExpiry(opportunity.ExpiresAt)
		if err != nil {
			return domain.NetworkOpportunity{}, fmt.Errorf("%w: expires_at must be RFC3339 or a stored UTC timestamp", ErrNetworkValidation)
		}
		opportunity.ExpiresAt = expiresAt
	}
	_, profile, err := s.repo.GetAccountProfile(accountID)
	if err != nil {
		return domain.NetworkOpportunity{}, err
	}
	return s.repo.UpdateOpportunity(profile.ID, opportunityID, opportunity)
}

func (s *BuilderNetworkService) MyOpportunities(accountID int64) ([]domain.NetworkOpportunity, error) {
	_, profile, err := s.repo.GetAccountProfile(accountID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListOpportunitiesByProfile(profile.ID)
}

func (s *BuilderNetworkService) PublicOpportunities() ([]domain.NetworkOpportunity, error) {
	return s.repo.ListPublicOpportunities()
}

func (s *BuilderNetworkService) BrowsePublicOpportunities(query domain.NetworkDiscoveryQuery) (domain.NetworkPage[domain.NetworkOpportunity], error) {
	return s.repo.BrowsePublicOpportunities(query)
}

func (s *BuilderNetworkService) PublicOpportunityBySlug(slug string) (domain.NetworkOpportunity, error) {
	return s.repo.GetPublicOpportunityBySlug(strings.TrimSpace(slug))
}

func (s *BuilderNetworkService) CreateContactRequest(accountID int64, request domain.NetworkContactRequest) (domain.NetworkContactRequest, error) {
	request.Message = strings.TrimSpace(request.Message)
	if request.TargetID <= 0 || request.Message == "" || len([]rune(request.Message)) > 500 {
		return domain.NetworkContactRequest{}, fmt.Errorf("%w: target_id and a message of up to 500 characters are required", ErrNetworkValidation)
	}
	if request.TargetType != "profile" && request.TargetType != "opportunity" {
		return domain.NetworkContactRequest{}, fmt.Errorf("%w: target_type must be profile or opportunity", ErrNetworkValidation)
	}
	if len(request.RelatedWorkIDs) > 3 {
		return domain.NetworkContactRequest{}, fmt.Errorf("%w: no more than three works may be attached", ErrNetworkValidation)
	}
	_, profile, err := s.repo.GetAccountProfile(accountID)
	if err != nil {
		return domain.NetworkContactRequest{}, err
	}
	contacts, shareEmail, shareWeChat, sharePhone, err := s.repo.GetContacts(profile.ID)
	if err != nil {
		return domain.NetworkContactRequest{}, err
	}
	if request.ShareEmail && (!shareEmail || contacts.Email == "") ||
		request.ShareWeChat && (!shareWeChat || contacts.WeChat == "") ||
		request.SharePhone && (!sharePhone || contacts.Phone == "") {
		return domain.NetworkContactRequest{}, fmt.Errorf("%w: selected contact details must be configured and allowed for exchange in your workspace", ErrNetworkValidation)
	}
	return s.repo.CreateContactRequest(profile.ID, request)
}

func (s *BuilderNetworkService) ContactRequests(accountID int64, incoming bool) ([]domain.NetworkContactRequest, error) {
	return s.repo.ListContactRequests(accountID, incoming)
}

func (s *BuilderNetworkService) AcceptRequest(requestID, accountID int64) error {
	return s.repo.UpdateContactRequestStatus(requestID, accountID, "accepted")
}

func (s *BuilderNetworkService) DeclineRequest(requestID, accountID int64) error {
	return s.repo.UpdateContactRequestStatus(requestID, accountID, "declined")
}

func (s *BuilderNetworkService) WithdrawRequest(requestID, accountID int64) error {
	return s.repo.UpdateContactRequestStatus(requestID, accountID, "withdrawn")
}

func (s *BuilderNetworkService) UpdateContactRequestProgress(requestID, accountID int64, progress string) error {
	progress = strings.TrimSpace(progress)
	switch progress {
	case "contacted", "discussing", "collaborating", "completed", "closed":
	default:
		return fmt.Errorf("%w: progress_status is invalid", ErrNetworkValidation)
	}
	return s.repo.UpdateContactRequestProgress(requestID, accountID, progress)
}

func (s *BuilderNetworkService) ContactExchange(requestID, accountID int64) (domain.NetworkContactExchange, error) {
	return s.repo.GetContactExchange(requestID, accountID)
}

func (s *BuilderNetworkService) RevokeContactExchange(requestID, accountID int64) error {
	return s.repo.RevokeContactExchange(requestID, accountID)
}

func (s *BuilderNetworkService) newSession(account domain.NetworkAccount) (domain.NetworkSession, error) {
	token, err := randomToken(32)
	if err != nil {
		return domain.NetworkSession{}, err
	}
	expiresAt := time.Now().UTC().Add(30 * 24 * time.Hour).Format("2006-01-02 15:04:05")
	if err := s.repo.CreateSession(account.ID, hashToken(token), expiresAt); err != nil {
		return domain.NetworkSession{}, err
	}
	_, profile, err := s.repo.GetAccountProfile(account.ID)
	if err != nil {
		return domain.NetworkSession{}, err
	}
	profile.AccountID = account.ID
	return domain.NetworkSession{Token: token, Account: account, Profile: profile}, nil
}

func randomToken(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func validAvailability(value string) bool {
	return value == "open" || value == "limited" || value == "unavailable"
}

func validWorkStatus(value string) bool {
	return value == "draft" || value == "published" || value == "archived"
}

func validOpportunityStatus(value string) bool {
	switch value {
	case "draft", "open", "paused", "filled", "closed", "expired":
		return true
	default:
		return false
	}
}

func parseOpportunityExpiry(value string) (string, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC().Format("2006-01-02 15:04:05"), nil
	}
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.UTC)
	if err != nil {
		return "", err
	}
	return parsed.UTC().Format("2006-01-02 15:04:05"), nil
}

func validateWorkMembers(ownerProfileID int64, members []domain.NetworkWorkMember) error {
	if len(members) > 20 {
		return fmt.Errorf("%w: no more than 20 collaborators may be attached", ErrNetworkValidation)
	}
	seen := map[int64]struct{}{ownerProfileID: {}}
	for _, member := range members {
		if member.ProfileID <= 0 {
			return fmt.Errorf("%w: collaborator profile_id must be positive", ErrNetworkValidation)
		}
		if member.ProfileID == ownerProfileID {
			continue
		}
		if _, exists := seen[member.ProfileID]; exists {
			return fmt.Errorf("%w: collaborator profiles must be unique", ErrNetworkValidation)
		}
		seen[member.ProfileID] = struct{}{}
		if strings.TrimSpace(member.Role) == "" {
			return fmt.Errorf("%w: collaborator role is required", ErrNetworkValidation)
		}
		if len([]rune(strings.TrimSpace(member.Role))) > 80 || len([]rune(strings.TrimSpace(member.Contribution))) > 500 {
			return fmt.Errorf("%w: collaborator role or contribution is too long", ErrNetworkValidation)
		}
		if member.SortOrder < 0 {
			return fmt.Errorf("%w: collaborator sort_order cannot be negative", ErrNetworkValidation)
		}
	}
	return nil
}

func slugWithSuffix(title string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(title) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteByte('-')
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		slug = "work"
	}
	return fmt.Sprintf("%s-%d", slug, time.Now().UnixNano())
}
