package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port            string
	DBDriver        string
	SqliteDBPath    string
	JWTSecret       string
	JWTExpiry       time.Duration
	JWTIssuer       string
	CORSAllowOrigin string
	LogLevel        string
	LogFormat       string
	AnalyticsCleanupInterval          time.Duration
	AnalyticsProcessedOutboxRetention time.Duration
	AnalyticsDeadLetterRetention      time.Duration
	AnalyticsEventRetention           time.Duration
}

func LoadConfig() (Config, error) {
	cfg := Config{
		Port:            getOrDefault("PORT", "8080"),
		DBDriver:        getOrDefault("DB_DRIVER", "sqlite"),
		SqliteDBPath:    getOrDefault("SQLITE_DB_PATH", "./sessions.db"),
		JWTIssuer:       getOrDefault("JWT_ISSUER", "realtime-session-coordination"),
		CORSAllowOrigin: os.Getenv("CORS_ALLOW_ORIGIN"),
		LogLevel:        getOrDefault("LOG_LEVEL", "info"),
		LogFormat:       getOrDefault("LOG_FORMAT", "json"),
		AnalyticsCleanupInterval:          parseDurationOrDefault("ANALYTICS_CLEANUP_INTERVAL", 10*time.Minute),
		AnalyticsProcessedOutboxRetention: parseDurationOrDefault("ANALYTICS_PROCESSED_OUTBOX_RETENTION", 24*time.Hour),
		AnalyticsDeadLetterRetention:      parseDurationOrDefault("ANALYTICS_DEAD_LETTER_RETENTION", 7*24*time.Hour),
		AnalyticsEventRetention:           parseDurationOrDefault("ANALYTICS_EVENT_RETENTION", 14*24*time.Hour),
	}

	secret, err := requiredEnv("JWT_SECRET")
	if err != nil {
		return Config{}, err
	}
	cfg.JWTSecret = secret

	expiryHoursRaw, err := requiredEnv("JWT_EXPIRY_HOURS")
	if err != nil {
		return Config{}, err
	}

	expiryHours, convErr := strconv.Atoi(expiryHoursRaw)
	if convErr != nil || expiryHours <= 0 {
		return Config{}, fmt.Errorf("JWT_EXPIRY_HOURS must be a positive integer")
	}
	cfg.JWTExpiry = time.Duration(expiryHours) * time.Hour

	switch cfg.DBDriver {
	case "memory", "sqlite":
		// valid
	default:
		return Config{}, fmt.Errorf("DB_DRIVER must be 'memory' or 'sqlite'")
	}

	return cfg, nil
}

func requiredEnv(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func getOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func parseDurationOrDefault(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
