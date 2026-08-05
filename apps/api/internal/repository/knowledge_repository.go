package repository

import (
	"database/sql"
	"strconv"
	"strings"

	"voidlabai/apps/api/internal/domain"
)

type KnowledgeSpaceUpsertParams struct {
	Title            string
	Slug             string
	Description      string
	CoverLabel       string
	Icon             string
	ThemeTint        string
	VisibilityMode   string
	DirectorySummary string
	IntroMarkdown    string
	TokenHint        string
	CoverURL         string
	Status           string
	UserID           int64
}

type KnowledgeEntryUpsertParams struct {
	SpaceID              int64
	Title                string
	Slug                 string
	SectionName          string
	SortOrder            int
	EstimatedReadMinutes int
	PublicSummary        string
	ContentMarkdown      string
	CoverURL             string
	IsPreview            bool
	Status               string
	UserID               int64
}

type KnowledgeAccessTokenCreateParams struct {
	Name        string
	AccessLevel string
	ScopeType   string
	SpaceIDs    []int64
	TokenHash   string
	ExpiresAt   string
	CreatedBy   int64
}

type KnowledgeAccessTokenRecord struct {
	domain.KnowledgeAccessToken
	TokenHash string
}

type KnowledgeAccessLogCreateParams struct {
	SpaceID   int64
	EntryID   *int64
	TokenID   *int64
	Action    string
	RequestIP string
	UserAgent string
}

