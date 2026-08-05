package service

import (
	"database/sql"
	"errors"
	"strings"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

type EventService struct {
	repo *repository.EventRepository
}

type EventInput struct {
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	Summary   string `json:"summary"`
	City      string `json:"city"`
	Location  string `json:"location"`
	EventType string `json:"event_type"`
	EventTime string `json:"event_time"`
	CoverURL  string `json:"cover_url"`
	Content   string `json:"content"`
	Status    string `json:"status"`
}

func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) List() ([]domain.Event, error) {
	return s.repo.List()
}

func (s *EventService) ListPublished() ([]domain.Event, error) {
	return s.repo.ListPublished()
}

func (s *EventService) GetByID(id int64) (domain.Event, error) {
	return s.repo.GetByID(id)
}

func (s *EventService) GetPublishedBySlug(slug string) (domain.Event, error) {
	return s.repo.GetPublishedBySlug(strings.TrimSpace(slug))
}

func (s *EventService) Create(input EventInput, userID int64) (int64, error) {
	params, err := validateEventCreateInput(input, userID)
	if err != nil {
		return 0, err
	}

	return s.repo.Create(params)
}

func (s *EventService) Update(id int64, input EventInput, userID int64) error {
	current, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("event not found")
		}
		return err
	}

	params, err := validateEventUpdateInput(current, input, userID)
	if err != nil {
		return err
	}

	return s.repo.Update(id, params)
}

func (s *EventService) UpdateStatus(id int64, input ContentStatusInput, userID int64) error {
	current, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("event not found")
		}
		return err
	}

	status, err := normalizeUpdateContentStatus(current.Status, input.Status)
	if err != nil {
		return err
	}

	return s.repo.UpdateStatus(id, status, userID)
}

func (s *EventService) Delete(id int64) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("event not found")
		}
		return err
	}

	return s.repo.Delete(id)
}

func validateEventCreateInput(input EventInput, userID int64) (repository.EventUpsertParams, error) {
	title := strings.TrimSpace(input.Title)
	slug := strings.TrimSpace(input.Slug)
	status, err := normalizeCreateContentStatus(input.Status)

	if title == "" {
		return repository.EventUpsertParams{}, errors.New("title is required")
	}

	if slug == "" {
		return repository.EventUpsertParams{}, errors.New("slug is required")
	}

	if err != nil {
		return repository.EventUpsertParams{}, err
	}

	return repository.EventUpsertParams{
		Title:     title,
		Slug:      slug,
		Summary:   strings.TrimSpace(input.Summary),
		City:      strings.TrimSpace(input.City),
		Location:  strings.TrimSpace(input.Location),
		EventType: strings.TrimSpace(input.EventType),
		EventTime: strings.TrimSpace(input.EventTime),
		CoverURL:  strings.TrimSpace(input.CoverURL),
		Content:   strings.TrimSpace(input.Content),
		Status:    status,
		UserID:    userID,
	}, nil
}

func validateEventUpdateInput(current domain.Event, input EventInput, userID int64) (repository.EventUpsertParams, error) {
	title := strings.TrimSpace(input.Title)
	slug := strings.TrimSpace(input.Slug)
	status, err := normalizeUpdateContentStatus(current.Status, input.Status)

	if title == "" {
		return repository.EventUpsertParams{}, errors.New("title is required")
	}

	if slug == "" {
		return repository.EventUpsertParams{}, errors.New("slug is required")
	}

	if err != nil {
		return repository.EventUpsertParams{}, err
	}

	return repository.EventUpsertParams{
		Title:     title,
		Slug:      slug,
		Summary:   strings.TrimSpace(input.Summary),
		City:      strings.TrimSpace(input.City),
		Location:  strings.TrimSpace(input.Location),
		EventType: strings.TrimSpace(input.EventType),
		EventTime: strings.TrimSpace(input.EventTime),
		CoverURL:  strings.TrimSpace(input.CoverURL),
		Content:   strings.TrimSpace(input.Content),
		Status:    status,
		UserID:    userID,
	}, nil
}
