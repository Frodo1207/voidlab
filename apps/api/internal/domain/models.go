package domain

type User struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Role        string `json:"role"`
	DisplayName string `json:"display_name"`
	IsActive    bool   `json:"is_active"`
}

type ManagedUser struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Role        string `json:"role"`
	DisplayName string `json:"display_name"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type AgentToken struct {
	ID         int64    `json:"id"`
	Name       string   `json:"name"`
	Scopes     []string `json:"scopes"`
	IsActive   bool     `json:"is_active"`
	LastUsedAt string   `json:"last_used_at"`
	CreatedBy  int64    `json:"created_by"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

type Actor struct {
	Type         string   `json:"type"`
	ID           int64    `json:"id"`
	Username     string   `json:"username"`
	Role         string   `json:"role"`
	AgentTokenID *int64   `json:"agent_token_id,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
}

type AuthSession struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type Article struct {
	ID         int64    `json:"id"`
	Title      string   `json:"title"`
	Slug       string   `json:"slug"`
	Summary    string   `json:"summary"`
	Category   string   `json:"category"`
	Audience   string   `json:"audience"`
	Tags       []string `json:"tags"`
	CoverURL   string   `json:"cover_url"`
	Content    string   `json:"content"`
	SourceName string   `json:"source_name"`
	SourceURL  string   `json:"source_url"`
	Featured   bool     `json:"featured"`
	Status     string   `json:"status"`
	UpdatedAt  string   `json:"updated_at"`
}

type Event struct {
	ID                    int64  `json:"id"`
	Title                 string `json:"title"`
	Slug                  string `json:"slug"`
	Summary               string `json:"summary"`
	City                  string `json:"city"`
	Location              string `json:"location"`
	EventType             string `json:"event_type"`
	EventTime             string `json:"event_time"`
	Content               string `json:"content"`
	CoverURL              string `json:"cover_url"`
	Status                string `json:"status"`
	SignupMode            string `json:"signup_mode"`
	SignupStatus          string `json:"signup_status"`
	SignupEnabled         bool   `json:"signup_enabled"`
	SignupStartsAt        string `json:"signup_starts_at"`
	SignupDeadline        string `json:"signup_deadline"`
	Capacity              int64  `json:"capacity"`
	SignupCount           int64  `json:"signup_count"`
	AllowSignupDuringLive bool   `json:"allow_signup_during_live"`
	ExternalSignupURL     string `json:"external_signup_url"`
	SignupButtonLabel     string `json:"signup_button_label"`
	SignupSuccessMessage  string `json:"signup_success_message"`
	SignupClosedReason    string `json:"signup_closed_reason"`
	UpdatedAt             string `json:"updated_at"`
}

type Builder struct {
	ID                 int64    `json:"id"`
	Name               string   `json:"name"`
	Slug               string   `json:"slug"`
	Title              string   `json:"title"`
	City               string   `json:"city"`
	Role               string   `json:"role"`
	Intro              string   `json:"intro"`
	Story              string   `json:"story"`
	Expertise          []string `json:"expertise"`
	FocusAreas         []string `json:"focus_areas"`
	CollaborationModes []string `json:"collaboration_modes"`
	Contactable        bool     `json:"contactable"`
	Featured           bool     `json:"featured"`
	CoverURL           string   `json:"cover_url"`
	Status             string   `json:"status"`
	UpdatedAt          string   `json:"updated_at"`
}

type NetworkSession struct {
	Token   string         `json:"token"`
	Account NetworkAccount `json:"account"`
	Profile NetworkProfile `json:"profile"`
}

type NetworkAgentToken struct {
	ID         int64    `json:"id"`
	ProfileID  int64    `json:"profile_id"`
	Name       string   `json:"name"`
	Scopes     []string `json:"scopes"`
	IsActive   bool     `json:"is_active"`
	LastUsedAt string   `json:"last_used_at"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

type NetworkAgentPrincipal struct {
	Account NetworkAccount
	Token   NetworkAgentToken
}

type NetworkAccount struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type NetworkInviteCode struct {
	ID        int64  `json:"id"`
	MaxUses   int    `json:"max_uses"`
	UsedCount int    `json:"used_count"`
	ExpiresAt string `json:"expires_at"`
	Status    string `json:"status"`
	CreatedBy int64  `json:"created_by"`
	CreatedAt string `json:"created_at"`
}

type NetworkProfile struct {
	ID            int64    `json:"id"`
	AccountID     int64    `json:"account_id,omitempty"`
	Slug          string   `json:"slug"`
	DisplayName   string   `json:"display_name"`
	PrimaryRole   string   `json:"primary_role"`
	Headline      string   `json:"headline"`
	Bio           string   `json:"bio"`
	City          string   `json:"city"`
	Skills        []string `json:"skills"`
	Collaboration []string `json:"collaboration_modes"`
	Availability  string   `json:"availability"`
	AvailableFrom string   `json:"available_from"`
	Email         string   `json:"email,omitempty"`
	WeChat        string   `json:"wechat,omitempty"`
	Phone         string   `json:"phone,omitempty"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

type NetworkWork struct {
	ID         int64               `json:"id"`
	ProfileID  int64               `json:"profile_id"`
	Slug       string              `json:"slug"`
	Title      string              `json:"title"`
	Summary    string              `json:"summary"`
	Body       string              `json:"body"`
	CoverURL   string              `json:"cover_url"`
	DemoURL    string              `json:"demo_url"`
	RepoURL    string              `json:"repo_url"`
	Tags       []string            `json:"tags"`
	Status     string              `json:"status"`
	CreatedAt  string              `json:"created_at"`
	UpdatedAt  string              `json:"updated_at"`
	AuthorName string              `json:"author_name,omitempty"`
	Members    []NetworkWorkMember `json:"members,omitempty"`
}

type NetworkWorkMember struct {
	ProfileID    int64  `json:"profile_id"`
	DisplayName  string `json:"display_name,omitempty"`
	Role         string `json:"role"`
	Contribution string `json:"contribution"`
	SortOrder    int    `json:"sort_order"`
	IsOwner      bool   `json:"is_owner"`
}

type NetworkDiscoveryQuery struct {
	Query        string
	Role         string
	Skill        string
	City         string
	Availability string
	Type         string
	Remote       *bool
	Page         int
	PageSize     int
}

type NetworkPage[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type NetworkOpportunity struct {
	ID              int64                      `json:"id"`
	ProjectID       int64                      `json:"project_id"`
	ProfileID       int64                      `json:"profile_id"`
	Slug            string                     `json:"slug"`
	Title           string                     `json:"title"`
	Summary         string                     `json:"summary"`
	OpportunityType string                     `json:"opportunity_type"`
	LookingFor      []string                   `json:"looking_for_roles"`
	RequiredSkills  []string                   `json:"required_skills"`
	EngagementMode  string                     `json:"engagement_mode"`
	BudgetHint      string                     `json:"budget_hint"`
	DurationHint    string                     `json:"duration_hint"`
	CityScope       string                     `json:"city_scope"`
	RemoteAllowed   bool                       `json:"remote_allowed"`
	Status          string                     `json:"status"`
	ExpiresAt       string                     `json:"expires_at"`
	CreatedAt       string                     `json:"created_at"`
	UpdatedAt       string                     `json:"updated_at"`
	OwnerName       string                     `json:"owner_name,omitempty"`
	ProjectPlan     *NetworkProjectPlan        `json:"project_plan,omitempty"`
	TeamIntro       string                     `json:"team_intro,omitempty"`
	TeamMembers     []NetworkOpportunityMember `json:"team_members,omitempty"`
}

type NetworkProject struct {
	ID             int64                      `json:"id"`
	OwnerProfileID int64                      `json:"owner_profile_id"`
	Slug           string                     `json:"slug"`
	Title          string                     `json:"title"`
	Summary        string                     `json:"summary"`
	Status         string                     `json:"status"`
	ProjectPlan    *NetworkProjectPlan        `json:"project_plan,omitempty"`
	TeamIntro      string                     `json:"team_intro"`
	OwnerName      string                     `json:"owner_name,omitempty"`
	Members        []NetworkOpportunityMember `json:"members"`
	Opportunities  []NetworkOpportunity       `json:"opportunities,omitempty"`
	CreatedAt      string                     `json:"created_at"`
	UpdatedAt      string                     `json:"updated_at"`
}

type NetworkProjectPlan struct {
	Milestones []NetworkProjectMilestone   `json:"milestones"`
	Sections   []NetworkProjectPlanSection `json:"sections"`
}

type NetworkProjectMilestone struct {
	Title              string                `json:"title"`
	Description        string                `json:"description"`
	Order              int                   `json:"order"`
	Status             string                `json:"status"`
	Timing             *NetworkProjectTiming `json:"timing,omitempty"`
	Goals              []string              `json:"goals"`
	Deliverables       []string              `json:"deliverables"`
	AcceptanceCriteria []string              `json:"acceptance_criteria"`
	OwnerMemberIDs     []int64               `json:"owner_member_ids"`
}

type NetworkProjectTiming struct {
	Kind          string `json:"kind"`
	Label         string `json:"label,omitempty"`
	StartAt       string `json:"start_at,omitempty"`
	EndAt         string `json:"end_at,omitempty"`
	DurationValue int    `json:"duration_value,omitempty"`
	DurationUnit  string `json:"duration_unit,omitempty"`
}

type NetworkProjectPlanSection struct {
	Type      string   `json:"type"`
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Items     []string `json:"items"`
	SortOrder int      `json:"sort_order"`
}

type NetworkOpportunityMember struct {
	ProfileID    int64  `json:"profile_id"`
	DisplayName  string `json:"display_name,omitempty"`
	Slug         string `json:"slug,omitempty"`
	PrimaryRole  string `json:"primary_role,omitempty"`
	Headline     string `json:"headline,omitempty"`
	Role         string `json:"role"`
	Contribution string `json:"contribution"`
	SortOrder    int    `json:"sort_order"`
}

type NetworkContactRequest struct {
	ID              int64   `json:"id"`
	FromProfileID   int64   `json:"from_profile_id"`
	TargetType      string  `json:"target_type"`
	TargetID        int64   `json:"target_id"`
	Message         string  `json:"message"`
	RelatedWorkIDs  []int64 `json:"related_work_ids"`
	ShareEmail      bool    `json:"share_email"`
	ShareWeChat     bool    `json:"share_wechat"`
	SharePhone      bool    `json:"share_phone"`
	Status          string  `json:"status"`
	ProgressStatus  string  `json:"progress_status"`
	ExchangeRevoked bool    `json:"exchange_revoked"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	ApplicantName   string  `json:"applicant_name,omitempty"`
	ApplicantRole   string  `json:"applicant_role,omitempty"`
}

type NetworkContactExchange struct {
	RequestID  int64                 `json:"request_id"`
	Applicant  NetworkContactDetails `json:"applicant"`
	Recipient  NetworkContactDetails `json:"recipient"`
	RevealedAt string                `json:"revealed_at"`
}

type NetworkContactDetails struct {
	Email  string `json:"email,omitempty"`
	WeChat string `json:"wechat,omitempty"`
	Phone  string `json:"phone,omitempty"`
}

type MediaAsset struct {
	ID            int64  `json:"id"`
	FileName      string `json:"file_name"`
	ObjectURL     string `json:"object_url"`
	ContentType   string `json:"content_type"`
	FileSize      int64  `json:"-"`
	FileSizeLabel string `json:"file_size"`
	CreatedAt     string `json:"created_at"`
}

type MediaGallery struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
	ItemCount   int64  `json:"item_count"`
	CoverURL    string `json:"cover_url"`
	UpdatedAt   string `json:"updated_at"`
}

