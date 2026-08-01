package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr    string
	Env         string
	CORSOrigins []string

	DatabaseURL string
	AutoMigrate bool
	Seed        bool
}

func Load() Config {
	addr := getenv("HTTP_ADDR", ":8080")
	env := getenv("APP_ENV", "development")
	origins := []string{
		getenv("CORS_WEB", "https://ecomerce.nvnhan0810.com"),
		getenv("CORS_MERCHANT", "https://ecomerce-merchant.nvnhan0810.com"),
		getenv("CORS_ADMIN", "https://ecomerce-admin.nvnhan0810.com"),
		"http://localhost:5173",
		"http://localhost:5174",
		"http://localhost:5175",
	}

	return Config{
		HTTPAddr:    addr,
		Env:         env,
		CORSOrigins: origins,
		DatabaseURL: buildDatabaseURL(),
		AutoMigrate: getenvBool("DB_AUTO_MIGRATE", true),
		Seed:        getenvBool("DB_SEED", true),
	}
}

func buildDatabaseURL() string {
	if raw := strings.TrimSpace(os.Getenv("DATABASE_URL")); raw != "" {
		return raw
	}
	host := getenv("DB_HOST", "127.0.0.1")
	port := getenv("DB_PORT", "5432")
	db := getenv("DB_DATABASE", "ecomerce")
	user := url.UserPassword(
		getenv("DB_USERNAME", "postgres"),
		os.Getenv("DB_PASSWORD"),
	).String()
	ssl := getenv("DB_SSLMODE", "disable")
	return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s", user, host, port, db, ssl)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
