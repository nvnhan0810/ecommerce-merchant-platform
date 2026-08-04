package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nvnhan0810/ecomerce-api/internal/platform/storage"
)

type Config struct {
	HTTPAddr    string
	Env         string
	CORSOrigins []string

	DatabaseURL string
	AutoMigrate bool
	Seed        bool

	JWTSecret              string
	JWTTTL                 time.Duration
	AdminBootstrapPassword string
	DeliveryWebhookSecret  string

	S3            storage.Config
	PublicAPIBase string
	WebBaseURL    string
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

	ttlHours := getenvInt("JWT_TTL_HOURS", 24)
	webBase := getenv("CORS_WEB", "https://ecomerce.nvnhan0810.com")

	return Config{
		HTTPAddr:               addr,
		Env:                    env,
		CORSOrigins:            origins,
		DatabaseURL:            buildDatabaseURL(),
		AutoMigrate:            getenvBool("DB_AUTO_MIGRATE", true),
		Seed:                   getenvBool("DB_SEED", true),
		JWTSecret:              getenv("JWT_SECRET", "ecomerce-dev-jwt-secret-change-me"),
		JWTTTL:                 time.Duration(ttlHours) * time.Hour,
		AdminBootstrapPassword: getenv("ADMIN_BOOTSTRAP_PASSWORD", "Admin@123456"),
		DeliveryWebhookSecret:  getenv("DELIVERY_WEBHOOK_SECRET", "delivery-webhook-dev-secret"),
		S3: storage.Config{
			Endpoint:     getenv("S3_ENDPOINT", "http://seaweedfs-s3.seaweedfs.svc.cluster.local:8333"),
			Region:       getenv("S3_REGION", "us-east-1"),
			AccessKey:    getenv("S3_ACCESS_KEY", ""),
			SecretKey:    getenv("S3_SECRET_KEY", ""),
			Bucket:       getenv("S3_BUCKET", ""),
			UsePathStyle: getenvBool("S3_USE_PATH_STYLE", true),
		},
		PublicAPIBase: strings.TrimRight(getenv("PUBLIC_API_BASE_URL", "https://ecomerce-api.nvnhan0810.com"), "/"),
		WebBaseURL:    strings.TrimRight(webBase, "/"),
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

func getenvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
