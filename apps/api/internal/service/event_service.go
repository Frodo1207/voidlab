package service

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

type EventService struct {
	repo *repository.EventRepository
}

const (
	EventSignupModeInternal = "internal"
	EventSignupModeExternal = "external"
	EventSignupModeClosed   = "closed"
)

const (
	EventSignupStatusOpen       = "open"
	EventSignupStatusNotStarted = "not_started"
	EventSignupStatusClosed     = "closed"
	EventSignupStatusFull       = "full"
	EventSignupStatusLive       = "live_locked"
	EventSignupStatusEnded      = "ended"
	EventSignupStatusExternal   = "external"
)

type EventInput struct {
	Title                 string `json:"title"`
	Slug                  string `json:"slug"`
	Summary               string `json:"summary"`
	City                  string `json:"city"`
	Location              string `json:"location"`
	EventType             string `json:"event_type"`
	EventTime             string `json:"event_time"`
	CoverURL              string `json:"cover_url"`
	Content               string `json:"content"`
	Status                string `json:"status"`
	SignupMode            string `json:"signup_mode"`
	SignupEnabled         bool   `json:"signup_enabled"`
	SignupStartsAt        string `json:"signup_starts_at"`
	SignupDeadline        string `json:"signup_deadline"`
	Capacity              int64  `json:"capacity"`
	AllowSignupDuringLive bool   `json:"allow_signup_during_live"`
	ExternalSignupURL     string `json:"external_signup_url"`
	SignupButtonLabel     string `json:"signup_button_label"`
	SignupSuccessMessage  string `json:"signup_success_message"`
	SignupClosedReason    string `json:"signup_closed_reason"`
}

func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) List() ([]domain.Event, error) {
	events, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	for index := range events {
		events[index].SignupStatus = DeriveEventSignupStatus(events[index])
	}
	return events, nil
}

func (s *EventService) ListPublished() ([]domain.Event, error) {
	events, err := s.repo.ListPublished()
	if err != nil {
		return nil, err
	}
	for index := range events {
		events[index].SignupStatus = DeriveEventSignupStatus(events[index])
	}
	return events, nil
}

func (s *EventService) GetByID(id int64) (domain.Event, error) {
	event, err := s.repo.GetByID(id)
	if err != nil {
		return domain.Event{}, err
	}
	event.SignupStatus = DeriveEventSignupStatus(event)
	return event, nil
}

