package dailynews

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"voidlabai/apps/api/internal/storage"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type senderFunc func(context.Context, string, string) (string, error)

func (f senderFunc) Send(ctx context.Context, text, uuid string) (string, error) {
	return f(ctx, text, uuid)
}

func fixture(count int) Snapshot {
	s := Snapshot{SchemaVersion: 1, Items: make([]Item, 0, count),
		Attribution: json.RawMessage(`{"name":"AI HOT","canonical":"https://aihot.virxact.com"}`)}
	for i := 0; i < count; i++ {
		item := Item{ID: fmt.Sprintf("item-%d", i), Title: fmt.Sprintf("Title %d", i),
			Summary: fmt.Sprintf("Summary %d", i), Category: "ai-models",
			DiscoveredAt: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}
		item.Source.Name = "Official"
		item.Links.AIHOT = fmt.Sprintf("https://aihot.virxact.com/item/%d", i)
		item.Links.Original = fmt.Sprintf("https://example.com/news/%d", i)
		s.Items = append(s.Items, item)
	}
	return s
}

func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Etag": []string{`"v1"`}},
		Body: io.NopCloser(strings.NewReader(body))}
}

func fixtureClient(t *testing.T, snapshot Snapshot) *http.Client {
	t.Helper()
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{Transport: transportFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != itemsURL || req.Method != http.MethodGet ||
			req.Header.Get("Authorization") != "" || req.Header.Get("Cookie") != "" {
			t.Errorf("unexpected anonymous request: %s %s", req.Method, req.URL)
		}
		return response(200, string(data)), nil
	})}
}

func database(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "dailynews.db"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	return db
}

