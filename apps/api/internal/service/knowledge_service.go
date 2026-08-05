package service

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

type KnowledgeSpaceInput struct {
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
}

type KnowledgeEntryInput struct {
	SpaceID              int64  `json:"space_id"`
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
}

type KnowledgeAccessTokenCreateInput struct {
	SpaceID     int64   `json:"space_id"`
	SpaceIDs    []int64 `json:"space_ids"`
	Name        string  `json:"name"`
	AccessLevel string  `json:"access_level"`
	ExpiresAt   string  `json:"expires_at"`
}

type KnowledgeAccessTokenStatusInput struct {
	IsActive bool `json:"is_active"`
}

type KnowledgeService struct {
	spaceRepo     *repository.KnowledgeSpaceRepository
	entryRepo     *repository.KnowledgeEntryRepository
	tokenRepo     *repository.KnowledgeAccessTokenRepository
	accessLogRepo *repository.KnowledgeAccessLogRepository
}

func NewKnowledgeService(
	spaceRepo *repository.KnowledgeSpaceRepository,
	entryRepo *repository.KnowledgeEntryRepository,
	tokenRepo *repository.KnowledgeAccessTokenRepository,
	accessLogRepo *repository.KnowledgeAccessLogRepository,
) *KnowledgeService {
	return &KnowledgeService{
		spaceRepo:     spaceRepo,
		entryRepo:     entryRepo,
		tokenRepo:     tokenRepo,
		accessLogRepo: accessLogRepo,
	}
}

func (s *KnowledgeService) ListSpaces() ([]domain.KnowledgeSpace, error) {
	return s.spaceRepo.List()
}

func (s *KnowledgeService) ListPublishedSpaces() ([]domain.KnowledgeSpace, error) {
	return s.spaceRepo.ListPublished()
}

func (s *KnowledgeService) GetSpaceByID(id int64) (domain.KnowledgeSpace, error) {
	return s.spaceRepo.GetByID(id)
}

func (s *KnowledgeService) GetPublishedSpaceBySlug(slug string) (domain.KnowledgeSpace, error) {
	return s.spaceRepo.GetPublishedBySlug(strings.TrimSpace(slug))
}

func (s *KnowledgeService) CreateSpace(input KnowledgeSpaceInput, userID int64) (int64, error) {
	params, err := validateKnowledgeSpaceInput(input, userID)
	if err != nil {
		return 0, err
	}
	return s.spaceRepo.Create(params)
}

func (s *KnowledgeService) UpdateSpace(id int64, input KnowledgeSpaceInput, userID int64) error {
	current, err := s.spaceRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("knowledge space not found")
		}
		return err
	}

	params, err := validateKnowledgeSpaceUpdateInput(current, input, userID)
	if err != nil {
		return err
	}
	return s.spaceRepo.Update(id, params)
}

func (s *KnowledgeService) DeleteSpace(id int64) error {
	if _, err := s.spaceRepo.GetByID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("knowledge space not found")
		}
		return err
	}
	return s.spaceRepo.Delete(id)
}

func (s *KnowledgeService) ListEntries(spaceID *int64) ([]domain.KnowledgeEntry, error) {
	if spaceID == nil {
		return s.entryRepo.List()
	}
	return s.entryRepo.ListBySpace(*spaceID)
}

func (s *KnowledgeService) ListPublishedEntriesBySpace(spaceID int64) ([]domain.KnowledgeEntry, error) {
	return s.entryRepo.ListPublishedBySpace(spaceID)
}

func (s *KnowledgeService) GetEntryByID(id int64) (domain.KnowledgeEntry, error) {
	return s.entryRepo.GetByID(id)
}

func (s *KnowledgeService) GetPublishedEntryBySpaceAndSlug(spaceID int64, slug string) (domain.KnowledgeEntry, error) {
	return s.entryRepo.GetPublishedBySpaceAndSlug(spaceID, strings.TrimSpace(slug))
}

func (s *KnowledgeService) CreateEntry(input KnowledgeEntryInput, userID int64) (int64, error) {
	params, err := s.validateKnowledgeEntryInput(input, userID)
	if err != nil {
		return 0, err
	}
	return s.entryRepo.Create(params)
}

func (s *KnowledgeService) UpdateEntry(id int64, input KnowledgeEntryInput, userID int64) error {
	current, err := s.entryRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("knowledge entry not found")
		}
		return err
	}

	params, err := s.validateKnowledgeEntryUpdateInput(current, input, userID)
	if err != nil {
		return err
	}
	return s.entryRepo.Update(id, params)
}

