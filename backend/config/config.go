package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv    string
	AppPort   string
	AppSecret string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Cloudflare R2
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
	R2PublicURL       string

	// Gotenberg
	GotenbergURL string

	// Web Clients
	WebPublicURL string
	WebAdminURL  string
}

func LoadConfig() (*Config, error) {
	// Try loading from .env if present
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	cfg := &Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		AppPort:           getEnv("APP_PORT", "8080"),
		AppSecret:         getEnv("APP_SECRET", "supersecretjwtkey_min_32_characters_long_sidak_2026"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres"),
		DBName:            getEnv("DB_NAME", "sidak_db"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		R2AccountID:       getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:      getEnv("R2_BUCKET_NAME", "sidak-storage"),
		R2PublicURL:       getEnv("R2_PUBLIC_URL", "https://cdn.kelurahan.go.id"),
		GotenbergURL:      getEnv("GOTENBERG_URL", "http://localhost:3000"),
		WebPublicURL:      getEnv("WEB_PUBLIC_URL", "http://localhost:3001"),
		WebAdminURL:       getEnv("WEB_ADMIN_URL", "http://localhost:3001/admin"),
	}

	return cfg, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