func newTestService(t *testing.T, db *sql.DB, opts Options) *Service {
	t.Helper()
	if opts.HTTPClient == nil {
		opts.HTTPClient = fixtureClient(t, fixture(8))
	}
	s, err := NewService(context.Background(), db, opts)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func generate(t *testing.T, s *Service, date string) Task {
	t.Helper()
	task, err := s.Generate(context.Background(), date)
	if err != nil {
		t.Fatal(err)
	}
	return task
}

func assertCount(t *testing.T, db *sql.DB, table string, want int) {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != want {
		t.Fatalf("%s count = %d, want %d", table, n, want)
	}
}

func seedDefaultGallery(t *testing.T, db *sql.DB, urls []string) {
	t.Helper()
	if _, err := db.Exec(`
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
		CREATE TABLE IF NOT EXISTS media_galleries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL DEFAULT '',
			is_default INTEGER NOT NULL DEFAULT 0,
			created_by INTEGER NOT NULL DEFAULT 1,
			updated_by INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS media_gallery_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			gallery_id INTEGER NOT NULL,
			media_asset_id INTEGER NOT NULL,
			alt_text TEXT NOT NULL DEFAULT '',
			sort_order INTEGER NOT NULL DEFAULT 10,
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(gallery_id, media_asset_id)
		);
	`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO media_galleries (id, title, slug, description, is_default, created_by, updated_by)
		VALUES (1, '基础图库', 'base-gallery', 'default', 1, 1, 1)
		ON CONFLICT(id) DO NOTHING
	`); err != nil {
		t.Fatal(err)
	}
	for index, rawURL := range urls {
		result, err := db.Exec(`
			INSERT INTO media_assets (object_key, object_url, file_name, content_type, file_size, uploaded_by)
			VALUES (?, ?, ?, 'image/jpeg', 1024, 1)
		`, fmt.Sprintf("asset-%d", index+1), rawURL, fmt.Sprintf("cover-%d.jpg", index+1))
		if err != nil {
			t.Fatal(err)
		}
		assetID, err := result.LastInsertId()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`
			INSERT INTO media_gallery_items (gallery_id, media_asset_id, alt_text, sort_order, is_active)
			VALUES (1, ?, '', ?, 1)
		`, assetID, (index+1)*10); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMigrationGenerationAndDateDeduplication(t *testing.T) {
	ctx := context.Background()
	db := database(t)
	s := newTestService(t, db, Options{})
	task := generate(t, s, "2026-10-12")
	if task.Status != "pending" || task.SendStatus != "pending" || len(task.Snapshot.Items) != 8 ||
		len(task.Articles) != maxDraftArticles {
		t.Fatalf("unexpected task: %+v", task)
	}
	for _, article := range task.Articles {
		if article.Status != "draft" || !strings.Contains(article.Content, "第三方原文版权归原作者所有") ||
			!strings.Contains(article.Content, "https://example.com/news/") ||
			strings.Contains(article.Content, "AI HOT") {
			t.Fatalf("invalid draft: %+v", article)
		}
	}
	// Reinitialization is harmless and an existing date requires no upstream request.
	s2 := newTestService(t, db, Options{HTTPClient: &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("duplicate date must not collect")
		return nil, nil
	})}})
	again := generate(t, s2, task.Date)
	if again.ID != task.ID {
		t.Fatal("duplicate date created a different task")
	}
	assertCount(t, db, "articles", maxDraftArticles)
	list, err := s.List(ctx, 10, 0)
	if err != nil || len(list) != 1 || list[0].ID != task.ID {
		t.Fatalf("list: %+v %v", list, err)
	}
	if _, err := s.Get(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("get missing: %v", err)
	}
	if _, err := s.Generate(ctx, "2026-02-30"); err == nil {
		t.Fatal("invalid date accepted")
	}
	if _, err := s.List(ctx, 101, 0); err == nil {
		t.Fatal("unbounded list accepted")
	}
	if err := s.RunScheduler(ctx, nil); !errors.Is(err, ErrSchedulerDisabled) {
		t.Fatalf("scheduler unexpectedly enabled: %v", err)
	}
}

func TestETagCacheSurvivesRestartAndPollFloor(t *testing.T) {
	db := database(t)
	now := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	calls := 0
	data, _ := json.Marshal(fixture(8))
	client := &http.Client{Transport: transportFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if req.URL.String() != itemsURL || req.Header.Get("Cookie") != "" || req.Header.Get("Authorization") != "" {
			t.Fatal("wrong endpoint or credentials")
		}
		if calls == 1 {
			if req.Header.Get("If-None-Match") != "" {
				t.Fatal("unexpected initial ETag")
			}
			return response(200, string(data)), nil
		}
		if req.Header.Get("If-None-Match") != `"v1"` {
			t.Fatal("missing persisted ETag")
		}
		return response(304, ""), nil
	})}
	opts := Options{HTTPClient: client, Now: func() time.Time { return now }}
	s := newTestService(t, db, opts)
	first, err := s.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	s = newTestService(t, db, opts)
	if _, err := s.Collect(context.Background()); err != nil || calls != 1 {
		t.Fatalf("poll floor: calls=%d err=%v", calls, err)
	}
	now = now.Add(time.Minute)
	second, err := s.Collect(context.Background())
	if err != nil || calls != 2 || len(second.Items) != len(first.Items) || string(second.Attribution) != string(first.Attribution) {
		t.Fatalf("304 cache reuse: %+v calls=%d err=%v", second, calls, err)
	}
}

func TestCollectHTTPAnonymousETagAndRedirect(t *testing.T) {
	var mu sync.Mutex
	var requests []http.Header
	var paths []string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, r.Header.Clone())
		paths = append(paths, r.URL.RequestURI())
		call := len(requests)
		mu.Unlock()
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}
		switch call {
		case 1:
			w.Header().Set("ETag", `"http-v1"`)
			w.Header().Set("Set-Cookie", "session=private; Path=/")
			json.NewEncoder(w).Encode(fixture(2))
		case 2:
			w.WriteHeader(http.StatusNotModified)
		default:
			http.Redirect(w, r, "/must-not-follow", http.StatusFound)
		}
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := server.Client()
	baseTransport := client.Transport
	client.Transport = transportFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != itemsURL {
			t.Errorf("unexpected upstream URL: %s", req.URL)
		}
		// Route only the test's connection to localhost; production URL stays fixed.
		local := req.Clone(req.Context())
		copiedURL := *req.URL
		copiedURL.Host, copiedURL.Scheme = target.Host, target.Scheme
		local.URL = &copiedURL
		return baseTransport.RoundTrip(local)
	})
	client.Jar, err = cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	upstream, _ := url.Parse(itemsURL)
	client.Jar.SetCookies(upstream, []*http.Cookie{{Name: "credential", Value: "private"}})
	now := time.Now()
	db := database(t)
	s := newTestService(t, db, Options{HTTPClient: client, Now: func() time.Time { return now }})
	first, err := s.Collect(context.Background())
	if err != nil || len(first.Items) != 2 {
		t.Fatalf("HTTP collection: %+v %v", first, err)
	}
	now = now.Add(time.Minute)
	second, err := s.Collect(context.Background())
	if err != nil || len(second.Items) != 2 {
		t.Fatalf("HTTP 304: %+v %v", second, err)
	}
	now = now.Add(time.Minute)
	if _, err := s.Collect(context.Background()); err == nil || !strings.Contains(err.Error(), "302") {
		t.Fatalf("redirect was accepted: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 3 || requests[1].Get("If-None-Match") != `"http-v1"` {
		t.Fatalf("redirect followed or ETag missing: %+v", requests)
	}
	for i, header := range requests {
		if header.Get("Cookie") != "" || header.Get("Authorization") != "" ||
			paths[i] != "/api/v1/items?mode=selected&window=24h&limit=8" {
			t.Fatalf("invalid anonymous HTTP request: %s %+v", paths[i], header)
		}
	}
}

func TestCollectionRejectsInvalidResponses(t *testing.T) {
	data, _ := json.Marshal(fixture(9))
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"unexpected304", 304, ""},
		{"rateLimited", 429, `{"secret":"do not expose"}`},
		{"invalidJSON", 200, "invalid"},
		{"missingItems", 200, `{"schemaVersion":1}`},
		{"tooManyItems", 200, string(data)},
		{"tooLarge", 200, strings.Repeat("x", (1<<20)+1)},
		{"missingVersion", 200, `{"items":[]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := database(t)
			s := newTestService(t, db, Options{HTTPClient: &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
				return response(tc.status, tc.body), nil
			})}})
			if _, err := s.Collect(context.Background()); err == nil || strings.Contains(err.Error(), "secret") {
				t.Fatalf("invalid collection error: %v", err)
			}
			assertCount(t, db, "dailynews_cache", 0)
		})
	}
}