func (s *KnowledgeService) DeleteEntry(id int64) error {
	if _, err := s.entryRepo.GetByID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("knowledge entry not found")
		}
		return err
	}
	return s.entryRepo.Delete(id)
}

func (s *KnowledgeService) ListAccessTokens(spaceID *int64) ([]domain.KnowledgeAccessToken, error) {
	return s.tokenRepo.List(spaceID)
}

func (s *KnowledgeService) CreateAccessToken(input KnowledgeAccessTokenCreateInput, createdBy int64) (domain.KnowledgeAccessToken, string, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domain.KnowledgeAccessToken{}, "", errors.New("token name is required")
	}

	accessLevel, scopeType, normalizedSpaceIDs, err := s.normalizeKnowledgeAccessScope(input)
	if err != nil {
		return domain.KnowledgeAccessToken{}, "", err
	}

	expiresAt, err := normalizeKnowledgeExpiresAt(input.ExpiresAt)
	if err != nil {
		return domain.KnowledgeAccessToken{}, "", err
	}

	plainToken, tokenHash, err := generateKnowledgeAccessToken()
	if err != nil {
		return domain.KnowledgeAccessToken{}, "", err
	}

	id, err := s.tokenRepo.Create(repository.KnowledgeAccessTokenCreateParams{
		Name:        name,
		AccessLevel: accessLevel,
		ScopeType:   scopeType,
		SpaceIDs:    normalizedSpaceIDs,
		TokenHash:   tokenHash,
		ExpiresAt:   expiresAt,
		CreatedBy:   createdBy,
	})
	if err != nil {
		return domain.KnowledgeAccessToken{}, "", err
	}

	record, err := s.tokenRepo.GetByID(id)
	if err != nil {
		return domain.KnowledgeAccessToken{}, "", err
	}

	return record, plainToken, nil
}

func (s *KnowledgeService) UpdateAccessTokenStatus(id int64, input KnowledgeAccessTokenStatusInput) error {
	if _, err := s.tokenRepo.GetByID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("knowledge access token not found")
		}
		return err
	}
	return s.tokenRepo.UpdateStatus(id, input.IsActive)
}

func (s *KnowledgeService) VerifySpaceToken(space domain.KnowledgeSpace, token string, requestIP string, userAgent string) (domain.KnowledgeAccessToken, string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.KnowledgeAccessToken{}, "", errors.New("token is required")
	}

	record, err := s.tokenRepo.GetByHash(hashKnowledgeAccessToken(token))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.KnowledgeAccessToken{}, "", errors.New("invalid knowledge access token")
		}
		return domain.KnowledgeAccessToken{}, "", err
	}

	if !record.IsActive || knowledgeAccessTokenExpired(record.ExpiresAt) || !knowledgeAccessAppliesToSpace(record.KnowledgeAccessToken, space.ID) {
		return domain.KnowledgeAccessToken{}, "", errors.New("invalid knowledge access token")
	}

	grant := buildKnowledgeGrant(record.TokenHash)
	tokenID := record.ID
	_ = s.accessLogRepo.Create(repository.KnowledgeAccessLogCreateParams{
		SpaceID:   space.ID,
		TokenID:   &tokenID,
		Action:    "verify",
		RequestIP: requestIP,
		UserAgent: userAgent,
	})

	return record.KnowledgeAccessToken, grant, nil
}

func (s *KnowledgeService) CanReadSpaceContent(space domain.KnowledgeSpace, grant string) (bool, error) {
	if space.VisibilityMode == "public" {
		return true, nil
	}

	tokenRecord, err := s.resolveGrant(space, grant)
	if err != nil {
		return false, err
	}
	return tokenRecord != nil, nil
}

func (s *KnowledgeService) CanReadEntry(space domain.KnowledgeSpace, entry domain.KnowledgeEntry, grant string, requestIP string, userAgent string) (bool, error) {
	if space.VisibilityMode == "public" || entry.IsPreview {
		_ = s.accessLogRepo.Create(repository.KnowledgeAccessLogCreateParams{
			SpaceID:   space.ID,
			EntryID:   toOptionalInt64(entry.ID),
			Action:    "view_entry",
			RequestIP: requestIP,
			UserAgent: userAgent,
		})
		return true, nil
	}

	tokenRecord, err := s.resolveGrant(space, grant)
	if err != nil {
		return false, err
	}
	if tokenRecord == nil {
		return false, nil
	}

	tokenID := tokenRecord.ID
	_ = s.accessLogRepo.Create(repository.KnowledgeAccessLogCreateParams{
		SpaceID:   space.ID,
		EntryID:   toOptionalInt64(entry.ID),
		TokenID:   &tokenID,
		Action:    "view_entry",
		RequestIP: requestIP,
		UserAgent: userAgent,
	})
	return true, nil
}