type MediaGalleryItem struct {
	ID            int64  `json:"id"`
	GalleryID     int64  `json:"gallery_id"`
	MediaAssetID  int64  `json:"media_asset_id"`
	FileName      string `json:"file_name"`
	ObjectURL     string `json:"object_url"`
	ContentType   string `json:"content_type"`
	FileSize      int64  `json:"-"`
	FileSizeLabel string `json:"file_size"`
	AltText       string `json:"alt_text"`
	SortOrder     int    `json:"sort_order"`
	IsActive      bool   `json:"is_active"`
	CreatedAt     string `json:"created_at"`
}

type Lead struct {
	ID         int64     `json:"id"`
	SourceType string    `json:"source_type"`
	SourceID   *int64    `json:"source_id,omitempty"`
	Name       string    `json:"name"`
	Contact    string    `json:"contact"`
	Message    string    `json:"message"`
	Status     string    `json:"status"`
	Notes      string    `json:"notes"`
	DedupeKey  string    `json:"dedupe_key"`
	OwnerID    *int64    `json:"owner_id,omitempty"`
	CreatedAt  string    `json:"created_at"`
	UpdatedAt  string    `json:"updated_at"`
	Logs       []LeadLog `json:"logs,omitempty"`
}

type LeadLog struct {
	ID        int64  `json:"id"`
	LeadID    int64  `json:"lead_id"`
	Action    string `json:"action"`
	Content   string `json:"content"`
	CreatedBy int64  `json:"created_by"`
	CreatedAt string `json:"created_at"`
}

