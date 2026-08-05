package service

import (
	"database/sql"
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"

	"voidlabai/apps/api/internal/config"
	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

type KnowledgeAssetService struct {
	spaceRepo    *repository.KnowledgeSpaceRepository
	assetRepo    *repository.KnowledgeAssetRepository
	mediaService *MediaService
	uploadsDir   string
}

func NewKnowledgeAssetService(
	cfg config.Config,
	spaceRepo *repository.KnowledgeSpaceRepository,
	assetRepo *repository.KnowledgeAssetRepository,
	mediaService *MediaService,
) *KnowledgeAssetService {
	return &KnowledgeAssetService{
		spaceRepo:    spaceRepo,
		assetRepo:    assetRepo,
		mediaService: mediaService,
		uploadsDir:   cfg.UploadsDir,
	}
}

func (s *KnowledgeAssetService) ListBySpace(spaceID int64) ([]domain.KnowledgeAsset, error) {
	return s.assetRepo.ListBySpace(spaceID)
}

func (s *KnowledgeAssetService) Upload(spaceID int64, fileHeader *multipart.FileHeader, userID int64) (domain.KnowledgeAsset, error) {
	if spaceID <= 0 {
		return domain.KnowledgeAsset{}, errors.New("knowledge space not found")
	}
	if _, err := s.spaceRepo.GetByID(spaceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.KnowledgeAsset{}, errors.New("knowledge space not found")
		}
		return domain.KnowledgeAsset{}, err
	}

	mediaRecord, err := s.mediaService.Upload(fileHeader, userID)
	if err != nil {
		return domain.KnowledgeAsset{}, err
	}

	assetID, err := s.assetRepo.Create(repository.KnowledgeAssetCreateParams{
		SpaceID:      spaceID,
		MediaAssetID: mediaRecord.ID,
		CreatedBy:    userID,
	})
	if err != nil {
		return domain.KnowledgeAsset{}, err
	}

	return s.assetRepo.GetByID(assetID)
}

func (s *KnowledgeAssetService) GetStorageBySpaceAndID(spaceID int64, assetID int64) (repository.KnowledgeAssetStorageRecord, string, error) {
	record, err := s.assetRepo.GetStorageBySpaceAndID(spaceID, assetID)
	if err != nil {
		return repository.KnowledgeAssetStorageRecord{}, "", err
	}

	path := filepath.Join(s.uploadsDir, record.ObjectKey)
	if _, err := os.Stat(path); err != nil {
		return repository.KnowledgeAssetStorageRecord{}, "", err
	}

	return record, path, nil
}
