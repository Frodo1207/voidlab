package dailynews

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const (
	itemsURL         = "https://aihot.virxact.com/api/v1/items?mode=selected&window=24h&limit=8"
	maxDraftArticles = 5
)

var (
	ErrNotFound          = errors.New("dailynews: task not found")
	ErrInvalidState      = errors.New("dailynews: invalid state transition")
	ErrSchedulerDisabled = errors.New("dailynews: scheduler is disabled")
	ErrSendInProgress    = errors.New("dailynews: send is in progress")
	ErrInvalidInput      = errors.New("dailynews: invalid input")
)

type Options struct {
	HTTPClient       *http.Client
	Sender           Sender
	Now              func() time.Time
	SchedulerEnabled bool
	// ScheduleTime is a Beijing wall-clock time, defaulting to 09:00.
	ScheduleTime string
	// SendTime is the earliest automatic delivery time in Beijing, defaulting to 10:00.
	SendTime  string
	PublicURL string
}

type Item struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Summary      string    `json:"summary"`
	Category     string    `json:"category"`
	PublishedAt  time.Time `json:"publishedAt"`
	DiscoveredAt time.Time `json:"discoveredAt"`
	Source       struct {
		Name string `json:"name"`
	} `json:"source"`
	Links struct {
		AIHOT    string `json:"aihot"`
		Original string `json:"original"`
	} `json:"links"`
}

type Snapshot struct {
	SchemaVersion int             `json:"schemaVersion"`
	Items         []Item          `json:"items"`
	Attribution   json.RawMessage `json:"attribution,omitempty"`
	Canonical     json.RawMessage `json:"canonical,omitempty"`
}

type Article struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Slug     string `json:"slug"`
	Summary  string `json:"summary"`
	CoverURL string `json:"cover_url"`
	Content  string `json:"content"`
	Status   string `json:"status"`
}

type Task struct {
	ID           int64         `json:"id"`
	Date         string        `json:"date"`
	Status       string        `json:"status"`
	SendStatus   string        `json:"send_status"`
	SendAttempts int           `json:"send_attempts"`
	SendUUID     string        `json:"send_uuid"`
	MessageID    string        `json:"message_id"`
	Message      string        `json:"message"`
	LastError    string        `json:"last_error"`
	ReviewerID   int64         `json:"reviewer_id"`
	ReviewNote   string        `json:"review_note"`
	CreatedAt    string        `json:"created_at"`
	UpdatedAt    string        `json:"updated_at"`
	Snapshot     Snapshot      `json:"snapshot"`
	Articles     []Article     `json:"articles,omitempty"`
	RunnerReport *RunnerReport `json:"runner_report,omitempty"`
	CopyText     string        `json:"-"`
}

