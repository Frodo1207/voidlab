package repository

import (
	"database/sql"
	"strings"

	"voidlabai/apps/api/internal/domain"
)

type LeadListFilter struct {
	SourceType string
	SourceID   *int64
	Status     string
}

type LeadCreateParams struct {
	SourceType string
	SourceID   *int64
	Name       string
	Contact    string
	Message    string
	Status     string
	Notes      string
	DedupeKey  string
	OwnerID    *int64
}

type LeadLogCreateParams struct {
	LeadID    int64
	Action    string
	Content   string
	CreatedBy int64
}

func (r *LeadRepository) List(filter LeadListFilter) ([]domain.Lead, error) {
	query := `
                SELECT id, source_type, source_id, name, contact, message, status, notes, dedupe_key, owner_id,
                       strftime('%Y-%m-%d %H:%M', created_at),
                       strftime('%Y-%m-%d %H:%M', updated_at)
                FROM leads
        `
	clauses := make([]string, 0, 3)
	args := make([]any, 0, 3)

	if filter.SourceType != "" {
		clauses = append(clauses, "source_type = ?")
		args = append(args, filter.SourceType)
	}
	if filter.SourceID != nil {
		clauses = append(clauses, "source_id = ?")
		args = append(args, *filter.SourceID)
	}
	if filter.Status != "" {
		clauses = append(clauses, "status = ?")
		args = append(args, filter.Status)
	}
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY created_at DESC, id DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]domain.Lead, 0)
	for rows.Next() {
		record, scanErr := scanLead(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *LeadRepository) GetByID(id int64) (domain.Lead, error) {
	row := r.db.QueryRow(`
                SELECT id, source_type, source_id, name, contact, message, status, notes, dedupe_key, owner_id,
                       strftime('%Y-%m-%d %H:%M', created_at),
                       strftime('%Y-%m-%d %H:%M', updated_at)
                FROM leads
                WHERE id = ?
                LIMIT 1
        `, id)

	return scanLead(row)
}

func (r *LeadRepository) Create(params LeadCreateParams) (int64, error) {
	result, err := r.db.Exec(`
                INSERT INTO leads (
                        source_type, source_id, name, contact, message, status, notes, dedupe_key, owner_id
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `,
		params.SourceType,
		params.SourceID,
		params.Name,
		params.Contact,
		params.Message,
		params.Status,
		params.Notes,
		params.DedupeKey,
		params.OwnerID,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *LeadRepository) UpdateStatus(id int64, status string) error {
	_, err := r.db.Exec(`
                UPDATE leads
                SET status = ?, updated_at = CURRENT_TIMESTAMP
                WHERE id = ?
        `, status, id)
	return err
}

func (r *LeadRepository) ListLogs(leadID int64) ([]domain.LeadLog, error) {
	rows, err := r.db.Query(`
                SELECT id, lead_id, action, content, created_by, strftime('%Y-%m-%d %H:%M', created_at)
                FROM lead_logs
                WHERE lead_id = ?
                ORDER BY created_at DESC, id DESC
        `, leadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]domain.LeadLog, 0)
	for rows.Next() {
		record, scanErr := scanLeadLog(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		logs = append(logs, record)
	}

	return logs, rows.Err()
}

func (r *LeadRepository) AddLog(params LeadLogCreateParams) (int64, error) {
	result, err := r.db.Exec(`
                INSERT INTO lead_logs (lead_id, action, content, created_by)
                VALUES (?, ?, ?, ?)
        `,
		params.LeadID,
		params.Action,
		params.Content,
		params.CreatedBy,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *LeadRepository) CountAll() (int64, error) {
	return countTableRows(r.db, "leads")
}

func (r *LeadRepository) StatusCounts() (domain.LeadStatusStats, error) {
	rows, err := r.db.Query(`
		SELECT status, COUNT(1)
		FROM leads
		GROUP BY status
	`)
	if err != nil {
		return domain.LeadStatusStats{}, err
	}
	defer rows.Close()

	stats := domain.LeadStatusStats{}
	for rows.Next() {
		var status string
		var count int64

		if scanErr := rows.Scan(&status, &count); scanErr != nil {
			return domain.LeadStatusStats{}, scanErr
		}

		switch strings.TrimSpace(status) {
		case "new":
			stats.New = count
		case "contacted":
			stats.Contacted = count
		case "following":
			stats.Following = count
		case "converted":
			stats.Converted = count
		case "invalid":
			stats.Invalid = count
		}
	}

	return stats, rows.Err()
}

func (r *LeadRepository) ListActionable(limit int) ([]domain.Lead, error) {
	rows, err := r.db.Query(`
                SELECT id, source_type, source_id, name, contact, message, status, notes, dedupe_key, owner_id,
		       strftime('%Y-%m-%d %H:%M', created_at),
		       strftime('%Y-%m-%d %H:%M', updated_at)
		FROM leads
                WHERE status IN ('new', 'applied', 'approved', 'waitlisted', 'contacted', 'following')
		ORDER BY updated_at DESC, id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]domain.Lead, 0)
	for rows.Next() {
		record, scanErr := scanLead(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

type leadScanner interface {
	Scan(dest ...any) error
}

func scanLead(scanner leadScanner) (domain.Lead, error) {
	var lead domain.Lead
	var sourceID sql.NullInt64
	var ownerID sql.NullInt64
	var createdAt sql.NullString
	var updatedAt sql.NullString

	err := scanner.Scan(
		&lead.ID,
		&lead.SourceType,
		&sourceID,
		&lead.Name,
		&lead.Contact,
		&lead.Message,
		&lead.Status,
		&lead.Notes,
		&lead.DedupeKey,
		&ownerID,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.Lead{}, err
	}

	if sourceID.Valid {
		lead.SourceID = &sourceID.Int64
	}
	if ownerID.Valid {
		lead.OwnerID = &ownerID.Int64
	}
	lead.CreatedAt = createdAt.String
	lead.UpdatedAt = updatedAt.String

	return lead, nil
}

func scanLeadLog(scanner leadScanner) (domain.LeadLog, error) {
	var log domain.LeadLog
	var createdAt sql.NullString

	err := scanner.Scan(
		&log.ID,
		&log.LeadID,
		&log.Action,
		&log.Content,
		&log.CreatedBy,
		&createdAt,
	)
	if err != nil {
		return domain.LeadLog{}, err
	}

	log.CreatedAt = createdAt.String
	return log, nil
}
