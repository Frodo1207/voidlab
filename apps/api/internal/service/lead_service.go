package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

type LeadService struct {
	repo *repository.LeadRepository
}

type LeadListInput struct {
	SourceType string
	SourceID   *int64
	Status     string
}

type LeadInput struct {
	SourceType string `json:"source_type"`
	SourceID   *int64 `json:"source_id"`
	Name       string `json:"name"`
	Contact    string `json:"contact"`
	Message    string `json:"message"`
	Status     string `json:"status"`
	Notes      string `json:"notes"`
	OwnerID    *int64 `json:"owner_id"`
}

type LeadStatusInput struct {
	Status string `json:"status"`
}

type LeadLogInput struct {
	Action  string `json:"action"`
	Content string `json:"content"`
}

func NewLeadService(repo *repository.LeadRepository) *LeadService {
	return &LeadService{repo: repo}
}

func (s *LeadService) List(input LeadListInput) ([]domain.Lead, error) {
	filter := repository.LeadListFilter{}

	sourceType := strings.TrimSpace(input.SourceType)
	if sourceType != "" {
		if !isValidLeadSourceType(sourceType) {
			return nil, errors.New("invalid lead source type")
		}
		filter.SourceType = sourceType
	}

	if input.SourceID != nil {
		filter.SourceID = input.SourceID
	}

	status := strings.TrimSpace(input.Status)
	if status != "" {
		if !isValidLeadStatus(status) {
			return nil, errors.New("invalid lead status")
		}
		filter.Status = status
	}

	return s.repo.List(filter)
}

func (s *LeadService) GetByID(id int64) (domain.Lead, error) {
	lead, err := s.repo.GetByID(id)
	if err != nil {
		return domain.Lead{}, err
	}

	logs, err := s.repo.ListLogs(id)
	if err != nil {
		return domain.Lead{}, err
	}

	lead.Logs = logs
	return lead, nil
}

func (s *LeadService) Create(input LeadInput) (int64, error) {
	params, err := validateLeadInput(input)
	if err != nil {
		return 0, err
	}

	return s.repo.Create(params)
}

func (s *LeadService) UpdateStatus(id int64, input LeadStatusInput) error {
	current, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("lead not found")
		}
		return err
	}

	status, err := normalizeUpdateLeadStatus(current.Status, input.Status)
	if err != nil {
		return err
	}

	return s.repo.UpdateStatus(id, status)
}

func (s *LeadService) AddLog(id int64, input LeadLogInput, userID int64) error {
	if _, err := s.repo.GetByID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("lead not found")
		}
		return err
	}

	action := strings.TrimSpace(input.Action)
	if action == "" {
		action = "note"
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return errors.New("log content is required")
	}

	_, err := s.repo.AddLog(repository.LeadLogCreateParams{
		LeadID:    id,
		Action:    action,
		Content:   content,
		CreatedBy: userID,
	})
	return err
}

func validateLeadInput(input LeadInput) (repository.LeadCreateParams, error) {
	sourceType := strings.TrimSpace(input.SourceType)
	if !isValidLeadSourceType(sourceType) {
		return repository.LeadCreateParams{}, errors.New("invalid lead source type")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return repository.LeadCreateParams{}, errors.New("name is required")
	}

	contact := strings.TrimSpace(input.Contact)
	if contact == "" {
		return repository.LeadCreateParams{}, errors.New("contact is required")
	}

	status, err := normalizeCreateLeadStatus(input.Status)
	if err != nil {
		return repository.LeadCreateParams{}, err
	}

	dedupeKey := buildLeadDedupeKey(sourceType, input.SourceID, contact)

	return repository.LeadCreateParams{
		SourceType: sourceType,
		SourceID:   input.SourceID,
		Name:       name,
		Contact:    contact,
		Message:    strings.TrimSpace(input.Message),
		Status:     status,
		Notes:      strings.TrimSpace(input.Notes),
		DedupeKey:  dedupeKey,
		OwnerID:    input.OwnerID,
	}, nil
}

func isValidLeadSourceType(value string) bool {
	switch value {
	case "contact", "event", "builder":
		return true
	default:
		return false
	}
}

func isValidLeadStatus(value string) bool {
	switch value {
	case "new", "applied", "approved", "waitlisted", "rejected", "checked_in", "contacted", "following", "converted", "invalid":
		return true
	default:
		return false
	}
}

func buildLeadDedupeKey(sourceType string, sourceID *int64, contact string) string {
	if sourceType != "event" || sourceID == nil {
		return ""
	}

	normalizedContact := strings.ToLower(strings.TrimSpace(contact))
	if normalizedContact == "" {
		return ""
	}

	return fmt.Sprintf("%s#event#%d", normalizedContact, *sourceID)
}
