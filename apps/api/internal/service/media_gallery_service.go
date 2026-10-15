package service

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/dustin/go-humanize"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

type MediaGalleryService struct {
	repo *repository.MediaGalleryRepository
}

type MediaGalleryInput struct {
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
}

type MediaGalleryItemsAppendInput struct {
	MediaAssetIDs []int64 `json:"media_asset_ids"`
}

type MediaGalleryItemUpdateInput struct {
	AltText   string `json:"alt_text"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

func NewMediaGalleryService(repo *repository.MediaGalleryRepository) *MediaGalleryService {
	return &MediaGalleryService{repo: repo}
}

func (s *MediaGalleryService) List() ([]domain.MediaGallery, error) {
	return s.repo.List()
}

func (s *MediaGalleryService) GetByID(id int64) (domain.MediaGallery, error) {
	return s.repo.GetByID(id)
}

func (s *MediaGalleryService) Create(input MediaGalleryInput, userID int64) (int64, error) {
	params, err := validateMediaGalleryInput(input, userID)
	if err != nil {
		return 0, err
	}
	return s.repo.Create(params)
}

func (s *MediaGalleryService) Update(id int64, input MediaGalleryInput, userID int64) error {
	if _, err := s.repo.GetByID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("gallery not found")
		}
		return err
	}

	params, err := validateMediaGalleryInput(input, userID)
	if err != nil {
		return err
	}
	return s.repo.Update(id, params)
}

func (s *MediaGalleryService) Delete(id int64) error {
	if _, err := s.repo.GetByID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("gallery not found")
		}
		return err
	}
	return s.repo.Delete(id)
}

func (s *MediaGalleryService) ListItems(galleryID int64) ([]domain.MediaGalleryItem, error) {
	if _, err := s.repo.GetByID(galleryID); err != nil {
		return nil, err
	}

	items, err := s.repo.ListItems(galleryID)
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index].FileSizeLabel = humanize.Bytes(uint64(items[index].FileSize))
	}
	return items, nil
}

func (s *MediaGalleryService) AddItems(galleryID int64, input MediaGalleryItemsAppendInput) error {
	if _, err := s.repo.GetByID(galleryID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("gallery not found")
		}
		return err
	}

	if len(input.MediaAssetIDs) == 0 {
		return errors.New("media_asset_ids is required")
	}

	seen := make(map[int64]struct{}, len(input.MediaAssetIDs))
	for _, mediaAssetID := range input.MediaAssetIDs {
		if mediaAssetID <= 0 {
			return errors.New("media_asset_ids contains invalid id")
		}
		if _, exists := seen[mediaAssetID]; exists {
			continue
		}
		seen[mediaAssetID] = struct{}{}

		sortOrder, err := s.repo.NextSortOrder(galleryID)
		if err != nil {
			return err
		}
		if err := s.repo.AddItem(repository.MediaGalleryItemCreateParams{
			GalleryID:    galleryID,
			MediaAssetID: mediaAssetID,
			SortOrder:    sortOrder,
			IsActive:     true,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *MediaGalleryService) UpdateItem(galleryID, itemID int64, input MediaGalleryItemUpdateInput) error {
	if _, err := s.repo.GetByID(galleryID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("gallery not found")
		}
		return err
	}

	current, err := s.repo.GetItemByID(galleryID, itemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("gallery item not found")
		}
		return err
	}

	sortOrder := input.SortOrder
	if sortOrder <= 0 {
		sortOrder = current.SortOrder
	}

	return s.repo.UpdateItem(galleryID, itemID, repository.MediaGalleryItemUpdateParams{
		AltText:   strings.TrimSpace(input.AltText),
		SortOrder: sortOrder,
		IsActive:  input.IsActive,
	})
}

func (s *MediaGalleryService) DeleteItem(galleryID, itemID int64) error {
	if _, err := s.repo.GetItemByID(galleryID, itemID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("gallery item not found")
		}
		return err
	}
	return s.repo.DeleteItem(galleryID, itemID)
}

func validateMediaGalleryInput(input MediaGalleryInput, userID int64) (repository.MediaGalleryUpsertParams, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return repository.MediaGalleryUpsertParams{}, errors.New("title is required")
	}

	slug := strings.TrimSpace(input.Slug)
	if slug == "" {
		slug = buildKnowledgeImportSlug(title)
	}
	if slug == "" {
		return repository.MediaGalleryUpsertParams{}, errors.New("slug is required")
	}

	return repository.MediaGalleryUpsertParams{
		Title:       title,
		Slug:        slug,
		Description: strings.TrimSpace(input.Description),
		IsDefault:   input.IsDefault,
		UserID:      userID,
	}, nil
}
