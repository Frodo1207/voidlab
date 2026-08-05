package main

import (
	"log"

	"voidlabai/apps/api/internal/config"
	apihttp "voidlabai/apps/api/internal/http"
	"voidlabai/apps/api/internal/storage"
)

func main() {
	cfg := config.Load()

	db, err := storage.OpenSQLite(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("bootstrap sqlite failed: %v", err)
	}
	defer db.Close()

	router := apihttp.NewRouter(cfg, db)

	log.Printf("voidlab api listening on %s", cfg.Address())
	if err := router.Run(cfg.Address()); err != nil {
		log.Fatalf("start api failed: %v", err)
	}
}