type SiteConfig struct {
	ID              int64  `json:"id"`
	ConfigKey       string `json:"config_key"`
	ConfigValueJSON string `json:"config_value_json"`
	UpdatedBy       int64  `json:"updated_by"`
	UpdatedAt       string `json:"updated_at"`
}

type AuditLog struct {
	ID            int64  `json:"id"`
	ActorType     string `json:"actor_type"`
	ActorID       int64  `json:"actor_id"`
	ActorUsername string `json:"actor_username"`
	ActorRole     string `json:"actor_role"`
	AgentTokenID  *int64 `json:"agent_token_id,omitempty"`
	Action        string `json:"action"`
	EntityType    string `json:"entity_type"`
	EntityID      *int64 `json:"entity_id,omitempty"`
	EntityLabel   string `json:"entity_label"`
	DetailJSON    string `json:"detail_json"`
	CreatedAt     string `json:"created_at"`
}

type LeadStatusStats struct {
	New       int64 `json:"new"`
	Contacted int64 `json:"contacted"`
	Following int64 `json:"following"`
	Converted int64 `json:"converted"`
	Invalid   int64 `json:"invalid"`
}

type DashboardStats struct {
	ArticleCount           int64           `json:"article_count"`
	PublishedArticleCount  int64           `json:"published_article_count"`
	EventCount             int64           `json:"event_count"`
	PublishedEventCount    int64           `json:"published_event_count"`
	BuilderCount           int64           `json:"builder_count"`
	PublishedBuilderCount  int64           `json:"published_builder_count"`
	LeadCount              int64           `json:"lead_count"`
	ActionableLeadCount    int64           `json:"actionable_lead_count"`
	LeadStatusDistribution LeadStatusStats `json:"lead_status_distribution"`
	RecentActivities       []AuditLog      `json:"recent_activities"`
	RecentActionableLeads  []Lead          `json:"recent_actionable_leads"`
}

