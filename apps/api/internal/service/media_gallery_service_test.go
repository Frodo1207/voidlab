package service

import (
	"database/sql"
	"path/filepath"
	"testing"

	"voidlabai/apps/api/internal/repository"
	"voidlabai/apps/api/internal/storage"

	_ "modernc.org/sqlite"
)

func TestOpenSQLiteSeedsDefaultMediaGalleryFromExistingImages(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "gallery-seed.db")

	seedDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer seedDB.Close()

	if _, err := seedDB.Exec(`
		CREATE TABLE media_assets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			object_key TEXT NOT NULL,
			object_url TEXT NOT NULL,
			file_name TEXT NOT NULL,
			content_type TEXT NOT NULL,
			file_size INTEGER NOT NULL DEFAULT 0,
			uploaded_by INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		t.Fatal(err)
	}

	if _, err := seedDB.Exec(`
		INSERT INTO media_assets (object_key, object_url, file_name, content_type, file_size)
		VALUES
			('img-1', 'https://voidlab.club/uploads/existing-1.jpg', 'existing-1.jpg', 'image/jpeg', 1024),
			('doc-1', 'https://voidlab.club/uploads/readme.txt', 'readme.txt', 'text/plain', 64)
	`); err != nil {
		t.Fatal(err)
	}

	opened, err := storage.OpenSQLite(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()

	var galleryCount int
	if err := opened.QueryRow(`SELECT COUNT(1) FROM media_galleries WHERE slug = 'base-gallery'`).Scan(&galleryCount); err != nil {
		t.Fatal(err)
	}
	if galleryCount != 1 {
		t.Fatalf("want base gallery seeded, got %d", galleryCount)
	}

	var itemCount int
	if err := opened.QueryRow(`
		SELECT COUNT(1)
		FROM media_gallery_items i
		INNER JOIN media_galleries g ON g.id = i.gallery_id
		WHERE g.slug = 'base-gallery'
	`).Scan(&itemCount); err != nil {
		t.Fatal(err)
	}
	if itemCount != 1 {
		t.Fatalf("want 1 image imported into base gallery, got %d", itemCount)
	}
}

func TestMediaGalleryServiceAddUpdateAndDeleteItem(t *testing.T) {
	db, err := storage.OpenSQLite(filepath.Join(t.TempDir(), "gallery-service.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	result, err := db.Exec(`
		INSERT INTO media_assets (object_key, object_url, file_name, content_type, file_size, uploaded_by)
		VALUES ('img-2', 'https://voidlab.club/uploads/cover-2.jpg', 'cover-2.jpg', 'image/jpeg', 2048, 1)
	`)
	if err != nil {
		t.Fatal(err)
	}
	assetID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	service := NewMediaGalleryService(repository.NewMediaGalleryRepository(db))
	galleries, err := service.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(galleries) == 0 {
		t.Fatal("want base gallery to exist")
	}
	galleryID := galleries[0].ID

	if err := service.AddItems(galleryID, MediaGalleryItemsAppendInput{MediaAssetIDs: []int64{assetID}}); err != nil {
		t.Fatal(err)
	}

	items, err := service.ListItems(galleryID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("want 1 gallery item, got %d", len(items))
	}

	item := items[0]
	if err := service.UpdateItem(galleryID, item.ID, MediaGalleryItemUpdateInput{
		AltText:   "科技场景封面",
		SortOrder: 30,
		IsActive:  false,
	}); err != nil {
		t.Fatal(err)
	}

	items, err = service.ListItems(galleryID)
	if err != nil {
		t.Fatal(err)
	}
	if items[0].AltText != "科技场景封面" || items[0].SortOrder != 30 || items[0].IsActive {
		t.Fatalf("unexpected gallery item after update: %+v", items[0])
	}

	if err := service.DeleteItem(galleryID, items[0].ID); err != nil {
		t.Fatal(err)
	}

	items, err = service.ListItems(galleryID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("want gallery items to be empty after delete, got %d", len(items))
	}
}
