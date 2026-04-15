package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPPort                 string
	DatabaseURL              string
	MigrationsPath           string
	Env                      string
	JWTSecret                string
	OTPSecret                string // HMAC key for OTP code hashing; must differ from JWT in production.
	DevOTPCode               string
	OTPProvider              string
	OTPAPIKey                string
	OTPAPISecret             string
	OTPSenderID              string
	CORSOrigins              []string
	RunMigrations            bool
	ObjectStorePublicBaseURL string
	ObjectStoreBucket        string
	ObjectStoreRegion        string
	RazorpayKeyID            string
	RazorpayWebhookSecret    string
	RedisURL                 string
	UpstashRESTURL           string
	UpstashRESTToken         string
	AuthOTPRequestLimit      int
	AuthOTPVerifyLimit       int
	RSVPOTPRequestLimit      int
	RSVPOTPVerifyLimit       int
	PublicRSVPSubmitLimit    int
	RateLimitWindowSec       int
	OTPMaxAttempts           int
	LogLevel                 string
	SentryDSN                string
	BetterstackHeartbeatURL  string
	FrontendSentryDSN        string
	AuthCookieDomain         string
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
	env := getenv("ENV", "development")
	isProd := strings.EqualFold(strings.TrimSpace(env), "production")
	if isProd {
		if strings.TrimSpace(secret) == "" || secret == "dev-insecure-change-me" {
			log.Fatal("JWT_SECRET must be set to a strong secret in production")
		}
		if len(secret) < 32 {
			log.Fatal("JWT_SECRET must be at least 32 characters in production")
		}
	}
	otpSecret := strings.TrimSpace(os.Getenv("OTP_SECRET"))
	if otpSecret == "" {
		otpSecret = secret
	}
	if isProd {
		if otpSecret == "" || otpSecret == "dev-insecure-change-me" {
			log.Fatal("OTP_SECRET must be set to a strong secret in production (independent of JWT rotation)")
		}
		if len(otpSecret) < 32 {
			log.Fatal("OTP_SECRET must be at least 32 characters in production")
		}
		if otpSecret == secret {
			log.Fatal("OTP_SECRET must differ from JWT_SECRET in production so rotating JWT does not break in-flight OTP verification")
		}
	}
	return &Config{
		HTTPPort:                 port,
		DatabaseURL:              dsn,
		MigrationsPath:           migrations,
		Env:                      env,
		JWTSecret:                secret,
		OTPSecret:                otpSecret,
		DevOTPCode:               getenv("DEV_OTP_CODE", "123456"),
		OTPProvider:              getenv("OTP_PROVIDER", ""),
		OTPAPIKey:                getenv("OTP_API_KEY", ""),
		OTPAPISecret:             getenv("OTP_API_SECRET", ""),
		OTPSenderID:              getenv("OTP_SENDER_ID", ""),
		CORSOrigins:              []string{cors},
		RunMigrations:            runMig,
		ObjectStorePublicBaseURL: getenv("OBJECT_STORE_PUBLIC_BASE_URL", "http://127.0.0.1:9000/utsav"),
		ObjectStoreBucket:        getenv("OBJECT_STORE_BUCKET", "utsav"),
		ObjectStoreRegion:        getenv("OBJECT_STORE_REGION", "auto"),
		RazorpayKeyID:            getenv("RAZORPAY_KEY_ID", "rzp_test_stub"),
		RazorpayWebhookSecret:    getenv("RAZORPAY_WEBHOOK_SECRET", "rzp_webhook_secret_stub"),
		RedisURL:                 getenv("REDIS_URL", ""),
		UpstashRESTURL:           getenv("UPSTASH_REDIS_REST_URL", ""),
		UpstashRESTToken:         getenv("UPSTASH_REDIS_REST_TOKEN", ""),
		AuthOTPRequestLimit:      mustAtoi(getenv("AUTH_OTP_REQUEST_LIMIT", "5"), 5),
		AuthOTPVerifyLimit:       mustAtoi(getenv("AUTH_OTP_VERIFY_LIMIT", "10"), 10),
		RSVPOTPRequestLimit:      mustAtoi(getenv("RSVP_OTP_REQUEST_LIMIT", "10"), 10),
		RSVPOTPVerifyLimit:       mustAtoi(getenv("RSVP_OTP_VERIFY_LIMIT", "20"), 20),
		PublicRSVPSubmitLimit:    mustAtoi(getenv("PUBLIC_RSVP_SUBMIT_LIMIT", "30"), 30),
		RateLimitWindowSec:       mustAtoi(getenv("RATE_LIMIT_WINDOW", "900"), 900),
		OTPMaxAttempts:           mustAtoi(getenv("OTP_MAX_ATTEMPTS", "5"), 5),
		LogLevel:                 getenv("LOG_LEVEL", "info"),
		SentryDSN:                getenv("SENTRY_DSN", ""),
		BetterstackHeartbeatURL:  getenv("BETTERSTACK_HEARTBEAT_URL", ""),
		FrontendSentryDSN:        getenv("NEXT_PUBLIC_SENTRY_DSN", ""),
		AuthCookieDomain:         getenv("AUTH_COOKIE_DOMAIN", ""),
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

func mustAtoi(v string, def int) int {
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
