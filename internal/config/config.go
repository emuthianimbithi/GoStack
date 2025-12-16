package config

import (
	"log"
	"os"
	"strconv"
)

type AppConfig struct {
	HTTP    HTTPConfig
	DB      DBConfig
	Auth    AuthConfig
	Email   EmailConfig
	OTLP    OTLPConfig
	Modules ModuleConfig
	Billing BillingConfig
}

type EmailConfig struct {
	SendGridKey string
	FromEmail   string
	FromName    string
}

type BillingConfig struct {
	StripeSecretKey     string
	StripeWebhookSecret string
	MpesaConsumerKey    string
	MpesaConsumerSecret string
	MpesaPassKey        string
	MpesaShortCode      string
	MpesaCallbackURL    string
}

type ModuleConfig struct {
	AuditEnabled bool
	// BillingEnabled bool // Future
	// NotificationsEnabled bool // Future
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
		Email: EmailConfig{
			SendGridKey: getEnv("SENDGRID_API_KEY", ""),
			FromEmail:   getEnv("EMAIL_FROM_ADDRESS", "no-reply@gostack.com"),
			FromName:    getEnv("EMAIL_FROM_NAME", "GoStack"),
		},
		Billing: BillingConfig{
			StripeSecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
			StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
			MpesaConsumerKey:    getEnv("MPESA_CONSUMER_KEY", ""),
			MpesaConsumerSecret: getEnv("MPESA_CONSUMER_SECRET", ""),
			MpesaPassKey:        getEnv("MPESA_PASSKEY", ""),
			MpesaShortCode:      getEnv("MPESA_SHORTCODE", ""),
			MpesaCallbackURL:    getEnv("MPESA_CALLBACK_URL", "https://api.gostack.com/api/v1/callbacks/mpesa"),
		},
		OTLP: OTLPConfig{
			Endpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
			ServiceName: getEnv("OTEL_SERVICE_NAME", "api-service"),
		},
		Modules: ModuleConfig{
			AuditEnabled: getEnvBool("ENABLE_AUDIT", true),
		},
	}
}

func getEnvBool(key string, fallback bool) bool {
	valStr := getEnv(key, "")
	if valStr == "" {
		return fallback
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		log.Printf("Invalid boolean for env %s, utilizing fallback: %v", key, err)
		return fallback
	}
	return val
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