func (s *EventService) GetPublishedBySlug(slug string) (domain.Event, error) {
	event, err := s.repo.GetPublishedBySlug(strings.TrimSpace(slug))
	if err != nil {
		return domain.Event{}, err
	}
	event.SignupStatus = DeriveEventSignupStatus(event)
	return event, nil
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

	signupParams, err := validateEventSignupConfig(input)
	if err != nil {
		return repository.EventUpsertParams{}, err
	}

	return repository.EventUpsertParams{
		Title:                 title,
		Slug:                  slug,
		Summary:               strings.TrimSpace(input.Summary),
		City:                  strings.TrimSpace(input.City),
		Location:              strings.TrimSpace(input.Location),
		EventType:             strings.TrimSpace(input.EventType),
		EventTime:             strings.TrimSpace(input.EventTime),
		CoverURL:              strings.TrimSpace(input.CoverURL),
		Content:               strings.TrimSpace(input.Content),
		Status:                status,
		SignupMode:            signupParams.SignupMode,
		SignupEnabled:         signupParams.SignupEnabled,
		SignupStartsAt:        signupParams.SignupStartsAt,
		SignupDeadline:        signupParams.SignupDeadline,
		Capacity:              signupParams.Capacity,
		AllowSignupDuringLive: signupParams.AllowSignupDuringLive,
		ExternalSignupURL:     signupParams.ExternalSignupURL,
		SignupButtonLabel:     signupParams.SignupButtonLabel,
		SignupSuccessMessage:  signupParams.SignupSuccessMessage,
		SignupClosedReason:    signupParams.SignupClosedReason,
		UserID:                userID,
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

	signupParams, err := validateEventSignupConfig(input)
	if err != nil {
		return repository.EventUpsertParams{}, err
	}

	return repository.EventUpsertParams{
		Title:                 title,
		Slug:                  slug,
		Summary:               strings.TrimSpace(input.Summary),
		City:                  strings.TrimSpace(input.City),
		Location:              strings.TrimSpace(input.Location),
		EventType:             strings.TrimSpace(input.EventType),
		EventTime:             strings.TrimSpace(input.EventTime),
		CoverURL:              strings.TrimSpace(input.CoverURL),
		Content:               strings.TrimSpace(input.Content),
		Status:                status,
		SignupMode:            signupParams.SignupMode,
		SignupEnabled:         signupParams.SignupEnabled,
		SignupStartsAt:        signupParams.SignupStartsAt,
		SignupDeadline:        signupParams.SignupDeadline,
		Capacity:              signupParams.Capacity,
		AllowSignupDuringLive: signupParams.AllowSignupDuringLive,
		ExternalSignupURL:     signupParams.ExternalSignupURL,
		SignupButtonLabel:     signupParams.SignupButtonLabel,
		SignupSuccessMessage:  signupParams.SignupSuccessMessage,
		SignupClosedReason:    signupParams.SignupClosedReason,
		UserID:                userID,
	}, nil
}

func DeriveEventSignupStatus(event domain.Event) string {
	if event.Status != "published" {
		return EventSignupStatusClosed
	}

	if !event.SignupEnabled {
		return EventSignupStatusClosed
	}

	switch event.SignupMode {
	case EventSignupModeExternal:
		return EventSignupStatusExternal
	case EventSignupModeClosed:
		return EventSignupStatusClosed
	}

	now := time.Now()
	eventTime := parseEventDateTime(event.EventTime)
	signupStartsAt := parseEventDateTime(event.SignupStartsAt)
	signupDeadline := parseEventDateTime(event.SignupDeadline)

	if signupStartsAt != nil && now.Before(*signupStartsAt) {
		return EventSignupStatusNotStarted
	}

	if signupDeadline != nil && now.After(*signupDeadline) {
		return EventSignupStatusClosed
	}

	if eventTime != nil && now.After(*eventTime) {
		return EventSignupStatusEnded
	}

	if event.Capacity > 0 && event.SignupCount >= event.Capacity {
		return EventSignupStatusFull
	}

	if eventTime != nil && isSameCalendarDay(now, *eventTime) && !event.AllowSignupDuringLive {
		return EventSignupStatusLive
	}

	return EventSignupStatusOpen
}

func validateEventSignupConfig(input EventInput) (repository.EventUpsertParams, error) {
	signupMode := strings.TrimSpace(input.SignupMode)
	if signupMode == "" {
		signupMode = EventSignupModeInternal
	}
	if signupMode != EventSignupModeInternal && signupMode != EventSignupModeExternal && signupMode != EventSignupModeClosed {
		return repository.EventUpsertParams{}, errors.New("invalid signup mode")
	}

	if input.Capacity < 0 {
		return repository.EventUpsertParams{}, errors.New("capacity cannot be negative")
	}

	signupStartsAt := strings.TrimSpace(input.SignupStartsAt)
	signupDeadline := strings.TrimSpace(input.SignupDeadline)
	if signupStartsAt != "" && parseEventDateTime(signupStartsAt) == nil {
		return repository.EventUpsertParams{}, errors.New("signup_starts_at must be a valid datetime")
	}
	if signupDeadline != "" && parseEventDateTime(signupDeadline) == nil {
		return repository.EventUpsertParams{}, errors.New("signup_deadline must be a valid datetime")
	}

	externalSignupURL := strings.TrimSpace(input.ExternalSignupURL)
	if signupMode == EventSignupModeExternal && externalSignupURL == "" {
		return repository.EventUpsertParams{}, errors.New("external_signup_url is required for external signup mode")
	}

	return repository.EventUpsertParams{
		SignupMode:            signupMode,
		SignupEnabled:         input.SignupEnabled,
		SignupStartsAt:        signupStartsAt,
		SignupDeadline:        signupDeadline,
		Capacity:              input.Capacity,
		AllowSignupDuringLive: input.AllowSignupDuringLive,
		ExternalSignupURL:     externalSignupURL,
		SignupButtonLabel:     strings.TrimSpace(input.SignupButtonLabel),
		SignupSuccessMessage:  strings.TrimSpace(input.SignupSuccessMessage),
		SignupClosedReason:    strings.TrimSpace(input.SignupClosedReason),
	}, nil
}

func parseEventDateTime(value string) *time.Time {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return nil
	}

	layouts := []string{
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, normalized, time.Local); err == nil {
			return &parsed
		}
	}
	return nil
}

func isSameCalendarDay(left time.Time, right time.Time) bool {
	return left.Year() == right.Year() && left.YearDay() == right.YearDay()
}