func TestCrossDateDeduplicationAllowsNoDrafts(t *testing.T) {
	db := database(t)
	s := newTestService(t, db, Options{})
	task := generate(t, s, "2026-10-12")
	if task.Articles[0].Title != "Title 0" {
		t.Fatalf("want Title 0, got %s", task.Articles[0].Title)
	}
	task, err := s.Generate(context.Background(), "2026-10-13")
	if err != nil || len(task.Articles) != 3 || task.Articles[0].Title != "Title 5" {
		t.Fatalf("want remaining drafts on second day: %+v %v", task, err)
	}
	if task, err := s.Generate(context.Background(), "2026-10-14"); err != nil || len(task.Articles) != 0 {
		t.Fatalf("want no duplicate drafts: %+v %v", task, err)
	}
	assertCount(t, db, "articles", 8)
	assertCount(t, db, "dailynews_tasks", 3)
	tasks, err := s.List(context.Background(), 2, 1)
	if err != nil || len(tasks) != 2 || tasks[0].Date != "2026-10-13" {
		t.Fatalf("pagination: %+v %v", tasks, err)
	}
}

func TestAppendItemsExtendsPendingTaskToFiveArticles(t *testing.T) {
	ctx := context.Background()
	db := database(t)
	snapshot := fixture(8)
	s := newTestService(t, db, Options{HTTPClient: fixtureClient(t, Snapshot{
		SchemaVersion: snapshot.SchemaVersion,
		Items:         snapshot.Items[:3],
		Attribution:   snapshot.Attribution,
	})})
	task := generate(t, s, "2026-10-12")
	if len(task.Articles) != 3 {
		t.Fatalf("want 3 initial articles, got %d", len(task.Articles))
	}
	task, err := s.AppendItems(ctx, task.ID, []Item{snapshot.Items[3], snapshot.Items[4], snapshot.Items[4]})
	if err != nil {
		t.Fatal(err)
	}
	if len(task.Articles) != 5 || len(task.Snapshot.Items) != 5 || !strings.Contains(task.Message, "Title 4") {
		t.Fatalf("append failed: %+v", task)
	}
	if strings.Contains(task.Message, "/insights/") {
		t.Fatal("preview should not contain published links")
	}
	if _, err := s.Review(ctx, task.ID, 1, true, ""); err != nil {
		t.Fatal(err)
	}
	task, err = s.Publish(ctx, task.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(task.Message, "https://voidlab.club/insights/") || !strings.Contains(task.Message, "原始来源") {
		t.Fatalf("published digest should prefer site links: %s", task.Message)
	}
	if _, err := s.AppendItems(ctx, task.ID, []Item{snapshot.Items[5]}); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("published task accepted append: %v", err)
	}
}

