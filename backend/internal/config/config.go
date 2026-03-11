package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port        string
	Environment string
	LogLevel    string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// Auth
	JWTSecret string

	// CORS
	CORSOrigins []string

	// Firebase Cloud Messaging
	FCMCredentialsFile string
	FCMCredentialsJSON string
}

func Load() (*Config, error) {
	// Load .env file if it exists (non-fatal if missing)
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		CORSOrigins:        []string{getEnv("CORS_ORIGINS", "http://localhost:3000")},
		FCMCredentialsFile: os.Getenv("FCM_CREDENTIALS_FILE"),
		FCMCredentialsJSON: os.Getenv("FCM_CREDENTIALS_JSON"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
