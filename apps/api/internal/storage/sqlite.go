package storage

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

func OpenSQLite(databasePath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, fmt.Errorf("create sqlite dir: %w", err)
	}

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite 在轻量内容后台里够用，但需要先把并发策略收紧，
	// 否则 public 读请求和后台写请求并发时，容易出现瞬时 busy/locked。
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec(`
                PRAGMA journal_mode = WAL;
                PRAGMA busy_timeout = 5000;
                PRAGMA synchronous = NORMAL;
                PRAGMA foreign_keys = ON;
        `); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("configure sqlite pragmas: %w", err)
	}

	if _, err := db.Exec(`
                CREATE TABLE IF NOT EXISTS users (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        username TEXT NOT NULL UNIQUE,
                        password_hash TEXT NOT NULL,
                        role TEXT NOT NULL DEFAULT 'editor',
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS media_assets (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        object_key TEXT NOT NULL,
                        object_url TEXT NOT NULL,
                        file_name TEXT NOT NULL,
                        content_type TEXT NOT NULL,
                        file_size INTEGER NOT NULL DEFAULT 0,
                        uploaded_by INTEGER,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS articles (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        title TEXT NOT NULL,
                        slug TEXT NOT NULL UNIQUE,
                        summary TEXT NOT NULL DEFAULT '',
                        category TEXT NOT NULL DEFAULT '',
                        audience TEXT NOT NULL DEFAULT '',
                        tags_json TEXT NOT NULL DEFAULT '[]',
                        cover_url TEXT NOT NULL DEFAULT '',
                        cover_media_id INTEGER,
                        content TEXT NOT NULL DEFAULT '',
                        source_name TEXT NOT NULL DEFAULT '',
                        source_url TEXT NOT NULL DEFAULT '',
                        featured INTEGER NOT NULL DEFAULT 0,
                        status TEXT NOT NULL DEFAULT 'draft',
                        published_at DATETIME,
                        created_by INTEGER,
                        updated_by INTEGER,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS events (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        title TEXT NOT NULL,
                        slug TEXT NOT NULL UNIQUE,
                        event_time DATETIME,
                        location TEXT NOT NULL DEFAULT '',
                        city TEXT NOT NULL DEFAULT '',
                        event_type TEXT NOT NULL DEFAULT '',
                        status TEXT NOT NULL DEFAULT 'draft',
                        summary TEXT NOT NULL DEFAULT '',
                        cover_url TEXT NOT NULL DEFAULT '',
                        cover_media_id INTEGER,
                        content TEXT NOT NULL DEFAULT '',
                        signup_mode TEXT NOT NULL DEFAULT 'internal',
                        signup_enabled INTEGER NOT NULL DEFAULT 1,
                        signup_starts_at DATETIME,
                        signup_deadline DATETIME,
                        capacity INTEGER NOT NULL DEFAULT 0,
                        allow_signup_during_live INTEGER NOT NULL DEFAULT 0,
                        external_signup_url TEXT NOT NULL DEFAULT '',
                        signup_button_label TEXT NOT NULL DEFAULT '',
                        signup_success_message TEXT NOT NULL DEFAULT '',
                        signup_closed_reason TEXT NOT NULL DEFAULT '',
                        created_by INTEGER,
                        updated_by INTEGER,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS builders (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        name TEXT NOT NULL,
                        slug TEXT NOT NULL UNIQUE,
                        title TEXT NOT NULL DEFAULT '',
                        city TEXT NOT NULL DEFAULT '',
                        role TEXT NOT NULL DEFAULT '',
                        intro TEXT NOT NULL DEFAULT '',
                        story TEXT NOT NULL DEFAULT '',
                        expertise_json TEXT NOT NULL DEFAULT '[]',
                        focus_areas_json TEXT NOT NULL DEFAULT '[]',
                        collaboration_modes_json TEXT NOT NULL DEFAULT '[]',
                        availability_note TEXT NOT NULL DEFAULT '',
                        open_for TEXT NOT NULL DEFAULT '',
                        contactable INTEGER NOT NULL DEFAULT 0,
                        featured INTEGER NOT NULL DEFAULT 0,
                        cover_url TEXT NOT NULL DEFAULT '',
                        cover_media_id INTEGER,
                        status TEXT NOT NULL DEFAULT 'draft',
                        created_by INTEGER,
                        updated_by INTEGER,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS leads (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        source_type TEXT NOT NULL,
                        source_id INTEGER,
                        name TEXT NOT NULL,
                        contact TEXT NOT NULL,
                        message TEXT NOT NULL DEFAULT '',
                        status TEXT NOT NULL DEFAULT 'new',
                        notes TEXT NOT NULL DEFAULT '',
                        dedupe_key TEXT NOT NULL DEFAULT '',
                        owner_id INTEGER,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS lead_logs (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        lead_id INTEGER NOT NULL,
                        action TEXT NOT NULL DEFAULT 'note',
                        content TEXT NOT NULL DEFAULT '',
                        created_by INTEGER NOT NULL,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS site_configs (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        config_key TEXT NOT NULL UNIQUE,
                        config_value_json TEXT NOT NULL DEFAULT '{}',
                        updated_by INTEGER NOT NULL DEFAULT 1,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS audit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			actor_type TEXT NOT NULL DEFAULT 'human',
			actor_id INTEGER NOT NULL,
			actor_username TEXT NOT NULL DEFAULT '',
			actor_role TEXT NOT NULL DEFAULT '',
			agent_token_id INTEGER,
			action TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id INTEGER,
			entity_label TEXT NOT NULL DEFAULT '',
			detail_json TEXT NOT NULL DEFAULT '{}',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS agent_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			scopes_json TEXT NOT NULL DEFAULT '[]',
			is_active INTEGER NOT NULL DEFAULT 1,
			last_used_at DATETIME,
			created_by INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS knowledge_spaces (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        title TEXT NOT NULL,
                        slug TEXT NOT NULL UNIQUE,
                        description TEXT NOT NULL DEFAULT '',
                        cover_label TEXT NOT NULL DEFAULT '',
                        icon TEXT NOT NULL DEFAULT '',
                        theme_tint TEXT NOT NULL DEFAULT '',
                        visibility_mode TEXT NOT NULL DEFAULT 'directory_only',
                        directory_summary TEXT NOT NULL DEFAULT '',
                        intro_markdown TEXT NOT NULL DEFAULT '',
                        token_hint TEXT NOT NULL DEFAULT '',
                        cover_url TEXT NOT NULL DEFAULT '',
                        cover_media_id INTEGER,
                        status TEXT NOT NULL DEFAULT 'draft',
                        created_by INTEGER,
                        updated_by INTEGER,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS knowledge_entries (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        space_id INTEGER NOT NULL,
                        title TEXT NOT NULL,
                        slug TEXT NOT NULL,
                        section_name TEXT NOT NULL DEFAULT '',
                        sort_order INTEGER NOT NULL DEFAULT 0,
                        estimated_read_minutes INTEGER NOT NULL DEFAULT 0,
                        public_summary TEXT NOT NULL DEFAULT '',
                        content_markdown TEXT NOT NULL DEFAULT '',
                        cover_url TEXT NOT NULL DEFAULT '',
                        cover_media_id INTEGER,
                        is_preview INTEGER NOT NULL DEFAULT 0,
                        status TEXT NOT NULL DEFAULT 'draft',
                        created_by INTEGER,
                        updated_by INTEGER,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        UNIQUE(space_id, slug)
                );

                CREATE TABLE IF NOT EXISTS knowledge_access_tokens (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        space_id INTEGER NOT NULL,
                        name TEXT NOT NULL,
                        token_hash TEXT NOT NULL UNIQUE,
                        is_active INTEGER NOT NULL DEFAULT 1,
                        expires_at DATETIME,
                        created_by INTEGER NOT NULL DEFAULT 1,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS knowledge_access_logs (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        space_id INTEGER NOT NULL,
                        entry_id INTEGER,
                        token_id INTEGER,
                        action TEXT NOT NULL DEFAULT 'verify',
                        request_ip TEXT NOT NULL DEFAULT '',
                        user_agent TEXT NOT NULL DEFAULT '',
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS knowledge_access_passes (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        legacy_token_id INTEGER UNIQUE,
                        name TEXT NOT NULL,
                        token_hash TEXT NOT NULL UNIQUE,
                        access_level TEXT NOT NULL DEFAULT 'basic',
                        scope_type TEXT NOT NULL DEFAULT 'single_space',
                        is_active INTEGER NOT NULL DEFAULT 1,
                        expires_at DATETIME,
                        created_by INTEGER NOT NULL DEFAULT 1,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS knowledge_access_pass_spaces (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        pass_id INTEGER NOT NULL,
                        space_id INTEGER NOT NULL,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        UNIQUE(pass_id, space_id)
                );

                CREATE TABLE IF NOT EXISTS knowledge_assets (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        space_id INTEGER NOT NULL,
                        media_asset_id INTEGER NOT NULL UNIQUE,
                        created_by INTEGER NOT NULL DEFAULT 1,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );
        `); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate sqlite: %w", err)
	}

	if err := migrateUserSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate user schema: %w", err)
	}

	if err := migrateMediaAssetSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate media asset schema: %w", err)
	}

	if err := migrateAuditLogSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate audit schema: %w", err)
	}

	if err := migrateEventSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate event schema: %w", err)
	}

	if err := migrateLeadSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate lead schema: %w", err)
	}

	if err := migrateKnowledgeAccessPassSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate knowledge access pass schema: %w", err)
	}

	if err := seedDefaultUsers(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("seed default users: %w", err)
	}

	if err := seedSiteConfigs(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("seed site configs: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return db, nil
}

func seedDefaultUsers(db *sql.DB) error {
	defaultUsers := []struct {
		username string
		password string
		role     string
	}{
		{username: "admin", password: "admin", role: "admin"},
		{username: "editor", password: "editor", role: "editor"},
		{username: "ops", password: "ops", role: "ops"},
	}

	for _, user := range defaultUsers {
		var count int
		if err := db.QueryRow(`SELECT COUNT(1) FROM users WHERE username = ?`, user.username).Scan(&count); err != nil {
			return err
		}

		if count > 0 {
			continue
		}

		hash := sha256.Sum256([]byte(user.password))
		if _, err := db.Exec(
			`INSERT INTO users (username, password_hash, role, is_active) VALUES (?, ?, ?, 1)`,
			user.username,
			hex.EncodeToString(hash[:]),
			user.role,
		); err != nil {
			return err
		}
	}

	return nil
}

func migrateUserSchema(db *sql.DB) error {
	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN is_active INTEGER NOT NULL DEFAULT 1`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return err
		}
	}

	return nil
}

func migrateMediaAssetSchema(db *sql.DB) error {
	if _, err := db.Exec(`ALTER TABLE media_assets ADD COLUMN file_size INTEGER NOT NULL DEFAULT 0`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return err
		}
	}

	if _, err := db.Exec(`ALTER TABLE media_assets ADD COLUMN uploaded_by INTEGER`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return err
		}
	}

	return nil
}

func migrateAuditLogSchema(db *sql.DB) error {
	if _, err := db.Exec(`ALTER TABLE audit_logs ADD COLUMN actor_type TEXT NOT NULL DEFAULT 'human'`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return err
		}
	}

	if _, err := db.Exec(`ALTER TABLE audit_logs ADD COLUMN agent_token_id INTEGER`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return err
		}
	}

	return nil
}

func migrateEventSchema(db *sql.DB) error {
	statements := []string{
		`ALTER TABLE events ADD COLUMN signup_mode TEXT NOT NULL DEFAULT 'internal'`,
		`ALTER TABLE events ADD COLUMN signup_enabled INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE events ADD COLUMN signup_starts_at DATETIME`,
		`ALTER TABLE events ADD COLUMN signup_deadline DATETIME`,
		`ALTER TABLE events ADD COLUMN capacity INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE events ADD COLUMN allow_signup_during_live INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE events ADD COLUMN external_signup_url TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE events ADD COLUMN signup_button_label TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE events ADD COLUMN signup_success_message TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE events ADD COLUMN signup_closed_reason TEXT NOT NULL DEFAULT ''`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
				return err
			}
		}
	}

	return nil
}

func migrateLeadSchema(db *sql.DB) error {
	if _, err := db.Exec(`ALTER TABLE leads ADD COLUMN dedupe_key TEXT NOT NULL DEFAULT ''`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return err
		}
	}

	if _, err := db.Exec(`
                UPDATE leads
                SET dedupe_key = CASE
                        WHEN source_type = 'event' AND source_id IS NOT NULL
                                THEN lower(trim(contact)) || '#event#' || source_id
                        ELSE ''
                END
                WHERE dedupe_key = ''
        `); err != nil {
		return err
	}

	// 历史库里可能已经存在重复报名，这里保留最新一条作为主记录，
	// 旧记录让出 dedupe_key，避免唯一索引在迁移时直接失败。
	if _, err := db.Exec(`
                UPDATE leads AS current
                SET dedupe_key = ''
                WHERE current.dedupe_key <> ''
                  AND current.source_type = 'event'
                  AND current.source_id IS NOT NULL
                  AND EXISTS (
                        SELECT 1
                        FROM leads AS newer
                        WHERE newer.id > current.id
                          AND newer.source_type = current.source_type
                          AND newer.source_id = current.source_id
                          AND lower(trim(newer.contact)) = lower(trim(current.contact))
                  )
        `); err != nil {
		return err
	}

	if _, err := db.Exec(`
                CREATE UNIQUE INDEX IF NOT EXISTS idx_leads_event_dedupe
                ON leads(dedupe_key)
                WHERE dedupe_key <> ''
        `); err != nil {
		return err
	}

	return nil
}

func migrateKnowledgeAccessPassSchema(db *sql.DB) error {
	if _, err := db.Exec(`ALTER TABLE knowledge_access_passes ADD COLUMN legacy_token_id INTEGER UNIQUE`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return err
		}
	}

	if _, err := db.Exec(`
                INSERT INTO knowledge_access_passes (
                        legacy_token_id, name, token_hash, access_level, scope_type, is_active, expires_at, created_by, created_at, updated_at
                )
                SELECT t.id, t.name, t.token_hash, 'basic', 'single_space', t.is_active, t.expires_at, t.created_by, t.created_at, t.updated_at
                FROM knowledge_access_tokens t
                WHERE NOT EXISTS (
                        SELECT 1 FROM knowledge_access_passes p WHERE p.legacy_token_id = t.id
                )
        `); err != nil {
		return err
	}

	if _, err := db.Exec(`
                INSERT INTO knowledge_access_pass_spaces (pass_id, space_id)
                SELECT p.id, t.space_id
                FROM knowledge_access_passes p
                INNER JOIN knowledge_access_tokens t ON t.id = p.legacy_token_id
                WHERE NOT EXISTS (
                        SELECT 1 FROM knowledge_access_pass_spaces ps
                        WHERE ps.pass_id = p.id AND ps.space_id = t.space_id
                )
        `); err != nil {
		return err
	}

	return nil
}

func seedSiteConfigs(db *sql.DB) error {
	defaults := []struct {
		key   string
		value string
	}{
		{
			key: "home_banner",
			value: `{
  "titleText": "探索 AI 的边界",
  "subtitle": "我们构建 AI 资产，不做玩具。",
  "primaryCtaLabel": "连接社交网络",
  "primaryCtaPath": "/builders",
  "secondaryCtaLabel": "查看社区活动",
  "secondaryCtaPath": "/events",
  "statusLabel": "系统状态 // 社区指标"
}`,
		},
		{
			key: "home_featured",
			value: `{
  "communityCount": "1000",
  "communityCountSuffix": "+",
  "eventCount": "10",
  "eventCountSuffix": "+",
  "eventsDescription": "",
  "buildersDescription": "社区里不只有活动和内容，还有一批可以被连接、被协作，也能直接发起合作的人。",
  "insightsDescription": "剥离噪音，只提供值得关注的行业信号与落地经验。"
}`,
		},
		{
			key: "contact_channels",
			value: `[
  {
    "title": "官方小红书",
    "desc": "关注 VOID LAB 的活动回顾、现场照片、社区动态和新活动发布。",
    "account": "小红书号 VOIDLAB_AI",
    "buttonText": "关注小红书",
    "link": "#"
  },
  {
    "title": "加入社区群",
    "desc": "加入 VOID LAB 微信群，获取活动通知、结识伙伴和线下聚会信息。",
    "account": "企业微信入群二维码",
    "buttonText": "查看二维码",
    "link": "#"
  },
  {
    "title": "官方邮箱",
    "desc": "合作、赞助、学校/社区组织、媒体和正式事项联系。",
    "account": "JOIN@VOIDLAB.AI",
    "buttonText": "发邮件",
    "link": "mailto:join@voidlab.ai"
  },
  {
    "title": "官方小助手",
    "desc": "咨询、报名、入群、活动问题和成员对接支持。",
    "account": "企业微信客服链接",
    "buttonText": "添加小助手",
    "link": "#"
  }
]`,
		},
		{
			key: "footer_config",
			value: `{
  "slogan": "连接共创者，重塑 AI 资产。",
  "navLinks": [
    { "label": "首页", "path": "/" },
    { "label": "活动", "path": "/events" },
    { "label": "社交网络", "path": "/builders" },
    { "label": "资讯", "path": "/insights" },
    { "label": "联系我们", "path": "/contact" }
  ],
  "legalText": "VOIDLAB.AI © 2026. All rights reserved."
}`,
		},
		{
			key: "global_cta",
			value: `{
  "eyebrow": "NEXT ACTION",
  "title": "准备把想法变成活动、内容或合作项目？",
  "description": "如果你想和 VOIDLAB 一起办活动、发起合作、加入网络或咨询 AI 落地方案，可以直接从这里进入下一步。",
  "primaryLabel": "提交合作意向",
  "primaryPath": "/contact",
  "secondaryLabel": "查看活动",
  "secondaryPath": "/events"
}`,
		},
		{
			key: "featured_content_slots",
			value: `{
  "eventsTitle": "社区活动",
  "eventsViewAllLabel": "查看全部 [全部活动]",
  "eventsLimit": 5,
  "buildersTitle": "社交网络",
  "buildersViewAllLabel": "查看全部网络 [全部成员]",
  "buildersLimit": 6,
  "insightsTitle": "资讯中心",
  "insightsViewAllLabel": "查看全部资讯 [全部内容]",
  "insightsLimit": 5
}`,
		},
	}

	for _, item := range defaults {
		if _, err := db.Exec(
			`INSERT INTO site_configs (config_key, config_value_json, updated_by)
                         VALUES (?, ?, 1)
                         ON CONFLICT(config_key) DO NOTHING`,
			item.key,
			item.value,
		); err != nil {
			return err
		}
	}

	return nil
}
