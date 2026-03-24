package main

import (
	"context"

	_ "github.com/georgifotev1/nuvelaone-api/docs"
	"github.com/hibiken/asynq"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"github.com/georgifotev1/nuvelaone-api/config"
	"github.com/georgifotev1/nuvelaone-api/pkg/database"
	"github.com/georgifotev1/nuvelaone-api/pkg/mailer"
	"github.com/georgifotev1/nuvelaone-api/pkg/ratelimiter"
	"github.com/georgifotev1/nuvelaone-api/pkg/redis"
)

// @title           NuvelaOne API
// @version         1.0.0
// @description     REST API for NuvelaOne application
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http https
// @accept          json
// @produce         json
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log := zap.Must(zap.NewProduction()).Sugar()
	defer log.Sync()

	db, err := database.Connect(cfg.DB.URL)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	log.Info("database connection established")

	redisOpt := redis.NewRedisOpts(cfg.Redis.Addr, cfg.Redis.PW, cfg.Redis.DB)
	redisClient := redis.NewRedisClient(redisOpt)
	defer redisClient.Close()
	log.Info("redis connection established")

	taskServer := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 10,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			log.Errorw("task failed", "type", task.Type(), "error", err)
		}),
	})

	scheduler := asynq.NewScheduler(redisOpt, nil)

	var rateLimiter ratelimiter.Limiter
	if cfg.RateLimiter.Enabled {
		rateLimiter = ratelimiter.NewFixedWindowLimiter(
			cfg.RateLimiter.RequestsPerTimeFrame,
			cfg.RateLimiter.TimeFrame,
		)
		log.Info("rate limiter enabled")
	}

	var mail mailer.Mailer
	var taskClient *asynq.Client
	if cfg.Resend.APIKey != "" && cfg.Resend.FromEmail != "" {
		mail = mailer.NewResendClient(
			cfg.Resend.APIKey,
			cfg.Resend.FromEmail,
			cfg.Resend.DevEmailOverride,
		)
		taskClient = asynq.NewClient(redisOpt)
		defer taskClient.Close()
	}

	app := &application{
		config:      cfg,
		db:          db,
		redis:       redisClient,
		rateLimiter: rateLimiter,
		mailer:      mail,
		logger:      log,
		taskClient:  taskClient,
		taskServer:  taskServer,
		scheduler:   scheduler,
	}

	if err := app.run(); err != nil {
		log.Fatal("server error", zap.Error(err))
	}
}
