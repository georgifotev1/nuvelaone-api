package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/georgifotev1/nuvelaone-api/config"
	"github.com/georgifotev1/nuvelaone-api/internal/cache"
	"github.com/georgifotev1/nuvelaone-api/internal/handler"
	"github.com/georgifotev1/nuvelaone-api/internal/middleware"
	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
	"github.com/georgifotev1/nuvelaone-api/pkg/ratelimiter"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type application struct {
	config      *config.Config
	db          *pgxpool.Pool
	redis       *redis.Client
	rateLimiter ratelimiter.Limiter
	logger      *zap.SugaredLogger
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   strings.Split(app.config.CORS.AllowedOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(chiMiddleware.Timeout(60 * time.Second))

	if app.rateLimiter != nil {
		r.Use(middleware.RateLimit(app.rateLimiter))
	}

	c := cache.New(app.redis)

	userRepo := repository.NewUserRepository(app.db, c.Users)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			jsonutil.Write(w, http.StatusOK, map[string]string{"status": "ok"})
		})
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/api/v1/swagger/doc.json"),
		))

		// Protected routes — uncomment to enable JWT auth:
		// r.Use(middleware.JWTAuth(cfg.Auth.JWTSecret))
		r.Route("/users", userHandler.Routes)
	})

	return r
}

func (app *application) run() error {
	srv := &http.Server{
		Addr:         ":" + app.config.Address.Port,
		Handler:      app.mount(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		app.logger.Info("server starting on port " + app.config.Address.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.logger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.logger.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("forced shutdown: %w", err)
	}

	app.logger.Info("server stopped")
	return nil
}