type KnowledgeSpace struct {
	ID               int64  `json:"id"`
	Title            string `json:"title"`
	Slug             string `json:"slug"`
	Description      string `json:"description"`
	CoverLabel       string `json:"cover_label"`
	Icon             string `json:"icon"`
	ThemeTint        string `json:"theme_tint"`
	VisibilityMode   string `json:"visibility_mode"`
	DirectorySummary string `json:"directory_summary"`
	IntroMarkdown    string `json:"intro_markdown"`
	TokenHint        string `json:"token_hint"`
	CoverURL         string `json:"cover_url"`
	Status           string `json:"status"`
	EntryCount       int64  `json:"entry_count"`
	SectionCount     int64  `json:"section_count"`
	UpdatedAt        string `json:"updated_at"`
}

type KnowledgeEntry struct {
	ID                   int64  `json:"id"`
	SpaceID              int64  `json:"space_id"`
	SpaceSlug            string `json:"space_slug,omitempty"`
	Title                string `json:"title"`
	Slug                 string `json:"slug"`
	SectionName          string `json:"section_name"`
	SortOrder            int    `json:"sort_order"`
	EstimatedReadMinutes int    `json:"estimated_read_minutes"`
	PublicSummary        string `json:"public_summary"`
	ContentMarkdown      string `json:"content_markdown"`
	CoverURL             string `json:"cover_url"`
	IsPreview            bool   `json:"is_preview"`
	Status               string `json:"status"`
	UpdatedAt            string `json:"updated_at"`
}

type KnowledgeAccessToken struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	AccessLevel string  `json:"access_level"`
	ScopeType   string  `json:"scope_type"`
	SpaceIDs    []int64 `json:"space_ids"`
	IsActive    bool    `json:"is_active"`
	ExpiresAt   string  `json:"expires_at"`
	CreatedBy   int64   `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type KnowledgeAsset struct {
	ID           int64  `json:"id"`
	SpaceID      int64  `json:"space_id"`
	MediaAssetID int64  `json:"media_asset_id"`
	FileName     string `json:"file_name"`
	ContentType  string `json:"content_type"`
	CreatedAt    string `json:"created_at"`
}
