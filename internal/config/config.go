package config

import "os"

type Config struct {
	Port      string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	DBSSLMode string
}

func Load() Config {
	return Config{
		Port:      env("APP_PORT", "8080"),
		DBHost:    env("DB_HOST", "localhost"),
		DBPort:    env("DB_PORT", "5432"),
		DBUser:    env("DB_USER", "postgres"),
		DBPass:    env("DB_PASSWORD", "postgres"),
		DBName:    env("DB_NAME", "org_structure"),
		DBSSLMode: env("DB_SSLMODE", "disable"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
