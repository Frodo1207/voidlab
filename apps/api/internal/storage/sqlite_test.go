package storage

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
)

func TestOpenSQLiteMigratesLegacyOpportunityDetails(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "legacy.sqlite")
	legacyDB, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatalf("open legacy database: %v", err)
	}
	_, err = legacyDB.Exec(`
		CREATE TABLE network_accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE network_profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id INTEGER NOT NULL UNIQUE REFERENCES network_accounts(id) ON DELETE CASCADE,
			slug TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL,
			primary_role TEXT NOT NULL,
			headline TEXT NOT NULL DEFAULT '',
			bio TEXT NOT NULL DEFAULT '',
			city TEXT NOT NULL DEFAULT '',
			skills_json TEXT NOT NULL DEFAULT '[]',
			collaboration_modes_json TEXT NOT NULL DEFAULT '[]',
			availability TEXT NOT NULL DEFAULT 'open',
			available_from TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'published',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO network_accounts (id, email, password_hash)
		VALUES (1, 'legacy@example.test', 'hash');
		INSERT INTO network_profiles (id, account_id, slug, display_name, primary_role)
		VALUES (1, 1, 'legacy-builder', 'Legacy Builder', 'Engineer');
		CREATE TABLE network_opportunities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			profile_id INTEGER NOT NULL,
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
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO network_opportunities (profile_id, slug, title, opportunity_type)
		VALUES (1, 'legacy-project', 'Legacy project', 'collaboration');
	`)
	if err != nil {
		t.Fatalf("create legacy opportunity table: %v", err)
	}
	if err := legacyDB.Close(); err != nil {
		t.Fatalf("close legacy database: %v", err)
	}

	db, err := OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("open and migrate legacy database: %v", err)
	}
	defer db.Close()

	columns := make(map[string]bool)
	rows, err := db.Query(`PRAGMA table_info(network_opportunities)`)
	if err != nil {
		t.Fatalf("inspect migrated table: %v", err)
	}
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			_ = rows.Close()
			t.Fatalf("scan migrated column: %v", err)
		}
		columns[name] = true
	}
	if err := rows.Close(); err != nil {
		t.Fatalf("close schema rows: %v", err)
	}
	for _, name := range []string{"project_plan_json", "team_intro", "team_members_json", "project_id"} {
		if !columns[name] {
			t.Errorf("expected migrated column %q", name)
		}
	}
	if !networkOpportunityProjectIDNotNull(t, db) {
		t.Fatal("expected project_id to be NOT NULL after all migrations")
	}

	var title, teamIntro, teamMembers string
	var projectID sql.NullInt64
	if err := db.QueryRow(`
		SELECT title, team_intro, team_members_json, project_id
		FROM network_opportunities WHERE slug = 'legacy-project'
	`).Scan(&title, &teamIntro, &teamMembers, &projectID); err != nil {
		t.Fatalf("read legacy opportunity after migration: %v", err)
	}
	if title != "Legacy project" || teamIntro != "" || teamMembers != "[]" || !projectID.Valid {
		t.Fatalf("legacy record changed unexpectedly: title=%q team_intro=%q team_members=%q project_id=%v",
			title, teamIntro, teamMembers, projectID)
	}

	var projectCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM network_projects WHERE id = ?`, projectID.Int64).Scan(&projectCount); err != nil {
		t.Fatalf("count migrated Project: %v", err)
	}
	if projectCount != 1 {
		t.Fatalf("expected one migrated Project, got %d", projectCount)
	}
	if err := migrateBuilderNetworkProjects(db); err != nil {
		t.Fatalf("rerun Project migration: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM network_projects`).Scan(&projectCount); err != nil {
		t.Fatalf("count Projects after rerun: %v", err)
	}
	if projectCount != 1 {
		t.Fatalf("Project migration created duplicates: got %d Projects", projectCount)
	}
}

