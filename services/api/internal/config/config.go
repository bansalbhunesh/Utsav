package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	HTTPPort                 string
	DatabaseURL              string
	MigrationsPath           string
	JWTSecret                string
	DevOTPCode               string
	CORSOrigins              []string
	RunMigrations            bool
	ObjectStorePublicBaseURL string
	ObjectStoreBucket        string
	ObjectStoreRegion        string
	RazorpayKeyID            string
	RazorpayWebhookSecret    string
}

func Load() (*Config, error) {
	port := getenv("HTTP_PORT", "8080")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://utsav:utsav@localhost:5432/utsav?sslmode=disable"
	}
	migrations := getenv("MIGRATIONS_PATH", "../../db/migrations")
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-insecure-change-me"
	}
	cors := getenv("CORS_ORIGIN", "http://localhost:3000")
	runMig := getenv("RUN_MIGRATIONS", "true") == "true"
	return &Config{
		HTTPPort:                 port,
		DatabaseURL:              dsn,
		MigrationsPath:           migrations,
		JWTSecret:                secret,
		DevOTPCode:               getenv("DEV_OTP_CODE", "123456"),
		CORSOrigins:              []string{cors},
		RunMigrations:            runMig,
		ObjectStorePublicBaseURL: getenv("OBJECT_STORE_PUBLIC_BASE_URL", "http://127.0.0.1:9000/utsav"),
		ObjectStoreBucket:        getenv("OBJECT_STORE_BUCKET", "utsav"),
		ObjectStoreRegion:        getenv("OBJECT_STORE_REGION", "auto"),
		RazorpayKeyID:            getenv("RAZORPAY_KEY_ID", "rzp_test_stub"),
		RazorpayWebhookSecret:    getenv("RAZORPAY_WEBHOOK_SECRET", "rzp_webhook_secret_stub"),
	}, nil
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func MustPort(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Errorf("invalid port: %w", err))
	}
	return n
}