func (r *KnowledgeSpaceRepository) List() ([]domain.KnowledgeSpace, error) {
	rows, err := r.db.Query(`
                SELECT id, title, slug, description, cover_label, icon, theme_tint, visibility_mode,
                       directory_summary, intro_markdown, token_hint, cover_url, status,
                       (SELECT COUNT(1) FROM knowledge_entries WHERE space_id = knowledge_spaces.id),
                       (SELECT COUNT(DISTINCT section_name) FROM knowledge_entries WHERE space_id = knowledge_spaces.id AND section_name != ''),
                       strftime('%Y-%m-%d %H:%M', updated_at)
                FROM knowledge_spaces
                ORDER BY updated_at DESC, id DESC
        `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	spaces := make([]domain.KnowledgeSpace, 0)
	for rows.Next() {
		record, scanErr := scanKnowledgeSpace(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		spaces = append(spaces, record)
	}

	return spaces, rows.Err()
}

func (r *KnowledgeSpaceRepository) ListPublished() ([]domain.KnowledgeSpace, error) {
	rows, err := r.db.Query(`
                SELECT id, title, slug, description, cover_label, icon, theme_tint, visibility_mode,
                       directory_summary, intro_markdown, token_hint, cover_url, status,
                       (SELECT COUNT(1) FROM knowledge_entries WHERE space_id = knowledge_spaces.id AND status = 'published'),
                       (SELECT COUNT(DISTINCT section_name) FROM knowledge_entries WHERE space_id = knowledge_spaces.id AND status = 'published' AND section_name != ''),
                       strftime('%Y-%m-%d %H:%M', updated_at)
                FROM knowledge_spaces
                WHERE status = 'published' AND visibility_mode != 'private_hidden'
                ORDER BY updated_at DESC, id DESC
        `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	spaces := make([]domain.KnowledgeSpace, 0)
	for rows.Next() {
		record, scanErr := scanKnowledgeSpace(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		spaces = append(spaces, record)
	}

	return spaces, rows.Err()
}

func (r *KnowledgeSpaceRepository) GetByID(id int64) (domain.KnowledgeSpace, error) {
	row := r.db.QueryRow(`
                SELECT id, title, slug, description, cover_label, icon, theme_tint, visibility_mode,
                       directory_summary, intro_markdown, token_hint, cover_url, status,
                       (SELECT COUNT(1) FROM knowledge_entries WHERE space_id = knowledge_spaces.id),
                       (SELECT COUNT(DISTINCT section_name) FROM knowledge_entries WHERE space_id = knowledge_spaces.id AND section_name != ''),
                       strftime('%Y-%m-%d %H:%M', updated_at)
                FROM knowledge_spaces
                WHERE id = ?
                LIMIT 1
        `, id)

	return scanKnowledgeSpace(row)
}

func (r *KnowledgeSpaceRepository) GetBySlug(slug string) (domain.KnowledgeSpace, error) {
	row := r.db.QueryRow(`
                SELECT id, title, slug, description, cover_label, icon, theme_tint, visibility_mode,
                       directory_summary, intro_markdown, token_hint, cover_url, status,
                       (SELECT COUNT(1) FROM knowledge_entries WHERE space_id = knowledge_spaces.id),
                       (SELECT COUNT(DISTINCT section_name) FROM knowledge_entries WHERE space_id = knowledge_spaces.id AND section_name != ''),
                       strftime('%Y-%m-%d %H:%M', updated_at)
                FROM knowledge_spaces
                WHERE slug = ?
                LIMIT 1
        `, slug)

	return scanKnowledgeSpace(row)
}

func (r *KnowledgeSpaceRepository) GetPublishedBySlug(slug string) (domain.KnowledgeSpace, error) {
	row := r.db.QueryRow(`
                SELECT id, title, slug, description, cover_label, icon, theme_tint, visibility_mode,
                       directory_summary, intro_markdown, token_hint, cover_url, status,
                       (SELECT COUNT(1) FROM knowledge_entries WHERE space_id = knowledge_spaces.id AND status = 'published'),
                       (SELECT COUNT(DISTINCT section_name) FROM knowledge_entries WHERE space_id = knowledge_spaces.id AND status = 'published' AND section_name != ''),
                       strftime('%Y-%m-%d %H:%M', updated_at)
                FROM knowledge_spaces
                WHERE slug = ? AND status = 'published' AND visibility_mode != 'private_hidden'
                LIMIT 1
        `, slug)

	return scanKnowledgeSpace(row)
}

func (r *KnowledgeSpaceRepository) Create(params KnowledgeSpaceUpsertParams) (int64, error) {
	result, err := r.db.Exec(`
                INSERT INTO knowledge_spaces (
                        title, slug, description, cover_label, icon, theme_tint, visibility_mode,
                        directory_summary, intro_markdown, token_hint, cover_url, status, created_by, updated_by
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `,
		params.Title,
		params.Slug,
		params.Description,
		params.CoverLabel,
		params.Icon,
		params.ThemeTint,
		params.VisibilityMode,
		params.DirectorySummary,
		params.IntroMarkdown,
		params.TokenHint,
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

func (r *KnowledgeSpaceRepository) Update(id int64, params KnowledgeSpaceUpsertParams) error {
	_, err := r.db.Exec(`
                UPDATE knowledge_spaces
                SET title = ?, slug = ?, description = ?, cover_label = ?, icon = ?, theme_tint = ?, visibility_mode = ?,
                    directory_summary = ?, intro_markdown = ?, token_hint = ?, cover_url = ?, status = ?,
                    updated_by = ?, updated_at = CURRENT_TIMESTAMP
                WHERE id = ?
        `,
		params.Title,
		params.Slug,
		params.Description,
		params.CoverLabel,
		params.Icon,
		params.ThemeTint,
		params.VisibilityMode,
		params.DirectorySummary,
		params.IntroMarkdown,
		params.TokenHint,
		params.CoverURL,
		params.Status,
		params.UserID,
		id,
	)
	return err
}

func (r *KnowledgeSpaceRepository) Delete(id int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM knowledge_access_logs WHERE space_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM knowledge_access_pass_spaces WHERE space_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`
                DELETE FROM knowledge_access_passes
                WHERE scope_type != 'all_published'
                  AND id NOT IN (SELECT DISTINCT pass_id FROM knowledge_access_pass_spaces)
        `); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM knowledge_access_tokens WHERE space_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM knowledge_entries WHERE space_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM knowledge_spaces WHERE id = ?`, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *KnowledgeEntryRepository) List() ([]domain.KnowledgeEntry, error) {
	rows, err := r.db.Query(`
                SELECT e.id, e.space_id, s.slug, e.title, e.slug, e.section_name, e.sort_order,
                       e.estimated_read_minutes, e.public_summary, e.content_markdown, e.cover_url,
                       e.is_preview, e.status, strftime('%Y-%m-%d %H:%M', e.updated_at)
                FROM knowledge_entries e
                INNER JOIN knowledge_spaces s ON s.id = e.space_id
                ORDER BY e.space_id ASC, e.sort_order ASC, e.id ASC
        `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]domain.KnowledgeEntry, 0)
	for rows.Next() {
		record, scanErr := scanKnowledgeEntry(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		entries = append(entries, record)
	}

	return entries, rows.Err()
}

func (r *KnowledgeEntryRepository) ListBySpace(spaceID int64) ([]domain.KnowledgeEntry, error) {
	return r.listBySpace(spaceID, false)
}

func (r *KnowledgeEntryRepository) ListPublishedBySpace(spaceID int64) ([]domain.KnowledgeEntry, error) {
	return r.listBySpace(spaceID, true)
}

func (r *KnowledgeEntryRepository) listBySpace(spaceID int64, onlyPublished bool) ([]domain.KnowledgeEntry, error) {
	query := `
                SELECT e.id, e.space_id, s.slug, e.title, e.slug, e.section_name, e.sort_order,
                       e.estimated_read_minutes, e.public_summary, e.content_markdown, e.cover_url,
                       e.is_preview, e.status, strftime('%Y-%m-%d %H:%M', e.updated_at)
                FROM knowledge_entries e
                INNER JOIN knowledge_spaces s ON s.id = e.space_id
                WHERE e.space_id = ?
        `
	args := []any{spaceID}
	if onlyPublished {
		query += ` AND e.status = 'published'`
	}
	query += ` ORDER BY e.sort_order ASC, e.id ASC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]domain.KnowledgeEntry, 0)
	for rows.Next() {
		record, scanErr := scanKnowledgeEntry(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		entries = append(entries, record)
	}

	return entries, rows.Err()
}

func (r *KnowledgeEntryRepository) GetByID(id int64) (domain.KnowledgeEntry, error) {
	row := r.db.QueryRow(`
                SELECT e.id, e.space_id, s.slug, e.title, e.slug, e.section_name, e.sort_order,
                       e.estimated_read_minutes, e.public_summary, e.content_markdown, e.cover_url,
                       e.is_preview, e.status, strftime('%Y-%m-%d %H:%M', e.updated_at)
                FROM knowledge_entries e
                INNER JOIN knowledge_spaces s ON s.id = e.space_id
                WHERE e.id = ?
                LIMIT 1
        `, id)

	return scanKnowledgeEntry(row)
}

func (r *KnowledgeEntryRepository) GetBySpaceAndSlug(spaceID int64, slug string) (domain.KnowledgeEntry, error) {
	row := r.db.QueryRow(`
                SELECT e.id, e.space_id, s.slug, e.title, e.slug, e.section_name, e.sort_order,
                       e.estimated_read_minutes, e.public_summary, e.content_markdown, e.cover_url,
                       e.is_preview, e.status, strftime('%Y-%m-%d %H:%M', e.updated_at)
                FROM knowledge_entries e
                INNER JOIN knowledge_spaces s ON s.id = e.space_id
                WHERE e.space_id = ? AND e.slug = ?
                LIMIT 1
        `, spaceID, slug)

	return scanKnowledgeEntry(row)
}

func (r *KnowledgeEntryRepository) GetPublishedBySpaceAndSlug(spaceID int64, slug string) (domain.KnowledgeEntry, error) {
	row := r.db.QueryRow(`
                SELECT e.id, e.space_id, s.slug, e.title, e.slug, e.section_name, e.sort_order,
                       e.estimated_read_minutes, e.public_summary, e.content_markdown, e.cover_url,
                       e.is_preview, e.status, strftime('%Y-%m-%d %H:%M', e.updated_at)
                FROM knowledge_entries e
                INNER JOIN knowledge_spaces s ON s.id = e.space_id
                WHERE e.space_id = ? AND e.slug = ? AND e.status = 'published'
                LIMIT 1
        `, spaceID, slug)

	return scanKnowledgeEntry(row)
}

func (r *KnowledgeEntryRepository) Create(params KnowledgeEntryUpsertParams) (int64, error) {
	result, err := r.db.Exec(`
                INSERT INTO knowledge_entries (
                        space_id, title, slug, section_name, sort_order, estimated_read_minutes,
                        public_summary, content_markdown, cover_url, is_preview, status, created_by, updated_by
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `,
		params.SpaceID,
		params.Title,
		params.Slug,
		params.SectionName,
		params.SortOrder,
		params.EstimatedReadMinutes,
		params.PublicSummary,
		params.ContentMarkdown,
		params.CoverURL,
		boolToInt(params.IsPreview),
		params.Status,
		params.UserID,
		params.UserID,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *KnowledgeEntryRepository) Update(id int64, params KnowledgeEntryUpsertParams) error {
	_, err := r.db.Exec(`
                UPDATE knowledge_entries
                SET space_id = ?, title = ?, slug = ?, section_name = ?, sort_order = ?, estimated_read_minutes = ?,
                    public_summary = ?, content_markdown = ?, cover_url = ?, is_preview = ?, status = ?,
                    updated_by = ?, updated_at = CURRENT_TIMESTAMP
                WHERE id = ?
        `,
		params.SpaceID,
		params.Title,
		params.Slug,
		params.SectionName,
		params.SortOrder,
		params.EstimatedReadMinutes,
		params.PublicSummary,
		params.ContentMarkdown,
		params.CoverURL,
		boolToInt(params.IsPreview),
		params.Status,
		params.UserID,
		id,
	)
	return err
}

func (r *KnowledgeEntryRepository) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM knowledge_entries WHERE id = ?`, id)
	return err
}

func (r *KnowledgeAccessTokenRepository) List(spaceID *int64) ([]domain.KnowledgeAccessToken, error) {
	query := `
                SELECT p.id, p.name, p.access_level, p.scope_type,
                       COALESCE(GROUP_CONCAT(ps.space_id), ''),
                       p.is_active, strftime('%Y-%m-%d %H:%M', p.expires_at),
                       p.created_by, strftime('%Y-%m-%d %H:%M', p.created_at), strftime('%Y-%m-%d %H:%M', p.updated_at)
                FROM knowledge_access_passes p
                LEFT JOIN knowledge_access_pass_spaces ps ON ps.pass_id = p.id
        `
	args := make([]any, 0, 1)
	if spaceID != nil {
		query += ` WHERE p.scope_type = 'all_published' OR EXISTS (
                        SELECT 1 FROM knowledge_access_pass_spaces fps WHERE fps.pass_id = p.id AND fps.space_id = ?
                )`
		args = append(args, *spaceID)
	}
	query += ` GROUP BY p.id ORDER BY p.created_at DESC, p.id DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]domain.KnowledgeAccessToken, 0)
	for rows.Next() {
		record, scanErr := scanKnowledgeAccessToken(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		tokens = append(tokens, record)
	}

	return tokens, rows.Err()
}

func (r *KnowledgeAccessTokenRepository) GetByID(id int64) (domain.KnowledgeAccessToken, error) {
	row := r.db.QueryRow(`
                SELECT p.id, p.name, p.access_level, p.scope_type,
                       COALESCE(GROUP_CONCAT(ps.space_id), ''),
                       p.is_active, strftime('%Y-%m-%d %H:%M', p.expires_at),
                       p.created_by, strftime('%Y-%m-%d %H:%M', p.created_at), strftime('%Y-%m-%d %H:%M', p.updated_at)
                FROM knowledge_access_passes p
                LEFT JOIN knowledge_access_pass_spaces ps ON ps.pass_id = p.id
                WHERE p.id = ?
                GROUP BY p.id
                LIMIT 1
        `, id)

	return scanKnowledgeAccessToken(row)
}

func (r *KnowledgeAccessTokenRepository) GetByHash(tokenHash string) (KnowledgeAccessTokenRecord, error) {
	row := r.db.QueryRow(`
                SELECT p.id, p.name, p.access_level, p.scope_type,
                       COALESCE(GROUP_CONCAT(ps.space_id), ''),
                       p.token_hash, p.is_active, strftime('%Y-%m-%d %H:%M', p.expires_at),
                       p.created_by, strftime('%Y-%m-%d %H:%M', p.created_at), strftime('%Y-%m-%d %H:%M', p.updated_at)
                FROM knowledge_access_passes p
                LEFT JOIN knowledge_access_pass_spaces ps ON ps.pass_id = p.id
                WHERE p.token_hash = ?
                GROUP BY p.id
                LIMIT 1
        `, tokenHash)

	return scanKnowledgeAccessTokenRecord(row)
}

func (r *KnowledgeAccessTokenRepository) ListActiveBySpace(spaceID int64) ([]KnowledgeAccessTokenRecord, error) {
	rows, err := r.db.Query(`
                SELECT p.id, p.name, p.access_level, p.scope_type,
                       COALESCE(GROUP_CONCAT(ps.space_id), ''),
                       p.token_hash, p.is_active, strftime('%Y-%m-%d %H:%M', p.expires_at),
                       p.created_by, strftime('%Y-%m-%d %H:%M', p.created_at), strftime('%Y-%m-%d %H:%M', p.updated_at)
                FROM knowledge_access_passes p
                LEFT JOIN knowledge_access_pass_spaces ps ON ps.pass_id = p.id
                WHERE p.is_active = 1
                  AND (
                        p.scope_type = 'all_published'
                        OR EXISTS (
                                SELECT 1 FROM knowledge_access_pass_spaces fps
                                WHERE fps.pass_id = p.id AND fps.space_id = ?
                        )
                  )
                GROUP BY p.id
                ORDER BY p.created_at DESC, p.id DESC
        `, spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]KnowledgeAccessTokenRecord, 0)
	for rows.Next() {
		record, scanErr := scanKnowledgeAccessTokenRecord(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *KnowledgeAccessTokenRepository) Create(params KnowledgeAccessTokenCreateParams) (int64, error) {
	var expiresAt any
	if params.ExpiresAt != "" {
		expiresAt = params.ExpiresAt
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
                INSERT INTO knowledge_access_passes (name, token_hash, access_level, scope_type, is_active, expires_at, created_by)
                VALUES (?, ?, ?, ?, 1, ?, ?)
        `, params.Name, params.TokenHash, params.AccessLevel, params.ScopeType, expiresAt, params.CreatedBy)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, spaceID := range params.SpaceIDs {
		if _, err := tx.Exec(`
                        INSERT INTO knowledge_access_pass_spaces (pass_id, space_id)
                        VALUES (?, ?)
                `, id, spaceID); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *KnowledgeAccessTokenRepository) UpdateStatus(id int64, isActive bool) error {
	_, err := r.db.Exec(`
                UPDATE knowledge_access_passes
                SET is_active = ?, updated_at = CURRENT_TIMESTAMP
                WHERE id = ?
        `, boolToInt(isActive), id)
	return err
}

func (r *KnowledgeAccessLogRepository) Create(params KnowledgeAccessLogCreateParams) error {
	_, err := r.db.Exec(`
                INSERT INTO knowledge_access_logs (space_id, entry_id, token_id, action, request_ip, user_agent)
                VALUES (?, ?, ?, ?, ?, ?)
        `,
		params.SpaceID,
		params.EntryID,
		params.TokenID,
		params.Action,
		params.RequestIP,
		params.UserAgent,
	)
	return err
}

type knowledgeSpaceScanner interface {
	Scan(dest ...any) error
}

func scanKnowledgeSpace(scanner knowledgeSpaceScanner) (domain.KnowledgeSpace, error) {
	var record domain.KnowledgeSpace
	var updatedAt sql.NullString
	err := scanner.Scan(
		&record.ID,
		&record.Title,
		&record.Slug,
		&record.Description,
		&record.CoverLabel,
		&record.Icon,
		&record.ThemeTint,
		&record.VisibilityMode,
		&record.DirectorySummary,
		&record.IntroMarkdown,
		&record.TokenHint,
		&record.CoverURL,
		&record.Status,
		&record.EntryCount,
		&record.SectionCount,
		&updatedAt,
	)
	if err != nil {
		return domain.KnowledgeSpace{}, err
	}

	record.UpdatedAt = updatedAt.String
	return record, nil
}

type knowledgeEntryScanner interface {
	Scan(dest ...any) error
}

func scanKnowledgeEntry(scanner knowledgeEntryScanner) (domain.KnowledgeEntry, error) {
	var record domain.KnowledgeEntry
	var isPreview int
	var updatedAt sql.NullString

	err := scanner.Scan(
		&record.ID,
		&record.SpaceID,
		&record.SpaceSlug,
		&record.Title,
		&record.Slug,
		&record.SectionName,
		&record.SortOrder,
		&record.EstimatedReadMinutes,
		&record.PublicSummary,
		&record.ContentMarkdown,
		&record.CoverURL,
		&isPreview,
		&record.Status,
		&updatedAt,
	)
	if err != nil {
		return domain.KnowledgeEntry{}, err
	}

	record.IsPreview = isPreview == 1
	record.UpdatedAt = updatedAt.String
	return record, nil
}

type knowledgeAccessTokenScanner interface {
	Scan(dest ...any) error
}

func scanKnowledgeAccessToken(scanner knowledgeAccessTokenScanner) (domain.KnowledgeAccessToken, error) {
	var record domain.KnowledgeAccessToken
	var scopeType string
	var accessLevel string
	var spaceIDs string
	var isActive int
	var expiresAt sql.NullString
	var createdAt sql.NullString
	var updatedAt sql.NullString

	err := scanner.Scan(
		&record.ID,
		&record.Name,
		&accessLevel,
		&scopeType,
		&spaceIDs,
		&isActive,
		&expiresAt,
		&record.CreatedBy,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.KnowledgeAccessToken{}, err
	}

	record.AccessLevel = accessLevel
	record.ScopeType = scopeType
	record.SpaceIDs = parseKnowledgeSpaceIDs(spaceIDs)
	record.IsActive = isActive == 1
	record.ExpiresAt = expiresAt.String
	record.CreatedAt = createdAt.String
	record.UpdatedAt = updatedAt.String
	return record, nil
}

func scanKnowledgeAccessTokenRecord(scanner knowledgeAccessTokenScanner) (KnowledgeAccessTokenRecord, error) {
	var record KnowledgeAccessTokenRecord
	var scopeType string
	var accessLevel string
	var spaceIDs string
	var isActive int
	var expiresAt sql.NullString
	var createdAt sql.NullString
	var updatedAt sql.NullString

	err := scanner.Scan(
		&record.ID,
		&record.Name,
		&accessLevel,
		&scopeType,
		&spaceIDs,
		&record.TokenHash,
		&isActive,
		&expiresAt,
		&record.CreatedBy,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return KnowledgeAccessTokenRecord{}, err
	}

	record.AccessLevel = accessLevel
	record.ScopeType = scopeType
	record.SpaceIDs = parseKnowledgeSpaceIDs(spaceIDs)
	record.IsActive = isActive == 1
	record.ExpiresAt = expiresAt.String
	record.CreatedAt = createdAt.String
	record.UpdatedAt = updatedAt.String
	return record, nil
}

func parseKnowledgeSpaceIDs(raw string) []int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []int64{}
	}

	parts := strings.Split(raw, ",")
	result := make([]int64, 0, len(parts))
	seen := make(map[int64]struct{}, len(parts))
	for _, part := range parts {
		value, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || value <= 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