func (s *KnowledgeService) validateKnowledgeEntryInput(input KnowledgeEntryInput, userID int64) (repository.KnowledgeEntryUpsertParams, error) {
	params, err := s.buildKnowledgeEntryParams(input, userID)
	if err != nil {
		return repository.KnowledgeEntryUpsertParams{}, err
	}

	status, err := normalizeCreateContentStatus(input.Status)
	if err != nil {
		return repository.KnowledgeEntryUpsertParams{}, err
	}
	params.Status = status
	return params, nil
}

func (s *KnowledgeService) buildKnowledgeEntryParams(input KnowledgeEntryInput, userID int64) (repository.KnowledgeEntryUpsertParams, error) {
	if input.SpaceID <= 0 {
		return repository.KnowledgeEntryUpsertParams{}, errors.New("space_id is required")
	}
	if _, err := s.spaceRepo.GetByID(input.SpaceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.KnowledgeEntryUpsertParams{}, errors.New("knowledge space not found")
		}
		return repository.KnowledgeEntryUpsertParams{}, err
	}

	title := strings.TrimSpace(input.Title)
	slug := strings.TrimSpace(input.Slug)
	if title == "" {
		return repository.KnowledgeEntryUpsertParams{}, errors.New("title is required")
	}
	if slug == "" {
		return repository.KnowledgeEntryUpsertParams{}, errors.New("slug is required")
	}

	return repository.KnowledgeEntryUpsertParams{
		SpaceID:              input.SpaceID,
		Title:                title,
		Slug:                 slug,
		SectionName:          strings.TrimSpace(input.SectionName),
		SortOrder:            input.SortOrder,
		EstimatedReadMinutes: input.EstimatedReadMinutes,
		PublicSummary:        strings.TrimSpace(input.PublicSummary),
		ContentMarkdown:      strings.TrimSpace(input.ContentMarkdown),
		CoverURL:             strings.TrimSpace(input.CoverURL),
		IsPreview:            input.IsPreview,
		UserID:               userID,
	}, nil
}

func (s *KnowledgeService) validateKnowledgeEntryUpdateInput(current domain.KnowledgeEntry, input KnowledgeEntryInput, userID int64) (repository.KnowledgeEntryUpsertParams, error) {
	params, err := s.buildKnowledgeEntryParams(input, userID)
	if err != nil {
		return repository.KnowledgeEntryUpsertParams{}, err
	}

	status, err := normalizeUpdateContentStatus(current.Status, input.Status)
	if err != nil {
		return repository.KnowledgeEntryUpsertParams{}, err
	}
	params.Status = status
	return params, nil
}

func validateKnowledgeSpaceInput(input KnowledgeSpaceInput, userID int64) (repository.KnowledgeSpaceUpsertParams, error) {
	params, err := buildKnowledgeSpaceParams(input, userID)
	if err != nil {
		return repository.KnowledgeSpaceUpsertParams{}, err
	}

	status, err := normalizeCreateContentStatus(input.Status)
	if err != nil {
		return repository.KnowledgeSpaceUpsertParams{}, err
	}
	params.Status = status
	return params, nil
}

func buildKnowledgeSpaceParams(input KnowledgeSpaceInput, userID int64) (repository.KnowledgeSpaceUpsertParams, error) {
	title := strings.TrimSpace(input.Title)
	slug := strings.TrimSpace(input.Slug)
	if title == "" {
		return repository.KnowledgeSpaceUpsertParams{}, errors.New("title is required")
	}
	if slug == "" {
		return repository.KnowledgeSpaceUpsertParams{}, errors.New("slug is required")
	}

	visibilityMode, err := normalizeKnowledgeVisibilityMode(input.VisibilityMode)
	if err != nil {
		return repository.KnowledgeSpaceUpsertParams{}, err
	}

	return repository.KnowledgeSpaceUpsertParams{
		Title:            title,
		Slug:             slug,
		Description:      strings.TrimSpace(input.Description),
		CoverLabel:       strings.TrimSpace(input.CoverLabel),
		Icon:             strings.TrimSpace(input.Icon),
		ThemeTint:        strings.TrimSpace(input.ThemeTint),
		VisibilityMode:   visibilityMode,
		DirectorySummary: strings.TrimSpace(input.DirectorySummary),
		IntroMarkdown:    strings.TrimSpace(input.IntroMarkdown),
		TokenHint:        strings.TrimSpace(input.TokenHint),
		CoverURL:         strings.TrimSpace(input.CoverURL),
		UserID:           userID,
	}, nil
}