type RunnerReport struct {
	Date           string   `json:"date"`
	Collect        string   `json:"collect"`
	TopUp          string   `json:"top_up"`
	Approve        string   `json:"approve"`
	Publish        string   `json:"publish"`
	GenerateCovers string   `json:"generate_covers"`
	UploadAssets   string   `json:"upload_assets"`
	BindCovers     string   `json:"bind_covers"`
	BuildDigest    string   `json:"build_digest"`
	SendFeishu     string   `json:"send_feishu"`
	Failures       []string `json:"failures,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
	Complete       bool     `json:"complete"`
	RecordedAt     string   `json:"recorded_at,omitempty"`
}

type DeliveryDigest struct {
	Date        string
	ItemCount   int
	ReadingText string
	CopyText    string
	Articles    []DigestArticle
}

type DigestArticle struct {
	Title       string
	Summary     string
	URL         string
	SourceName  string
	OriginalURL string
}

type Service struct {
	db        *sql.DB
	client    *http.Client
	sender    Sender
	now       func() time.Time
	location  *time.Location
	enabled   bool
	schedule  string
	sendTime  string
	publicURL string
	collectMu sync.Mutex
}

// NewService migrates only missing tables. It never starts a goroutine or owns db.
func NewService(ctx context.Context, db *sql.DB, opts Options) (*Service, error) {
	if ctx == nil || db == nil {
		return nil, errors.New("dailynews: context and database are required")
	}
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.ScheduleTime == "" {
		opts.ScheduleTime = "09:00"
	}
	if _, err := time.Parse("15:04", opts.ScheduleTime); err != nil {
		return nil, errors.New("dailynews: invalid schedule time")
	}
	if opts.SendTime == "" {
		opts.SendTime = "10:00"
	}
	if parsed, err := time.Parse("15:04", opts.SendTime); err != nil || parsed.Format("15:04") != opts.SendTime {
		return nil, errors.New("dailynews: invalid send time")
	}
	if opts.PublicURL == "" {
		opts.PublicURL = "https://voidlab.club"
	}
	if !validLink(opts.PublicURL) {
		return nil, errors.New("dailynews: invalid public URL")
	}
	client := http.Client{Timeout: 15 * time.Second}
	if opts.HTTPClient != nil {
		client = *opts.HTTPClient
	}
	client.Jar = nil
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	s := &Service{db: db, client: &client, sender: opts.Sender, now: opts.Now,
		location: location, enabled: opts.SchedulerEnabled, schedule: opts.ScheduleTime, sendTime: opts.SendTime,
		publicURL: strings.TrimRight(opts.PublicURL, "/")}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS articles (
 id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, slug TEXT NOT NULL UNIQUE,
 summary TEXT NOT NULL DEFAULT '', category TEXT NOT NULL DEFAULT '', audience TEXT NOT NULL DEFAULT '',
 tags_json TEXT NOT NULL DEFAULT '[]', cover_url TEXT NOT NULL DEFAULT '', cover_media_id INTEGER,
 content TEXT NOT NULL DEFAULT '', source_name TEXT NOT NULL DEFAULT '', source_url TEXT NOT NULL DEFAULT '',
 featured INTEGER NOT NULL DEFAULT 0, status TEXT NOT NULL DEFAULT 'draft', published_at DATETIME,
 created_by INTEGER, updated_by INTEGER, created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
 updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS dailynews_cache (
 url TEXT PRIMARY KEY, etag TEXT NOT NULL, payload TEXT NOT NULL, fetched_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS dailynews_tasks (
 id INTEGER PRIMARY KEY AUTOINCREMENT, date TEXT NOT NULL UNIQUE,
 status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','rejected','published')),
 snapshot_json TEXT NOT NULL, reviewer_id INTEGER NOT NULL DEFAULT 0, review_note TEXT NOT NULL DEFAULT '',
 send_status TEXT NOT NULL DEFAULT 'pending' CHECK(send_status IN ('pending','sending','failed','sent')),
 send_attempts INTEGER NOT NULL DEFAULT 0, send_uuid TEXT NOT NULL UNIQUE,
 send_text TEXT NOT NULL DEFAULT '', send_copy_text TEXT NOT NULL DEFAULT '',
 message_id TEXT NOT NULL DEFAULT '', last_error TEXT NOT NULL DEFAULT '',
 send_started_at INTEGER NOT NULL DEFAULT 0,
 created_at TEXT NOT NULL, updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS dailynews_task_articles (
 task_id INTEGER NOT NULL REFERENCES dailynews_tasks(id), position INTEGER NOT NULL,
 item_id TEXT NOT NULL UNIQUE, source_url TEXT NOT NULL UNIQUE,
 article_id INTEGER NOT NULL UNIQUE REFERENCES articles(id),
 PRIMARY KEY(task_id,position)
);`)
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `ALTER TABLE dailynews_tasks ADD COLUMN send_copy_text TEXT NOT NULL DEFAULT ''`); err != nil &&
		!strings.Contains(err.Error(), "duplicate column name") {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `ALTER TABLE dailynews_tasks ADD COLUMN runner_report_json TEXT NOT NULL DEFAULT ''`); err != nil &&
		!strings.Contains(err.Error(), "duplicate column name") {
		return nil, err
	}
	// Refresh unpublished previews, but never rewrite a frozen published message.
	rows, err := tx.QueryContext(ctx, `SELECT `+taskColumns+` FROM dailynews_tasks
WHERE send_text='' OR send_copy_text='' OR status!='published'`)
	if err != nil {
		return nil, err
	}
	var legacy []Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		legacy = append(legacy, task)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	for _, task := range legacy {
		articles, err := readArticles(ctx, tx, task.ID)
		if err != nil {
			return nil, err
		}
		message := s.renderMessage(task.Date, task.Snapshot, articles)
		copyText := s.renderCopyText(task.Date, task.Snapshot, articles)
		if task.Status != "published" {
			copyText = message
		} else if strings.TrimSpace(task.Message) != "" {
			message = task.Message
		}
		if _, err := tx.ExecContext(ctx, `UPDATE dailynews_tasks SET send_text=?,send_copy_text=? WHERE id=?`,
			message, copyText, task.ID); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return s, nil
}

