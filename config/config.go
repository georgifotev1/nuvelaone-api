package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	Env            string
	ApiUrl         string
	AllowedOrigins string
}

// Load reads .env (if present) and returns a Config.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "change-me-in-production"),
		Env:            getEnv("ENV", "development"),
		ApiUrl:         getEnv("API_URL", "localhost:8080"),
		AllowedOrigins: getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:5174"),
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
