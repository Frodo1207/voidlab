package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const builderNetworkProjectMigrationVersion = 1

const emptyNetworkProjectPlanJSON = `{"milestones":[],"sections":[]}`

type legacyProjectOpportunity struct {
	id              int64
	profileID       int64
	slug            string
	title           string
	summary         string
	status          string
	planJSON        string
	teamIntro       string
	teamMembersJSON string
	createdAt       string
	updatedAt       string
}

type legacyProjectMember struct {
	ProfileID    int64  `json:"profile_id"`
	Role         string `json:"role"`
	Contribution string `json:"contribution"`
}

func migrateBuilderNetworkProjects(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS builder_network_schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("create BuilderNetwork migration table: %w", err)
	}

	var applied int
	err := db.QueryRow(`
		SELECT version FROM builder_network_schema_migrations WHERE version = ?
	`, builderNetworkProjectMigrationVersion).Scan(&applied)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check BuilderNetwork migration version: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin Project migration: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS network_projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_profile_id INTEGER NOT NULL REFERENCES network_profiles(id) ON DELETE RESTRICT,
			slug TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			summary TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'draft'
				CHECK(status IN ('draft', 'published', 'archived')),
			project_plan_json TEXT NOT NULL DEFAULT '{"milestones":[],"sections":[]}',
			team_intro TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_network_projects_status_created
			ON network_projects(status, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_network_projects_owner_status
			ON network_projects(owner_profile_id, status);

		CREATE TABLE IF NOT EXISTS network_project_members (
			project_id INTEGER NOT NULL REFERENCES network_projects(id) ON DELETE CASCADE,
			profile_id INTEGER NOT NULL REFERENCES network_profiles(id) ON DELETE CASCADE,
			role TEXT NOT NULL,
			contribution TEXT NOT NULL DEFAULT '',
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(project_id, profile_id)
		);
		CREATE INDEX IF NOT EXISTS idx_network_project_members_profile
			ON network_project_members(profile_id, project_id);
	`); err != nil {
		return fmt.Errorf("create Project tables: %w", err)
	}

	hasProjectID, err := hasNetworkOpportunityProjectID(tx)
	if err != nil {
		return fmt.Errorf("inspect Opportunity project_id: %w", err)
	}
	if !hasProjectID {
		if _, err := tx.Exec(`
			ALTER TABLE network_opportunities
			ADD COLUMN project_id INTEGER REFERENCES network_projects(id) ON DELETE RESTRICT
		`); err != nil {
			return fmt.Errorf("add Opportunity project_id: %w", err)
		}
	}

	if err := backfillLegacyNetworkProjects(tx); err != nil {
		return err
	}
	if err := validateNetworkProjectMigration(tx); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO builder_network_schema_migrations(version, name)
		VALUES (?, 'split-projects-from-opportunities')
	`, builderNetworkProjectMigrationVersion); err != nil {
		return fmt.Errorf("record Project migration: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit Project migration: %w", err)
	}
	return nil
}

func hasNetworkOpportunityProjectID(tx *sql.Tx) (bool, error) {
	rows, err := tx.Query(`PRAGMA table_info(network_opportunities)`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return false, err
		}
		if name == "project_id" {
			return true, nil
		}
	}
	return false, rows.Err()
}

