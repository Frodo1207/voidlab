package repository

import (
	"database/sql"

	"voidlabai/apps/api/internal/domain"
)

type EventUpsertParams struct {
	Title     string
	Slug      string
	Summary   string
	City      string
	Location  string
	EventType string
	EventTime string
	CoverURL  string
	Content   string
	Status    string
	UserID    int64
}

func (r *EventRepository) List() ([]domain.Event, error) {
	rows, err := r.db.Query(`
		SELECT id, title, slug, summary, city, location, event_type, event_time,
		       content, cover_url, status, strftime('%Y-%m-%d %H:%M', updated_at)
		FROM events
		ORDER BY updated_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domain.Event, 0)
	for rows.Next() {
		record, scanErr := scanEvent(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		events = append(events, record)
	}

	return events, rows.Err()
}

func (r *EventRepository) ListPublished() ([]domain.Event, error) {
	rows, err := r.db.Query(`
		SELECT id, title, slug, summary, city, location, event_type, event_time,
		       content, cover_url, status, strftime('%Y-%m-%d %H:%M', updated_at)
		FROM events
		WHERE status = 'published'
		ORDER BY event_time DESC, updated_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domain.Event, 0)
	for rows.Next() {
		record, scanErr := scanEvent(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		events = append(events, record)
	}

	return events, rows.Err()
}

func (r *EventRepository) GetByID(id int64) (domain.Event, error) {
	row := r.db.QueryRow(`
		SELECT id, title, slug, summary, city, location, event_type, event_time,
		       content, cover_url, status, strftime('%Y-%m-%d %H:%M', updated_at)
		FROM events
		WHERE id = ?
		LIMIT 1
	`, id)

	return scanEvent(row)
}

func (r *EventRepository) GetPublishedBySlug(slug string) (domain.Event, error) {
	row := r.db.QueryRow(`
		SELECT id, title, slug, summary, city, location, event_type, event_time,
		       content, cover_url, status, strftime('%Y-%m-%d %H:%M', updated_at)
		FROM events
		WHERE slug = ? AND status = 'published'
		LIMIT 1
	`, slug)

	return scanEvent(row)
}

func (r *EventRepository) Create(params EventUpsertParams) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO events (
			title, slug, event_time, location, city, event_type, status, summary,
			cover_url, content, created_by, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		params.Title,
		params.Slug,
		params.EventTime,
		params.Location,
		params.City,
		params.EventType,
		params.Status,
		params.Summary,
		params.CoverURL,
		params.Content,
		params.UserID,
		params.UserID,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *EventRepository) Update(id int64, params EventUpsertParams) error {
	_, err := r.db.Exec(`
		UPDATE events
		SET title = ?, slug = ?, event_time = ?, location = ?, city = ?, event_type = ?,
		    status = ?, summary = ?, cover_url = ?, content = ?, updated_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`,
		params.Title,
		params.Slug,
		params.EventTime,
		params.Location,
		params.City,
		params.EventType,
		params.Status,
		params.Summary,
		params.CoverURL,
		params.Content,
		params.UserID,
		id,
	)
	return err
}

func (r *EventRepository) UpdateStatus(id int64, status string, userID int64) error {
	_, err := r.db.Exec(`
		UPDATE events
		SET status = ?, updated_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`,
		status,
		userID,
		id,
	)
	return err
}

func (r *EventRepository) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM events WHERE id = ?`, id)
	return err
}

func (r *EventRepository) CountAll() (int64, error) {
	return countTableRows(r.db, "events")
}

func (r *EventRepository) CountByStatus(status string) (int64, error) {
	return countTableRowsByStatus(r.db, "events", status)
}

type eventScanner interface {
	Scan(dest ...any) error
}

func scanEvent(scanner eventScanner) (domain.Event, error) {
	var event domain.Event
	var eventTime sql.NullString
	var updatedAt sql.NullString

	err := scanner.Scan(
		&event.ID,
		&event.Title,
		&event.Slug,
		&event.Summary,
		&event.City,
		&event.Location,
		&event.EventType,
		&eventTime,
		&event.Content,
		&event.CoverURL,
		&event.Status,
		&updatedAt,
	)
	if err != nil {
		return domain.Event{}, err
	}

	event.EventTime = eventTime.String
	event.UpdatedAt = updatedAt.String

	return event, nil
}
