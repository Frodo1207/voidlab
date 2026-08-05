package service

import (
	"database/sql"
	"errors"
	"strings"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

type BuilderService struct {
	repo *repository.BuilderRepository
}

type BuilderInput struct {
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
	CoverURL           string   `json:"cover_url"`
	Contactable        bool     `json:"contactable"`
	Featured           bool     `json:"featured"`
	Status             string   `json:"status"`
}

func NewBuilderService(repo *repository.BuilderRepository) *BuilderService {
	return &BuilderService{repo: repo}
}

func (s *BuilderService) List() ([]domain.Builder, error) {
	return s.repo.List()
}

func (s *BuilderService) ListPublished() ([]domain.Builder, error) {
	return s.repo.ListPublished()
}

func (s *BuilderService) GetByID(id int64) (domain.Builder, error) {
	return s.repo.GetByID(id)
}

func (s *BuilderService) GetPublishedBySlug(slug string) (domain.Builder, error) {
	return s.repo.GetPublishedBySlug(strings.TrimSpace(slug))
}

func (s *BuilderService) Create(input BuilderInput, userID int64) (int64, error) {
	params, err := validateBuilderCreateInput(input, userID)
	if err != nil {
		return 0, err
	}

	return s.repo.Create(params)
}

func (s *BuilderService) Update(id int64, input BuilderInput, userID int64) error {
	current, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("builder not found")
		}
		return err
	}

	params, err := validateBuilderUpdateInput(current, input, userID)
	if err != nil {
		return err
	}

	return s.repo.Update(id, params)
}

func (s *BuilderService) UpdateStatus(id int64, input ContentStatusInput, userID int64) error {
	current, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("builder not found")
		}
		return err
	}

	status, err := normalizeUpdateContentStatus(current.Status, input.Status)
	if err != nil {
		return err
	}

	return s.repo.UpdateStatus(id, status, userID)
}

func (s *BuilderService) Delete(id int64) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("builder not found")
		}
		return err
	}

	return s.repo.Delete(id)
}

func validateBuilderCreateInput(input BuilderInput, userID int64) (repository.BuilderUpsertParams, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.TrimSpace(input.Slug)
	status, err := normalizeCreateContentStatus(input.Status)

	if name == "" {
		return repository.BuilderUpsertParams{}, errors.New("name is required")
	}

	if slug == "" {
		return repository.BuilderUpsertParams{}, errors.New("slug is required")
	}

	if err != nil {
		return repository.BuilderUpsertParams{}, err
	}

	return repository.BuilderUpsertParams{
		Name:               name,
		Slug:               slug,
		Title:              strings.TrimSpace(input.Title),
		City:               strings.TrimSpace(input.City),
		Role:               strings.TrimSpace(input.Role),
		Intro:              strings.TrimSpace(input.Intro),
		Story:              strings.TrimSpace(input.Story),
		Expertise:          input.Expertise,
		FocusAreas:         input.FocusAreas,
		CollaborationModes: input.CollaborationModes,
		CoverURL:           strings.TrimSpace(input.CoverURL),
		Contactable:        input.Contactable,
		Featured:           input.Featured,
		Status:             status,
		UserID:             userID,
	}, nil
}

func validateBuilderUpdateInput(current domain.Builder, input BuilderInput, userID int64) (repository.BuilderUpsertParams, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.TrimSpace(input.Slug)
	status, err := normalizeUpdateContentStatus(current.Status, input.Status)

	if name == "" {
		return repository.BuilderUpsertParams{}, errors.New("name is required")
	}

	if slug == "" {
		return repository.BuilderUpsertParams{}, errors.New("slug is required")
	}

	if err != nil {
		return repository.BuilderUpsertParams{}, err
	}

	return repository.BuilderUpsertParams{
		Name:               name,
		Slug:               slug,
		Title:              strings.TrimSpace(input.Title),
		City:               strings.TrimSpace(input.City),
		Role:               strings.TrimSpace(input.Role),
		Intro:              strings.TrimSpace(input.Intro),
		Story:              strings.TrimSpace(input.Story),
		Expertise:          input.Expertise,
		FocusAreas:         input.FocusAreas,
		CollaborationModes: input.CollaborationModes,
		CoverURL:           strings.TrimSpace(input.CoverURL),
		Contactable:        input.Contactable,
		Featured:           input.Featured,
		Status:             status,
		UserID:             userID,
	}, nil
}
