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
	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/handler"
	"github.com/georgifotev1/nuvelaone-api/internal/middleware"
	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/internal/tasks"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
	"github.com/georgifotev1/nuvelaone-api/pkg/mailer"
	"github.com/georgifotev1/nuvelaone-api/pkg/ratelimiter"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/hibiken/asynq"
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
	mailer      mailer.Mailer
	logger      *zap.SugaredLogger
	taskClient  *asynq.Client
	taskServer  *asynq.Server
	scheduler   *asynq.Scheduler // added
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
	txManager := txmanager.NewTxManager(app.db)

	tenantRepo := repository.NewTenantRepository(app.db, c.Tenants)
	userRepo := repository.NewUserRepository(app.db, c.Users)
	tokenRepo := repository.NewTokenRepository(app.db)
	invitationRepo := repository.NewInvitationRepository(app.db)
	serviceRepo := repository.NewServiceRepository(app.db, c.Services)
	customerRepo := repository.NewCustomerRepository(app.db)
	eventRepo := repository.NewEventRepository(app.db)

	userSvc := service.NewUserService(userRepo)
	authSvc := service.NewAuthService(userRepo, tenantRepo, tokenRepo, txManager, app.taskClient, app.logger, service.AuthConfig{
		AccessSecret:    app.config.Auth.JWTSecret,
		AccessTokenTTL:  app.config.Auth.AccessTokenTTL,
		RefreshTokenTTL: app.config.Auth.RefreshTokenTTL,
	})
	invitationSvc := service.NewInvitationService(
		invitationRepo,
		userRepo,
		tenantRepo,
		txManager,
		app.taskClient,
		app.logger,
		service.InvitationConfig{
			InvitationExpiry: 48 * time.Hour,
			AppBaseURL:       app.config.AppBaseURL,
		})
	tenantSvc := service.NewTenantService(tenantRepo, app.logger)
	serviceSvc := service.NewServiceService(serviceRepo, txManager)
	customerSvc := service.NewCustomerService(customerRepo)
	eventSvc := service.NewEventService(eventRepo, serviceRepo, customerRepo, userRepo, txManager)

	userHandler := handler.NewUserHandler(userSvc)
	authHandler := handler.NewAuthHandler(authSvc, app.config.Auth.RefreshTokenTTL)
	invitationHandler := handler.NewInvitationHandler(invitationSvc)
	tenantHandler := handler.NewTenantHandler(tenantSvc)
	serviceHandler := handler.NewServiceHandler(serviceSvc)
	customerHandler := handler.NewCustomerHandler(customerSvc)
	eventHandler := handler.NewEventHandler(eventSvc)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			jsonutil.Write(w, http.StatusOK, map[string]string{"status": "ok"})
		})
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/api/v1/swagger/doc.json"),
		))

		r.Route("/auth", authHandler.Routes)

		r.Route("/me", func(r chi.Router) {
			r.Use(middleware.JWTAuth(app.config.Auth.JWTSecret))
			r.Get("/", userHandler.GetMe)
		})

		r.Route("/users", func(r chi.Router) {
			r.Use(middleware.JWTAuth(app.config.Auth.JWTSecret))
			r.Use(middleware.RequireRole(domain.RoleOwner, domain.RoleAdmin))
			r.Get("/", userHandler.List)
			r.Get("/{id}", userHandler.GetByID)
			r.Put("/{id}", userHandler.Update)
			r.Delete("/{id}", userHandler.Delete)
		})

		r.Route("/invitations", func(r chi.Router) {
			r.Post("/accept", invitationHandler.Accept)
			r.Group(func(r chi.Router) {
				r.Use(middleware.JWTAuth(app.config.Auth.JWTSecret))
				r.Use(middleware.RequireRole(domain.RoleOwner, domain.RoleAdmin))
				r.Post("/", invitationHandler.Create)
				r.Get("/", invitationHandler.List)
				r.Delete("/{id}", invitationHandler.Revoke)
				r.Post("/{id}/resend", invitationHandler.Resend)
			})
		})

		r.Route("/tenants", func(r chi.Router) {
			r.Use(middleware.JWTAuth(app.config.Auth.JWTSecret))
			r.Get("/me", tenantHandler.GetMyTenant)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(domain.RoleOwner, domain.RoleAdmin))
				r.Put("/", tenantHandler.Update)
			})
		})

		r.Route("/services", func(r chi.Router) {
			r.Use(middleware.JWTAuth(app.config.Auth.JWTSecret))
			r.Get("/", serviceHandler.List)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(domain.RoleOwner, domain.RoleAdmin))
				r.Post("/", serviceHandler.Create)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", serviceHandler.GetByID)
					r.Put("/", serviceHandler.Update)
					r.Delete("/", serviceHandler.Delete)
				})
			})
		})

		r.Route("/customers", func(r chi.Router) {
			r.Use(middleware.JWTAuth(app.config.Auth.JWTSecret))
			r.Get("/", customerHandler.List)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(domain.RoleOwner, domain.RoleAdmin))
				r.Post("/", customerHandler.Create)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", customerHandler.GetByID)
					r.Put("/", customerHandler.Update)
					r.Delete("/", customerHandler.Delete)
				})
			})
		})

		r.Route("/events", func(r chi.Router) {
			r.Use(middleware.JWTAuth(app.config.Auth.JWTSecret))
			r.Get("/", eventHandler.List)
			r.Post("/", eventHandler.Create)
			r.Put("/{id}", eventHandler.Update)
		})
	})

	return r
}

func (app *application) run() error {
	go func() {
		if err := app.taskServer.Run(tasks.Register(tasks.HandlerDeps{
			Mailer:    app.mailer,
			TokenRepo: repository.NewTokenRepository(app.db),
			Logger:    app.logger,
		})); err != nil {
			app.logger.Errorw("task server error", "error", err)
		}
	}()

	if _, err := app.scheduler.Register("0 0 * * *", tasks.NewCleanupExpiredTokensTask()); err != nil {
		app.logger.Fatalw("failed to register cleanup task", "error", err)
	}

	go func() {
		if err := app.scheduler.Run(); err != nil {
			app.logger.Errorw("scheduler error", "error", err)
		}
	}()
	srv := &http.Server{
		Addr:         ":" + app.config.Address.Port,
		Handler:      app.mount(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		app.logger.Infow("server starting", "port", app.config.Address.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.logger.Fatalw("server error", "error", err)
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
	app.logger.Info("http server stopped")

	app.taskServer.Shutdown()
	app.logger.Info("task server stopped")

	app.scheduler.Shutdown()
	app.logger.Info("scheduler stopped")

	return nil
}
