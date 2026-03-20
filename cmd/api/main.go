package main

import (
	_ "github.com/georgifotev1/nuvelaone-api/docs"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"github.com/georgifotev1/nuvelaone-api/config"
	"github.com/georgifotev1/nuvelaone-api/pkg/database"
	"github.com/georgifotev1/nuvelaone-api/pkg/ratelimiter"
	"github.com/georgifotev1/nuvelaone-api/pkg/redis"
)

// @title           NuvelaOne API
// @version         1.0.0
// @description     REST API for NuvelaOne application
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	db, err := database.Connect(cfg.DB.URL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	logger.Info("database connection established")

	redisClient := redis.NewRedisClient(cfg.Redis.Addr, cfg.Redis.PW, cfg.Redis.DB)
	defer redisClient.Close()
	logger.Info("redis connection established")

	var rateLimiter ratelimiter.Limiter
	if cfg.RateLimiter.Enabled {
		rateLimiter = ratelimiter.NewFixedWindowLimiter(
			cfg.RateLimiter.RequestsPerTimeFrame,
			cfg.RateLimiter.TimeFrame,
		)
		logger.Info("rate limiter enabled")
	}

	app := &application{
		config:      cfg,
		db:          db,
		redis:       redisClient,
		rateLimiter: rateLimiter,
		logger:      logger,
	}

	if err := app.run(); err != nil {
		logger.Fatal("server error", zap.Error(err))
	}
}
