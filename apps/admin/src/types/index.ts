export type ContentStatus = "draft" | "published" | "archived";
export type UserRole = "admin" | "editor" | "ops";
export type ActorType = "human" | "agent";
export type LeadSourceType = "contact" | "event" | "builder";
export type LeadStatus = "new" | "contacted" | "following" | "converted" | "invalid";
export type SiteConfigKey =
  | "home_banner"
  | "home_featured"
  | "contact_channels"
  | "footer_config"
  | "global_cta"
  | "featured_content_slots";

export interface UserProfile {
  id: number;
  username: string;
  role: UserRole;
  displayName: string;
  isActive?: boolean;
}

export interface AuthSession {
  token: string;
  user: UserProfile;
}

export interface ArticleRecord {
  id: number;
  title: string;
  slug: string;
  summary: string;
  category: string;
  audience: string;
  tags: string[];
  coverUrl: string;
  content: string;
  sourceName: string;
  sourceUrl: string;
  featured: boolean;
  status: ContentStatus;
  updatedAt: string;
}

export interface EventRecord {
  id: number;
  title: string;
  slug: string;
  summary: string;
  city: string;
  location: string;
  eventType: string;
  eventTime: string;
  content: string;
  coverUrl: string;
  status: ContentStatus;
  updatedAt: string;
}

export interface BuilderRecord {
  id: number;
  name: string;
  slug: string;
  title: string;
  city: string;
  role: string;
  intro: string;
  story: string;
  expertise: string[];
  focusAreas: string[];
  collaborationModes: string[];
  contactable: boolean;
  featured: boolean;
  coverUrl: string;
  status: ContentStatus;
  updatedAt: string;
}

export interface MediaAssetRecord {
  id: number;
  fileName: string;
  objectUrl: string;
  contentType: string;
  fileSizeLabel: string;
  createdAt: string;
}

export interface LeadLogRecord {
  id: number;
  leadId: number;
  action: string;
  content: string;
  createdBy: number;
  createdAt: string;
}

export interface LeadRecord {
  id: number;
  sourceType: LeadSourceType;
  sourceId?: number;
  name: string;
  contact: string;
  message: string;
  status: LeadStatus;
  notes: string;
  ownerId?: number;
  createdAt: string;
  updatedAt: string;
  logs: LeadLogRecord[];
}

export interface HomeBannerConfig {
  titleText: string;
  subtitle: string;
  primaryCtaLabel: string;
  primaryCtaPath: string;
  secondaryCtaLabel: string;
  secondaryCtaPath: string;
  statusLabel: string;
}

export interface HomeFeaturedConfig {
  communityCount: string;
  communityCountSuffix: string;
  eventCount: string;
  eventCountSuffix: string;
  eventsDescription: string;
  buildersDescription: string;
  insightsDescription: string;
}

export interface ContactChannelConfig {
  title: string;
  desc: string;
  account: string;
  buttonText: string;
  link: string;
}

export interface FooterNavItemConfig {
  label: string;
  path: string;
}

export interface FooterConfig {
  slogan: string;
  navLinks: FooterNavItemConfig[];
  legalText: string;
}

export interface GlobalCtaConfig {
  eyebrow: string;
  title: string;
  description: string;
  primaryLabel: string;
  primaryPath: string;
  secondaryLabel: string;
  secondaryPath: string;
}

export interface FeaturedContentSlotsConfig {
  eventsTitle: string;
  eventsViewAllLabel: string;
  eventsLimit: number;
  buildersTitle: string;
  buildersViewAllLabel: string;
  buildersLimit: number;
  insightsTitle: string;
  insightsViewAllLabel: string;
  insightsLimit: number;
}

export interface SiteConfigRecord<T = unknown> {
  id: number;
  configKey: SiteConfigKey;
  configValue: T;
  updatedBy: number;
  updatedAt: string;
}

export interface AuditLogRecord {
  id: number;
  actorType: ActorType;
  actorId: number;
  actorUsername: string;
  actorRole: string;
  agentTokenId?: number;
  action: string;
  entityType: string;
  entityId?: number;
  entityLabel: string;
  detail: unknown;
  createdAt: string;
}

export interface LeadStatusStatsRecord {
  new: number;
  contacted: number;
  following: number;
  converted: number;
  invalid: number;
}

export interface DashboardStatsRecord {
  articleCount: number;
  publishedArticleCount: number;
  eventCount: number;
  publishedEventCount: number;
  builderCount: number;
  publishedBuilderCount: number;
  leadCount: number;
  actionableLeadCount: number;
  leadStatusDistribution: LeadStatusStatsRecord;
  recentActivities: AuditLogRecord[];
  recentActionableLeads: LeadRecord[];
}

export type KnowledgeVisibilityMode = "public" | "directory_only" | "private_hidden";

export interface KnowledgeSpaceRecord {
  id: number;
  title: string;
  slug: string;
  description: string;
  coverLabel: string;
  icon: string;
  themeTint: string;
  visibilityMode: KnowledgeVisibilityMode;
  directorySummary: string;
  introMarkdown: string;
  tokenHint: string;
  coverUrl: string;
  status: ContentStatus;
  entryCount: number;
  sectionCount: number;
  updatedAt: string;
}

export interface KnowledgeEntryRecord {
  id: number;
  spaceId: number;
  spaceSlug: string;
  title: string;
  slug: string;
  sectionName: string;
  sortOrder: number;
  estimatedReadMinutes: number;
  publicSummary: string;
  contentMarkdown: string;
  coverUrl: string;
  isPreview: boolean;
  status: ContentStatus;
  updatedAt: string;
}

export interface KnowledgeMarkdownImportResult {
  title: string;
  slug: string;
  sectionName: string;
  estimatedReadMinutes: number;
  publicSummary: string;
  contentMarkdown: string;
  coverUrl: string;
  isPreview: boolean;
  status: ContentStatus;
}

export interface KnowledgeAssetRecord {
  id: number;
  spaceId: number;
  mediaAssetId: number;
  fileName: string;
  contentType: string;
  createdAt: string;
}

export interface KnowledgeAssetUploadResult {
  asset: KnowledgeAssetRecord;
  markdownUrl: string;
  markdownSnippet: string;
  publicUrl: string;
}

export interface KnowledgeAccessTokenRecord {
  id: number;
  name: string;
  accessLevel: "basic" | "pro" | "vip";
  scopeType: "single_space" | "multi_space" | "all_published";
  spaceIds: number[];
  isActive: boolean;
  expiresAt: string;
  createdBy: number;
  createdAt: string;
  updatedAt: string;
}

export interface ManagedUserRecord {
  id: number;
  username: string;
  role: UserRole;
  displayName: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export type AgentScope =
  | "articles:read"
  | "articles:write"
  | "events:read"
  | "events:write"
  | "builders:read"
  | "builders:write"
  | "knowledge:read"
  | "knowledge:write"
  | "knowledge_tokens:read"
  | "knowledge_tokens:write"
  | "media:read"
  | "media:write";

export interface AgentTokenRecord {
  id: number;
  name: string;
  scopes: AgentScope[];
  isActive: boolean;
  lastUsedAt: string;
  createdBy: number;
  createdAt: string;
  updatedAt: string;
}
