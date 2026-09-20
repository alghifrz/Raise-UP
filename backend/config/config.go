package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds application configuration loaded from the environment.
type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	JWTExpiresIn       time.Duration
	CORSAllowedOrigins []string
	ShutdownTimeout    time.Duration

	WhatsApp WhatsAppConfig

	// Optional seed credentials (used by cmd/seed only).
	AdminEmail    string
	AdminPassword string
	AdminName     string
	AdminRole     string
}

// WhatsAppConfig holds Meta WhatsApp Cloud API settings.
// When required credentials are missing, Enabled is forced to false.
type WhatsAppConfig struct {
	Enabled            bool
	APIVersion         string
	PhoneNumberID      string
	BusinessAccountID  string
	AccessToken        string
	VerifyToken        string
	AppSecret          string
}

// Load reads configuration from environment variables.
// It optionally loads a .env file when present (cwd or parent directory).
func Load() (*Config, error) {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		JWTExpiresIn:    time.Hour,
		ShutdownTimeout: 10 * time.Second,
		AdminEmail:      os.Getenv("ADMIN_EMAIL"),
		AdminPassword:   os.Getenv("ADMIN_PASSWORD"),
		AdminName:       getEnv("ADMIN_NAME", "Administrator"),
		AdminRole:       getEnv("ADMIN_ROLE", "SUPER_ADMIN"),
		WhatsApp: WhatsAppConfig{
			Enabled:           parseBoolEnv("WHATSAPP_ENABLED", false),
			APIVersion:        getEnv("WHATSAPP_API_VERSION", "v21.0"),
			PhoneNumberID:     strings.TrimSpace(os.Getenv("WHATSAPP_PHONE_NUMBER_ID")),
			BusinessAccountID: strings.TrimSpace(os.Getenv("WHATSAPP_BUSINESS_ACCOUNT_ID")),
			AccessToken:       strings.TrimSpace(os.Getenv("WHATSAPP_ACCESS_TOKEN")),
			VerifyToken:       strings.TrimSpace(os.Getenv("WHATSAPP_VERIFY_TOKEN")),
			AppSecret:         strings.TrimSpace(os.Getenv("WHATSAPP_APP_SECRET")),
		},
	}

	origins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	cfg.CORSAllowedOrigins = splitAndTrim(origins)

	if timeoutStr := os.Getenv("SHUTDOWN_TIMEOUT_SECONDS"); timeoutStr != "" {
		seconds, err := strconv.Atoi(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid SHUTDOWN_TIMEOUT_SECONDS: %w", err)
		}
		cfg.ShutdownTimeout = time.Duration(seconds) * time.Second
	}

	if expires := os.Getenv("JWT_EXPIRES_IN"); expires != "" {
		d, err := parseDuration(expires)
		if err != nil {
			return nil, fmt.Errorf("invalid JWT_EXPIRES_IN: %w", err)
		}
		cfg.JWTExpiresIn = d
	}

	cfg.WhatsApp.Enabled = cfg.WhatsApp.Enabled && cfg.WhatsApp.IsConfigured()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// IsConfigured reports whether the minimum credentials for send/receive are present.
func (w WhatsAppConfig) IsConfigured() bool {
	return w.PhoneNumberID != "" &&
		w.AccessToken != "" &&
		w.VerifyToken != "" &&
		w.AppSecret != ""
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWTSecret) < 16 {
		return fmt.Errorf("JWT_SECRET must be at least 16 characters")
	}
	if c.Port == "" {
		return fmt.Errorf("PORT is required")
	}
	if c.JWTExpiresIn <= 0 {
		return fmt.Errorf("JWT_EXPIRES_IN must be positive")
	}
	return nil
}

func parseDuration(value string) (time.Duration, error) {
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}
	return time.ParseDuration(value)
}

func parseBoolEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