func TestMigrateBuilderNetworkProjectsBackfillsMembersAndPreservesRequestTarget(t *testing.T) {
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "project-migration.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	prepareNullableNetworkOpportunityProjectID(t, db)
	seedProjectMigrationProfiles(t, db, 1, 2, 3)

	planJSON := `{"milestones":[{"title":"MVP","owner_member_ids":[2]}],"sections":[{"type":"intro","title":"Scope","items":["v1"]}]}`
	membersJSON := `[
		{"profile_id":2,"display_name":"Old name","slug":"old-builder","role":"Product","contribution":"Discovery","sort_order":0},
		{"profile_id":3,"display_name":"Another old name","slug":"another-builder","role":"Engineer","contribution":"Implementation","sort_order":1}
	]`
	result, err := db.Exec(`
		INSERT INTO network_opportunities (
			profile_id, slug, title, summary, opportunity_type, status,
			project_plan_json, team_intro, team_members_json
		) VALUES (1, 'legacy-workflow', 'Workflow', 'Legacy summary', 'collaboration', 'open', ?, 'Legacy team', ?)
	`, planJSON, membersJSON)
	if err != nil {
		t.Fatalf("insert legacy Opportunity: %v", err)
	}
	opportunityID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read Opportunity ID: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO network_contact_requests (
			from_profile_id, target_type, target_id, target_profile_id, message
		) VALUES (2, 'opportunity', ?, 1, 'Existing application')
	`, opportunityID); err != nil {
		t.Fatalf("insert existing application: %v", err)
	}
	if _, err := db.Exec(`DELETE FROM builder_network_schema_migrations WHERE version = ?`,
		builderNetworkProjectMigrationVersion); err != nil {
		t.Fatalf("reset migration marker: %v", err)
	}

	if err := migrateBuilderNetworkProjects(db); err != nil {
		t.Fatalf("migrate legacy Project data: %v", err)
	}

	var projectID int64
	var ownerID int64
	var slug, title, summary, status, storedPlan, teamIntro string
	if err := db.QueryRow(`
		SELECT p.id, p.owner_profile_id, p.slug, p.title, p.summary, p.status,
		       p.project_plan_json, p.team_intro
		FROM network_projects p
		JOIN network_opportunities o ON o.project_id = p.id
		WHERE o.id = ?
	`, opportunityID).Scan(&projectID, &ownerID, &slug, &title, &summary, &status, &storedPlan, &teamIntro); err != nil {
		t.Fatalf("read migrated Project: %v", err)
	}
	if ownerID != 1 || slug != "legacy-workflow" || title != "Workflow" ||
		summary != "Legacy summary" || status != "published" ||
		storedPlan != planJSON || teamIntro != "Legacy team" {
		t.Fatalf("Project mapping mismatch: owner=%d slug=%q title=%q summary=%q status=%q plan=%q team=%q",
			ownerID, slug, title, summary, status, storedPlan, teamIntro)
	}

	memberRows, err := db.Query(`
		SELECT profile_id, role, contribution, sort_order
		FROM network_project_members WHERE project_id = ? ORDER BY sort_order
	`, projectID)
	if err != nil {
		t.Fatalf("read migrated team members: %v", err)
	}
	type memberResult struct {
		profileID    int64
		role         string
		contribution string
		sortOrder    int
	}
	var members []memberResult
	for memberRows.Next() {
		var member memberResult
		if err := memberRows.Scan(&member.profileID, &member.role, &member.contribution, &member.sortOrder); err != nil {
			_ = memberRows.Close()
			t.Fatalf("scan migrated team member: %v", err)
		}
		members = append(members, member)
	}
	if err := memberRows.Close(); err != nil {
		t.Fatalf("close migrated member rows: %v", err)
	}
	if len(members) != 2 || members[0] != (memberResult{2, "Product", "Discovery", 0}) ||
		members[1] != (memberResult{3, "Engineer", "Implementation", 1}) {
		t.Fatalf("migrated members do not match legacy order and content: %#v", members)
	}

	var storedOpportunityID, targetID int64
	var storedOpportunitySlug, targetType string
	if err := db.QueryRow(`
		SELECT o.id, o.slug, r.target_type, r.target_id
		FROM network_opportunities o
		JOIN network_contact_requests r ON r.target_id = o.id
		WHERE o.id = ?
	`, opportunityID).Scan(&storedOpportunityID, &storedOpportunitySlug, &targetType, &targetID); err != nil {
		t.Fatalf("read migrated Opportunity and application: %v", err)
	}
	if storedOpportunityID != opportunityID || storedOpportunitySlug != "legacy-workflow" ||
		targetType != "opportunity" || targetID != opportunityID {
		t.Fatalf("legacy application target changed: opportunity=%d slug=%q target_type=%q target_id=%d",
			storedOpportunityID, storedOpportunitySlug, targetType, targetID)
	}

	if err := migrateBuilderNetworkProjects(db); err != nil {
		t.Fatalf("rerun Project migration: %v", err)
	}
	var projectCount, memberCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM network_projects`).Scan(&projectCount); err != nil {
		t.Fatalf("count Projects after rerun: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM network_project_members`).Scan(&memberCount); err != nil {
		t.Fatalf("count members after rerun: %v", err)
	}
	if projectCount != 1 || memberCount != 2 {
		t.Fatalf("rerun created duplicate data: projects=%d members=%d", projectCount, memberCount)
	}
}

