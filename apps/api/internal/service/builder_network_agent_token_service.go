package service

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

var allowedNetworkAgentScopes = map[string]struct{}{
	"network_profiles:read":          {},
	"network_profiles:write":         {},
	"network_works:read":             {},
	"network_works:write":            {},
	"network_projects:read":          {},
	"network_projects:write":         {},
	"network_opportunities:read":     {},
	"network_opportunities:write":    {},
	"network_contact_requests:read":  {},
	"network_contact_requests:write": {},
	"network_contact_exchange:read":  {},
	"network_contact_exchange:write": {},
}

type NetworkAgentTokenCreateInput struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

type NetworkAgentTokenCreateResult struct {
	Token  string                   `json:"token"`
	Record domain.NetworkAgentToken `json:"record"`
}

type BuilderNetworkAgentTokenService struct {
	networkRepo *repository.BuilderNetworkRepository
	tokenRepo   *repository.BuilderNetworkAgentTokenRepository
}

func NewBuilderNetworkAgentTokenService(
	networkRepo *repository.BuilderNetworkRepository,
	tokenRepo *repository.BuilderNetworkAgentTokenRepository,
) *BuilderNetworkAgentTokenService {
	return &BuilderNetworkAgentTokenService{networkRepo: networkRepo, tokenRepo: tokenRepo}
}

func (s *BuilderNetworkAgentTokenService) List(accountID int64) ([]domain.NetworkAgentToken, error) {
	_, profile, err := s.networkRepo.GetAccountProfile(accountID)
	if err != nil {
		return nil, err
	}
	return s.tokenRepo.List(profile.ID)
}

func (s *BuilderNetworkAgentTokenService) Create(accountID int64, input NetworkAgentTokenCreateInput) (NetworkAgentTokenCreateResult, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" || len([]rune(name)) > 80 {
		return NetworkAgentTokenCreateResult{}, fmt.Errorf("%w: token name must be 1 to 80 characters", ErrNetworkValidation)
	}
	scopes, err := normalizeNetworkAgentScopes(input.Scopes)
	if err != nil {
		return NetworkAgentTokenCreateResult{}, err
	}
	_, profile, err := s.networkRepo.GetAccountProfile(accountID)
	if err != nil {
		return NetworkAgentTokenCreateResult{}, err
	}
	raw, err := randomToken(32)
	if err != nil {
		return NetworkAgentTokenCreateResult{}, err
	}
	raw = "vla_network_" + raw
	id, err := s.tokenRepo.Create(profile.ID, name, hashToken(raw), scopes)
	if err != nil {
		return NetworkAgentTokenCreateResult{}, err
	}
	record, err := s.tokenRepo.GetByID(profile.ID, id)
	if err != nil {
		return NetworkAgentTokenCreateResult{}, err
	}
	return NetworkAgentTokenCreateResult{Token: raw, Record: record}, nil
}

func (s *BuilderNetworkAgentTokenService) UpdateStatus(accountID, tokenID int64, isActive bool) (domain.NetworkAgentToken, error) {
	if tokenID <= 0 {
		return domain.NetworkAgentToken{}, fmt.Errorf("%w: invalid token id", ErrNetworkValidation)
	}
	_, profile, err := s.networkRepo.GetAccountProfile(accountID)
	if err != nil {
		return domain.NetworkAgentToken{}, err
	}
	if err := s.tokenRepo.UpdateStatus(profile.ID, tokenID, isActive); err != nil {
		return domain.NetworkAgentToken{}, err
	}
	return s.tokenRepo.GetByID(profile.ID, tokenID)
}

func (s *BuilderNetworkAgentTokenService) ResolveToken(token string) (domain.NetworkAgentPrincipal, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.NetworkAgentPrincipal{}, ErrNetworkUnauthorized
	}
	principal, err := s.tokenRepo.ResolveToken(hashToken(token))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NetworkAgentPrincipal{}, ErrNetworkUnauthorized
	}
	return principal, err
}

func NetworkAgentScopeOptions() []string {
	scopes := make([]string, 0, len(allowedNetworkAgentScopes))
	for scope := range allowedNetworkAgentScopes {
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)
	return scopes
}

func normalizeNetworkAgentScopes(input []string) ([]string, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("%w: at least one agent scope is required", ErrNetworkValidation)
	}
	seen := make(map[string]struct{}, len(input))
	scopes := make([]string, 0, len(input))
	for _, raw := range input {
		scope := strings.TrimSpace(raw)
		if scope == "" {
			continue
		}
		if _, ok := allowedNetworkAgentScopes[scope]; !ok {
			return nil, fmt.Errorf("%w: invalid agent scope", ErrNetworkValidation)
		}
		if _, exists := seen[scope]; exists {
			continue
		}
		seen[scope] = struct{}{}
		scopes = append(scopes, scope)
	}
	if len(scopes) == 0 {
		return nil, fmt.Errorf("%w: at least one agent scope is required", ErrNetworkValidation)
	}
	return scopes, nil
}