func TestPublishKeepsCoverBindingOutsideDailyNewsService(t *testing.T) {
	ctx := context.Background()
	db := database(t)
	seedDefaultGallery(t, db, []string{
		"https://voidlab.club/uploads/gallery-cover-1.jpg",
		"https://voidlab.club/uploads/gallery-cover-2.jpg",
	})
	s := newTestService(t, db, Options{})
	task := generate(t, s, "2026-10-12")
	if _, err := db.Exec(`UPDATE articles SET cover_url='https://voidlab.club/uploads/manual-cover.jpg' WHERE id=?`, task.Articles[0].ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Review(ctx, task.ID, 1, true, ""); err != nil {
		t.Fatal(err)
	}
	task, err := s.Publish(ctx, task.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(task.Articles) != maxDraftArticles {
		t.Fatalf("want %d published articles, got %d", maxDraftArticles, len(task.Articles))
	}
	if task.Articles[0].CoverURL != "https://voidlab.club/uploads/manual-cover.jpg" {
		t.Fatalf("existing cover should be preserved, got %s", task.Articles[0].CoverURL)
	}
	for index, article := range task.Articles[1:] {
		if article.CoverURL != "" {
			t.Fatalf("article %d should wait for formal cover binding, got %s", index+1, article.CoverURL)
		}
	}
}

func TestGenerationTransactionRollback(t *testing.T) {
	for _, count := range []int{8} {
		t.Run(fmt.Sprintf("count-%d", count), func(t *testing.T) {
			db := database(t)
			s := newTestService(t, db, Options{HTTPClient: fixtureClient(t, fixture(count))})
			if count == 8 {
				_, err := db.Exec(`CREATE TRIGGER fail_second BEFORE INSERT ON articles
WHEN NEW.title='Title 1' BEGIN SELECT RAISE(ABORT,'test failure'); END`)
				if err != nil {
					t.Fatal(err)
				}
			}
			if _, err := s.Generate(context.Background(), "2026-10-12"); err == nil {
				t.Fatal("generation should fail")
			}
			assertCount(t, db, "articles", 0)
			assertCount(t, db, "dailynews_tasks", 0)
			assertCount(t, db, "dailynews_task_articles", 0)
		})
	}
}

func TestConcurrentGenerationAcrossServices(t *testing.T) {
	db := database(t)
	s1 := newTestService(t, db, Options{})
	s2 := newTestService(t, db, Options{})
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s := s1
			if i%2 != 0 {
				s = s2
			}
			if _, err := s.Generate(context.Background(), "2026-10-12"); err != nil {
				t.Errorf("concurrent generation: %v", err)
			}
		}(i)
	}
	wg.Wait()
	assertCount(t, db, "dailynews_tasks", 1)
	assertCount(t, db, "articles", maxDraftArticles)
}

func TestReviewPublishAndSendRetry(t *testing.T) {
	ctx := context.Background()
	db := database(t)
	var uuids, texts []string
	s := newTestService(t, db, Options{Sender: senderFunc(func(ctx context.Context, text, uuid string) (string, error) {
		uuids, texts = append(uuids, uuid), append(texts, text)
		if len(uuids) == 1 {
			return "", errors.New("secret provider error")
		}
		return "message-1", nil
	})})
	task := generate(t, s, "2026-10-12")
	if _, err := s.Publish(ctx, task.ID, 1); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("unreviewed publish: %v", err)
	}
	if _, err := s.Send(ctx, task.ID); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("unpublished send: %v", err)
	}
	if _, err := s.Review(ctx, task.ID, 2, true, "approved"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Review(ctx, task.ID, 2, false, "reject"); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("second review accepted: %v", err)
	}
	task, err := s.Publish(ctx, task.ID, 3)
	if err != nil || task.Status != "published" || task.ReviewerID != 2 {
		t.Fatalf("publish: %+v %v", task, err)
	}
	for _, article := range task.Articles {
		if article.Status != "published" {
			t.Fatal("article not published")
		}
	}
	if _, err := s.Publish(ctx, task.ID, 3); err != nil {
		t.Fatalf("idempotent publish: %v", err)
	}
	task, err = s.Send(ctx, task.ID)
	if err == nil || task.SendStatus != "failed" || task.SendAttempts != 1 || !strings.Contains(task.LastError, "secret") {
		t.Fatalf("failed delivery: %+v %v", task, err)
	}
	// Retry after restart must preserve the exact text even if articles were edited.
	if _, err := db.Exec(`UPDATE articles SET title='Edited title'`); err != nil {
		t.Fatal(err)
	}
	s = newTestService(t, db, Options{Sender: s.sender})
	task, err = s.Send(ctx, task.ID)
	if err != nil || task.SendStatus != "sent" || task.MessageID != "message-1" || task.SendAttempts != 2 {
		t.Fatalf("retry: %+v %v", task, err)
	}
	if len(uuids) != 2 || uuids[0] != uuids[1] || texts[0] != texts[1] ||
		!strings.Contains(texts[0], "VOIDLAB 每日资讯") || !strings.Contains(texts[0], "https://voidlab.club/insights/") {
		t.Fatal("retry text/UUID or digest is incorrect")
	}
	if _, err := s.Send(ctx, task.ID); err != nil || len(uuids) != 2 {
		t.Fatal("sent task delivered again")
	}
}

