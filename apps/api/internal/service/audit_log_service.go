package service

import (
	"encoding/json"
	"strings"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

type AuditLogInput struct {
	ActorType     string
	ActorID       int64
	ActorUsername string
	ActorRole     string
	AgentTokenID  *int64
	Action        string
	EntityType    string
	EntityID      *int64
	EntityLabel   string
	Detail        any
}

type AuditLogService struct {
	repo *repository.AuditLogRepository
}

func NewAuditLogService(repo *repository.AuditLogRepository) *AuditLogService {
	return &AuditLogService{repo: repo}
}

func (s *AuditLogService) List(limit int) ([]domain.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 200 {
		limit = 200
	}

	return s.repo.List(limit)
}

func (s *AuditLogService) Record(input AuditLogInput) error {
	if strings.TrimSpace(input.ActorType) == "" {
		return nil
	}

	action := strings.TrimSpace(input.Action)
	entityType := strings.TrimSpace(input.EntityType)
	if action == "" || entityType == "" {
		return nil
	}

	detailJSON := "{}"
	if input.Detail != nil {
		if encoded, err := json.Marshal(input.Detail); err == nil {
			detailJSON = string(encoded)
		}
	}

	_, err := s.repo.Create(repository.AuditLogCreateParams{
		ActorType:     strings.TrimSpace(input.ActorType),
		ActorID:       input.ActorID,
		ActorUsername: strings.TrimSpace(input.ActorUsername),
		ActorRole:     strings.TrimSpace(input.ActorRole),
		AgentTokenID:  input.AgentTokenID,
		Action:        action,
		EntityType:    entityType,
		EntityID:      input.EntityID,
		EntityLabel:   strings.TrimSpace(input.EntityLabel),
		DetailJSON:    detailJSON,
	})
	return err
}
