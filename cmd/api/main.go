package main

import (
	_ "github.com/georgifotev1/nuvelaone-api/docs"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"github.com/georgifotev1/nuvelaone-api/config"
	"github.com/georgifotev1/nuvelaone-api/pkg/database"
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

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	logger.Info("database connection established")

	app := &application{
		config: cfg,
		db:     db,
		logger: logger,
	}

	if err := app.run(); err != nil {
		logger.Fatal("server error", zap.Error(err))
	}
}
