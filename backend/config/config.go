package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL       string
	JWTSecret         string
	EncryptionKey     []byte // exactly 32 bytes for AES-256
	Port              string
	Environment       string
	RateLimitRequests int
	RateLimitWindow   int
}

// Load reads environment variables and fails fast if anything required is missing or invalid.
func Load() (*Config, error) {
	// Load .env in development; ignore error if file doesn't exist
	_ = godotenv.Load()

	cfg := &Config{}

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}
	if len(cfg.JWTSecret) < 32 {
		fmt.Println("WARNING: JWT_SECRET should be at least 32 characters")
	}

	encKeyStr := os.Getenv("ENCRYPTION_KEY")
	if encKeyStr == "" {
		return nil, errors.New("ENCRYPTION_KEY is required")
	}
	key, err := base64.StdEncoding.DecodeString(encKeyStr)
	if err != nil {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be valid base64: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("ENCRYPTION_KEY must decode to exactly 32 bytes (got %d)", len(key))
	}
	cfg.EncryptionKey = key

	cfg.Port = os.Getenv("PORT")
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	cfg.Environment = os.Getenv("ENVIRONMENT")
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}

	cfg.RateLimitRequests = 5
	if v := os.Getenv("RATE_LIMIT_REQUESTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.RateLimitRequests = n
		}
	}

	cfg.RateLimitWindow = 60
	if v := os.Getenv("RATE_LIMIT_WINDOW"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.RateLimitWindow = n
		}
	}

	return cfg, nil
}