func TestSparseDailyPreviewPublicationAndDelivery(t *testing.T) {
	for _, count := range []int{0, 1, maxDraftArticles} {
		t.Run(fmt.Sprintf("items-%d", count), func(t *testing.T) {
			ctx := context.Background()
			db := database(t)
			var delivered string
			s := newTestService(t, db, Options{HTTPClient: fixtureClient(t, fixture(count)),
				Sender: senderFunc(func(_ context.Context, text, _ string) (string, error) {
					delivered = text
					return "message", nil
				})})
			task, err := s.Generate(ctx, "2026-10-12")
			if err != nil || len(task.Articles) != count || task.Message == "" || task.Status != "pending" {
				t.Fatalf("sparse generation: %+v %v", task, err)
			}
			preview := task.Message
			if strings.Contains(preview, "/insights/") || strings.Contains(preview, "/insight/") ||
				strings.Contains(preview, "## 站内文章") {
				t.Fatal("unpublished preview contains internal article links")
			}
			if !strings.Contains(preview, "VOIDLAB 每日资讯") || !strings.Contains(preview, "过去 24 小时") ||
				!strings.Contains(preview, "第三方原文版权归原作者") {
				t.Fatalf("invalid Chinese preview: %s", preview)
			}
			if count == 0 && !strings.Contains(preview, "暂无精选资讯") {
				t.Fatal("empty day has no explanation")
			}
			if count > 0 && (!strings.Contains(preview, "1. Title 0") ||
				!strings.Contains(preview, "原文链接：https://example.com/news/0") ||
				!strings.Contains(task.Articles[0].Content, "### 来源") ||
				strings.Contains(preview, "AI HOT") ||
				strings.Contains(task.Articles[0].Content, "AI HOT")) {
				t.Fatal("missing Markdown sources")
			}
			list, err := s.List(ctx, 10, 0)
			if err != nil || list[0].Message != preview {
				t.Fatalf("list preview mismatch: %+v %v", list, err)
			}
			if _, err := s.Review(ctx, task.ID, 1, true, ""); err != nil {
				t.Fatal(err)
			}
			// Publication must use the latest title and slug, not the generation-time draft.
			if _, err := db.Exec(`UPDATE articles SET title='Changed',slug='updated-' || id`); err != nil {
				t.Fatal(err)
			}
			task, err = s.Publish(ctx, task.ID, 1)
			if err != nil {
				t.Fatal(err)
			}
			frozen := task.Message
			if count > 0 {
				for _, article := range task.Articles {
					want := fmt.Sprintf("站内链接：https://voidlab.club/insights/%s", article.Slug)
					if !strings.Contains(frozen, want) || article.Status != "published" {
						t.Fatalf("latest published article missing: %s", frozen)
					}
				}
			} else if frozen != preview {
				t.Fatal("empty day message changed at publication")
			}
			if _, err := db.Exec(`UPDATE articles SET title='After publication',slug='later-' || id`); err != nil {
				t.Fatal(err)
			}
			again, err := s.Publish(ctx, task.ID, 1)
			if err != nil || again.Message != frozen {
				t.Fatalf("repeated publication changed frozen message: %+v %v", again, err)
			}
			s = newTestService(t, db, Options{Sender: s.sender})
			task, err = s.Send(ctx, task.ID)
			if err != nil || delivered != task.CopyText || task.Message != frozen {
				t.Fatalf("delivery differs from published message: %+v %v", task, err)
			}
		})
	}
}

func TestSendTimeValidation(t *testing.T) {
	for _, sendTime := range []string{"25:00", "9:00", "12:60", "noon"} {
		if _, err := NewService(context.Background(), database(t), Options{SendTime: sendTime}); err == nil {
			t.Fatalf("invalid SendTime accepted: %s", sendTime)
		}
	}
	s := newTestService(t, database(t), Options{})
	if s.sendTime != "10:00" || s.enabled {
		t.Fatal("unexpected scheduler defaults")
	}
}

