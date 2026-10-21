package repository

import (
	"database/sql"
	"encoding/json"
	"errors"

	"voidlabai/apps/api/internal/domain"
)

type BuilderNetworkAgentTokenRepository struct {
	db *sql.DB
}

func NewBuilderNetworkAgentTokenRepository(db *sql.DB) *BuilderNetworkAgentTokenRepository {
	return &BuilderNetworkAgentTokenRepository{db: db}
}

func (r *BuilderNetworkAgentTokenRepository) Create(profileID int64, name, tokenHash string, scopes []string) (int64, error) {
	scopesJSON, err := json.Marshal(scopes)
	if err != nil {
		return 0, err
	}
	result, err := r.db.Exec(`
		INSERT INTO network_agent_tokens (profile_id, name, token_hash, scopes_json, is_active, updated_at)
		VALUES (?, ?, ?, ?, 1, CURRENT_TIMESTAMP)
	`, profileID, name, tokenHash, string(scopesJSON))
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *BuilderNetworkAgentTokenRepository) List(profileID int64) ([]domain.NetworkAgentToken, error) {
	rows, err := r.db.Query(`
		SELECT id, profile_id, name, scopes_json, is_active,
		       COALESCE(strftime('%Y-%m-%d %H:%M', last_used_at), ''),
		       strftime('%Y-%m-%d %H:%M', created_at),
		       strftime('%Y-%m-%d %H:%M', updated_at)
		FROM network_agent_tokens
		WHERE profile_id = ?
		ORDER BY created_at DESC, id DESC
	`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tokens := make([]domain.NetworkAgentToken, 0)
	for rows.Next() {
		token, err := scanNetworkAgentToken(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (r *BuilderNetworkAgentTokenRepository) GetByID(profileID, tokenID int64) (domain.NetworkAgentToken, error) {
	return scanNetworkAgentToken(r.db.QueryRow(`
		SELECT id, profile_id, name, scopes_json, is_active,
		       COALESCE(strftime('%Y-%m-%d %H:%M', last_used_at), ''),
		       strftime('%Y-%m-%d %H:%M', created_at),
		       strftime('%Y-%m-%d %H:%M', updated_at)
		FROM network_agent_tokens
		WHERE profile_id = ? AND id = ?
	`, profileID, tokenID))
}

func (r *BuilderNetworkAgentTokenRepository) UpdateStatus(profileID, tokenID int64, isActive bool) error {
	result, err := r.db.Exec(`
		UPDATE network_agent_tokens
		SET is_active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE profile_id = ? AND id = ?
	`, isActive, profileID, tokenID)
	if err != nil {
		return err
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if updated == 0 {
		return ErrNetworkNotFound
	}
	return nil
}

func (r *BuilderNetworkAgentTokenRepository) ResolveToken(tokenHash string) (domain.NetworkAgentPrincipal, error) {
	var principal domain.NetworkAgentPrincipal
	var scopesJSON string
	err := r.db.QueryRow(`
		SELECT t.id, t.profile_id, t.name, t.scopes_json, t.is_active,
		       COALESCE(strftime('%Y-%m-%d %H:%M', t.last_used_at), ''),
		       strftime('%Y-%m-%d %H:%M', t.created_at),
		       strftime('%Y-%m-%d %H:%M', t.updated_at),
		       a.id, a.email, a.status, a.created_at
		FROM network_agent_tokens t
		JOIN network_profiles p ON p.id = t.profile_id
		JOIN network_accounts a ON a.id = p.account_id
		WHERE t.token_hash = ? AND t.is_active = 1 AND a.status = 'active'
	`, tokenHash).Scan(
		&principal.Token.ID, &principal.Token.ProfileID, &principal.Token.Name, &scopesJSON,
		&principal.Token.IsActive, &principal.Token.LastUsedAt, &principal.Token.CreatedAt,
		&principal.Token.UpdatedAt, &principal.Account.ID, &principal.Account.Email,
		&principal.Account.Status, &principal.Account.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NetworkAgentPrincipal{}, sql.ErrNoRows
	}
	if err != nil {
		return domain.NetworkAgentPrincipal{}, err
	}
	if err := json.Unmarshal([]byte(scopesJSON), &principal.Token.Scopes); err != nil {
		return domain.NetworkAgentPrincipal{}, err
	}
	if principal.Token.Scopes == nil {
		principal.Token.Scopes = []string{}
	}
	if _, err := r.db.Exec(`
		UPDATE network_agent_tokens
		SET last_used_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, principal.Token.ID); err != nil {
		return domain.NetworkAgentPrincipal{}, err
	}
	return principal, nil
}

type networkAgentTokenScanner interface {
	Scan(dest ...any) error
}

func scanNetworkAgentToken(scanner networkAgentTokenScanner) (domain.NetworkAgentToken, error) {
	var token domain.NetworkAgentToken
	var scopesJSON string
	err := scanner.Scan(&token.ID, &token.ProfileID, &token.Name, &scopesJSON, &token.IsActive,
		&token.LastUsedAt, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return domain.NetworkAgentToken{}, err
	}
	if err := json.Unmarshal([]byte(scopesJSON), &token.Scopes); err != nil {
		return domain.NetworkAgentToken{}, err
	}
	if token.Scopes == nil {
		token.Scopes = []string{}
	}
	return token, nil
}
