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

	healthH := handler.NewHealthHandler(&dbPingAdapter{db: db})
	r.Get("/health", healthH.Health)

	userRepo := repo.NewUserRepo(db)
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	authH := handler.NewAuthHandler(authSvc)

	r.Route("/api/v1", func(api chi.Router) {
		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/register", authH.Register)
			auth.Post("/login", authH.Login)
			auth.Post("/refresh", authH.Refresh)
		})

		api.Group(func(protected chi.Router) {
			protected.Use(middleware.Auth(authSvc))

			soulRepo := repo.NewSoulRepo(db)
			soulSvc := service.NewSoulService(soulRepo)
			soulH := handler.NewSoulHandler(soulSvc)

			protected.Route("/souls", func(sr chi.Router) {
				sr.Post("/", soulH.Create)
				sr.Get("/", soulH.List)
				sr.Get("/{soulID}", soulH.Get)
				sr.Put("/{soulID}", soulH.Update)
				sr.Delete("/{soulID}", soulH.Delete)

				msgRepo := repo.NewMessageRepo(db)
				chatSvc := service.NewChatService(msgRepo, soulRepo, cfg.DeepSeekURL, cfg.DeepSeekAPIKey)
				ttsSvc := service.NewTTSService(cfg.SiliconFlowURL, cfg.SiliconFlowKey)
				chatH := handler.NewChatHandler(chatSvc, ttsSvc)

				sr.Post("/{soulID}/chat", chatH.Send)
				sr.Get("/{soulID}/messages", chatH.History)
				sr.Post("/{soulID}/speak", chatH.Speak)
			})
		})
	})

	// Serve embedded static files (index.html, style.css, app.js)
	staticFS, _ := fs.Sub(web.StaticFS, ".")
	fileServer := http.FileServer(http.FS(staticFS))
	r.Handle("/*", fileServer)

	return r
}
