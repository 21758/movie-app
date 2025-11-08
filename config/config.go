package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	BaseURL         string
	AuthToken       string
	DBURL           string
	BoxOfficeURL    string
	BoxOfficeAPIKey string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:            getEnv("PORT", "8080"),
		BaseURL:         getEnv("BASE_URL", "http://127.0.0.1:8080"),
		AuthToken:       getEnv("AUTH_TOKEN", ""),
		DBURL:           getEnv("DB_URL", ""),
		BoxOfficeURL:    getEnv("BOXOFFICE_URL", ""),
		BoxOfficeAPIKey: getEnv("BOXOFFICE_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