func TestMigrateBuilderNetworkProjectsRollsBackInvalidMember(t *testing.T) {
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "project-migration-rollback.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	prepareNullableNetworkOpportunityProjectID(t, db)
	seedProjectMigrationProfiles(t, db, 1)

	if _, err := db.Exec(`
		INSERT INTO network_opportunities (
			profile_id, slug, title, opportunity_type, status, team_members_json
		) VALUES (1, 'invalid-team', 'Invalid team', 'collaboration', 'open',
		          '[{"profile_id":999,"role":"Missing profile"}]')
	`); err != nil {
		t.Fatalf("insert invalid legacy Opportunity: %v", err)
	}
	if _, err := db.Exec(`DELETE FROM builder_network_schema_migrations WHERE version = ?`,
		builderNetworkProjectMigrationVersion); err != nil {
		t.Fatalf("reset migration marker: %v", err)
	}

	if err := migrateBuilderNetworkProjects(db); err == nil {
		t.Fatal("expected migration to reject a missing team Profile")
	}

	var projectCount, migrationCount, linkedCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM network_projects`).Scan(&projectCount); err != nil {
		t.Fatalf("count Projects after rollback: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM builder_network_schema_migrations WHERE version = ?`,
		builderNetworkProjectMigrationVersion).Scan(&migrationCount); err != nil {
		t.Fatalf("count migration marker after rollback: %v", err)
	}
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM network_opportunities WHERE project_id IS NOT NULL
	`).Scan(&linkedCount); err != nil {
		t.Fatalf("count linked Opportunities after rollback: %v", err)
	}
	if projectCount != 0 || migrationCount != 0 || linkedCount != 0 {
		t.Fatalf("failed migration left partial data: projects=%d marker=%d linked=%d",
			projectCount, migrationCount, linkedCount)
	}
}

func TestMigrateBuilderNetworkOpportunityProjectConstraintPreservesData(t *testing.T) {
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "project-constraint-migration.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	prepareNullableNetworkOpportunityProjectID(t, db)
	seedProjectMigrationProfiles(t, db, 1, 2)

	projectResult, err := db.Exec(`
		INSERT INTO network_projects(owner_profile_id, slug, title, summary, status)
		VALUES (1, 'constraint-project', 'Constraint project', 'Project summary', 'published')
	`)
	if err != nil {
		t.Fatalf("insert Project: %v", err)
	}
	projectID, err := projectResult.LastInsertId()
	if err != nil {
		t.Fatalf("read Project ID: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO network_opportunities (
			id, profile_id, slug, title, summary, opportunity_type,
			looking_for_roles_json, required_skills_json, engagement_mode,
			budget_hint, duration_hint, city_scope, remote_allowed, status,
			expires_at, project_plan_json, team_intro, team_members_json,
			created_at, updated_at, project_id
		) VALUES (
			42, 1, 'constraint-opportunity', 'Opportunity title', 'Opportunity summary', 'collaboration',
			'["Engineer"]', '["Go"]', 'remote', 'negotiable', '3 weeks', 'anywhere', 1, 'open',
			'2030-01-01', '{"milestones":[],"sections":[]}', 'Team intro', '[]',
			'2026-01-02 03:04:05', '2026-02-03 04:05:06', ?
		)
	`, projectID); err != nil {
		t.Fatalf("insert Opportunity: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO network_contact_requests (
			from_profile_id, target_type, target_id, target_profile_id, message
		) VALUES (2, 'opportunity', 42, 1, 'Preserve application')
	`); err != nil {
		t.Fatalf("insert application: %v", err)
	}
	if _, err := db.Exec(`DELETE FROM builder_network_schema_migrations WHERE version = ?`,
		builderNetworkProjectConstraintMigrationVersion); err != nil {
		t.Fatalf("reset constraint migration marker: %v", err)
	}

	if err := migrateBuilderNetworkOpportunityProjectConstraint(db); err != nil {
		t.Fatalf("apply Opportunity Project constraint migration: %v", err)
	}

	if !networkOpportunityProjectIDNotNull(t, db) {
		t.Fatal("expected project_id to be NOT NULL")
	}

	var profileID int64
	var slug, title, summary, rolesJSON, skillsJSON, engagementMode, budget, duration, city string
	var status, expiry, plan, teamIntro, membersJSON, createdAt, updatedAt string
	var remote int
	var storedProjectID int64
	if err := db.QueryRow(`
		SELECT profile_id, slug, title, summary, looking_for_roles_json, required_skills_json,
		       engagement_mode, budget_hint, duration_hint, city_scope, remote_allowed, status,
		       expires_at, project_plan_json, team_intro, team_members_json, created_at, updated_at, project_id
		FROM network_opportunities WHERE id = 42
	`).Scan(&profileID, &slug, &title, &summary, &rolesJSON, &skillsJSON, &engagementMode,
		&budget, &duration, &city, &remote, &status, &expiry, &plan, &teamIntro, &membersJSON,
		&createdAt, &updatedAt, &storedProjectID); err != nil {
		t.Fatalf("read migrated Opportunity: %v", err)
	}
	if profileID != 1 || slug != "constraint-opportunity" || title != "Opportunity title" ||
		summary != "Opportunity summary" || rolesJSON != `["Engineer"]` || skillsJSON != `["Go"]` ||
		engagementMode != "remote" || budget != "negotiable" || duration != "3 weeks" ||
		city != "anywhere" || remote != 1 || status != "open" || expiry != "2030-01-01T00:00:00Z" ||
		plan != `{"milestones":[],"sections":[]}` || teamIntro != "Team intro" || membersJSON != "[]" ||
		createdAt != "2026-01-02T03:04:05Z" || updatedAt != "2026-02-03T04:05:06Z" || storedProjectID != projectID {
		t.Fatalf("Opportunity data changed during table rebuild: %#v", []any{
			profileID, slug, title, summary, rolesJSON, skillsJSON, engagementMode, budget,
			duration, city, remote, status, expiry, plan, teamIntro, membersJSON,
			createdAt, updatedAt, storedProjectID,
		})
	}

	var targetType string
	var targetID int64
	if err := db.QueryRow(`
		SELECT target_type, target_id FROM network_contact_requests WHERE message = 'Preserve application'
	`).Scan(&targetType, &targetID); err != nil {
		t.Fatalf("read existing application target: %v", err)
	}
	if targetType != "opportunity" || targetID != 42 {
		t.Fatalf("application target changed: target_type=%q target_id=%d", targetType, targetID)
	}

	var retainedIndexCount int
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'index' AND name IN (
			'idx_network_opportunities_status', 'idx_network_opportunities_profile'
		)
	`).Scan(&retainedIndexCount); err != nil {
		t.Fatalf("count preserved Opportunity indexes: %v", err)
	}
	if retainedIndexCount != 2 {
		t.Fatalf("expected both Opportunity indexes after rebuild, got %d", retainedIndexCount)
	}
	if _, err := db.Exec(`DELETE FROM network_projects WHERE id = ?`, projectID); err == nil {
		t.Fatal("expected Project deletion to be restricted while Opportunities exist")
	}

	if _, err := db.Exec(`
		INSERT INTO network_opportunities(profile_id, slug, title, opportunity_type, project_id)
		VALUES (1, 'null-project-link', 'Invalid', 'collaboration', NULL)
	`); err == nil {
		t.Fatal("expected NULL project_id insertion to fail")
	}
	if err := migrateBuilderNetworkOpportunityProjectConstraint(db); err != nil {
		t.Fatalf("rerun constraint migration: %v", err)
	}
	var opportunityCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM network_opportunities`).Scan(&opportunityCount); err != nil {
		t.Fatalf("count Opportunities after rerun: %v", err)
	}
	if opportunityCount != 1 {
		t.Fatalf("constraint migration changed Opportunity count on rerun: %d", opportunityCount)
	}

	var projectForeignKeyFound bool
	foreignKeys, err := db.Query(`PRAGMA foreign_key_list(network_opportunities)`)
	if err != nil {
		t.Fatalf("inspect Opportunity foreign keys: %v", err)
	}
	for foreignKeys.Next() {
		var id, sequence int
		var table, from, to, onUpdate, onDelete, match string
		if err := foreignKeys.Scan(&id, &sequence, &table, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			_ = foreignKeys.Close()
			t.Fatalf("scan Opportunity foreign key: %v", err)
		}
		if table == "network_projects" && from == "project_id" && onDelete == "RESTRICT" {
			projectForeignKeyFound = true
		}
	}
	if err := foreignKeys.Close(); err != nil {
		t.Fatalf("close Opportunity foreign keys: %v", err)
	}
	if !projectForeignKeyFound {
		t.Fatal("expected project_id foreign key to network_projects with RESTRICT delete behavior")
	}
}

func TestMigrateBuilderNetworkOpportunityProjectConstraintRejectsUnlinkedRows(t *testing.T) {
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "project-constraint-rollback.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	prepareNullableNetworkOpportunityProjectID(t, db)
	seedProjectMigrationProfiles(t, db, 1)

	if _, err := db.Exec(`
		INSERT INTO network_opportunities(profile_id, slug, title, opportunity_type, project_id)
		VALUES (1, 'unlinked-project', 'Unlinked', 'collaboration', NULL)
	`); err != nil {
		t.Fatalf("insert legacy unlinked Opportunity: %v", err)
	}
	if _, err := db.Exec(`DELETE FROM builder_network_schema_migrations WHERE version = ?`,
		builderNetworkProjectConstraintMigrationVersion); err != nil {
		t.Fatalf("reset constraint migration marker: %v", err)
	}

	if err := migrateBuilderNetworkOpportunityProjectConstraint(db); err == nil {
		t.Fatal("expected migration to reject unlinked Opportunities")
	}
	var markerCount, opportunityCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM builder_network_schema_migrations WHERE version = ?`,
		builderNetworkProjectConstraintMigrationVersion).Scan(&markerCount); err != nil {
		t.Fatalf("count constraint migration marker: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM network_opportunities WHERE slug = 'unlinked-project'`).Scan(&opportunityCount); err != nil {
		t.Fatalf("count unlinked Opportunity after rollback: %v", err)
	}
	if markerCount != 0 || opportunityCount != 1 || networkOpportunityProjectIDNotNull(t, db) {
		t.Fatalf("failed migration changed schema or data: marker=%d opportunity=%d",
			markerCount, opportunityCount)
	}
}