func validateKnowledgeSpaceUpdateInput(current domain.KnowledgeSpace, input KnowledgeSpaceInput, userID int64) (repository.KnowledgeSpaceUpsertParams, error) {
	params, err := buildKnowledgeSpaceParams(input, userID)
	if err != nil {
		return repository.KnowledgeSpaceUpsertParams{}, err
	}

	status, err := normalizeUpdateContentStatus(current.Status, input.Status)
	if err != nil {
		return repository.KnowledgeSpaceUpsertParams{}, err
	}
	params.Status = status
	return params, nil
}

func normalizeKnowledgeVisibilityMode(mode string) (string, error) {
	mode = strings.TrimSpace(mode)
	if mode == "" {
		return "directory_only", nil
	}
	switch mode {
	case "public", "directory_only", "private_hidden":
		return mode, nil
	default:
		return "", errors.New("invalid visibility mode")
	}
}

func normalizeKnowledgeExpiresAt(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	if _, err := parseKnowledgeTime(value); err != nil {
		return "", errors.New("invalid expires_at")
	}
	return value, nil
}

func generateKnowledgeAccessToken() (string, string, error) {
	randomBytes := make([]byte, 24)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}

	plainToken := "vla_pass_" + hex.EncodeToString(randomBytes)
	return plainToken, hashKnowledgeAccessToken(plainToken), nil
}

func hashKnowledgeAccessToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func buildKnowledgeGrant(tokenHash string) string {
	hash := sha256.Sum256([]byte("knowledge-pass:" + strings.TrimSpace(tokenHash)))
	return "vla_kg_" + hex.EncodeToString(hash[:])
}

func (s *KnowledgeService) resolveGrant(space domain.KnowledgeSpace, grant string) (*repository.KnowledgeAccessTokenRecord, error) {
	grant = strings.TrimSpace(grant)
	if grant == "" {
		return nil, nil
	}

	records, err := s.tokenRepo.ListActiveBySpace(space.ID)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if knowledgeAccessTokenExpired(record.ExpiresAt) {
			continue
		}
		if buildKnowledgeGrant(record.TokenHash) == grant {
			copied := record
			return &copied, nil
		}
	}

	return nil, nil
}

func knowledgeAccessTokenExpired(expiresAt string) bool {
	if strings.TrimSpace(expiresAt) == "" {
		return false
	}

	parsed, err := parseKnowledgeTime(expiresAt)
	if err != nil {
		return false
	}
	return time.Now().After(parsed)
}

func parseKnowledgeTime(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

func toOptionalInt64(value int64) *int64 {
	return &value
}

func (s *KnowledgeService) normalizeKnowledgeAccessScope(input KnowledgeAccessTokenCreateInput) (string, string, []int64, error) {
	accessLevel := strings.ToLower(strings.TrimSpace(input.AccessLevel))
	if accessLevel == "" {
		accessLevel = "basic"
	}

	spaceIDs := make([]int64, 0, len(input.SpaceIDs)+1)
	if input.SpaceID > 0 {
		spaceIDs = append(spaceIDs, input.SpaceID)
	}
	spaceIDs = append(spaceIDs, input.SpaceIDs...)
	spaceIDs = uniquePositiveInt64s(spaceIDs)

	switch accessLevel {
	case "basic":
		if len(spaceIDs) != 1 {
			return "", "", nil, errors.New("basic token requires exactly one space")
		}
		if err := s.assertKnowledgeSpacesExist(spaceIDs); err != nil {
			return "", "", nil, err
		}
		return accessLevel, "single_space", spaceIDs, nil
	case "pro":
		if len(spaceIDs) == 0 {
			return "", "", nil, errors.New("pro token requires at least one space")
		}
		if err := s.assertKnowledgeSpacesExist(spaceIDs); err != nil {
			return "", "", nil, err
		}
		return accessLevel, "multi_space", spaceIDs, nil
	case "vip":
		return accessLevel, "all_published", nil, nil
	default:
		return "", "", nil, errors.New("invalid access_level")
	}
}

func (s *KnowledgeService) assertKnowledgeSpacesExist(spaceIDs []int64) error {
	for _, spaceID := range spaceIDs {
		if _, err := s.spaceRepo.GetByID(spaceID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("knowledge space not found")
			}
			return err
		}
	}
	return nil
}

func uniquePositiveInt64s(values []int64) []int64 {
	result := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func knowledgeAccessAppliesToSpace(record domain.KnowledgeAccessToken, spaceID int64) bool {
	switch record.ScopeType {
	case "all_published":
		return true
	case "single_space", "multi_space":
		for _, candidate := range record.SpaceIDs {
			if candidate == spaceID {
				return true
			}
		}
		return false
	default:
		return false
	}
}
