package router

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/wozai/wozai/internal/config"
	"github.com/wozai/wozai/internal/handler"
	"github.com/wozai/wozai/internal/middleware"
	"github.com/wozai/wozai/internal/repo"
	"github.com/wozai/wozai/internal/service"
	"github.com/wozai/wozai/web"
)

type dbPingAdapter struct {
	db *sql.DB
}

func (a *dbPingAdapter) Ping(ctx context.Context) error {
	return a.db.PingContext(ctx)
}

// New creates and configures the application router.
func New(db *sql.DB, cfg *config.Config) chi.Router {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.SecureHeaders)
	r.Use(middleware.MaxBodySize(1 << 20)) // 1MB

	limiter := middleware.NewRateLimiter(10, 60, 1*time.Second)
	r.Use(limiter.Middleware)

	// Repositories
	userRepo := repo.NewUserRepo(db)
	soulRepo := repo.NewSoulRepo(db)
	msgRepo := repo.NewMessageRepo(db)
	soulHistoryRepo := repo.NewSoulHistoryRepo(db)
	auditLogRepo := repo.NewAuditLogRepo(db)

	// Audit logger
	auditLogger := middleware.NewAuditLogger(auditLogRepo)

	// Services
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	// Multi-model AI providers
	providers := make(map[string]service.AIProvider)
	providers["deepseek"] = service.NewDeepSeekProvider(cfg.DeepSeekURL, cfg.DeepSeekAPIKey)
	if cfg.OpenAIAPIKey != "" {
		providers["openai"] = service.NewOpenAIProvider(cfg.OpenAIURL, cfg.OpenAIAPIKey, "")
	}
	if cfg.ZhipuAPIKey != "" {
		providers["zhipu"] = service.NewZhipuProvider(cfg.ZhipuURL, cfg.ZhipuAPIKey)
	}
	if cfg.GemmaAPIKey != "" {
		providers["gemma"] = service.NewGemmaProvider(cfg.GemmaURL, cfg.GemmaAPIKey)
	}

	chatSvc := service.NewChatService(msgRepo, soulRepo, providers, cfg.AIProvider, cfg.EnableSentiment)
	soulSvc := service.NewSoulService(soulRepo, soulHistoryRepo)
	ttsSvc := service.NewTTSService(cfg.SiliconFlowURL, cfg.SiliconFlowKey)
	profileSvc := service.NewProfileService(userRepo)

	// Handlers
	healthH := handler.NewHealthHandler(&dbPingAdapter{db: db})
	authH := handler.NewAuthHandler(authSvc, auditLogger)
	soulH := handler.NewSoulHandler(soulSvc, auditLogger)
	chatH := handler.NewChatHandler(chatSvc, ttsSvc)
	profileH := handler.NewProfileHandler(profileSvc)
	adminH := handler.NewAdminHandler(userRepo)

	// Routes
	r.Get("/health", healthH.Health)

	r.Route("/api/v1", func(api chi.Router) {
		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/register", authH.Register)
			auth.Post("/login", authH.Login)
			auth.Post("/refresh", authH.Refresh)
		})

		api.Group(func(protected chi.Router) {
			protected.Use(middleware.Auth(authSvc))

			// Profile
			protected.Get("/profile", profileH.GetProfile)
			protected.Put("/profile", profileH.UpdateProfile)

			// Stats
			protected.Get("/stats", adminH.UserStats)
			protected.Get("/admin/stats", adminH.Stats)

			// AI providers
			protected.Get("/providers", chatH.Providers)

			// Souls
			protected.Route("/souls", func(sr chi.Router) {
				sr.Post("/", soulH.Create)
				sr.Get("/", soulH.List)
				sr.Get("/{soulID}", soulH.Get)
				sr.Put("/{soulID}", soulH.Update)
				sr.Delete("/{soulID}", soulH.Delete)

				sr.Get("/{soulID}/history", soulH.EditHistory)
				sr.Post("/{soulID}/chat", chatH.Send)
				sr.Get("/{soulID}/messages", chatH.History)
				sr.Post("/{soulID}/speak", chatH.Speak)
			})
		})
	})

	// Serve embedded static files (index.html, style.css, app.js, manifest.json, sw.js)
	staticFS, _ := fs.Sub(web.StaticFS, ".")
	fileServer := http.FileServer(http.FS(staticFS))
	r.Handle("/*", fileServer)

	return r
}
