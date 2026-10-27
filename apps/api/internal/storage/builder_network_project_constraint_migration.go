package storage

import (
	"database/sql"
	"errors"
	"fmt"
)

const builderNetworkProjectConstraintMigrationVersion = 2

func migrateBuilderNetworkOpportunityProjectConstraint(db *sql.DB) error {
	var applied int
	err := db.QueryRow(`
		SELECT version FROM builder_network_schema_migrations WHERE version = ?
	`, builderNetworkProjectConstraintMigrationVersion).Scan(&applied)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check BuilderNetwork migration version %d: %w",
			builderNetworkProjectConstraintMigrationVersion, err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin Opportunity Project constraint migration: %w", err)
	}
	defer tx.Rollback()

	var unlinked int
	if err := tx.QueryRow(`
		SELECT COUNT(*) FROM network_opportunities WHERE project_id IS NULL
	`).Scan(&unlinked); err != nil {
		return fmt.Errorf("check unlinked Opportunities: %w", err)
	}
	if unlinked != 0 {
		return fmt.Errorf("cannot enforce project_id NOT NULL: %d Opportunities are unlinked", unlinked)
	}

	if _, err := tx.Exec(`
		CREATE TABLE network_opportunities_v2 (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			profile_id INTEGER NOT NULL REFERENCES network_profiles(id) ON DELETE CASCADE,
			slug TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			summary TEXT NOT NULL DEFAULT '',
			opportunity_type TEXT NOT NULL,
			looking_for_roles_json TEXT NOT NULL DEFAULT '[]',
			required_skills_json TEXT NOT NULL DEFAULT '[]',
			engagement_mode TEXT NOT NULL DEFAULT '',
			budget_hint TEXT NOT NULL DEFAULT '',
			duration_hint TEXT NOT NULL DEFAULT '',
			city_scope TEXT NOT NULL DEFAULT '',
			remote_allowed INTEGER NOT NULL DEFAULT 1,
			status TEXT NOT NULL DEFAULT 'open',
			expires_at DATETIME,
			project_plan_json TEXT NOT NULL DEFAULT '',
			team_intro TEXT NOT NULL DEFAULT '',
			team_members_json TEXT NOT NULL DEFAULT '[]',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			project_id INTEGER NOT NULL REFERENCES network_projects(id) ON DELETE RESTRICT
		)
	`); err != nil {
		return fmt.Errorf("create constrained Opportunity table: %w", err)
	}

	if _, err := tx.Exec(`
		INSERT INTO network_opportunities_v2 (
			id, profile_id, slug, title, summary, opportunity_type,
			looking_for_roles_json, required_skills_json, engagement_mode,
			budget_hint, duration_hint, city_scope, remote_allowed, status,
			expires_at, project_plan_json, team_intro, team_members_json,
			created_at, updated_at, project_id
		)
		SELECT
			id, profile_id, slug, title, summary, opportunity_type,
			looking_for_roles_json, required_skills_json, engagement_mode,
			budget_hint, duration_hint, city_scope, remote_allowed, status,
			expires_at, project_plan_json, team_intro, team_members_json,
			created_at, updated_at, project_id
		FROM network_opportunities
		ORDER BY id
	`); err != nil {
		return fmt.Errorf("copy Opportunities into constrained table: %w", err)
	}

	if _, err := tx.Exec(`DROP TABLE network_opportunities`); err != nil {
		return fmt.Errorf("replace old Opportunity table: %w", err)
	}
	if _, err := tx.Exec(`ALTER TABLE network_opportunities_v2 RENAME TO network_opportunities`); err != nil {
		return fmt.Errorf("rename constrained Opportunity table: %w", err)
	}
	if _, err := tx.Exec(`
		CREATE INDEX idx_network_opportunities_status
			ON network_opportunities(status, created_at DESC);
		CREATE INDEX idx_network_opportunities_profile
			ON network_opportunities(profile_id, status);
	`); err != nil {
		return fmt.Errorf("recreate Opportunity indexes: %w", err)
	}

	for _, table := range []string{"network_opportunities", "network_projects"} {
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

	if _, err := tx.Exec(`
		INSERT INTO builder_network_schema_migrations(version, name)
		VALUES (?, 'require-opportunity-project')
	`, builderNetworkProjectConstraintMigrationVersion); err != nil {
		return fmt.Errorf("record Opportunity Project constraint migration: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit Opportunity Project constraint migration: %w", err)
	}
	return nil
}
