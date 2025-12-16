package config

import (
	"log"
	"os"
	"strconv"
)

type AppConfig struct {
	HTTP HTTPConfig
	DB   DBConfig
	Auth AuthConfig
	OTLP OTLPConfig
}

type HTTPConfig struct {
	Addr string
}

type DBConfig struct {
	DSN string
}

type AuthConfig struct {
	JWTSecret       string
	AccessDuration  int
	RefreshDuration int
}

type OTLPConfig struct {
	Endpoint    string
	ServiceName string
}

func Load() AppConfig {
	return AppConfig{
		HTTP: HTTPConfig{
			Addr: getEnv("HTTP_ADDR", ":8080"),
		},
		DB: DBConfig{
			DSN: getEnv("DB_DSN", "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"),
		},
		Auth: AuthConfig{
			JWTSecret:       getEnv("JWT_SECRET", "super-secret-key"),
			AccessDuration:  getEnvInt("JWT_ACCESS_MINUTES", 15),
			RefreshDuration: getEnvInt("JWT_REFRESH_DAYS", 7),
		},
		OTLP: OTLPConfig{
			Endpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
			ServiceName: getEnv("OTEL_SERVICE_NAME", "api-service"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return fallback
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Printf("Invalid integer for env %s, utilizing fallback: %v", key, err)
		return fallback
	}
	return val
}
