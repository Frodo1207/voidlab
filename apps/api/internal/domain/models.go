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
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	Summary   string `json:"summary"`
	City      string `json:"city"`
	Location  string `json:"location"`
	EventType string `json:"event_type"`
	EventTime string `json:"event_time"`
	Content   string `json:"content"`
	CoverURL  string `json:"cover_url"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updated_at"`
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

type MediaAsset struct {
	ID            int64  `json:"id"`
	FileName      string `json:"file_name"`
	ObjectURL     string `json:"object_url"`
	ContentType   string `json:"content_type"`
	FileSize      int64  `json:"-"`
	FileSizeLabel string `json:"file_size"`
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
