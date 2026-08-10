package service

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

var allowedAgentScopes = map[string]struct{}{
	"articles:read":          {},
	"articles:write":         {},
	"events:read":            {},
	"events:write":           {},
	"leads:read":             {},
	"leads:write":            {},
	"builders:read":          {},
	"builders:write":         {},
	"knowledge:read":         {},
	"knowledge:write":        {},
	"knowledge_tokens:read":  {},
	"knowledge_tokens:write": {},
	"media:read":             {},
	"media:write":            {},
}

type AgentTokenCreateInput struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

type AgentTokenStatusInput struct {
	IsActive bool `json:"is_active"`
}

type AgentTokenService struct {
	repo *repository.AgentTokenRepository
}

func NewAgentTokenService(repo *repository.AgentTokenRepository) *AgentTokenService {
	return &AgentTokenService{repo: repo}
}

func (s *AgentTokenService) List() ([]domain.AgentToken, error) {
	return s.repo.List()
}

func (s *AgentTokenService) GetByID(id int64) (domain.AgentToken, error) {
	return s.repo.GetByID(id)
}

func (s *AgentTokenService) Create(input AgentTokenCreateInput, createdBy int64) (domain.AgentToken, string, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domain.AgentToken{}, "", errors.New("agent token name is required")
	}

	scopes, err := normalizeAgentScopes(input.Scopes)
	if err != nil {
		return domain.AgentToken{}, "", err
	}

	plainToken, tokenHash, err := generateAgentToken()
	if err != nil {
		return domain.AgentToken{}, "", err
	}

	id, err := s.repo.Create(repository.AgentTokenCreateParams{
		Name:      name,
		TokenHash: tokenHash,
		Scopes:    scopes,
		CreatedBy: createdBy,
	})
	if err != nil {
		return domain.AgentToken{}, "", err
	}

	record, err := s.repo.GetByID(id)
	if err != nil {
		return domain.AgentToken{}, "", err
	}

	return record, plainToken, nil
}

func (s *AgentTokenService) UpdateStatus(id int64, input AgentTokenStatusInput) error {
	if _, err := s.repo.GetByID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("agent token not found")
		}
		return err
	}

	return s.repo.UpdateStatus(id, input.IsActive)
}

func (s *AgentTokenService) ResolveToken(token string) (domain.AgentToken, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.AgentToken{}, ErrInvalidToken
	}

	record, err := s.repo.GetByHash(hashAgentToken(token))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.AgentToken{}, ErrInvalidToken
		}
		return domain.AgentToken{}, err
	}

	if !record.IsActive {
		return domain.AgentToken{}, ErrInvalidToken
	}

	_ = s.repo.TouchLastUsed(record.ID)
	return record.AgentToken, nil
}

func normalizeAgentScopes(input []string) ([]string, error) {
	if len(input) == 0 {
		return nil, errors.New("at least one scope is required")
	}

	seen := make(map[string]struct{}, len(input))
	scopes := make([]string, 0, len(input))
	for _, raw := range input {
		scope := strings.TrimSpace(raw)
		if scope == "" {
			continue
		}

		if _, ok := allowedAgentScopes[scope]; !ok {
			return nil, errors.New("invalid agent scope")
		}

		if _, exists := seen[scope]; exists {
			continue
		}

		seen[scope] = struct{}{}
		scopes = append(scopes, scope)
	}

	if len(scopes) == 0 {
		return nil, errors.New("at least one scope is required")
	}

	return scopes, nil
}

func generateAgentToken() (string, string, error) {
	randomBytes := make([]byte, 24)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}

	plainToken := "vla_agent_" + hex.EncodeToString(randomBytes)
	return plainToken, hashAgentToken(plainToken), nil
}

func hashAgentToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
