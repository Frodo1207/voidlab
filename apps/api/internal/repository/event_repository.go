package repository

import (
	"database/sql"

	"voidlabai/apps/api/internal/domain"
)

type EventUpsertParams struct {
	Title                 string
	Slug                  string
	Summary               string
	City                  string
	Location              string
	EventType             string
	EventTime             string
	CoverURL              string
	Content               string
	Status                string
	SignupMode            string
	SignupEnabled         bool
	SignupStartsAt        string
	SignupDeadline        string
	Capacity              int64
	AllowSignupDuringLive bool
	ExternalSignupURL     string
	SignupButtonLabel     string
	SignupSuccessMessage  string
	SignupClosedReason    string
	UserID                int64
}

func (r *EventRepository) List() ([]domain.Event, error) {
	rows, err := r.db.Query(`
                SELECT e.id, e.title, e.slug, e.summary, e.city, e.location, e.event_type, e.event_time,
                       e.content, e.cover_url, e.status, e.signup_mode, e.signup_enabled, e.signup_starts_at,
                       e.signup_deadline, e.capacity, e.allow_signup_during_live, e.external_signup_url,
                       e.signup_button_label, e.signup_success_message, e.signup_closed_reason,
                       COUNT(l.id) AS signup_count,
                       strftime('%Y-%m-%d %H:%M', e.updated_at)
                FROM events e
                LEFT JOIN leads l ON l.source_type = 'event' AND l.source_id = e.id
                GROUP BY e.id
                ORDER BY e.updated_at DESC, e.id DESC
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
                SELECT e.id, e.title, e.slug, e.summary, e.city, e.location, e.event_type, e.event_time,
                       e.content, e.cover_url, e.status, e.signup_mode, e.signup_enabled, e.signup_starts_at,
                       e.signup_deadline, e.capacity, e.allow_signup_during_live, e.external_signup_url,
                       e.signup_button_label, e.signup_success_message, e.signup_closed_reason,
                       COUNT(l.id) AS signup_count,
                       strftime('%Y-%m-%d %H:%M', e.updated_at)
                FROM events e
                LEFT JOIN leads l ON l.source_type = 'event' AND l.source_id = e.id
                WHERE e.status = 'published'
                GROUP BY e.id
                ORDER BY e.event_time DESC, e.updated_at DESC, e.id DESC
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
                SELECT e.id, e.title, e.slug, e.summary, e.city, e.location, e.event_type, e.event_time,
                       e.content, e.cover_url, e.status, e.signup_mode, e.signup_enabled, e.signup_starts_at,
                       e.signup_deadline, e.capacity, e.allow_signup_during_live, e.external_signup_url,
                       e.signup_button_label, e.signup_success_message, e.signup_closed_reason,
                       COUNT(l.id) AS signup_count,
                       strftime('%Y-%m-%d %H:%M', e.updated_at)
                FROM events e
                LEFT JOIN leads l ON l.source_type = 'event' AND l.source_id = e.id
                WHERE e.id = ?
                GROUP BY e.id
		LIMIT 1
	`, id)

	return scanEvent(row)
}

func (r *EventRepository) GetPublishedBySlug(slug string) (domain.Event, error) {
	row := r.db.QueryRow(`
                SELECT e.id, e.title, e.slug, e.summary, e.city, e.location, e.event_type, e.event_time,
                       e.content, e.cover_url, e.status, e.signup_mode, e.signup_enabled, e.signup_starts_at,
                       e.signup_deadline, e.capacity, e.allow_signup_during_live, e.external_signup_url,
                       e.signup_button_label, e.signup_success_message, e.signup_closed_reason,
                       COUNT(l.id) AS signup_count,
                       strftime('%Y-%m-%d %H:%M', e.updated_at)
                FROM events e
                LEFT JOIN leads l ON l.source_type = 'event' AND l.source_id = e.id
                WHERE e.slug = ? AND e.status = 'published'
                GROUP BY e.id
		LIMIT 1
	`, slug)

	return scanEvent(row)
}

func (r *EventRepository) Create(params EventUpsertParams) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO events (
			title, slug, event_time, location, city, event_type, status, summary,
                        cover_url, content, signup_mode, signup_enabled, signup_starts_at,
                        signup_deadline, capacity, allow_signup_during_live, external_signup_url,
                        signup_button_label, signup_success_message, signup_closed_reason,
                        created_by, updated_by
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		params.SignupMode,
		params.SignupEnabled,
		nullableString(params.SignupStartsAt),
		nullableString(params.SignupDeadline),
		params.Capacity,
		params.AllowSignupDuringLive,
		params.ExternalSignupURL,
		params.SignupButtonLabel,
		params.SignupSuccessMessage,
		params.SignupClosedReason,
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
                    status = ?, summary = ?, cover_url = ?, content = ?, signup_mode = ?, signup_enabled = ?,
                    signup_starts_at = ?, signup_deadline = ?, capacity = ?, allow_signup_during_live = ?,
                    external_signup_url = ?, signup_button_label = ?, signup_success_message = ?,
                    signup_closed_reason = ?, updated_by = ?, updated_at = CURRENT_TIMESTAMP
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
		params.SignupMode,
		params.SignupEnabled,
		nullableString(params.SignupStartsAt),
		nullableString(params.SignupDeadline),
		params.Capacity,
		params.AllowSignupDuringLive,
		params.ExternalSignupURL,
		params.SignupButtonLabel,
		params.SignupSuccessMessage,
		params.SignupClosedReason,
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
	var signupStartsAt sql.NullString
	var signupDeadline sql.NullString
	var capacity sql.NullInt64
	var signupCount sql.NullInt64
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
		&event.SignupMode,
		&event.SignupEnabled,
		&signupStartsAt,
		&signupDeadline,
		&capacity,
		&event.AllowSignupDuringLive,
		&event.ExternalSignupURL,
		&event.SignupButtonLabel,
		&event.SignupSuccessMessage,
		&event.SignupClosedReason,
		&signupCount,
		&updatedAt,
	)
	if err != nil {
		return domain.Event{}, err
	}

	event.EventTime = eventTime.String
	event.SignupStartsAt = signupStartsAt.String
	event.SignupDeadline = signupDeadline.String
	event.Capacity = capacity.Int64
	event.SignupCount = signupCount.Int64
	event.UpdatedAt = updatedAt.String

	return event, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
