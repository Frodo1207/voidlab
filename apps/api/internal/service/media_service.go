package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/google/uuid"

	"voidlabai/apps/api/internal/config"
	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

type MediaService struct {
	repo          *repository.MediaRepository
	uploadsDir    string
	publicBaseURL string
}

func NewMediaService(cfg config.Config, repo *repository.MediaRepository) *MediaService {
	return &MediaService{
		repo:          repo,
		uploadsDir:    cfg.UploadsDir,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}
}

func (s *MediaService) List() ([]domain.MediaAsset, error) {
	records, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	for index := range records {
		records[index].FileSizeLabel = humanize.Bytes(uint64(records[index].FileSize))
	}

	return records, nil
}

func (s *MediaService) Upload(fileHeader *multipart.FileHeader, userID int64) (domain.MediaAsset, error) {
	if fileHeader == nil {
		return domain.MediaAsset{}, fmt.Errorf("file is required")
	}

	if err := os.MkdirAll(s.uploadsDir, 0o755); err != nil {
		return domain.MediaAsset{}, fmt.Errorf("create uploads dir: %w", err)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return domain.MediaAsset{}, fmt.Errorf("open upload file: %w", err)
	}
	defer src.Close()

	extension := filepath.Ext(fileHeader.Filename)
	objectKey := fmt.Sprintf("%s-%s%s", time.Now().Format("20060102150405"), uuid.NewString(), extension)
	objectPath := filepath.Join(s.uploadsDir, objectKey)

	dst, err := os.Create(objectPath)
	if err != nil {
		return domain.MediaAsset{}, fmt.Errorf("create destination file: %w", err)
	}

	size, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	if copyErr != nil {
		return domain.MediaAsset{}, fmt.Errorf("write upload file: %w", copyErr)
	}
	if closeErr != nil {
		return domain.MediaAsset{}, fmt.Errorf("close upload file: %w", closeErr)
	}

	objectURL := fmt.Sprintf("%s/uploads/%s", s.publicBaseURL, objectKey)
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	id, err := s.repo.Create(repository.MediaAssetCreateParams{
		ObjectKey:   objectKey,
		ObjectURL:   objectURL,
		FileName:    fileHeader.Filename,
		ContentType: contentType,
		FileSize:    size,
		UploadedBy:  userID,
	})
	if err != nil {
		return domain.MediaAsset{}, err
	}

	record, err := s.repo.GetByID(id)
	if err != nil {
		return domain.MediaAsset{}, err
	}

	record.FileSizeLabel = humanize.Bytes(uint64(record.FileSize))
	return record, nil
}