func backfillLegacyNetworkProjects(tx *sql.Tx) error {
	rows, err := tx.Query(`
		SELECT id, profile_id, slug, title, COALESCE(summary, ''), status,
		       COALESCE(project_plan_json, ''), COALESCE(team_intro, ''),
		       COALESCE(team_members_json, '[]'),
		       COALESCE(created_at, ''), COALESCE(updated_at, '')
		FROM network_opportunities
		WHERE project_id IS NULL
		ORDER BY id
	`)
	if err != nil {
		return fmt.Errorf("read legacy Opportunities: %w", err)
	}

	opportunities := make([]legacyProjectOpportunity, 0)
	for rows.Next() {
		var opportunity legacyProjectOpportunity
		if err := rows.Scan(
			&opportunity.id,
			&opportunity.profileID,
			&opportunity.slug,
			&opportunity.title,
			&opportunity.summary,
			&opportunity.status,
			&opportunity.planJSON,
			&opportunity.teamIntro,
			&opportunity.teamMembersJSON,
			&opportunity.createdAt,
			&opportunity.updatedAt,
		); err != nil {
			_ = rows.Close()
			return fmt.Errorf("scan legacy Opportunity: %w", err)
		}
		opportunities = append(opportunities, opportunity)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return fmt.Errorf("iterate legacy Opportunities: %w", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close legacy Opportunity rows: %w", err)
	}

	for _, opportunity := range opportunities {
		if err := migrateLegacyNetworkProject(tx, opportunity); err != nil {
			return err
		}
	}
	return nil
}

func migrateLegacyNetworkProject(tx *sql.Tx, opportunity legacyProjectOpportunity) error {
	planJSON, err := normalizeLegacyProjectPlan(opportunity.planJSON)
	if err != nil {
		return fmt.Errorf("Opportunity %d has invalid project_plan_json: %w", opportunity.id, err)
	}
	members, err := parseLegacyProjectMembers(opportunity.teamMembersJSON)
	if err != nil {
		return fmt.Errorf("Opportunity %d has invalid team_members_json: %w", opportunity.id, err)
	}
	if err := ensureNetworkProfileExists(tx, opportunity.profileID, opportunity.id, "owner"); err != nil {
		return err
	}

	projectStatus := "published"
	if opportunity.status == "draft" {
		projectStatus = "draft"
	}
	result, err := tx.Exec(`
		INSERT INTO network_projects (
			owner_profile_id, slug, title, summary, status, project_plan_json, team_intro,
			created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?,
			COALESCE(NULLIF(?, ''), CURRENT_TIMESTAMP),
			COALESCE(NULLIF(?, ''), CURRENT_TIMESTAMP))
	`, opportunity.profileID, opportunity.slug, opportunity.title, opportunity.summary,
		projectStatus, planJSON, opportunity.teamIntro, opportunity.createdAt, opportunity.updatedAt)
	if err != nil {
		return fmt.Errorf("create Project for Opportunity %d: %w", opportunity.id, err)
	}
	projectID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("read Project ID for Opportunity %d: %w", opportunity.id, err)
	}

	seen := make(map[int64]struct{}, len(members))
	for index, member := range members {
		if member.ProfileID <= 0 || strings.TrimSpace(member.Role) == "" {
			return fmt.Errorf("Opportunity %d has a team member without profile_id or role", opportunity.id)
		}
		if _, exists := seen[member.ProfileID]; exists {
			return fmt.Errorf("Opportunity %d has duplicate team member profile %d", opportunity.id, member.ProfileID)
		}
		seen[member.ProfileID] = struct{}{}
		if err := ensureNetworkProfileExists(tx, member.ProfileID, opportunity.id, "team member"); err != nil {
			return err
		}
		if _, err := tx.Exec(`
			INSERT INTO network_project_members(project_id, profile_id, role, contribution, sort_order)
			VALUES (?, ?, ?, ?, ?)
		`, projectID, member.ProfileID, strings.TrimSpace(member.Role),
			strings.TrimSpace(member.Contribution), index); err != nil {
			return fmt.Errorf("migrate team member %d for Opportunity %d: %w", member.ProfileID, opportunity.id, err)
		}
	}

	result, err = tx.Exec(`
		UPDATE network_opportunities SET project_id = ? WHERE id = ? AND project_id IS NULL
	`, projectID, opportunity.id)
	if err != nil {
		return fmt.Errorf("link Opportunity %d to Project: %w", opportunity.id, err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check Opportunity %d link: %w", opportunity.id, err)
	}
	if updated != 1 {
		return fmt.Errorf("Opportunity %d was not linked to its Project", opportunity.id)
	}
	return nil
}

func normalizeLegacyProjectPlan(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return emptyNetworkProjectPlanJSON, nil
	}

	var plan map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return "", err
	}
	if plan == nil {
		return "", errors.New("expected a JSON object")
	}
	for _, key := range []string{"milestones", "sections"} {
		value, ok := plan[key]
		if !ok || len(value) == 0 || value[0] != '[' {
			return "", fmt.Errorf("%s must be an array", key)
		}
	}
	return raw, nil
}

func parseLegacyProjectMembers(raw string) ([]legacyProjectMember, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return nil, nil
	}
	var members []legacyProjectMember
	if err := json.Unmarshal([]byte(raw), &members); err != nil {
		return nil, err
	}
	if members == nil {
		return nil, nil
	}
	return members, nil
}

func ensureNetworkProfileExists(tx *sql.Tx, profileID, opportunityID int64, relationship string) error {
	var exists int
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM network_profiles WHERE id = ?)`, profileID).Scan(&exists); err != nil {
		return fmt.Errorf("check %s profile for Opportunity %d: %w", relationship, opportunityID, err)
	}
	if exists != 1 {
		return fmt.Errorf("Opportunity %d references missing %s Profile %d", opportunityID, relationship, profileID)
	}
	return nil
}

func validateNetworkProjectMigration(tx *sql.Tx) error {
	var unlinked int
	if err := tx.QueryRow(`
		SELECT COUNT(*) FROM network_opportunities WHERE project_id IS NULL
	`).Scan(&unlinked); err != nil {
		return fmt.Errorf("validate Opportunity Project links: %w", err)
	}
	if unlinked != 0 {
		return fmt.Errorf("Project migration left %d Opportunities without a Project", unlinked)
	}

	for _, table := range []string{"network_projects", "network_project_members", "network_opportunities"} {
		rows, err := tx.Query(`PRAGMA foreign_key_check(` + table + `)`)
		if err != nil {
			return fmt.Errorf("check %s foreign keys: %w", table, err)
		}
		if rows.Next() {
			var childTable string
			var rowID sql.NullInt64
			var parentTable string
			var foreignKeyID int
			scanErr := rows.Scan(&childTable, &rowID, &parentTable, &foreignKeyID)
			_ = rows.Close()
			if scanErr != nil {
				return fmt.Errorf("read %s foreign key violation: %w", table, scanErr)
			}
			return fmt.Errorf("%s has a foreign key violation at row %v referencing %s",
				childTable, rowID, parentTable)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return fmt.Errorf("read %s foreign key check: %w", table, err)
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("close %s foreign key rows: %w", table, err)
		}
	}
	return nil
}
