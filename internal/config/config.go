package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env                    string
	HTTPAddr               string
	DatabaseURL            string
	RedisAddr              string
	RedisPassword          string
	RedisDB                int
	RedisRequired          bool
	CORSAllowedOrigins     []string
	SupabaseURL            string
	SupabaseSecretKey      string
	JWTSecret              string
	AuthPasswordPrivateKey string
}

func Load() Config {
	loadDotEnv(".env")

	return Config{
		Env:                getEnv("APP_ENV", "development"),
		HTTPAddr:           getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://fitflow:fitflow_password@localhost:5432/fitflow_db?sslmode=disable"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            getEnvInt("REDIS_DB", 0),
		RedisRequired:      getEnvBool("REDIS_REQUIRED", false),
		CORSAllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
		// SupabaseURL:            getEnv("SUPABASE_URL", ""),
		SupabaseSecretKey:      getEnv("SUPABASE_SECRET_KEY", ""),
		JWTSecret:              getEnv("JWT_SECRET", "change-me-in-development"),
		AuthPasswordPrivateKey: getEnv("AUTH_PASSWORD_PRIVATE_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	firstLine := true
	for scanner.Scan() {
		line := scanner.Text()
		if firstLine {
			line = strings.TrimPrefix(line, "\ufeff")
			firstLine = false
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}

		if len(value) >= 2 {
			first := value[0]
			last := value[len(value)-1]
			if (first == '\'' && last == '\'') || (first == '"' && last == '"') {
				value = value[1 : len(value)-1]
			}
		}

		_ = os.Setenv(key, value)
	}
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
