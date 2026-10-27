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

func (s *BuilderNetworkService) MyProjects(accountID int64) ([]domain.NetworkProject, error) {
	_, profile, err := s.repo.GetAccountProfile(accountID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListProjectsByOwner(profile.ID)
}

func (s *BuilderNetworkService) ProjectByID(accountID, projectID int64) (domain.NetworkProject, error) {
	if projectID <= 0 {
		return domain.NetworkProject{}, fmt.Errorf("%w: invalid project id", ErrNetworkValidation)
	}
	_, profile, err := s.repo.GetAccountProfile(accountID)
	if err != nil {
		return domain.NetworkProject{}, err
	}
	return s.repo.GetProjectByOwner(profile.ID, projectID)
}

func (s *BuilderNetworkService) CreateProject(accountID int64, project domain.NetworkProject) (domain.NetworkProject, error) {
	project.Title = strings.TrimSpace(project.Title)
	project.Summary = strings.TrimSpace(project.Summary)
	project.TeamIntro = strings.TrimSpace(project.TeamIntro)
	if project.Title == "" || project.Summary == "" || len([]rune(project.Title)) > 100 || len([]rune(project.Summary)) > 3000 {
		return domain.NetworkProject{}, fmt.Errorf("%w: project title and summary are required and must fit limits", ErrNetworkValidation)
	}
	if project.Status == "" {
		project.Status = "draft"
	}
	if project.Status != "draft" && project.Status != "published" && project.Status != "archived" {
		return domain.NetworkProject{}, fmt.Errorf("%w: invalid project status", ErrNetworkValidation)
	}
	if len([]rune(project.TeamIntro)) > 2000 {
		return domain.NetworkProject{}, fmt.Errorf("%w: team_intro exceeds 2000 characters", ErrNetworkValidation)
	}
	if err := s.validateProjectDetails(&project); err != nil {
		return domain.NetworkProject{}, err
	}
	_, profile, err := s.repo.GetAccountProfile(accountID)
	if err != nil {
		return domain.NetworkProject{}, err
	}
	project.Slug = slugWithSuffix(project.Title)
	return s.repo.CreateProject(profile.ID, project)
}

func (s *BuilderNetworkService) UpdateProject(accountID, projectID int64, project domain.NetworkProject) (domain.NetworkProject, error) {
	if projectID <= 0 {
		return domain.NetworkProject{}, fmt.Errorf("%w: invalid project id", ErrNetworkValidation)
	}
	project.Title = strings.TrimSpace(project.Title)
	project.Summary = strings.TrimSpace(project.Summary)
	project.TeamIntro = strings.TrimSpace(project.TeamIntro)
	if project.Title == "" || project.Summary == "" || len([]rune(project.Title)) > 100 || len([]rune(project.Summary)) > 3000 {
		return domain.NetworkProject{}, fmt.Errorf("%w: project title and summary are required and must fit limits", ErrNetworkValidation)
	}
	if project.Status != "draft" && project.Status != "published" && project.Status != "archived" {
		return domain.NetworkProject{}, fmt.Errorf("%w: invalid project status", ErrNetworkValidation)
	}
	if len([]rune(project.TeamIntro)) > 2000 {
		return domain.NetworkProject{}, fmt.Errorf("%w: team_intro exceeds 2000 characters", ErrNetworkValidation)
	}
	if err := s.validateProjectDetails(&project); err != nil {
		return domain.NetworkProject{}, err
	}
	_, profile, err := s.repo.GetAccountProfile(accountID)
	if err != nil {
		return domain.NetworkProject{}, err
	}
	return s.repo.UpdateProject(profile.ID, projectID, project)
}

func (s *BuilderNetworkService) validateProjectDetails(project *domain.NetworkProject) error {
	opportunity := domain.NetworkOpportunity{
		ProjectPlan: project.ProjectPlan,
		TeamIntro:   project.TeamIntro,
		TeamMembers: project.Members,
	}
	if err := s.validateOpportunityProjectDetails(&opportunity); err != nil {
		return err
	}
	project.ProjectPlan = opportunity.ProjectPlan
	project.Members = opportunity.TeamMembers
	return nil
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
	if err := s.validateOpportunityProjectDetails(&opportunity); err != nil {
		return domain.NetworkOpportunity{}, err
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
	if err := s.validateOpportunityProjectDetails(&opportunity); err != nil {
		return domain.NetworkOpportunity{}, err
	}
	_, profile, err := s.repo.GetAccountProfile(accountID)
	if err != nil {
		return domain.NetworkOpportunity{}, err
	}
	return s.repo.UpdateOpportunity(profile.ID, opportunityID, opportunity)
}

func (s *BuilderNetworkService) validateOpportunityProjectDetails(opportunity *domain.NetworkOpportunity) error {
	opportunity.TeamIntro = strings.TrimSpace(opportunity.TeamIntro)
	if len([]rune(opportunity.TeamIntro)) > 2000 {
		return fmt.Errorf("%w: team_intro exceeds 2000 characters", ErrNetworkValidation)
	}

	if opportunity.ProjectPlan != nil {
		plan := opportunity.ProjectPlan
		if len(plan.Milestones) > 30 || len(plan.Sections) > 20 {
			return fmt.Errorf("%w: project plan allows up to 30 milestones and 20 sections", ErrNetworkValidation)
		}
		teamMemberIDs := make(map[int64]struct{}, len(opportunity.TeamMembers))
		for _, member := range opportunity.TeamMembers {
			teamMemberIDs[member.ProfileID] = struct{}{}
		}
		for i := range plan.Milestones {
			milestone := &plan.Milestones[i]
			milestone.Title = strings.TrimSpace(milestone.Title)
			milestone.Description = strings.TrimSpace(milestone.Description)
			if milestone.Title == "" || len([]rune(milestone.Title)) > 120 || len([]rune(milestone.Description)) > 2000 {
				return fmt.Errorf("%w: each milestone needs a title of up to 120 characters", ErrNetworkValidation)
			}
			if milestone.Status == "" {
				milestone.Status = "planned"
			}
			if milestone.Status != "planned" && milestone.Status != "in_progress" && milestone.Status != "completed" && milestone.Status != "blocked" {
				return fmt.Errorf("%w: invalid milestone status", ErrNetworkValidation)
			}
			if len(milestone.Goals) > 20 || len(milestone.Deliverables) > 20 || len(milestone.AcceptanceCriteria) > 20 {
				return fmt.Errorf("%w: milestone lists allow up to 20 items", ErrNetworkValidation)
			}
			milestone.Goals = cleanNetworkTextItems(milestone.Goals)
			milestone.Deliverables = cleanNetworkTextItems(milestone.Deliverables)
			milestone.AcceptanceCriteria = cleanNetworkTextItems(milestone.AcceptanceCriteria)
			for _, ownerID := range milestone.OwnerMemberIDs {
				if _, ok := teamMemberIDs[ownerID]; !ok {
					return fmt.Errorf("%w: milestone owners must be project team members", ErrNetworkValidation)
				}
			}
			if milestone.Timing != nil {
				timing := milestone.Timing
				timing.Kind = strings.TrimSpace(timing.Kind)
				timing.Label = strings.TrimSpace(timing.Label)
				timing.DurationUnit = strings.TrimSpace(timing.DurationUnit)
				switch timing.Kind {
				case "fixed_dates":
					if !validPlanDate(timing.StartAt) || !validPlanDate(timing.EndAt) {
						return fmt.Errorf("%w: fixed_dates timing requires valid start_at and end_at", ErrNetworkValidation)
					}
					start, _ := parsePlanDate(timing.StartAt)
					end, _ := parsePlanDate(timing.EndAt)
					if end.Before(start) {
						return fmt.Errorf("%w: timing end_at must not precede start_at", ErrNetworkValidation)
					}
				case "duration":
					if timing.DurationValue <= 0 || timing.DurationUnit == "" || len([]rune(timing.DurationUnit)) > 24 {
						return fmt.Errorf("%w: duration timing requires a positive duration and unit", ErrNetworkValidation)
					}
				case "relative", "flexible", "unspecified":
				default:
					return fmt.Errorf("%w: timing kind must be fixed_dates, duration, relative, flexible, or unspecified", ErrNetworkValidation)
				}
				if len([]rune(timing.Label)) > 120 {
					return fmt.Errorf("%w: timing label exceeds 120 characters", ErrNetworkValidation)
				}
			}
		}
		for i := range plan.Sections {
			section := &plan.Sections[i]
			section.Type = strings.TrimSpace(section.Type)
			section.Title = strings.TrimSpace(section.Title)
			section.Body = strings.TrimSpace(section.Body)
			if section.Type == "" {
				section.Type = "custom"
			}
			if section.Title == "" || len([]rune(section.Title)) > 100 || len([]rune(section.Body)) > 5000 || len(section.Items) > 30 {
				return fmt.Errorf("%w: each plan section needs a title and may contain up to 30 items", ErrNetworkValidation)
			}
			section.Items = cleanNetworkTextItems(section.Items)
		}
	}

	if len(opportunity.TeamMembers) > 20 {
		return fmt.Errorf("%w: project teams allow up to 20 members", ErrNetworkValidation)
	}
	profileIDs := make([]int64, 0, len(opportunity.TeamMembers))
	seen := make(map[int64]struct{}, len(opportunity.TeamMembers))
	for i := range opportunity.TeamMembers {
		member := &opportunity.TeamMembers[i]
		member.Role = strings.TrimSpace(member.Role)
		member.Contribution = strings.TrimSpace(member.Contribution)
		if member.ProfileID <= 0 || member.Role == "" || len([]rune(member.Role)) > 100 || len([]rune(member.Contribution)) > 1000 {
			return fmt.Errorf("%w: each team member needs a profile_id and role", ErrNetworkValidation)
		}
		if _, exists := seen[member.ProfileID]; exists {
			return fmt.Errorf("%w: duplicate project team member", ErrNetworkValidation)
		}
		seen[member.ProfileID] = struct{}{}
		profileIDs = append(profileIDs, member.ProfileID)
		member.SortOrder = i
	}
	if len(profileIDs) > 0 {
		exists, err := s.repo.PublicProfilesExist(profileIDs)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("%w: project team members must reference published profiles", ErrNetworkValidation)
		}
	}
	return nil
}

func cleanNetworkTextItems(items []string) []string {
	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			cleaned = append(cleaned, item)
		}
	}
	return cleaned
}

func validPlanDate(value string) bool {
	_, err := parsePlanDate(value)
	return err == nil
}

func parsePlanDate(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date")
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

func (s *BuilderNetworkService) PublicProjects() ([]domain.NetworkProject, error) {
	return s.repo.ListPublicProjects()
}

func (s *BuilderNetworkService) PublicProjectBySlug(slug string) (domain.NetworkProject, error) {
	return s.repo.GetPublicProjectBySlug(strings.TrimSpace(slug))
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
