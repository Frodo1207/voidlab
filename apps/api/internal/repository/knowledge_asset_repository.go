package repository

import (
	"database/sql"

	"voidlabai/apps/api/internal/domain"
)

type KnowledgeAssetCreateParams struct {
	SpaceID      int64
	MediaAssetID int64
	CreatedBy    int64
}

type KnowledgeAssetStorageRecord struct {
	domain.KnowledgeAsset
	ObjectKey string
}

func (r *KnowledgeAssetRepository) ListBySpace(spaceID int64) ([]domain.KnowledgeAsset, error) {
	rows, err := r.db.Query(`
                SELECT ka.id, ka.space_id, ka.media_asset_id, ma.file_name, ma.content_type,
                       strftime('%Y-%m-%d %H:%M', ka.created_at)
                FROM knowledge_assets ka
                INNER JOIN media_assets ma ON ma.id = ka.media_asset_id
                WHERE ka.space_id = ?
                ORDER BY ka.created_at DESC, ka.id DESC
        `, spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assets := make([]domain.KnowledgeAsset, 0)
	for rows.Next() {
		record, scanErr := scanKnowledgeAsset(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		assets = append(assets, record)
	}

	return assets, rows.Err()
}

func (r *KnowledgeAssetRepository) Create(params KnowledgeAssetCreateParams) (int64, error) {
	result, err := r.db.Exec(`
                INSERT INTO knowledge_assets (space_id, media_asset_id, created_by)
                VALUES (?, ?, ?)
        `, params.SpaceID, params.MediaAssetID, params.CreatedBy)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *KnowledgeAssetRepository) GetByID(id int64) (domain.KnowledgeAsset, error) {
	row := r.db.QueryRow(`
                SELECT ka.id, ka.space_id, ka.media_asset_id, ma.file_name, ma.content_type,
                       strftime('%Y-%m-%d %H:%M', ka.created_at)
                FROM knowledge_assets ka
                INNER JOIN media_assets ma ON ma.id = ka.media_asset_id
                WHERE ka.id = ?
                LIMIT 1
        `, id)

	return scanKnowledgeAsset(row)
}

func (r *KnowledgeAssetRepository) GetStorageBySpaceAndID(spaceID int64, assetID int64) (KnowledgeAssetStorageRecord, error) {
	row := r.db.QueryRow(`
                SELECT ka.id, ka.space_id, ka.media_asset_id, ma.file_name, ma.content_type, ma.object_key,
                       strftime('%Y-%m-%d %H:%M', ka.created_at)
                FROM knowledge_assets ka
                INNER JOIN media_assets ma ON ma.id = ka.media_asset_id
                WHERE ka.space_id = ? AND ka.id = ?
                LIMIT 1
        `, spaceID, assetID)

	return scanKnowledgeAssetStorage(row)
}

type knowledgeAssetScanner interface {
	Scan(dest ...any) error
}

func scanKnowledgeAsset(scanner knowledgeAssetScanner) (domain.KnowledgeAsset, error) {
	var asset domain.KnowledgeAsset
	var createdAt sql.NullString

	err := scanner.Scan(
		&asset.ID,
		&asset.SpaceID,
		&asset.MediaAssetID,
		&asset.FileName,
		&asset.ContentType,
		&createdAt,
	)
	if err != nil {
		return domain.KnowledgeAsset{}, err
	}

	asset.CreatedAt = createdAt.String
	return asset, nil
}

func scanKnowledgeAssetStorage(scanner knowledgeAssetScanner) (KnowledgeAssetStorageRecord, error) {
	var asset KnowledgeAssetStorageRecord
	var createdAt sql.NullString

	err := scanner.Scan(
		&asset.ID,
		&asset.SpaceID,
		&asset.MediaAssetID,
		&asset.FileName,
		&asset.ContentType,
		&asset.ObjectKey,
		&createdAt,
	)
	if err != nil {
		return KnowledgeAssetStorageRecord{}, err
	}

	asset.CreatedAt = createdAt.String
	return asset, nil
}
