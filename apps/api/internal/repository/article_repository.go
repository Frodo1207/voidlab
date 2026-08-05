package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"voidlabai/apps/api/internal/domain"
)

type ArticleUpsertParams struct {
	Title      string
	Slug       string
	Summary    string
	Category   string
	Audience   string
	Tags       []string
	CoverURL   string
	Content    string
	SourceName string
	SourceURL  string
	Featured   bool
	Status     string
	UserID     int64
}

func (r *ArticleRepository) List() ([]domain.Article, error) {
	rows, err := r.db.Query(`
		SELECT id, title, slug, summary, category, audience, tags_json, cover_url,
		       content, source_name, source_url, featured, status,
		       strftime('%Y-%m-%d %H:%M', updated_at)
		FROM articles
		ORDER BY updated_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]domain.Article, 0)
	for rows.Next() {
		record, scanErr := scanArticle(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		articles = append(articles, record)
	}

	return articles, rows.Err()
}

func (r *ArticleRepository) ListPublished() ([]domain.Article, error) {
	rows, err := r.db.Query(`
                SELECT id, title, slug, summary, category, audience, tags_json, cover_url,
                       content, source_name, source_url, featured, status,
                       strftime('%Y-%m-%d %H:%M', COALESCE(published_at, updated_at))
                FROM articles
                WHERE status = 'published'
                ORDER BY COALESCE(published_at, updated_at) DESC, id DESC
        `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]domain.Article, 0)
	for rows.Next() {
		record, scanErr := scanArticle(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		articles = append(articles, record)
	}

	return articles, rows.Err()
}

func (r *ArticleRepository) GetByID(id int64) (domain.Article, error) {
	row := r.db.QueryRow(`
		SELECT id, title, slug, summary, category, audience, tags_json, cover_url,
		       content, source_name, source_url, featured, status,
		       strftime('%Y-%m-%d %H:%M', updated_at)
		FROM articles
		WHERE id = ?
		LIMIT 1
	`, id)

	return scanArticle(row)
}

func (r *ArticleRepository) GetPublishedBySlug(slug string) (domain.Article, error) {
	row := r.db.QueryRow(`
                SELECT id, title, slug, summary, category, audience, tags_json, cover_url,
                       content, source_name, source_url, featured, status,
                       strftime('%Y-%m-%d %H:%M', COALESCE(published_at, updated_at))
                FROM articles
                WHERE slug = ? AND status = 'published'
                LIMIT 1
        `, slug)

	return scanArticle(row)
}

func (r *ArticleRepository) Create(params ArticleUpsertParams) (int64, error) {
	tagsJSON, err := json.Marshal(params.Tags)
	if err != nil {
		return 0, fmt.Errorf("marshal tags: %w", err)
	}

	result, err := r.db.Exec(`
		INSERT INTO articles (
			title, slug, summary, category, audience, tags_json, cover_url,
			content, source_name, source_url, featured, status, published_at,
			created_by, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CASE WHEN ? = 'published' THEN CURRENT_TIMESTAMP ELSE NULL END, ?, ?)
	`,
		params.Title,
		params.Slug,
		params.Summary,
		params.Category,
		params.Audience,
		string(tagsJSON),
		params.CoverURL,
		params.Content,
		params.SourceName,
		params.SourceURL,
		boolToInt(params.Featured),
		params.Status,
		params.Status,
		params.UserID,
		params.UserID,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *ArticleRepository) Update(id int64, params ArticleUpsertParams) error {
	tagsJSON, err := json.Marshal(params.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}

	_, err = r.db.Exec(`
		UPDATE articles
		SET title = ?, slug = ?, summary = ?, category = ?, audience = ?, tags_json = ?, cover_url = ?,
		    content = ?, source_name = ?, source_url = ?, featured = ?, status = ?,
		    published_at = CASE
		        WHEN ? = 'published' AND published_at IS NULL THEN CURRENT_TIMESTAMP
		        WHEN ? != 'published' THEN NULL
		        ELSE published_at
		    END,
		    updated_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`,
		params.Title,
		params.Slug,
		params.Summary,
		params.Category,
		params.Audience,
		string(tagsJSON),
		params.CoverURL,
		params.Content,
		params.SourceName,
		params.SourceURL,
		boolToInt(params.Featured),
		params.Status,
		params.Status,
		params.Status,
		params.UserID,
		id,
	)

	return err
}

func (r *ArticleRepository) UpdateStatus(id int64, status string, userID int64) error {
	_, err := r.db.Exec(`
		UPDATE articles
		SET status = ?,
		    published_at = CASE
		        WHEN ? = 'published' AND published_at IS NULL THEN CURRENT_TIMESTAMP
		        WHEN ? != 'published' THEN NULL
		        ELSE published_at
		    END,
		    updated_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`,
		status,
		status,
		status,
		userID,
		id,
	)

	return err
}

func (r *ArticleRepository) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM articles WHERE id = ?`, id)
	return err
}

func (r *ArticleRepository) CountAll() (int64, error) {
	return countTableRows(r.db, "articles")
}

func (r *ArticleRepository) CountByStatus(status string) (int64, error) {
	return countTableRowsByStatus(r.db, "articles", status)
}

type articleScanner interface {
	Scan(dest ...any) error
}

func scanArticle(scanner articleScanner) (domain.Article, error) {
	var article domain.Article
	var tagsJSON string
	var featured int
	var updatedAt sql.NullString

	err := scanner.Scan(
		&article.ID,
		&article.Title,
		&article.Slug,
		&article.Summary,
		&article.Category,
		&article.Audience,
		&tagsJSON,
		&article.CoverURL,
		&article.Content,
		&article.SourceName,
		&article.SourceURL,
		&featured,
		&article.Status,
		&updatedAt,
	)
	if err != nil {
		return domain.Article{}, err
	}

	if tagsJSON != "" {
		if unmarshalErr := json.Unmarshal([]byte(tagsJSON), &article.Tags); unmarshalErr != nil {
			return domain.Article{}, fmt.Errorf("unmarshal tags: %w", unmarshalErr)
		}
	}

	article.Featured = featured == 1
	article.UpdatedAt = updatedAt.String

	return article, nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