// Collect persists the ETag and payload together; cached reads enforce a 60s poll floor.
func (s *Service) Collect(ctx context.Context) (Snapshot, error) {
	s.collectMu.Lock()
	defer s.collectMu.Unlock()
	var etag, payload string
	var fetched int64
	err := s.db.QueryRowContext(ctx, `SELECT etag,payload,fetched_at FROM dailynews_cache WHERE url=?`, itemsURL).
		Scan(&etag, &payload, &fetched)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Snapshot{}, err
	}
	now := s.now().Unix()
	if payload != "" && now >= fetched && now-fetched < 60 {
		return decodeSnapshot([]byte(payload))
	}
	requestCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, itemsURL, nil)
	if err != nil {
		return Snapshot{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "voidlab-dailynews/1.0")
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		if requestCtx.Err() != nil {
			return Snapshot{}, requestCtx.Err()
		}
		return Snapshot{}, errors.New("dailynews: collection request failed")
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusNotModified:
		if payload == "" {
			return Snapshot{}, errors.New("dailynews: 304 without cached payload")
		}
	case http.StatusOK:
		data, readErr := io.ReadAll(io.LimitReader(resp.Body, (1<<20)+1))
		if readErr != nil || len(data) > 1<<20 {
			return Snapshot{}, errors.New("dailynews: cannot read bounded response")
		}
		payload = string(data)
		etag = resp.Header.Get("ETag")
	default:
		return Snapshot{}, fmt.Errorf("dailynews: HTTP status %d", resp.StatusCode)
	}
	snapshot, err := decodeSnapshot([]byte(payload))
	if err != nil {
		return Snapshot{}, err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO dailynews_cache(url,etag,payload,fetched_at)
VALUES(?,?,?,?) ON CONFLICT(url) DO UPDATE SET etag=excluded.etag,payload=excluded.payload,fetched_at=excluded.fetched_at`,
		itemsURL, etag, payload, now)
	return snapshot, err
}

func decodeSnapshot(data []byte) (Snapshot, error) {
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil || snapshot.SchemaVersion != 1 || snapshot.Items == nil || len(snapshot.Items) > 8 {
		return Snapshot{}, errors.New("dailynews: invalid AIHOT response")
	}
	for _, item := range snapshot.Items {
		if !validItem(item) {
			return Snapshot{}, errors.New("dailynews: invalid AIHOT item")
		}
	}
	return snapshot, nil
}

func validItem(item Item) bool {
	return strings.TrimSpace(item.ID) != "" &&
		strings.TrimSpace(item.Title) != "" &&
		!item.DiscoveredAt.IsZero() &&
		validLink(item.Links.AIHOT) &&
		validLink(item.Links.Original)
}

func validLink(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != "" && u.User == nil &&
		!strings.ContainsAny(raw, "\r\n")
}

// Generate is date-idempotent and creates up to maxDraftArticles unseen drafts, including none.
func (s *Service) Generate(ctx context.Context, date string) (Task, error) {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return Task{}, errors.New("dailynews: date must be YYYY-MM-DD")
	}
	if task, err := s.GetByDate(ctx, date); err == nil {
		return task, nil
	} else if !errors.Is(err, ErrNotFound) {
		return Task{}, err
	}
	snapshot, err := s.Collect(ctx)
	if err != nil {
		return Task{}, err
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return Task{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Task{}, err
	}
	defer tx.Rollback()
	now := s.now().UTC().Format(time.RFC3339Nano)
	result, err := tx.ExecContext(ctx, `INSERT INTO dailynews_tasks(date,snapshot_json,send_uuid,created_at,updated_at)
VALUES(?,?,?,?,?) ON CONFLICT(date) DO NOTHING`, date, string(data), sendUUID(date), now, now)
	if err != nil {
		return Task{}, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return Task{}, err
	}
	if n == 0 {
		if err := tx.Rollback(); err != nil {
			return Task{}, err
		}
		return s.GetByDate(ctx, date)
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		return Task{}, err
	}
	count := 0
	for _, item := range snapshot.Items {
		var used int
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM dailynews_task_articles WHERE item_id=? OR source_url=?)
OR EXISTS(SELECT 1 FROM articles WHERE source_url=? OR slug=?)`,
			item.ID, item.Links.Original, item.Links.Original, articleSlug(item)).Scan(&used)
		if err != nil {
			return Task{}, err
		}
		if used != 0 {
			continue
		}
		result, err = tx.ExecContext(ctx, `INSERT INTO articles(title,slug,summary,category,content,source_name,source_url,status)
VALUES(?,?,?,?,?,?,?,'draft')`, item.Title, articleSlug(item), item.Summary, item.Category,
			articleContent(item, snapshot), item.Source.Name, item.Links.Original)
		if err != nil {
			return Task{}, err
		}
		articleID, err := result.LastInsertId()
		if err != nil {
			return Task{}, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO dailynews_task_articles(task_id,position,item_id,source_url,article_id) VALUES(?,?,?,?,?)`,
			taskID, count, item.ID, item.Links.Original, articleID)
		if err != nil {
			return Task{}, err
		}
		count++
		if count == maxDraftArticles {
			break
		}
	}
	articles, err := readArticles(ctx, tx, taskID)
	if err != nil {
		return Task{}, err
	}
	preview := s.renderMessage(date, snapshot, articles)
	if _, err := tx.ExecContext(ctx, `UPDATE dailynews_tasks SET send_text=?,send_copy_text=? WHERE id=?`,
		preview, preview, taskID); err != nil {
		return Task{}, err
	}
	if err := tx.Commit(); err != nil {
		return Task{}, err
	}
	return s.Get(ctx, taskID)
}

// AppendItems adds manually curated items to an existing unpublished task until maxDraftArticles is reached.
func (s *Service) AppendItems(ctx context.Context, id int64, items []Item) (Task, error) {
	if len(items) == 0 {
		return Task{}, ErrInvalidInput
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Task{}, err
	}
	defer tx.Rollback()
	task, err := scanTask(tx.QueryRowContext(ctx, `SELECT `+taskColumns+` FROM dailynews_tasks WHERE id=?`, id))
	if err != nil {
		return Task{}, err
	}
	if task.Status != "pending" {
		return Task{}, ErrInvalidState
	}
	articles, err := readArticles(ctx, tx, id)
	if err != nil {
		return Task{}, err
	}
	position := len(articles)
	if position > maxDraftArticles {
		return Task{}, ErrInvalidState
	}
	existingItemIDs := make(map[string]struct{}, len(task.Snapshot.Items))
	existingSourceURLs := make(map[string]struct{}, len(task.Snapshot.Items))
	for _, item := range task.Snapshot.Items {
		existingItemIDs[item.ID] = struct{}{}
		existingSourceURLs[item.Links.Original] = struct{}{}
	}
	added := false
	for _, item := range items {
		if !validItem(item) {
			return Task{}, ErrInvalidInput
		}
		if position >= maxDraftArticles {
			break
		}
		if _, ok := existingItemIDs[item.ID]; ok {
			continue
		}
		if _, ok := existingSourceURLs[item.Links.Original]; ok {
			continue
		}
		var used int
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM dailynews_task_articles WHERE item_id=? OR source_url=?)
OR EXISTS(SELECT 1 FROM articles WHERE source_url=? OR slug=?)`,
			item.ID, item.Links.Original, item.Links.Original, articleSlug(item)).Scan(&used)
		if err != nil {
			return Task{}, err
		}
		if used != 0 {
			continue
		}
		result, err := tx.ExecContext(ctx, `INSERT INTO articles(title,slug,summary,category,content,source_name,source_url,status)
VALUES(?,?,?,?,?,?,?,'draft')`, item.Title, articleSlug(item), item.Summary, item.Category,
			articleContent(item, task.Snapshot), item.Source.Name, item.Links.Original)
		if err != nil {
			return Task{}, err
		}
		articleID, err := result.LastInsertId()
		if err != nil {
			return Task{}, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO dailynews_task_articles(task_id,position,item_id,source_url,article_id) VALUES(?,?,?,?,?)`,
			id, position, item.ID, item.Links.Original, articleID); err != nil {
			return Task{}, err
		}
		task.Snapshot.Items = append(task.Snapshot.Items, item)
		existingItemIDs[item.ID] = struct{}{}
		existingSourceURLs[item.Links.Original] = struct{}{}
		position++
		added = true
	}
	if !added {
		if err := tx.Rollback(); err != nil {
			return Task{}, err
		}
		return s.Get(ctx, id)
	}
	data, err := json.Marshal(task.Snapshot)
	if err != nil {
		return Task{}, err
	}
	articles, err = readArticles(ctx, tx, id)
	if err != nil {
		return Task{}, err
	}
	preview := s.renderMessage(task.Date, task.Snapshot, articles)
	if _, err := tx.ExecContext(ctx, `UPDATE dailynews_tasks SET snapshot_json=?,send_text=?,send_copy_text=?,updated_at=? WHERE id=?`,
		string(data), preview, preview, s.now().UTC().Format(time.RFC3339Nano), id); err != nil {
		return Task{}, err
	}
	if err := tx.Commit(); err != nil {
		return Task{}, err
	}
	return s.Get(ctx, id)
}

// Reset deletes one daily task and all articles created for it, allowing the date to be regenerated.
func (s *Service) Reset(ctx context.Context, id int64) ([]int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	task, err := scanTask(tx.QueryRowContext(ctx, `SELECT `+taskColumns+` FROM dailynews_tasks WHERE id=?`, id))
	if err != nil {
		return nil, err
	}
	if task.SendStatus == "sending" {
		return nil, ErrSendInProgress
	}
	rows, err := tx.QueryContext(ctx, `SELECT article_id FROM dailynews_task_articles WHERE task_id=? ORDER BY position`, id)
	if err != nil {
		return nil, err
	}
	articleIDs := make([]int64, 0, maxDraftArticles)
	for rows.Next() {
		var articleID int64
		if err := rows.Scan(&articleID); err != nil {
			rows.Close()
			return nil, err
		}
		articleIDs = append(articleIDs, articleID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	if _, err := tx.ExecContext(ctx, `DELETE FROM dailynews_task_articles WHERE task_id=?`, id); err != nil {
		return nil, err
	}
	if len(articleIDs) > 0 {
		args := make([]any, 0, len(articleIDs))
		placeholders := make([]string, 0, len(articleIDs))
		for _, articleID := range articleIDs {
			args = append(args, articleID)
			placeholders = append(placeholders, "?")
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM articles WHERE id IN (`+strings.Join(placeholders, ",")+`)`, args...); err != nil {
			return nil, err
		}
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM dailynews_tasks WHERE id=?`, id)
	if err != nil {
		return nil, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, ErrNotFound
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return articleIDs, nil
}

func articleSlug(item Item) string {
	base := slugify(item.Title)
	if base == "" {
		base = "daily-insight"
	}
	if len(base) > 48 {
		base = strings.Trim(base[:48], "-")
	}
	sum := sha256.Sum256([]byte(item.ID + "|" + item.Links.Original))
	return "insight-" + base + "-" + hex.EncodeToString(sum[:4])
}

func sendUUID(date string) string {
	sum := sha256.Sum256([]byte("voidlab-dailynews:" + date))
	h := hex.EncodeToString(sum[:16])
	return h[:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:]
}

func articleContent(item Item, snapshot Snapshot) string {
	return "## " + markdownText(item.Title) + "\n\n" + item.Summary +
		"\n\n### 来源\n\n- 来源：" + markdownText(item.Source.Name) +
		"\n- [第三方原文](" + markdownURL(item.Links.Original) + ")\n\n" + attributionText(snapshot)
}

func attributionText(snapshot Snapshot) string {
	_ = snapshot
	return "第三方原文版权归原作者所有。"
}

func markdownText(text string) string {
	return strings.NewReplacer("\\", "\\\\", "[", "\\[", "]", "\\]", "*", "\\*", "_", "\\_",
		"`", "\\`", "<", "&lt;", ">", "&gt;", "\r", " ", "\n", " ").Replace(text)
}

func markdownURL(raw string) string {
	return strings.NewReplacer("(", "%28", ")", "%29", " ", "%20", "<", "%3C", ">", "%3E").Replace(raw)
}

func slugify(text string) string {
	var b strings.Builder
	lastHyphen := false
	for _, r := range strings.ToLower(strings.TrimSpace(text)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		case !lastHyphen:
			b.WriteByte('-')
			lastHyphen = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func (s *Service) renderMessage(date string, snapshot Snapshot, articles []Article) string {
	var text strings.Builder
	fmt.Fprintf(&text, "VOIDLAB 每日资讯 | %s\n\n过去 24 小时 · %d 条资讯\n", date, len(snapshot.Items))
	if len(snapshot.Items) == 0 {
		text.WriteString("\n过去 24 小时暂无精选资讯。\n")
	}
	publishedArticles := make([]Article, 0, len(articles))
	for _, article := range articles {
		if article.Status == "published" {
			publishedArticles = append(publishedArticles, article)
		}
	}
	if len(publishedArticles) > 0 {
		for i, article := range publishedArticles {
			fmt.Fprintf(&text, "\n%d. %s\n%s\n站内链接：%s/insights/%s\n",
				i+1, article.Title, article.Summary, s.publicURL, url.PathEscape(article.Slug))
		}
		text.WriteString("\n原始来源\n")
		for i, item := range snapshot.Items {
			fmt.Fprintf(&text, "\n%d. %s · %s\n原文链接：%s\n",
				i+1, item.Title, item.Source.Name, item.Links.Original)
		}
		text.WriteString("\n" + attributionText(snapshot))
		return text.String()
	}
	for i, item := range snapshot.Items {
		fmt.Fprintf(&text, "\n%d. %s\n%s\n来源：%s\n原文链接：%s\n",
			i+1, item.Title, item.Summary, item.Source.Name, item.Links.Original)
	}
	text.WriteString("\n" + attributionText(snapshot))
	return text.String()
}

func (s *Service) renderCopyText(date string, snapshot Snapshot, articles []Article) string {
	var text strings.Builder
	fmt.Fprintf(&text, "VOIDLAB 每日资讯 | %s\n\n过去 24 小时 · %d 条资讯\n", date, len(snapshot.Items))
	publishedArticles := make([]Article, 0, len(articles))
	for _, article := range articles {
		if article.Status == "published" {
			publishedArticles = append(publishedArticles, article)
		}
	}
	if len(publishedArticles) == 0 {
		if len(snapshot.Items) == 0 {
			text.WriteString("\n过去 24 小时暂无精选资讯。\n")
			return text.String()
		}
		for i, item := range snapshot.Items {
			fmt.Fprintf(&text, "\n%d. %s\n%s\n原文链接：%s\n", i+1, item.Title, item.Summary, item.Links.Original)
		}
		return text.String()
	}
	for i, article := range publishedArticles {
		fmt.Fprintf(&text, "\n%d. %s\n%s\n站内链接：%s/insights/%s\n",
			i+1, article.Title, article.Summary, s.publicURL, url.PathEscape(article.Slug))
	}
	return text.String()
}

func (s *Service) buildDigest(task Task) DeliveryDigest {
	articles := make([]DigestArticle, 0, len(task.Articles))
	for _, article := range task.Articles {
		if article.Status != "published" {
			continue
		}
		sourceName := ""
		originalURL := ""
		if len(articles) < len(task.Snapshot.Items) {
			item := task.Snapshot.Items[len(articles)]
			sourceName = item.Source.Name
			originalURL = item.Links.Original
		}
		articles = append(articles, DigestArticle{
			Title:       article.Title,
			Summary:     article.Summary,
			URL:         s.publicURL + "/insights/" + url.PathEscape(article.Slug),
			SourceName:  sourceName,
			OriginalURL: originalURL,
		})
	}
	return DeliveryDigest{
		Date:        task.Date,
		ItemCount:   len(task.Snapshot.Items),
		ReadingText: s.renderMessage(task.Date, task.Snapshot, task.Articles),
		CopyText:    s.renderCopyText(task.Date, task.Snapshot, task.Articles),
		Articles:    articles,
	}
}

const taskColumns = `id,date,status,send_status,send_attempts,send_uuid,message_id,last_error,
reviewer_id,review_note,created_at,updated_at,snapshot_json,send_text,send_copy_text,runner_report_json`

type taskScanner interface{ Scan(...any) error }

func scanTask(row taskScanner) (Task, error) {
	var task Task
	var payload, runnerPayload string
	err := row.Scan(&task.ID, &task.Date, &task.Status, &task.SendStatus, &task.SendAttempts,
		&task.SendUUID, &task.MessageID, &task.LastError, &task.ReviewerID, &task.ReviewNote,
		&task.CreatedAt, &task.UpdatedAt, &payload, &task.Message, &task.CopyText, &runnerPayload)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, err
	}
	err = json.Unmarshal([]byte(payload), &task.Snapshot)
	if err != nil {
		return Task{}, err
	}
	if strings.TrimSpace(runnerPayload) != "" {
		var report RunnerReport
		if err := json.Unmarshal([]byte(runnerPayload), &report); err != nil {
			return Task{}, err
		}
		task.RunnerReport = &report
	}
	return task, nil
}

func (s *Service) GetByDate(ctx context.Context, date string) (Task, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, `SELECT id FROM dailynews_tasks WHERE date=?`, date).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, err
	}
	return s.Get(ctx, id)
}

func (s *Service) Get(ctx context.Context, id int64) (Task, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return Task{}, err
	}
	defer tx.Rollback()
	task, err := scanTask(tx.QueryRowContext(ctx, `SELECT `+taskColumns+` FROM dailynews_tasks WHERE id=?`, id))
	if err != nil {
		return Task{}, err
	}
	task.Articles, err = readArticles(ctx, tx, id)
	if err != nil {
		return Task{}, err
	}
	if task.Status != "published" {
		task.Message = s.renderMessage(task.Date, task.Snapshot, nil)
		task.CopyText = task.Message
	} else if task.Message == "" {
		task.Message = s.renderMessage(task.Date, task.Snapshot, task.Articles)
	}
	if task.Status == "published" && task.CopyText == "" {
		task.CopyText = s.renderCopyText(task.Date, task.Snapshot, task.Articles)
	}
	return task, tx.Commit()
}

func (s *Service) UpdateRunnerReport(ctx context.Context, date string, report RunnerReport) (Task, error) {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return Task{}, errors.New("dailynews: date must be YYYY-MM-DD")
	}
	task, err := s.GetByDate(ctx, date)
	if err != nil {
		return Task{}, err
	}
	report.Date = task.Date
	report.RecordedAt = s.now().UTC().Format(time.RFC3339Nano)
	payload, err := json.Marshal(report)
	if err != nil {
		return Task{}, err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE dailynews_tasks SET runner_report_json=?,updated_at=? WHERE id=?`,
		string(payload), report.RecordedAt, task.ID)
	if err != nil {
		return Task{}, err
	}
	return s.Get(ctx, task.ID)
}

