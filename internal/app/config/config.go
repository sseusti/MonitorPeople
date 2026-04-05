package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	HTTPPort         string
	DatabaseURL      string
	AdminLogin       string
	AdminPassword    string
	EntranceLogin    string
	EntrancePassword string
}

func Load() Config {
	return Config{
		HTTPPort:         envOrDefault("HTTP_PORT", "8080"),
		DatabaseURL:      defaultDatabaseURL(),
		AdminLogin:       envOrDefault("ADMIN_LOGIN", "admin"),
		AdminPassword:    envOrDefault("ADMIN_PASSWORD", "admin123"),
		EntranceLogin:    envOrDefault("ENTRANCE_LOGIN", "entrance"),
		EntrancePassword: envOrDefault("ENTRANCE_PASSWORD", "entrance123"),
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func defaultDatabaseURL() string {
	if value := strings.TrimSpace(os.Getenv("DATABASE_URL")); value != "" {
		return value
	}

	user := strings.TrimSpace(os.Getenv("USER"))
	if user == "" {
		user = "postgres"
	}
	return fmt.Sprintf("postgres://%s@localhost:5432/monitor_people?sslmode=disable", user)
}
