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

	// Local Storage
	StorageBasePath  string
	StoragePublicURL string

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
		AppEnv:           getEnv("APP_ENV", "development"),
		AppPort:          getEnv("APP_PORT", "8080"),
		AppSecret:        getEnv("APP_SECRET", "supersecretjwtkey_min_32_characters_long_sidak_2026"),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", "postgres"),
		DBName:           getEnv("DB_NAME", "sidak_db"),
		DBSSLMode:        getEnv("DB_SSLMODE", "disable"),
		StorageBasePath:  getEnv("STORAGE_BASE_PATH", "./uploads"),
		StoragePublicURL: getEnv("STORAGE_PUBLIC_URL", "http://localhost:8080/uploads"),
		GotenbergURL:     getEnv("GOTENBERG_URL", "http://localhost:3000"),
		WebPublicURL:     getEnv("WEB_PUBLIC_URL", "http://localhost:3001"),
		WebAdminURL:      getEnv("WEB_ADMIN_URL", "http://localhost:3001/admin"),
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
