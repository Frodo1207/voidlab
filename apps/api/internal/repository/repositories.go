package repository

import "database/sql"

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

type ArticleRepository struct {
	db *sql.DB
}

func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

type BuilderRepository struct {
	db *sql.DB
}

func NewBuilderRepository(db *sql.DB) *BuilderRepository {
	return &BuilderRepository{db: db}
}

type MediaRepository struct {
	db *sql.DB
}

func NewMediaRepository(db *sql.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

type LeadRepository struct {
	db *sql.DB
}

func NewLeadRepository(db *sql.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

type AuditLogRepository struct {
	db *sql.DB
}

func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

type AgentTokenRepository struct {
	db *sql.DB
}

func NewAgentTokenRepository(db *sql.DB) *AgentTokenRepository {
	return &AgentTokenRepository{db: db}
}

type KnowledgeSpaceRepository struct {
	db *sql.DB
}

func NewKnowledgeSpaceRepository(db *sql.DB) *KnowledgeSpaceRepository {
	return &KnowledgeSpaceRepository{db: db}
}

type KnowledgeEntryRepository struct {
	db *sql.DB
}

func NewKnowledgeEntryRepository(db *sql.DB) *KnowledgeEntryRepository {
	return &KnowledgeEntryRepository{db: db}
}

type KnowledgeAccessTokenRepository struct {
	db *sql.DB
}

func NewKnowledgeAccessTokenRepository(db *sql.DB) *KnowledgeAccessTokenRepository {
	return &KnowledgeAccessTokenRepository{db: db}
}

type KnowledgeAccessLogRepository struct {
	db *sql.DB
}

func NewKnowledgeAccessLogRepository(db *sql.DB) *KnowledgeAccessLogRepository {
	return &KnowledgeAccessLogRepository{db: db}
}

type KnowledgeAssetRepository struct {
	db *sql.DB
}

func NewKnowledgeAssetRepository(db *sql.DB) *KnowledgeAssetRepository {
	return &KnowledgeAssetRepository{db: db}
}