func TestLegacyPreviewBackfill(t *testing.T) {
	ctx := context.Background()
	db := database(t)
	s := newTestService(t, db, Options{})
	task := generate(t, s, "2026-10-12")
	if _, err := db.Exec(`UPDATE dailynews_tasks SET send_text='' WHERE id=?`, task.ID); err != nil {
		t.Fatal(err)
	}
	s = newTestService(t, db, Options{})
	detail, err := s.Get(ctx, task.ID)
	if err != nil || detail.Message != task.Message {
		t.Fatalf("legacy detail: %+v %v", detail, err)
	}
	list, err := s.List(ctx, 10, 0)
	if err != nil || list[0].Message != detail.Message {
		t.Fatalf("legacy list differs: %+v %v", list, err)
	}
	var persisted string
	if err := db.QueryRow(`SELECT send_text FROM dailynews_tasks WHERE id=?`, task.ID).Scan(&persisted); err != nil || persisted != task.Message {
		t.Fatalf("backfill not persisted: %s %v", persisted, err)
	}
}

func TestUnpublishedLegacyLinksRemovedWithoutRewritingFrozenMessage(t *testing.T) {
	ctx := context.Background()
	db := database(t)
	s := newTestService(t, db, Options{})
	task := generate(t, s, "2026-10-12")
	unsafe := "[Draft](https://voidlab.club/insight/unpublished)"
	if _, err := db.Exec(`UPDATE dailynews_tasks SET send_text=? WHERE id=?`, unsafe, task.ID); err != nil {
		t.Fatal(err)
	}
	for _, state := range []string{"pending", "approved", "rejected"} {
		if _, err := db.Exec(`UPDATE dailynews_tasks SET status=? WHERE id=?`, state, task.ID); err != nil {
			t.Fatal(err)
		}
		detail, err := s.Get(ctx, task.ID)
		list, listErr := s.List(ctx, 10, 0)
		if err != nil || listErr != nil || detail.Message != task.Message || list[0].Message != task.Message {
			t.Fatalf("legacy %s leaked unpublished link: %+v %v %v", state, detail, err, listErr)
		}
	}
	s = newTestService(t, db, Options{})
	var persisted string
	if err := db.QueryRow(`SELECT send_text FROM dailynews_tasks WHERE id=?`, task.ID).Scan(&persisted); err != nil || persisted != task.Message {
		t.Fatalf("legacy unsafe preview not migrated: %s %v", persisted, err)
	}
	if _, err := db.Exec(`UPDATE dailynews_tasks SET status='approved' WHERE id=?`, task.ID); err != nil {
		t.Fatal(err)
	}
	task, err := s.Publish(ctx, task.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	s = newTestService(t, db, Options{PublicURL: "https://different.example"})
	detail, err := s.Get(ctx, task.ID)
	if err != nil || detail.Message != task.Message {
		t.Fatalf("restart rewrote frozen message: %+v %v", detail, err)
	}
}

func TestAutomaticClaimCannotRetryFailedAttempt(t *testing.T) {
	ctx := context.Background()
	db := database(t)
	var mu sync.Mutex
	calls := 0
	s := newTestService(t, db, Options{Sender: senderFunc(func(context.Context, string, string) (string, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		return "", errors.New("failed")
	})})
	task := generate(t, s, "2026-10-12")
	if _, err := s.Review(ctx, task.ID, 1, true, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Publish(ctx, task.ID, 1); err != nil {
		t.Fatal(err)
	}
	s2 := newTestService(t, db, Options{Sender: s.sender})
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			service := s
			if i%2 != 0 {
				service = s2
			}
			// Simulates a scheduler that selected the ID before another attempt completed.
			service.send(ctx, task.ID, true)
		}(i)
	}
	wg.Wait()
	task, err := s.Get(ctx, task.ID)
	if err != nil || calls != 1 || task.SendAttempts != 1 || task.SendStatus != "failed" {
		t.Fatalf("automatic retries occurred: calls=%d task=%+v err=%v", calls, task, err)
	}
}

