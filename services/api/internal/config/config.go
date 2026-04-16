package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort                 string
	DatabaseURL              string
	// DatabaseReadURL optional read replica DSN for guest/public SELECT paths (same schema). Empty = use DatabaseURL only.
	DatabaseReadURL          string
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
	// PublicMetrics when false omits /metrics from the HTTP router (default off in production).
	PublicMetrics bool
	DBMaxConns               int
	DBMinConns               int
	DBStatementTimeoutMs    int
}

func Load() (*Config, error) {
	return loadWithMode(false)
}

// LoadWorker is [Load] with UTSAV_WORKER=1 semantics: same env parsing, but skips
// API-only production requirements (CORS, Razorpay, OTP provider, distinct OTP_SECRET).
// Use for cmd/worker background jobs only.
func LoadWorker() (*Config, error) {
	return loadWithMode(true)
}

func loadWithMode(workerMode bool) (*Config, error) {
	// Load optional local env file for development.
	// In production (Render) environment variables are injected by the platform.
	if envRaw := strings.TrimSpace(os.Getenv("ENV")); envRaw == "" || !strings.EqualFold(envRaw, "production") {
		_ = godotenv.Load(".env")
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = getenv("HTTP_PORT", "8080")
	}
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	dsnRead := strings.TrimSpace(os.Getenv("DATABASE_READ_URL"))
	migrations := getenv("MIGRATIONS_PATH", "../../db/migrations")
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	cors := strings.TrimSpace(os.Getenv("CORS_ORIGIN"))
	runMig := getenv("RUN_MIGRATIONS", "true") == "true"
	env := getenv("ENV", "development")
	isProd := strings.EqualFold(strings.TrimSpace(env), "production")
	if strings.TrimSpace(dsn) == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if strings.TrimSpace(secret) == "" {
		log.Fatal("JWT_SECRET is required")
	}
	if strings.TrimSpace(cors) == "" && !workerMode {
		log.Fatal("CORS_ORIGIN is required")
	}
	if strings.TrimSpace(cors) == "" && workerMode {
		cors = "http://127.0.0.1:1"
	}
	if isProd {
		if len(secret) < 32 {
			log.Fatal("JWT_SECRET must be at least 32 characters in production")
		}
	}
	otpSecret := strings.TrimSpace(os.Getenv("OTP_SECRET"))
	if otpSecret == "" {
		otpSecret = secret
		if !isProd {
			log.Printf("WARN: OTP_SECRET is unset; using JWT_SECRET for OTP hashing. Set OTP_SECRET so OTP verification stays valid when JWT_SECRET is rotated.")
		}
	}
	if isProd && !workerMode {
		if otpSecret == "" {
			log.Fatal("OTP_SECRET must be set to a strong secret in production (independent of JWT rotation)")
		}
		if len(otpSecret) < 32 {
			log.Fatal("OTP_SECRET must be at least 32 characters in production")
		}
		if otpSecret == secret {
			log.Fatal("OTP_SECRET must differ from JWT_SECRET in production so rotating JWT does not break in-flight OTP verification")
		}
	}
	razorpayKeyID := strings.TrimSpace(os.Getenv("RAZORPAY_KEY_ID"))
	razorpayWebhookSecret := strings.TrimSpace(os.Getenv("RAZORPAY_WEBHOOK_SECRET"))
	if isProd && !workerMode {
		if razorpayKeyID == "" {
			log.Fatal("RAZORPAY_KEY_ID is required in production")
		}
		if razorpayWebhookSecret == "" {
			log.Fatal("RAZORPAY_WEBHOOK_SECRET is required in production")
		}
	}

	metricsPublicRaw := strings.TrimSpace(os.Getenv("METRICS_PUBLIC"))
	publicMetrics := !isProd
	switch strings.ToLower(metricsPublicRaw) {
	case "1", "true", "yes", "on":
		publicMetrics = true
	case "0", "false", "no", "off":
		publicMetrics = false
	case "":
		// keep default above
	default:
		log.Printf("WARN: METRICS_PUBLIC=%q is not a boolean; using default public_metrics=%v", metricsPublicRaw, publicMetrics)
	}
	if isProd && !publicMetrics {
		log.Printf("INFO: Prometheus /metrics is not mounted (METRICS_PUBLIC unset or false). Set METRICS_PUBLIC=true only behind auth or a private listener.")
	}
	dbMax := mustAtoi(getenv("DB_MAX_CONNS", "20"), 20)
	if dbMax < 1 {
		dbMax = 20
	}
	dbMin := mustAtoi(getenv("DB_MIN_CONNS", "2"), 2)
	if dbMin < 0 {
		dbMin = 0
	}
	if dbMin > dbMax {
		dbMin = dbMax
	}
	dbStmtMs := mustAtoi(getenv("DB_STATEMENT_TIMEOUT_MS", "5000"), 5000)
	if dbStmtMs < 100 {
		dbStmtMs = 5000
	}
	corsOrigins := splitAndTrimCSV(cors)

	return &Config{
		HTTPPort:                 port,
		DatabaseURL:              dsn,
		DatabaseReadURL:          dsnRead,
		MigrationsPath:           migrations,
		Env:                      env,
		JWTSecret:                secret,
		OTPSecret:                otpSecret,
		DevOTPCode:               getenv("DEV_OTP_CODE", ""),
		OTPProvider:              getenv("OTP_PROVIDER", ""),
		OTPAPIKey:                getenv("OTP_API_KEY", ""),
		OTPAPISecret:             getenv("OTP_API_SECRET", ""),
		OTPSenderID:              getenv("OTP_SENDER_ID", ""),
		CORSOrigins:              corsOrigins,
		RunMigrations:            runMig,
		ObjectStorePublicBaseURL: getenv("OBJECT_STORE_PUBLIC_BASE_URL", ""),
		ObjectStoreBucket:        getenv("OBJECT_STORE_BUCKET", "utsav"),
		ObjectStoreRegion:        getenv("OBJECT_STORE_REGION", "auto"),
		RazorpayKeyID:            razorpayKeyID,
		RazorpayWebhookSecret:    razorpayWebhookSecret,
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
		PublicMetrics:            publicMetrics,
		DBMaxConns:               dbMax,
		DBMinConns:               dbMin,
		DBStatementTimeoutMs:     dbStmtMs,
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

func splitAndTrimCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
