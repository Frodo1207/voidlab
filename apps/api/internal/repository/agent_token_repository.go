package repository

import (
	"encoding/json"

	"voidlabai/apps/api/internal/domain"
)

type AgentTokenCreateParams struct {
	Name      string
	TokenHash string
	Scopes    []string
	CreatedBy int64
}

type AgentTokenRecord struct {
	domain.AgentToken
	TokenHash string
}

func (r *AgentTokenRepository) List() ([]domain.AgentToken, error) {
	rows, err := r.db.Query(`
		SELECT id, name, scopes_json, is_active,
		       COALESCE(strftime('%Y-%m-%d %H:%M', last_used_at), ''),
		       created_by,
		       strftime('%Y-%m-%d %H:%M', created_at),
		       strftime('%Y-%m-%d %H:%M', updated_at)
		FROM agent_tokens
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]domain.AgentToken, 0)
	for rows.Next() {
		record, scanErr := scanAgentToken(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *AgentTokenRepository) GetByID(id int64) (domain.AgentToken, error) {
	row := r.db.QueryRow(`
		SELECT id, name, scopes_json, is_active,
		       COALESCE(strftime('%Y-%m-%d %H:%M', last_used_at), ''),
		       created_by,
		       strftime('%Y-%m-%d %H:%M', created_at),
		       strftime('%Y-%m-%d %H:%M', updated_at)
		FROM agent_tokens
		WHERE id = ?
		LIMIT 1
	`, id)

	return scanAgentToken(row)
}

func (r *AgentTokenRepository) GetByHash(tokenHash string) (AgentTokenRecord, error) {
	row := r.db.QueryRow(`
		SELECT id, name, scopes_json, is_active,
		       COALESCE(strftime('%Y-%m-%d %H:%M', last_used_at), ''),
		       created_by,
		       strftime('%Y-%m-%d %H:%M', created_at),
		       strftime('%Y-%m-%d %H:%M', updated_at),
		       token_hash
		FROM agent_tokens
		WHERE token_hash = ?
		LIMIT 1
	`, tokenHash)

	var scopesJSON string
	var record AgentTokenRecord
	err := row.Scan(
		&record.ID,
		&record.Name,
		&scopesJSON,
		&record.IsActive,
		&record.LastUsedAt,
		&record.CreatedBy,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.TokenHash,
	)
	if err != nil {
		return AgentTokenRecord{}, err
	}

	record.Scopes = parseAgentScopes(scopesJSON)
	return record, nil
}

func (r *AgentTokenRepository) Create(params AgentTokenCreateParams) (int64, error) {
	scopesJSON, err := json.Marshal(params.Scopes)
	if err != nil {
		return 0, err
	}

	result, err := r.db.Exec(`
		INSERT INTO agent_tokens (name, token_hash, scopes_json, is_active, created_by, updated_at)
		VALUES (?, ?, ?, 1, ?, CURRENT_TIMESTAMP)
	`, params.Name, params.TokenHash, string(scopesJSON), params.CreatedBy)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *AgentTokenRepository) UpdateStatus(id int64, isActive bool) error {
	_, err := r.db.Exec(`
		UPDATE agent_tokens
		SET is_active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, isActive, id)
	return err
}

func (r *AgentTokenRepository) TouchLastUsed(id int64) error {
	_, err := r.db.Exec(`
		UPDATE agent_tokens
		SET last_used_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, id)
	return err
}

type agentTokenScanner interface {
	Scan(dest ...any) error
}

func scanAgentToken(scanner agentTokenScanner) (domain.AgentToken, error) {
	var record domain.AgentToken
	var scopesJSON string

	err := scanner.Scan(
		&record.ID,
		&record.Name,
		&scopesJSON,
		&record.IsActive,
		&record.LastUsedAt,
		&record.CreatedBy,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		return domain.AgentToken{}, err
	}

	record.Scopes = parseAgentScopes(scopesJSON)
	return record, nil
}

func parseAgentScopes(scopesJSON string) []string {
	scopes := make([]string, 0)
	_ = json.Unmarshal([]byte(scopesJSON), &scopes)
	return scopes
}
