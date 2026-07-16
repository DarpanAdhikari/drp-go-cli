package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DBDriver   string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	AppPort    string
	JWTSecret  string
	AccessTokenExpiryMinutes int
	RefreshTokenExpiryDays   int
	EncryptionKey string
	CORSOrigins   []string
	LogLevel      string
	RateLimitEnabled bool
	RateLimitRPS     int
	RateLimitBurst   int
}

func Load() *Config {
	return &Config{
		DBDriver:   getenv("DB_DRIVER", "postgres"),
		DBHost:     getenv("DB_HOST", "127.0.0.1"),
		DBPort:     getenv("DB_PORT", "5432"),
		DBUser:     getenv("DB_USER", "/tmp/test-crud"),
		DBPassword: getenv("DB_PASSWORD", ""),
		DBName:     getenv("DB_NAME", "postgres"),
		DBSSLMode:  getenv("DB_SSLMODE", "disable"),
		AppPort:    getenv("APP_PORT", "8080"),
		JWTSecret:  getenv("JWT_SECRET", "change-me-to-a-long-random-secret"),
		AccessTokenExpiryMinutes: getenvInt("ACCESS_TOKEN_EXPIRY_MINUTES", 15),
		RefreshTokenExpiryDays:   getenvInt("REFRESH_TOKEN_EXPIRY_DAYS", 30),
		EncryptionKey:            getenv("ENCRYPTION_KEY", ""),
		CORSOrigins:              splitAndTrim(getenv("CORS_ORIGINS", "*"), ","),
		LogLevel:                 getenv("LOG_LEVEL", "info"),
		RateLimitEnabled:         getenvBool("RATE_LIMIT_ENABLED", true),
		RateLimitRPS:             getenvInt("RATE_LIMIT_RPS", 10),
		RateLimitBurst:           getenvInt("RATE_LIMIT_BURST", 20),
	}
}

func (c *Config) DSN() string {
	if c.DBDriver == "mysql" {
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?parseTime=true",
			c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
		)
	}
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v, err := strconv.Atoi(getenv(key, ""))
	if err != nil {
		return fallback
	}
	return v
}

func getenvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return strings.ToLower(v) == "true" || v == "1"
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
