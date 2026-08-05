package service

import (
	"database/sql"
	"errors"
	"strings"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

type ArticleService struct {
	repo *repository.ArticleRepository
}

type ArticleInput struct {
	Title      string   `json:"title"`
	Slug       string   `json:"slug"`
	Summary    string   `json:"summary"`
	Category   string   `json:"category"`
	Audience   string   `json:"audience"`
	Tags       []string `json:"tags"`
	CoverURL   string   `json:"cover_url"`
	Content    string   `json:"content"`
	SourceName string   `json:"source_name"`
	SourceURL  string   `json:"source_url"`
	Featured   bool     `json:"featured"`
	Status     string   `json:"status"`
}

func NewArticleService(repo *repository.ArticleRepository) *ArticleService {
	return &ArticleService{repo: repo}
}

func (s *ArticleService) List() ([]domain.Article, error) {
	return s.repo.List()
}

func (s *ArticleService) ListPublished() ([]domain.Article, error) {
	return s.repo.ListPublished()
}

func (s *ArticleService) GetByID(id int64) (domain.Article, error) {
	return s.repo.GetByID(id)
}

func (s *ArticleService) GetPublishedBySlug(slug string) (domain.Article, error) {
	return s.repo.GetPublishedBySlug(strings.TrimSpace(slug))
}

func (s *ArticleService) Create(input ArticleInput, userID int64) (int64, error) {
	params, err := validateArticleCreateInput(input, userID)
	if err != nil {
		return 0, err
	}

	return s.repo.Create(params)
}

func (s *ArticleService) Update(id int64, input ArticleInput, userID int64) error {
	current, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("article not found")
		}
		return err
	}

	params, err := validateArticleUpdateInput(current, input, userID)
	if err != nil {
		return err
	}

	return s.repo.Update(id, params)
}

func (s *ArticleService) UpdateStatus(id int64, input ContentStatusInput, userID int64) error {
	current, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("article not found")
		}
		return err
	}

	status, err := normalizeUpdateContentStatus(current.Status, input.Status)
	if err != nil {
		return err
	}

	return s.repo.UpdateStatus(id, status, userID)
}

func (s *ArticleService) Delete(id int64) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("article not found")
		}
		return err
	}

	return s.repo.Delete(id)
}

func validateArticleCreateInput(input ArticleInput, userID int64) (repository.ArticleUpsertParams, error) {
	title := strings.TrimSpace(input.Title)
	slug := strings.TrimSpace(input.Slug)
	status, err := normalizeCreateContentStatus(input.Status)

	if title == "" {
		return repository.ArticleUpsertParams{}, errors.New("title is required")
	}

	if slug == "" {
		return repository.ArticleUpsertParams{}, errors.New("slug is required")
	}

	if err != nil {
		return repository.ArticleUpsertParams{}, err
	}

	return repository.ArticleUpsertParams{
		Title:      title,
		Slug:       slug,
		Summary:    strings.TrimSpace(input.Summary),
		Category:   strings.TrimSpace(input.Category),
		Audience:   strings.TrimSpace(input.Audience),
		Tags:       input.Tags,
		CoverURL:   strings.TrimSpace(input.CoverURL),
		Content:    strings.TrimSpace(input.Content),
		SourceName: strings.TrimSpace(input.SourceName),
		SourceURL:  strings.TrimSpace(input.SourceURL),
		Featured:   input.Featured,
		Status:     status,
		UserID:     userID,
	}, nil
}

func validateArticleUpdateInput(current domain.Article, input ArticleInput, userID int64) (repository.ArticleUpsertParams, error) {
	title := strings.TrimSpace(input.Title)
	slug := strings.TrimSpace(input.Slug)
	status, err := normalizeUpdateContentStatus(current.Status, input.Status)

	if title == "" {
		return repository.ArticleUpsertParams{}, errors.New("title is required")
	}

	if slug == "" {
		return repository.ArticleUpsertParams{}, errors.New("slug is required")
	}

	if err != nil {
		return repository.ArticleUpsertParams{}, err
	}

	return repository.ArticleUpsertParams{
		Title:      strings.TrimSpace(input.Title),
		Slug:       strings.TrimSpace(input.Slug),
		Summary:    strings.TrimSpace(input.Summary),
		Category:   strings.TrimSpace(input.Category),
		Audience:   strings.TrimSpace(input.Audience),
		Tags:       input.Tags,
		CoverURL:   strings.TrimSpace(input.CoverURL),
		Content:    strings.TrimSpace(input.Content),
		SourceName: strings.TrimSpace(input.SourceName),
		SourceURL:  strings.TrimSpace(input.SourceURL),
		Featured:   input.Featured,
		Status:     status,
		UserID:     userID,
	}, nil
}
