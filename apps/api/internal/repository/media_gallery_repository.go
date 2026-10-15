package repository

import (
	"database/sql"

	"voidlabai/apps/api/internal/domain"
)

type MediaGalleryUpsertParams struct {
	Title       string
	Slug        string
	Description string
	IsDefault   bool
	UserID      int64
}

type MediaGalleryItemCreateParams struct {
	GalleryID    int64
	MediaAssetID int64
	AltText      string
	SortOrder    int
	IsActive     bool
}

type MediaGalleryItemUpdateParams struct {
	AltText   string
	SortOrder int
	IsActive  bool
}

func (r *MediaGalleryRepository) List() ([]domain.MediaGallery, error) {
	rows, err := r.db.Query(`
		SELECT g.id, g.title, g.slug, g.description, g.is_default,
		       COALESCE(COUNT(i.id), 0) AS item_count,
		       COALESCE((
		                SELECT m2.object_url
		                FROM media_gallery_items i2
		                INNER JOIN media_assets m2 ON m2.id = i2.media_asset_id
		                WHERE i2.gallery_id = g.id AND i2.is_active = 1
		                ORDER BY i2.sort_order ASC, i2.id ASC
		                LIMIT 1
		              ), '') AS cover_url,
		       strftime('%Y-%m-%d %H:%M', g.updated_at)
		FROM media_galleries g
		LEFT JOIN media_gallery_items i ON i.gallery_id = g.id AND i.is_active = 1
		GROUP BY g.id, g.title, g.slug, g.description, g.is_default, g.updated_at
		ORDER BY g.is_default DESC, g.updated_at DESC, g.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	galleries := make([]domain.MediaGallery, 0)
	for rows.Next() {
		record, scanErr := scanMediaGallery(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		galleries = append(galleries, record)
	}

	return galleries, rows.Err()
}

func (r *MediaGalleryRepository) GetByID(id int64) (domain.MediaGallery, error) {
	row := r.db.QueryRow(`
		SELECT g.id, g.title, g.slug, g.description, g.is_default,
		       COALESCE(COUNT(i.id), 0) AS item_count,
		       COALESCE((
		                SELECT m2.object_url
		                FROM media_gallery_items i2
		                INNER JOIN media_assets m2 ON m2.id = i2.media_asset_id
		                WHERE i2.gallery_id = g.id AND i2.is_active = 1
		                ORDER BY i2.sort_order ASC, i2.id ASC
		                LIMIT 1
		              ), '') AS cover_url,
		       strftime('%Y-%m-%d %H:%M', g.updated_at)
		FROM media_galleries g
		LEFT JOIN media_gallery_items i ON i.gallery_id = g.id AND i.is_active = 1
		WHERE g.id = ?
		GROUP BY g.id, g.title, g.slug, g.description, g.is_default, g.updated_at
		LIMIT 1
	`, id)

	return scanMediaGallery(row)
}

func (r *MediaGalleryRepository) Create(params MediaGalleryUpsertParams) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO media_galleries (
			title, slug, description, is_default, created_by, updated_by
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		params.Title,
		params.Slug,
		params.Description,
		boolToInt(params.IsDefault),
		params.UserID,
		params.UserID,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *MediaGalleryRepository) Update(id int64, params MediaGalleryUpsertParams) error {
	_, err := r.db.Exec(`
		UPDATE media_galleries
		SET title = ?, slug = ?, description = ?, is_default = ?, updated_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`,
		params.Title,
		params.Slug,
		params.Description,
		boolToInt(params.IsDefault),
		params.UserID,
		id,
	)
	return err
}

func (r *MediaGalleryRepository) Delete(id int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM media_gallery_items WHERE gallery_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM media_galleries WHERE id = ?`, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MediaGalleryRepository) ListItems(galleryID int64) ([]domain.MediaGalleryItem, error) {
	rows, err := r.db.Query(`
		SELECT i.id, i.gallery_id, i.media_asset_id, m.file_name, m.object_url, m.content_type, m.file_size,
		       i.alt_text, i.sort_order, i.is_active, strftime('%Y-%m-%d %H:%M', i.created_at)
		FROM media_gallery_items i
		INNER JOIN media_assets m ON m.id = i.media_asset_id
		WHERE i.gallery_id = ?
		ORDER BY i.sort_order ASC, i.id ASC
	`, galleryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.MediaGalleryItem, 0)
	for rows.Next() {
		record, scanErr := scanMediaGalleryItem(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, record)
	}

	return items, rows.Err()
}

func (r *MediaGalleryRepository) GetItemByID(galleryID, itemID int64) (domain.MediaGalleryItem, error) {
	row := r.db.QueryRow(`
		SELECT i.id, i.gallery_id, i.media_asset_id, m.file_name, m.object_url, m.content_type, m.file_size,
		       i.alt_text, i.sort_order, i.is_active, strftime('%Y-%m-%d %H:%M', i.created_at)
		FROM media_gallery_items i
		INNER JOIN media_assets m ON m.id = i.media_asset_id
		WHERE i.gallery_id = ? AND i.id = ?
		LIMIT 1
	`, galleryID, itemID)
	return scanMediaGalleryItem(row)
}

func (r *MediaGalleryRepository) NextSortOrder(galleryID int64) (int, error) {
	var maxSort sql.NullInt64
	if err := r.db.QueryRow(`
		SELECT MAX(sort_order) FROM media_gallery_items WHERE gallery_id = ?
	`, galleryID).Scan(&maxSort); err != nil {
		return 0, err
	}
	if !maxSort.Valid {
		return 10, nil
	}
	return int(maxSort.Int64) + 10, nil
}

func (r *MediaGalleryRepository) AddItem(params MediaGalleryItemCreateParams) error {
	_, err := r.db.Exec(`
		INSERT INTO media_gallery_items (
			gallery_id, media_asset_id, alt_text, sort_order, is_active
		)
		SELECT ?, m.id, ?, ?, ?
		FROM media_assets m
		WHERE m.id = ?
		  AND (
			lower(m.content_type) LIKE 'image/%'
			OR lower(m.file_name) LIKE '%.jpg'
			OR lower(m.file_name) LIKE '%.jpeg'
			OR lower(m.file_name) LIKE '%.png'
			OR lower(m.file_name) LIKE '%.webp'
			OR lower(m.file_name) LIKE '%.gif'
			OR lower(m.file_name) LIKE '%.svg'
		  )
		ON CONFLICT(gallery_id, media_asset_id) DO NOTHING
	`,
		params.GalleryID,
		params.AltText,
		params.SortOrder,
		boolToInt(params.IsActive),
		params.MediaAssetID,
	)
	return err
}

func (r *MediaGalleryRepository) UpdateItem(galleryID, itemID int64, params MediaGalleryItemUpdateParams) error {
	_, err := r.db.Exec(`
		UPDATE media_gallery_items
		SET alt_text = ?, sort_order = ?, is_active = ?
		WHERE gallery_id = ? AND id = ?
	`,
		params.AltText,
		params.SortOrder,
		boolToInt(params.IsActive),
		galleryID,
		itemID,
	)
	return err
}

func (r *MediaGalleryRepository) DeleteItem(galleryID, itemID int64) error {
	_, err := r.db.Exec(`DELETE FROM media_gallery_items WHERE gallery_id = ? AND id = ?`, galleryID, itemID)
	return err
}

func scanMediaGallery(scanner mediaScanner) (domain.MediaGallery, error) {
	var gallery domain.MediaGallery
	var isDefault int
	var updatedAt sql.NullString

	err := scanner.Scan(
		&gallery.ID,
		&gallery.Title,
		&gallery.Slug,
		&gallery.Description,
		&isDefault,
		&gallery.ItemCount,
		&gallery.CoverURL,
		&updatedAt,
	)
	if err != nil {
		return domain.MediaGallery{}, err
	}

	gallery.IsDefault = isDefault == 1
	gallery.UpdatedAt = updatedAt.String
	return gallery, nil
}

func scanMediaGalleryItem(scanner mediaScanner) (domain.MediaGalleryItem, error) {
	var item domain.MediaGalleryItem
	var isActive int
	var createdAt sql.NullString

	err := scanner.Scan(
		&item.ID,
		&item.GalleryID,
		&item.MediaAssetID,
		&item.FileName,
		&item.ObjectURL,
		&item.ContentType,
		&item.FileSize,
		&item.AltText,
		&item.SortOrder,
		&isActive,
		&createdAt,
	)
	if err != nil {
		return domain.MediaGalleryItem{}, err
	}

	item.IsActive = isActive == 1
	item.CreatedAt = createdAt.String
	return item, nil
}