func prepareNullableNetworkOpportunityProjectID(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
		ALTER TABLE network_opportunities RENAME TO network_opportunities_constrained;
		CREATE TABLE network_opportunities (
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
			project_id INTEGER REFERENCES network_projects(id) ON DELETE RESTRICT
		);
		INSERT INTO network_opportunities SELECT * FROM network_opportunities_constrained;
		DROP TABLE network_opportunities_constrained;
		CREATE INDEX idx_network_opportunities_status ON network_opportunities(status, created_at DESC);
		CREATE INDEX idx_network_opportunities_profile ON network_opportunities(profile_id, status);
		DELETE FROM builder_network_schema_migrations WHERE version = 2;
	`)
	if err != nil {
		t.Fatalf("prepare pre-constraint Opportunity schema: %v", err)
	}
}

func networkOpportunityProjectIDNotNull(t *testing.T, db *sql.DB) bool {
	t.Helper()
	rows, err := db.Query(`PRAGMA table_info(network_opportunities)`)
	if err != nil {
		t.Fatalf("inspect Opportunity columns: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("scan Opportunity column: %v", err)
		}
		if name == "project_id" {
			return notNull == 1
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate Opportunity columns: %v", err)
	}
	return false
}

func seedProjectMigrationProfiles(t *testing.T, db *sql.DB, profileIDs ...int64) {
	t.Helper()
	for _, profileID := range profileIDs {
		email := fmt.Sprintf("project-migration-%d@example.test", profileID)
		slug := fmt.Sprintf("project-migration-%d", profileID)
		if _, err := db.Exec(`
			INSERT INTO network_accounts(id, email, password_hash) VALUES (?, ?, 'hash')
		`, profileID, email); err != nil {
			t.Fatalf("insert network account %d: %v", profileID, err)
		}
		if _, err := db.Exec(`
			INSERT INTO network_profiles(id, account_id, slug, display_name, primary_role)
			VALUES (?, ?, ?, ?, 'Engineer')
		`, profileID, profileID, slug, fmt.Sprintf("Builder %d", profileID)); err != nil {
			t.Fatalf("insert network Profile %d: %v", profileID, err)
		}
	}
}
