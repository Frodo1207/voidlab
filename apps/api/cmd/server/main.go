package main

import (
	"context"
	"log"
	"os"

	"voidlabai/apps/api/internal/config"
	"voidlabai/apps/api/internal/dailynews"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sendTime := os.Getenv("DAILY_NEWS_SEND_TIME")
	if sendTime == "" {
		sendTime = "10:30"
	}
	news, err := dailynews.NewService(ctx, db, dailynews.Options{
		Sender:           &dailynews.FeishuSender{},
		PublicURL:        os.Getenv("DAILY_NEWS_SITE_URL"),
		SchedulerEnabled: os.Getenv("DAILY_NEWS_ENABLED") == "true",
		ScheduleTime:     os.Getenv("DAILY_NEWS_COLLECT_TIME"),
		SendTime:         sendTime,
	})
	if err != nil {
		log.Fatalf("bootstrap daily news failed: %v", err)
	}
	if os.Getenv("DAILY_NEWS_ENABLED") == "true" {
		go func() {
			if err := news.RunScheduler(ctx, func(err error) {
				log.Printf("daily news scheduler: %v", err)
			}); err != nil && ctx.Err() == nil {
				log.Printf("daily news scheduler stopped: %v", err)
			}
		}()
	}
	router := apihttp.NewRouter(cfg, db, news)

	log.Printf("voidlab api listening on %s", cfg.Address())
	if err := router.Run(cfg.Address()); err != nil {
		log.Fatalf("start api failed: %v", err)
	}
}
