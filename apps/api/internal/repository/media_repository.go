package repository

import (
	"database/sql"

	"voidlabai/apps/api/internal/domain"
)

type MediaAssetCreateParams struct {
	ObjectKey   string
	ObjectURL   string
	FileName    string
	ContentType string
	FileSize    int64
	UploadedBy  int64
}

func (r *MediaRepository) List() ([]domain.MediaAsset, error) {
	rows, err := r.db.Query(`
		SELECT id, file_name, object_url, content_type, file_size,
		       strftime('%Y-%m-%d %H:%M', created_at)
		FROM media_assets
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assets := make([]domain.MediaAsset, 0)
	for rows.Next() {
		record, scanErr := scanMediaAsset(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		assets = append(assets, record)
	}

	return assets, rows.Err()
}

func (r *MediaRepository) Create(params MediaAssetCreateParams) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO media_assets (
			object_key, object_url, file_name, content_type, file_size, uploaded_by
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		params.ObjectKey,
		params.ObjectURL,
		params.FileName,
		params.ContentType,
		params.FileSize,
		params.UploadedBy,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *MediaRepository) GetByID(id int64) (domain.MediaAsset, error) {
	row := r.db.QueryRow(`
		SELECT id, file_name, object_url, content_type, file_size,
		       strftime('%Y-%m-%d %H:%M', created_at)
		FROM media_assets
		WHERE id = ?
		LIMIT 1
	`, id)

	return scanMediaAsset(row)
}

type mediaScanner interface {
	Scan(dest ...any) error
}

func scanMediaAsset(scanner mediaScanner) (domain.MediaAsset, error) {
	var asset domain.MediaAsset
	var createdAt sql.NullString

	err := scanner.Scan(
		&asset.ID,
		&asset.FileName,
		&asset.ObjectURL,
		&asset.ContentType,
		&asset.FileSize,
		&createdAt,
	)
	if err != nil {
		return domain.MediaAsset{}, err
	}

	asset.CreatedAt = createdAt.String
	return asset, nil
}
