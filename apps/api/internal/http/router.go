package http

import (
	"database/sql"
	stdhttp "net/http"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/config"
	"voidlabai/apps/api/internal/http/middleware"
	"voidlabai/apps/api/internal/repository"
	"voidlabai/apps/api/internal/service"
)

func NewRouter(cfg config.Config, db *sql.DB) *gin.Engine {
	router := gin.Default()
	router.Static("/uploads", cfg.UploadsDir)

	userRepo := repository.NewUserRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	eventRepo := repository.NewEventRepository(db)
	builderRepo := repository.NewBuilderRepository(db)
	mediaRepo := repository.NewMediaRepository(db)
	leadRepo := repository.NewLeadRepository(db)
	siteConfigRepo := repository.NewSiteConfigRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	agentTokenRepo := repository.NewAgentTokenRepository(db)
	knowledgeSpaceRepo := repository.NewKnowledgeSpaceRepository(db)
	knowledgeEntryRepo := repository.NewKnowledgeEntryRepository(db)
	knowledgeAccessTokenRepo := repository.NewKnowledgeAccessTokenRepository(db)
	knowledgeAccessLogRepo := repository.NewKnowledgeAccessLogRepository(db)
	knowledgeAssetRepo := repository.NewKnowledgeAssetRepository(db)

	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)
	agentTokenService := service.NewAgentTokenService(agentTokenRepo)
	auditLogService := service.NewAuditLogService(auditLogRepo)
	knowledgeService := service.NewKnowledgeService(knowledgeSpaceRepo, knowledgeEntryRepo, knowledgeAccessTokenRepo, knowledgeAccessLogRepo)
	dashboardService := service.NewDashboardService(articleRepo, eventRepo, builderRepo, leadRepo, auditLogRepo)
	authHandler := NewAuthHandler(authService)
	userHandler := NewUserHandler(userService, auditLogService)
	agentTokenHandler := NewAgentTokenHandler(agentTokenService, auditLogService)
	articleHandler := NewArticleHandler(service.NewArticleService(articleRepo), auditLogService)
	eventHandler := NewEventHandler(service.NewEventService(eventRepo), auditLogService)
	builderHandler := NewBuilderHandler(service.NewBuilderService(builderRepo), auditLogService)
	mediaService := service.NewMediaService(cfg, mediaRepo)
	mediaHandler := NewMediaHandler(mediaService, auditLogService)
	knowledgeAssetService := service.NewKnowledgeAssetService(cfg, knowledgeSpaceRepo, knowledgeAssetRepo, mediaService)
	knowledgeHandler := NewKnowledgeHandler(knowledgeService, knowledgeAssetService, auditLogService)
	leadHandler := NewLeadHandler(service.NewLeadService(leadRepo), auditLogService)
	siteConfigHandler := NewSiteConfigHandler(service.NewSiteConfigService(siteConfigRepo), auditLogService)
	auditLogHandler := NewAuditLogHandler(auditLogService)
	dashboardHandler := NewDashboardHandler(dashboardService)
	intakeHandler := NewIntakeHandler(
		service.NewLeadService(leadRepo),
		service.NewEventService(eventRepo),
		service.NewBuilderService(builderRepo),
	)

	router.GET("/healthz", func(ctx *gin.Context) {
		sqliteStatus := "ok"
		if err := db.Ping(); err != nil {
			sqliteStatus = err.Error()
		}

		ctx.JSON(stdhttp.StatusOK, gin.H{
			"service": "voidlab-api",
			"env":     cfg.Environment,
			"sqlite":  sqliteStatus,
			"minio": gin.H{
				"endpoint": cfg.MinioEndpoint,
				"bucket":   cfg.MinioBucket,
			},
		})
	})

	api := router.Group("/api/v1")
	{
		api.GET("/system/bootstrap", func(ctx *gin.Context) {
			ctx.JSON(stdhttp.StatusOK, gin.H{
				"message": "VOIDLAB Phase 1 API skeleton is ready",
				"modules": []string{
					"auth",
					"articles",
					"events",
					"builders",
					"media-assets",
					"leads",
					"knowledge-base",
				},
			})
		})

		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.RequireAuth(authService, agentTokenService), authHandler.Me)
			auth.POST("/logout", middleware.RequireAuth(authService, agentTokenService), authHandler.Logout)
		}

		public := api.Group("/public")
		{
			public.GET("/site-configs", siteConfigHandler.PublicList)
			public.GET("/articles", articleHandler.PublicList)
			public.GET("/articles/:slug", articleHandler.PublicDetail)
			public.GET("/events", eventHandler.PublicList)
			public.GET("/events/:slug", eventHandler.PublicDetail)
			public.GET("/builders", builderHandler.PublicList)
			public.GET("/builders/:slug", builderHandler.PublicDetail)
			public.GET("/knowledge/spaces", knowledgeHandler.PublicListSpaces)
			public.GET("/knowledge/spaces/:slug/toc", knowledgeHandler.PublicSpaceTOC)
			public.POST("/knowledge/spaces/:slug/verify-token", knowledgeHandler.PublicVerifyToken)
			public.GET("/knowledge/spaces/:slug/entries/:entrySlug", knowledgeHandler.PublicEntryDetail)
			public.GET("/knowledge/spaces/:slug/assets/:assetID", knowledgeHandler.PublicAsset)
		}

		api.POST("/contact/submit", intakeHandler.ContactSubmit)
		api.POST("/events/:id/rsvp", intakeHandler.EventRSVP)
		api.POST("/builders/inquiry", intakeHandler.BuilderInquiry)

		protected := api.Group("/")
		protected.Use(middleware.RequireAuth(authService, agentTokenService))
		{
			protected.GET("/dashboard/stats", dashboardHandler.Stats)

			protected.GET("/articles", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"articles:read", "articles:write"}), articleHandler.List)
			protected.GET("/articles/:id", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"articles:read", "articles:write"}), articleHandler.Detail)
			protected.POST("/articles", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"articles:write"}), articleHandler.Create)
			protected.PUT("/articles/:id", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"articles:write"}), articleHandler.Update)
			protected.PUT("/articles/:id/status", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"articles:write"}), articleHandler.UpdateStatus)
			protected.DELETE("/articles/:id", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"articles:write"}), articleHandler.Delete)

			protected.GET("/events", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"events:read", "events:write"}), eventHandler.List)
			protected.GET("/events/:id", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"events:read", "events:write"}), eventHandler.Detail)
			protected.POST("/events", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"events:write"}), eventHandler.Create)
			protected.PUT("/events/:id", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"events:write"}), eventHandler.Update)
			protected.PUT("/events/:id/status", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"events:write"}), eventHandler.UpdateStatus)
			protected.DELETE("/events/:id", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"events:write"}), eventHandler.Delete)

			protected.GET("/builders", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"builders:read", "builders:write"}), builderHandler.List)
			protected.GET("/builders/:id", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"builders:read", "builders:write"}), builderHandler.Detail)
			protected.POST("/builders", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"builders:write"}), builderHandler.Create)
			protected.PUT("/builders/:id", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"builders:write"}), builderHandler.Update)
			protected.PUT("/builders/:id/status", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"builders:write"}), builderHandler.UpdateStatus)
			protected.DELETE("/builders/:id", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"builders:write"}), builderHandler.Delete)

			protected.GET("/media", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"media:read", "media:write"}), mediaHandler.List)
			protected.POST("/media/upload", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"media:write"}), mediaHandler.Upload)

			protected.GET("/knowledge/spaces", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"knowledge:read", "knowledge:write"}), knowledgeHandler.ListSpaces)
			protected.POST("/knowledge/spaces", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"knowledge:write"}), knowledgeHandler.CreateSpace)
			protected.PUT("/knowledge/spaces/:id", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"knowledge:write"}), knowledgeHandler.UpdateSpace)
			protected.DELETE("/knowledge/spaces/:id", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"knowledge:write"}), knowledgeHandler.DeleteSpace)
			protected.GET("/knowledge/spaces/:id/assets", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"knowledge:read", "knowledge:write"}), knowledgeHandler.ListAssets)
			protected.POST("/knowledge/spaces/:id/assets", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"knowledge:write"}), knowledgeHandler.UploadAsset)
			protected.GET("/knowledge/entries", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"knowledge:read", "knowledge:write"}), knowledgeHandler.ListEntries)
			protected.POST("/knowledge/entries/import-markdown", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"knowledge:write"}), knowledgeHandler.ImportEntryMarkdown)
			protected.POST("/knowledge/entries", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"knowledge:write"}), knowledgeHandler.CreateEntry)
			protected.PUT("/knowledge/entries/:id", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"knowledge:write"}), knowledgeHandler.UpdateEntry)
			protected.DELETE("/knowledge/entries/:id", middleware.RequireRolesOrScopes([]string{"admin", "editor"}, []string{"knowledge:write"}), knowledgeHandler.DeleteEntry)
			protected.GET("/knowledge/access-tokens", middleware.RequireRolesOrScopes([]string{"admin"}, []string{"knowledge_tokens:read", "knowledge_tokens:write"}), knowledgeHandler.ListAccessTokens)
			protected.POST("/knowledge/access-tokens", middleware.RequireRolesOrScopes([]string{"admin"}, []string{"knowledge_tokens:write"}), knowledgeHandler.CreateAccessToken)
			protected.PUT("/knowledge/access-tokens/:id/status", middleware.RequireRolesOrScopes([]string{"admin"}, []string{"knowledge_tokens:write"}), knowledgeHandler.UpdateAccessTokenStatus)

			leadOperations := protected.Group("/")
			leadOperations.Use(middleware.RequireAnyRole("admin", "ops"))
			{
				leadOperations.GET("/leads", leadHandler.List)
				leadOperations.GET("/leads/:id", leadHandler.Detail)
				leadOperations.POST("/leads", leadHandler.Create)
				leadOperations.PUT("/leads/:id/status", leadHandler.UpdateStatus)
				leadOperations.POST("/leads/:id/logs", leadHandler.AddLog)
			}

			siteSettings := protected.Group("/")
			siteSettings.Use(middleware.RequireAnyRole("admin"))
			{
				siteSettings.GET("/agent-tokens", agentTokenHandler.List)
				siteSettings.POST("/agent-tokens", agentTokenHandler.Create)
				siteSettings.PUT("/agent-tokens/:id/status", agentTokenHandler.UpdateStatus)
				siteSettings.GET("/users", userHandler.List)
				siteSettings.POST("/users", userHandler.Create)
				siteSettings.PUT("/users/:id/role", userHandler.UpdateRole)
				siteSettings.PUT("/users/:id/status", userHandler.UpdateStatus)
				siteSettings.PUT("/users/:id/password", userHandler.ResetPassword)
				siteSettings.GET("/site-configs", siteConfigHandler.List)
				siteSettings.PUT("/site-configs/:key", siteConfigHandler.Update)
				siteSettings.GET("/audit-logs", auditLogHandler.List)
			}
		}
	}

	return router
}
