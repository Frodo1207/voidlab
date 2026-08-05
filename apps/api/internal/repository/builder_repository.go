package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"voidlabai/apps/api/internal/domain"
)

type BuilderUpsertParams struct {
	Name               string
	Slug               string
	Title              string
	City               string
	Role               string
	Intro              string
	Story              string
	Expertise          []string
	FocusAreas         []string
	CollaborationModes []string
	CoverURL           string
	Contactable        bool
	Featured           bool
	Status             string
	UserID             int64
}

func (r *BuilderRepository) List() ([]domain.Builder, error) {
	rows, err := r.db.Query(`
		SELECT id, name, slug, title, city, role, intro, story,
		       expertise_json, focus_areas_json, collaboration_modes_json,
		       contactable, featured, cover_url, status,
		       strftime('%Y-%m-%d %H:%M', updated_at)
		FROM builders
		ORDER BY updated_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	builders := make([]domain.Builder, 0)
	for rows.Next() {
		record, scanErr := scanBuilder(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		builders = append(builders, record)
	}

	return builders, rows.Err()
}

func (r *BuilderRepository) ListPublished() ([]domain.Builder, error) {
	rows, err := r.db.Query(`
                SELECT id, name, slug, title, city, role, intro, story,
                       expertise_json, focus_areas_json, collaboration_modes_json,
                       contactable, featured, cover_url, status,
                       strftime('%Y-%m-%d %H:%M', updated_at)
                FROM builders
                WHERE status = 'published'
                ORDER BY featured DESC, updated_at DESC, id DESC
        `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	builders := make([]domain.Builder, 0)
	for rows.Next() {
		record, scanErr := scanBuilder(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		builders = append(builders, record)
	}

	return builders, rows.Err()
}

func (r *BuilderRepository) GetByID(id int64) (domain.Builder, error) {
	row := r.db.QueryRow(`
		SELECT id, name, slug, title, city, role, intro, story,
		       expertise_json, focus_areas_json, collaboration_modes_json,
		       contactable, featured, cover_url, status,
		       strftime('%Y-%m-%d %H:%M', updated_at)
		FROM builders
		WHERE id = ?
		LIMIT 1
	`, id)

	return scanBuilder(row)
}

func (r *BuilderRepository) GetPublishedBySlug(slug string) (domain.Builder, error) {
	row := r.db.QueryRow(`
                SELECT id, name, slug, title, city, role, intro, story,
                       expertise_json, focus_areas_json, collaboration_modes_json,
                       contactable, featured, cover_url, status,
                       strftime('%Y-%m-%d %H:%M', updated_at)
                FROM builders
                WHERE slug = ? AND status = 'published'
                LIMIT 1
        `, slug)

	return scanBuilder(row)
}

func (r *BuilderRepository) Create(params BuilderUpsertParams) (int64, error) {
	expertiseJSON, focusAreasJSON, collaborationModesJSON, err := marshalBuilderJSON(params)
	if err != nil {
		return 0, err
	}

	result, err := r.db.Exec(`
		INSERT INTO builders (
			name, slug, title, city, role, intro, story, expertise_json, focus_areas_json,
			collaboration_modes_json, contactable, featured, cover_url, status, created_by, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		params.Name,
		params.Slug,
		params.Title,
		params.City,
		params.Role,
		params.Intro,
		params.Story,
		expertiseJSON,
		focusAreasJSON,
		collaborationModesJSON,
		boolToInt(params.Contactable),
		boolToInt(params.Featured),
		params.CoverURL,
		params.Status,
		params.UserID,
		params.UserID,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *BuilderRepository) Update(id int64, params BuilderUpsertParams) error {
	expertiseJSON, focusAreasJSON, collaborationModesJSON, err := marshalBuilderJSON(params)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		UPDATE builders
		SET name = ?, slug = ?, title = ?, city = ?, role = ?, intro = ?, story = ?,
		    expertise_json = ?, focus_areas_json = ?, collaboration_modes_json = ?,
		    contactable = ?, featured = ?, cover_url = ?, status = ?, updated_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`,
		params.Name,
		params.Slug,
		params.Title,
		params.City,
		params.Role,
		params.Intro,
		params.Story,
		expertiseJSON,
		focusAreasJSON,
		collaborationModesJSON,
		boolToInt(params.Contactable),
		boolToInt(params.Featured),
		params.CoverURL,
		params.Status,
		params.UserID,
		id,
	)

	return err
}

func (r *BuilderRepository) UpdateStatus(id int64, status string, userID int64) error {
	_, err := r.db.Exec(`
		UPDATE builders
		SET status = ?, updated_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`,
		status,
		userID,
		id,
	)

	return err
}

func (r *BuilderRepository) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM builders WHERE id = ?`, id)
	return err
}

func (r *BuilderRepository) CountAll() (int64, error) {
	return countTableRows(r.db, "builders")
}

func (r *BuilderRepository) CountByStatus(status string) (int64, error) {
	return countTableRowsByStatus(r.db, "builders", status)
}

func marshalBuilderJSON(params BuilderUpsertParams) (string, string, string, error) {
	expertiseJSON, err := json.Marshal(params.Expertise)
	if err != nil {
		return "", "", "", fmt.Errorf("marshal expertise: %w", err)
	}

	focusAreasJSON, err := json.Marshal(params.FocusAreas)
	if err != nil {
		return "", "", "", fmt.Errorf("marshal focus areas: %w", err)
	}

	collaborationModesJSON, err := json.Marshal(params.CollaborationModes)
	if err != nil {
		return "", "", "", fmt.Errorf("marshal collaboration modes: %w", err)
	}

	return string(expertiseJSON), string(focusAreasJSON), string(collaborationModesJSON), nil
}

type builderScanner interface {
	Scan(dest ...any) error
}

func scanBuilder(scanner builderScanner) (domain.Builder, error) {
	var builder domain.Builder
	var expertiseJSON string
	var focusAreasJSON string
	var collaborationModesJSON string
	var contactable int
	var featured int
	var updatedAt sql.NullString

	err := scanner.Scan(
		&builder.ID,
		&builder.Name,
		&builder.Slug,
		&builder.Title,
		&builder.City,
		&builder.Role,
		&builder.Intro,
		&builder.Story,
		&expertiseJSON,
		&focusAreasJSON,
		&collaborationModesJSON,
		&contactable,
		&featured,
		&builder.CoverURL,
		&builder.Status,
		&updatedAt,
	)
	if err != nil {
		return domain.Builder{}, err
	}

	if expertiseJSON != "" {
		if unmarshalErr := json.Unmarshal([]byte(expertiseJSON), &builder.Expertise); unmarshalErr != nil {
			return domain.Builder{}, fmt.Errorf("unmarshal expertise: %w", unmarshalErr)
		}
	}

	if focusAreasJSON != "" {
		if unmarshalErr := json.Unmarshal([]byte(focusAreasJSON), &builder.FocusAreas); unmarshalErr != nil {
			return domain.Builder{}, fmt.Errorf("unmarshal focus areas: %w", unmarshalErr)
		}
	}

	if collaborationModesJSON != "" {
		if unmarshalErr := json.Unmarshal([]byte(collaborationModesJSON), &builder.CollaborationModes); unmarshalErr != nil {
			return domain.Builder{}, fmt.Errorf("unmarshal collaboration modes: %w", unmarshalErr)
		}
	}

	builder.Contactable = contactable == 1
	builder.Featured = featured == 1
	builder.UpdatedAt = updatedAt.String

	return builder, nil
}
