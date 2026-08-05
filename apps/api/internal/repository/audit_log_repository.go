package repository

import (
	"database/sql"

	"voidlabai/apps/api/internal/domain"
)

type AuditLogCreateParams struct {
	ActorType     string
	ActorID       int64
	ActorUsername string
	ActorRole     string
	AgentTokenID  *int64
	Action        string
	EntityType    string
	EntityID      *int64
	EntityLabel   string
	DetailJSON    string
}

func (r *AuditLogRepository) List(limit int) ([]domain.AuditLog, error) {
	rows, err := r.db.Query(`
		SELECT id, actor_type, actor_id, actor_username, actor_role, agent_token_id, action, entity_type, entity_id,
		       entity_label, detail_json, strftime('%Y-%m-%d %H:%M', created_at)
		FROM audit_logs
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]domain.AuditLog, 0)
	for rows.Next() {
		record, scanErr := scanAuditLog(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *AuditLogRepository) Create(params AuditLogCreateParams) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO audit_logs (
			actor_type, actor_id, actor_username, actor_role, agent_token_id, action, entity_type, entity_id, entity_label, detail_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		params.ActorType,
		params.ActorID,
		params.ActorUsername,
		params.ActorRole,
		params.AgentTokenID,
		params.Action,
		params.EntityType,
		params.EntityID,
		params.EntityLabel,
		params.DetailJSON,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

type auditLogScanner interface {
	Scan(dest ...any) error
}

func scanAuditLog(scanner auditLogScanner) (domain.AuditLog, error) {
	var record domain.AuditLog
	var agentTokenID sql.NullInt64
	var entityID sql.NullInt64
	var createdAt sql.NullString

	err := scanner.Scan(
		&record.ID,
		&record.ActorType,
		&record.ActorID,
		&record.ActorUsername,
		&record.ActorRole,
		&agentTokenID,
		&record.Action,
		&record.EntityType,
		&entityID,
		&record.EntityLabel,
		&record.DetailJSON,
		&createdAt,
	)
	if err != nil {
		return domain.AuditLog{}, err
	}

	if entityID.Valid {
		record.EntityID = &entityID.Int64
	}
	if agentTokenID.Valid {
		record.AgentTokenID = &agentTokenID.Int64
	}
	record.CreatedAt = createdAt.String

	return record, nil
}
