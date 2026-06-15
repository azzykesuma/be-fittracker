package config

import (
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                    string
	HTTPAddr               string
	DatabaseURL            string
	RedisAddr              string
	RedisPassword          string
	RedisDB                int
	CORSAllowedOrigins     []string
	SupabaseURL            string
	SupabaseSecretKey      string
	JWTSecret              string
	AuthPasswordPrivateKey string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Env:                    getEnv("APP_ENV", "development"),
		HTTPAddr:               getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:            normalizeDatabaseURL(getEnv("DATABASE_URL", "postgres://fitflow:fitflow_password@localhost:5432/fitflow_db?sslmode=disable")),
		RedisAddr:              getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:          getEnv("REDIS_PASSWORD", ""),
		RedisDB:                getEnvInt("REDIS_DB", 0),
		CORSAllowedOrigins:     splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
		SupabaseURL:            getEnv("SUPABASE_URL", ""),
		SupabaseSecretKey:      getEnv("SUPABASE_SECRET_KEY", ""),
		JWTSecret:              getEnv("JWT_SECRET", "change-me-in-development"),
		AuthPasswordPrivateKey: getEnv("AUTH_PASSWORD_PRIVATE_KEY", ""),
	}
}

func normalizeDatabaseURL(value string) string {
	parsed, err := url.Parse(value)
	if err != nil || !strings.Contains(parsed.Hostname(), "supabase") {
		return value
	}

	query := parsed.Query()
	if query.Get("sslmode") == "" {
		query.Set("sslmode", "require")
	}
	if query.Get("connect_timeout") == "" {
		query.Set("connect_timeout", "15")
	}
	parsed.RawQuery = query.Encode()

	return parsed.String()
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}
