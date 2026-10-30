package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port    string
	AppEnv  string
	BaseURL string
	AppName string

	// Database
	DatabaseURL       string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	// Auth & Security
	SessionSecret        string
	CSRFSecret           string
	CookieSecure         bool
	CookieSameSite       string
	SessionMaxAgeSeconds int
	DefaultAdminEmail    string
	DefaultAdminPassword string

	// SMTP / Email
	SMTPHost                string
	SMTPPort                int
	SMTPUser                string
	SMTPPassword            string
	SMTPFromEmail           string
	SMTPFromName            string
	NotificationRecipient   string
	OutboxWorkerIntervalSec int
	EmailMaxRetries         int

	// Rate Limiting
	RateLimitRPS            float64
	RateLimitBurst          int
	ContactRateLimitMinutes int
	MaxBodySizeBytes        int64
}

func LoadConfig() (*Config, error) {
	// Try loading .env from current or parent directories
	_ = godotenv.Load(".env", "../.env", "../../.env")

	cfg := &Config{
		Port:    getEnv("PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),
		BaseURL: getEnv("BASE_URL", "http://localhost:8080"),
		AppName: getEnv("APP_NAME", "Spot Solutions"),

		DatabaseURL:       getEnv("DATABASE_URL", ""),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres"),
		DBName:            getEnv("DB_NAME", "spot_solution_db"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		DBMaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),

		SessionSecret:        getEnv("SESSION_SECRET", "super-secret-session-key-32-chars-long!"),
		CSRFSecret:           getEnv("CSRF_SECRET", "super-secret-csrf-key-32-chars-long!"),
		CookieSecure:         getEnvAsBool("COOKIE_SECURE", false),
		CookieSameSite:       getEnv("COOKIE_SAME_SITE", "strict"),
		SessionMaxAgeSeconds: getEnvAsInt("SESSION_MAX_AGE_SECONDS", 86400),
		DefaultAdminEmail:    getEnv("DEFAULT_ADMIN_EMAIL", "admin@nexoradynamics.com"),
		DefaultAdminPassword: getEnv("DEFAULT_ADMIN_PASSWORD", "AdminPass123!"),

		SMTPHost:                getEnv("SMTP_HOST", "localhost"),
		SMTPPort:                getEnvAsInt("SMTP_PORT", 1025),
		SMTPUser:                getEnv("SMTP_USER", getEnv("SMTP_USERNAME", "")),
		SMTPPassword:            getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail:           getEnv("SMTP_FROM_EMAIL", getEnv("SMTP_USER", getEnv("SMTP_USERNAME", "no-reply@spotsolutions.com"))),
		SMTPFromName:            getEnv("SMTP_FROM_NAME", "Spot Solutions"),
		NotificationRecipient:   getEnv("ADMIN_NOTIFICATION_EMAIL", getEnv("NOTIFICATION_RECIPIENT_EMAIL", getEnv("SMTP_USER", getEnv("SMTP_USERNAME", "leads@spotsolutions.com")))),
		OutboxWorkerIntervalSec: getEnvAsInt("EMAIL_OUTBOX_WORKER_INTERVAL_SEC", 15),
		EmailMaxRetries:         getEnvAsInt("EMAIL_MAX_RETRIES", 5),

		RateLimitRPS:            getEnvAsFloat("RATE_LIMIT_RPS", 5.0),
		RateLimitBurst:          getEnvAsInt("RATE_LIMIT_BURST", 10),
		ContactRateLimitMinutes: getEnvAsInt("CONTACT_RATE_LIMIT_MINUTES", 5),
		MaxBodySizeBytes:        getEnvAsInt64("MAX_BODY_SIZE_BYTES", 1048576), // 1MB
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	if c.DatabaseURL != "" {
		dsn := c.DatabaseURL
		// Enforce TLS/SSL for Neon connections if not explicitly set
		if strings.Contains(dsn, "neon.tech") && !strings.Contains(dsn, "sslmode=") {
			separator := "?"
			if strings.Contains(dsn, "?") {
				separator = "&"
			}
			dsn += separator + "sslmode=require"
		}
		return dsn
	}

	if c.DBHost == "" || c.DBUser == "" || c.DBName == "" {
		return ""
	}

	sslMode := c.DBSSLMode
	if strings.Contains(c.DBHost, "neon.tech") && (sslMode == "" || sslMode == "disable") {
		sslMode = "require"
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, sslMode,
	)
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsInt64(key string, defaultVal int64) int64 {
	valStr := getEnv(key, "")
	if val, err := strconv.ParseInt(valStr, 10, 64); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsFloat(key string, defaultVal float64) float64 {
	valStr := getEnv(key, "")
	if val, err := strconv.ParseFloat(valStr, 64); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valStr := getEnv(key, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valStr := getEnv(key, "")
	if val, err := time.ParseDuration(valStr); err == nil {
		return val
	}
	return defaultVal
}