func readArticles(ctx context.Context, tx *sql.Tx, id int64) ([]Article, error) {
	rows, err := tx.QueryContext(ctx, `SELECT a.id,a.title,a.slug,a.summary,a.cover_url,a.content,a.status
FROM dailynews_task_articles d JOIN articles a ON a.id=d.article_id WHERE d.task_id=? ORDER BY d.position`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	articles := make([]Article, 0, maxDraftArticles)
	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Slug, &a.Summary, &a.CoverURL, &a.Content, &a.Status); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, rows.Err()
}

// List returns newest dates first; articles are available through Get.
func (s *Service) List(ctx context.Context, limit, offset int) ([]Task, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, errors.New("dailynews: limit must be 1..100 and offset nonnegative")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+taskColumns+` FROM dailynews_tasks ORDER BY date DESC,id DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := make([]Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		if task.Status != "published" || task.Message == "" {
			task.Message = s.renderMessage(task.Date, task.Snapshot, nil)
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

// Review is a one-way decision on a pending task; authorization belongs to the caller.
func (s *Service) Review(ctx context.Context, id, reviewerID int64, approve bool, note string) (Task, error) {
	if reviewerID <= 0 {
		return Task{}, errors.New("dailynews: reviewer ID is required")
	}
	status := "rejected"
	if approve {
		status = "approved"
	}
	result, err := s.db.ExecContext(ctx, `UPDATE dailynews_tasks SET status=?,reviewer_id=?,review_note=?,updated_at=?
WHERE id=? AND status='pending'`, status, reviewerID, note, s.now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return Task{}, err
	}
	if n, err := result.RowsAffected(); err != nil {
		return Task{}, err
	} else if n == 0 {
		if _, err := s.Get(ctx, id); err != nil {
			return Task{}, err
		}
		return Task{}, ErrInvalidState
	}
	return s.Get(ctx, id)
}

// Publish atomically publishes all linked drafts and freezes their latest titles and links.
func (s *Service) Publish(ctx context.Context, id, userID int64) (Task, error) {
	if userID <= 0 {
		return Task{}, errors.New("dailynews: publisher ID is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Task{}, err
	}
	defer tx.Rollback()
	// Acquire the SQLite writer before reading the state.
	result, err := tx.ExecContext(ctx, `UPDATE dailynews_tasks SET status='published',updated_at=? WHERE id=? AND status='approved'`,
		s.now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return Task{}, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return Task{}, err
	}
	task, err := scanTask(tx.QueryRowContext(ctx, `SELECT `+taskColumns+` FROM dailynews_tasks WHERE id=?`, id))
	if err != nil {
		return Task{}, err
	}
	if n == 0 {
		if task.Status != "published" {
			return Task{}, ErrInvalidState
		}
		if err := tx.Rollback(); err != nil {
			return Task{}, err
		}
		return s.Get(ctx, id)
	}
	articles, err := readArticles(ctx, tx, id)
	if err != nil {
		return Task{}, err
	}
	var linked int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM dailynews_task_articles WHERE task_id=?`, id).Scan(&linked); err != nil {
		return Task{}, err
	}
	if len(articles) != linked || linked > maxDraftArticles {
		return Task{}, ErrInvalidState
	}
	result, err = tx.ExecContext(ctx, `UPDATE articles SET status='published',published_at=CURRENT_TIMESTAMP,
updated_by=?,updated_at=CURRENT_TIMESTAMP WHERE id IN (SELECT article_id FROM dailynews_task_articles WHERE task_id=?) AND status='draft'`, userID, id)
	if err != nil {
		return Task{}, err
	}
	if n, err = result.RowsAffected(); err != nil {
		return Task{}, err
	} else if n != int64(linked) {
		return Task{}, ErrInvalidState
	}
	articles, err = readArticles(ctx, tx, id)
	if err != nil {
		return Task{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE dailynews_tasks SET send_text=?,send_copy_text=? WHERE id=?`,
		s.renderMessage(task.Date, task.Snapshot, articles),
		s.renderCopyText(task.Date, task.Snapshot, articles), id); err != nil {
		return Task{}, err
	}
	if err := tx.Commit(); err != nil {
		return Task{}, err
	}
	return s.Get(ctx, id)
}

// Send claims a persisted lease. Crashed/ambiguous attempts reuse the same UUID and text.
func (s *Service) Send(ctx context.Context, id int64) (Task, error) {
	return s.send(ctx, id, false)
}

func (s *Service) send(ctx context.Context, id int64, automatic bool) (Task, error) {
	if s.sender == nil {
		return Task{}, errors.New("dailynews: sender is not configured")
	}
	now := s.now().Unix()
	condition := `(send_status IN ('pending','failed') OR (send_status='sending' AND send_started_at<=?))`
	if automatic {
		// Claim only the first attempt, even if another scheduler selected this task earlier.
		condition = `(send_status='pending' AND send_attempts=0 AND send_started_at<=?)`
	}
	result, err := s.db.ExecContext(ctx, `UPDATE dailynews_tasks SET send_status='sending',
send_attempts=send_attempts+1,send_started_at=?,last_error='',updated_at=?
WHERE id=? AND status='published' AND `+condition,
		now, s.now().UTC().Format(time.RFC3339Nano), id, now-120)
	if err != nil {
		return Task{}, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return Task{}, err
	}
	task, err := s.Get(ctx, id)
	if err != nil {
		return Task{}, err
	}
	if n == 0 {
		if automatic {
			return task, nil
		}
		if task.SendStatus == "sent" {
			return task, nil
		}
		if task.SendStatus == "sending" {
			return Task{}, ErrSendInProgress
		}
		return Task{}, ErrInvalidState
	}
	var text, copyText string
	if err := s.db.QueryRowContext(ctx, `SELECT send_text,send_copy_text FROM dailynews_tasks WHERE id=?`, id).Scan(&text, &copyText); err != nil {
		return Task{}, err
	}
	sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	digest := s.buildDigest(task)
	if strings.TrimSpace(text) != "" {
		digest.ReadingText = text
	}
	if strings.TrimSpace(copyText) != "" {
		digest.CopyText = copyText
	}
	if strings.TrimSpace(digest.ReadingText) == "" {
		digest.ReadingText = s.renderMessage(task.Date, task.Snapshot, task.Articles)
	}
	if strings.TrimSpace(digest.CopyText) == "" {
		digest.CopyText = digest.ReadingText
	}
	messageID := ""
	var sendErr error
	if sender, ok := s.sender.(interface {
		SendDigest(context.Context, DeliveryDigest, string) (string, error)
	}); ok {
		messageID, sendErr = sender.SendDigest(sendCtx, digest, task.SendUUID)
	} else {
		messageID, sendErr = s.sender.Send(sendCtx, digest.CopyText, task.SendUUID)
	}
	if sendErr == nil && strings.TrimSpace(messageID) == "" {
		sendErr = errors.New("dailynews: sender returned empty message ID")
	}
	status, lastError := "sent", ""
	if sendErr != nil {
		status, lastError, messageID = "failed", strings.TrimSpace(sendErr.Error()), ""
		if lastError == "" {
			lastError = "delivery failed; retry with the same UUID"
		}
	}
	// Persist failures even when the caller cancels during delivery.
	saveCtx, saveCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer saveCancel()
	result, err = s.db.ExecContext(saveCtx, `UPDATE dailynews_tasks SET send_status=?,message_id=?,last_error=?,updated_at=?
WHERE id=? AND send_status='sending' AND send_attempts=?`, status, messageID, lastError,
		s.now().UTC().Format(time.RFC3339Nano), id, task.SendAttempts)
	if err != nil {
		return Task{}, err
	}
	if n, err := result.RowsAffected(); err != nil {
		return Task{}, err
	} else if n != 1 {
		return Task{}, ErrSendInProgress
	}
	task, err = s.Get(saveCtx, id)
	if err != nil {
		return Task{}, err
	}
	if sendErr != nil {
		return task, errors.New("dailynews: delivery failed")
	}
	return task, nil
}

// RunScheduler generates drafts and sends published tasks once after SendTime.
// Approval, publication and all retries remain manual.
// onError must return promptly. The caller explicitly starts this method in its own goroutine.
func (s *Service) RunScheduler(ctx context.Context, onError func(error)) error {
	if !s.enabled {
		return ErrSchedulerDisabled
	}
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.runScheduled(ctx); err != nil && onError != nil {
			onError(err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (s *Service) runScheduled(ctx context.Context) error {
	now := s.now().In(s.location)
	date := now.Format("2006-01-02")
	var failures []error
	if now.Format("15:04") >= s.schedule {
		if _, err := s.Generate(ctx, date); err != nil {
			failures = append(failures, err)
		}
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM dailynews_tasks
WHERE status='published' AND send_status='pending' AND send_attempts=0
AND date=? AND ? >= ? ORDER BY id`,
		date, now.Format("15:04"), s.sendTime)
	if err != nil {
		return errors.Join(append(failures, err)...)
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return errors.Join(append(failures, err)...)
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return errors.Join(append(failures, err)...)
	}
	for _, id := range ids {
		if _, err := s.send(ctx, id, true); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}
