package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Address     AddressConfig
	DB          DBConfig
	Auth        AuthConfig
	CORS        CORSConfig
	Redis       RedisConfig
	RateLimiter RateLimiterConfig
	Env         string
	ApiUrl      string
}

type AddressConfig struct {
	Port string
}

type DBConfig struct {
	URL string
}

type AuthConfig struct {
	JWTSecret string
}

type CORSConfig struct {
	AllowedOrigins string
}

type RedisConfig struct {
	Addr string
	PW   string
	DB   int
	TTL  time.Duration
}

type RateLimiterConfig struct {
	Enabled              bool
	RequestsPerTimeFrame int
	TimeFrame            time.Duration
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	return &Config{
		Address: AddressConfig{
			Port: getEnv("PORT", "8080"),
		},
		DB: DBConfig{
			URL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable"),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:5174"),
		},
		Redis: RedisConfig{
			Addr: getEnv("REDIS_ADDR", "localhost:6379"),
			PW:   getEnv("REDIS_PASSWORD", ""),
			DB:   getEnvInt("REDIS_DB", 0),
			TTL:  getEnvDuration("CACHE_TTL", 5*time.Minute),
		},
		RateLimiter: RateLimiterConfig{
			Enabled:              getEnvBool("RATE_LIMITER_ENABLED", true),
			RequestsPerTimeFrame: getEnvInt("RATE_LIMITER_REQUESTS", 50),
			TimeFrame:            getEnvDuration("RATE_LIMITER_WINDOW", 1*time.Minute),
		},
		Env:    getEnv("ENV", "development"),
		ApiUrl: getEnv("API_URL", "localhost:8080"),
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		var i int
		if _, err := parseInt(val, &i); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1"
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return fallback
}

func parseInt(s string, i *int) (string, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return s, nil
		}
		n = n*10 + int(c-'0')
	}
	*i = n
	return s, nil
}