func TestSchedulerSendTimeGatesAndManualRetries(t *testing.T) {
	ctx := context.Background()
	db := database(t)
	now := time.Date(2026, 10, 12, 9, 0, 0, 0, time.FixedZone("Beijing", 8*3600))
	calls := 0
	s := newTestService(t, db, Options{SchedulerEnabled: true, SendTime: "10:30",
		Now: func() time.Time { return now },
		Sender: senderFunc(func(context.Context, string, string) (string, error) {
			calls++
			if calls == 1 {
				return "", errors.New("delivery failed")
			}
			return "message", nil
		})})
	task := generate(t, s, "2026-10-12")
	if err := s.runScheduled(ctx); err != nil || calls != 0 {
		t.Fatalf("before SendTime: calls=%d err=%v", calls, err)
	}
	now = now.Add(90 * time.Minute)
	if err := s.runScheduled(ctx); err != nil || calls != 0 {
		t.Fatalf("pending task was sent: %v", err)
	}
	if _, err := s.Review(ctx, task.ID, 1, true, ""); err != nil {
		t.Fatal(err)
	}
	if err := s.runScheduled(ctx); err != nil || calls != 0 {
		t.Fatalf("approved task was published/sent: %v", err)
	}
	if _, err := s.Publish(ctx, task.ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.runScheduled(ctx); err == nil || calls != 1 {
		t.Fatalf("first automatic send: calls=%d err=%v", calls, err)
	}
	s = newTestService(t, db, Options{SchedulerEnabled: true, SendTime: "10:30", Now: func() time.Time { return now }, Sender: s.sender})
	for i := 0; i < 3; i++ {
		if err := s.runScheduled(ctx); err != nil || calls != 1 {
			t.Fatalf("failed task automatically retried: calls=%d err=%v", calls, err)
		}
	}
	task, err := s.Send(ctx, task.ID)
	if err != nil || calls != 2 || task.SendStatus != "sent" {
		t.Fatalf("manual retry failed: %+v %v", task, err)
	}
	// A stale ambiguous send is also manual-only.
	if _, err := db.Exec(`UPDATE dailynews_tasks SET send_status='sending',send_started_at=0 WHERE id=?`, task.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.runScheduled(ctx); err != nil || calls != 2 {
		t.Fatalf("stale send automatically retried: calls=%d err=%v", calls, err)
	}
}

func TestSchedulerSendsOnlyTodayAndSkipsHistory(t *testing.T) {
	ctx := context.Background()
	db := database(t)
	now := time.Date(2026, 10, 12, 9, 59, 0, 0, time.FixedZone("Beijing", 8*3600))
	calls := 0
	s := newTestService(t, db, Options{Now: func() time.Time { return now }, SendTime: "10:00",
		Sender: senderFunc(func(context.Context, string, string) (string, error) {
			calls++
			return "message", nil
		})})
	task := generate(t, s, "2026-10-12")
	if _, err := s.Review(ctx, task.ID, 1, true, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Publish(ctx, task.ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.runScheduled(ctx); err != nil || calls != 0 {
		t.Fatalf("sent too early: %v", err)
	}
	now = now.Add(time.Minute)
	if err := s.runScheduled(ctx); err != nil || calls != 1 {
		t.Fatalf("not sent at SendTime: %v", err)
	}
	if err := s.runScheduled(ctx); err != nil || calls != 1 {
		t.Fatalf("successful delivery duplicated: %v", err)
	}
	task = generate(t, s, "2026-10-13")
	if _, err := s.Review(ctx, task.ID, 1, true, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Publish(ctx, task.ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.runScheduled(ctx); err != nil || calls != 1 {
		t.Fatalf("future date sent: %v", err)
	}
	now = now.Add(48 * time.Hour)
	for i := 0; i < 2; i++ {
		if err := s.runScheduled(ctx); err != nil || calls != 1 {
			t.Fatalf("historical task automatically delivered: calls=%d err=%v", calls, err)
		}
	}
	historical, err := s.Get(ctx, task.ID)
	if err != nil || historical.SendStatus != "pending" || historical.SendAttempts != 0 {
		t.Fatalf("historical task state changed: %+v %v", historical, err)
	}
	if sent, err := s.Send(ctx, task.ID); err != nil || calls != 2 || sent.SendStatus != "sent" {
		t.Fatalf("historical task manual delivery failed: %+v %v", sent, err)
	}
}

func TestPublishRollbackAndRejection(t *testing.T) {
	ctx := context.Background()
	db := database(t)
	s := newTestService(t, db, Options{})
	task := generate(t, s, "2026-10-12")
	if _, err := s.Review(ctx, task.ID, 1, false, "not ready"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Publish(ctx, task.ID, 1); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("rejected task published: %v", err)
	}
	task = generate(t, s, "2026-10-13")
	if _, err := s.Review(ctx, task.ID, 1, true, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE articles SET status='archived' WHERE id=?`, task.Articles[1].ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Publish(ctx, task.ID, 1); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("non-draft publish: %v", err)
	}
	task, err := s.Get(ctx, task.ID)
	if err != nil || task.Status != "approved" || task.Articles[0].Status != "draft" || strings.Contains(task.Message, "/insights/") {
		t.Fatalf("publish did not rollback: %+v %v", task, err)
	}
}

func TestConcurrentSendAndStaleLease(t *testing.T) {
	ctx := context.Background()
	db := database(t)
	started, release := make(chan struct{}), make(chan struct{})
	s := newTestService(t, db, Options{Sender: senderFunc(func(context.Context, string, string) (string, error) {
		close(started)
		<-release
		return "message", nil
	})})
	task := generate(t, s, "2026-10-12")
	if _, err := s.Review(ctx, task.ID, 1, true, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Publish(ctx, task.ID, 1); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := s.Send(ctx, task.ID); done <- err }()
	<-started
	s2 := newTestService(t, db, Options{Sender: s.sender})
	_, err := s2.Send(ctx, task.ID)
	close(release)
	if !errors.Is(err, ErrSendInProgress) {
		t.Fatalf("concurrent send was not blocked: %v", err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE dailynews_tasks SET send_status='sending',send_started_at=0 WHERE id=?`, task.ID); err != nil {
		t.Fatal(err)
	}
	s2.sender = senderFunc(func(_ context.Context, _ string, uuid string) (string, error) {
		if uuid != task.SendUUID {
			t.Fatal("stale lease changed UUID")
		}
		return "message", nil
	})
	if recovered, err := s2.Send(ctx, task.ID); err != nil || recovered.SendStatus != "sent" {
		t.Fatalf("stale lease recovery: %+v %v", recovered, err)
	}
}

func TestCanceledSendPersistsFailure(t *testing.T) {
	db := database(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := newTestService(t, db, Options{Sender: senderFunc(func(ctx context.Context, _, _ string) (string, error) {
		cancel()
		return "", ctx.Err()
	})})
	task := generate(t, s, "2026-10-12")
	if _, err := s.Review(ctx, task.ID, 1, true, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Publish(ctx, task.ID, 1); err != nil {
		t.Fatal(err)
	}
	task, err := s.Send(ctx, task.ID)
	if err == nil || task.SendStatus != "failed" {
		t.Fatalf("canceled send not persisted: %+v %v", task, err)
	}
}

func TestResetRemovesTaskAndArticlesAndAllowsRegenerate(t *testing.T) {
	ctx := context.Background()
	db := database(t)
	s := newTestService(t, db, Options{})
	first := generate(t, s, "2026-10-12")
	if len(first.Articles) != maxDraftArticles {
		t.Fatalf("want %d articles, got %d", maxDraftArticles, len(first.Articles))
	}
	deletedIDs, err := s.Reset(ctx, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(deletedIDs) != maxDraftArticles {
		t.Fatalf("want %d deleted ids, got %d", maxDraftArticles, len(deletedIDs))
	}
	if _, err := s.GetByDate(ctx, first.Date); !errors.Is(err, ErrNotFound) {
		t.Fatalf("reset should remove task: %v", err)
	}
	assertCount(t, db, "dailynews_tasks", 0)
	assertCount(t, db, "dailynews_task_articles", 0)
	assertCount(t, db, "articles", 0)
	second := generate(t, s, first.Date)
	if second.ID == first.ID {
		t.Fatal("reset should allow a fresh task to be created")
	}
	if len(second.Articles) != maxDraftArticles {
		t.Fatalf("want %d regenerated articles, got %d", maxDraftArticles, len(second.Articles))
	}
}

func TestResetRejectsSendingTask(t *testing.T) {
	db := database(t)
	s := newTestService(t, db, Options{})
	task := generate(t, s, "2026-10-12")
	if _, err := db.Exec(`UPDATE dailynews_tasks SET send_status='sending', send_started_at=1 WHERE id=?`, task.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Reset(context.Background(), task.ID); !errors.Is(err, ErrSendInProgress) {
		t.Fatalf("reset sending task: %v", err)
	}
	assertCount(t, db, "dailynews_tasks", 1)
	assertCount(t, db, "articles", maxDraftArticles)
}

func TestExistingStorageSchemaAndScheduler(t *testing.T) {
	db, err := storage.OpenSQLite(filepath.Join(t.TempDir(), "existing.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 10, 11, 17, 0, 0, 0, time.UTC) // Beijing 2026-10-12 01:00.
	s := newTestService(t, db, Options{Now: func() time.Time { return now },
		SchedulerEnabled: true, ScheduleTime: "00:00"})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.RunScheduler(ctx, func(err error) { t.Errorf("scheduler: %v", err) }) }()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := s.GetByDate(context.Background(), "2026-10-12"); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("scheduler exit: %v", err)
	}
	task, err := s.GetByDate(context.Background(), "2026-10-12")
	if err != nil || task.Status != "pending" || task.SendAttempts != 0 {
		t.Fatalf("scheduler must only generate drafts: %+v %v", task, err)
	}
}
