package repository

import (
	"database/sql"

	"voidlabai/apps/api/internal/domain"
)

type SiteConfigRepository struct {
	db *sql.DB
}

func NewSiteConfigRepository(db *sql.DB) *SiteConfigRepository {
	return &SiteConfigRepository{db: db}
}

func (r *SiteConfigRepository) List() ([]domain.SiteConfig, error) {
	rows, err := r.db.Query(`
                SELECT id, config_key, config_value_json, updated_by,
                       strftime('%Y-%m-%d %H:%M', updated_at)
                FROM site_configs
                ORDER BY config_key ASC
        `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]domain.SiteConfig, 0)
	for rows.Next() {
		var record domain.SiteConfig
		if scanErr := rows.Scan(
			&record.ID,
			&record.ConfigKey,
			&record.ConfigValueJSON,
			&record.UpdatedBy,
			&record.UpdatedAt,
		); scanErr != nil {
			return nil, scanErr
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *SiteConfigRepository) GetByKey(key string) (domain.SiteConfig, error) {
	row := r.db.QueryRow(`
                SELECT id, config_key, config_value_json, updated_by,
                       strftime('%Y-%m-%d %H:%M', updated_at)
                FROM site_configs
                WHERE config_key = ?
                LIMIT 1
        `, key)

	var record domain.SiteConfig
	err := row.Scan(
		&record.ID,
		&record.ConfigKey,
		&record.ConfigValueJSON,
		&record.UpdatedBy,
		&record.UpdatedAt,
	)
	return record, err
}

func (r *SiteConfigRepository) Upsert(key string, valueJSON string, updatedBy int64) error {
	_, err := r.db.Exec(`
                INSERT INTO site_configs (config_key, config_value_json, updated_by, updated_at)
                VALUES (?, ?, ?, CURRENT_TIMESTAMP)
                ON CONFLICT(config_key) DO UPDATE SET
                        config_value_json = excluded.config_value_json,
                        updated_by = excluded.updated_by,
                        updated_at = CURRENT_TIMESTAMP
        `, key, valueJSON, updatedBy)
	return err
}
