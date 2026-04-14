package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	sentry "github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"

	"github.com/bhune/utsav/services/api/internal/config"
	"github.com/bhune/utsav/services/api/internal/db"
	"github.com/bhune/utsav/services/api/internal/httpserver"
	"github.com/bhune/utsav/services/api/internal/media"
	"github.com/bhune/utsav/services/api/internal/middleware"
	"github.com/bhune/utsav/services/api/internal/migrate"
	"github.com/bhune/utsav/services/api/internal/otp"
	"github.com/bhune/utsav/services/api/internal/ratelimit"
	"github.com/bhune/utsav/services/api/internal/repository/authrepo"
	"github.com/bhune/utsav/services/api/internal/repository/billingrepo"
	"github.com/bhune/utsav/services/api/internal/repository/broadcastrepo"
	"github.com/bhune/utsav/services/api/internal/repository/eventrepo"
	"github.com/bhune/utsav/services/api/internal/repository/galleryrepo"
	"github.com/bhune/utsav/services/api/internal/repository/guestrepo"
	"github.com/bhune/utsav/services/api/internal/repository/memorybookrepo"
	"github.com/bhune/utsav/services/api/internal/repository/rsvprepo"
	"github.com/bhune/utsav/services/api/internal/repository/shagunrepo"
	"github.com/bhune/utsav/services/api/internal/repository/vendorrepo"
	"github.com/bhune/utsav/services/api/internal/repository/organiserrepo"
	"github.com/bhune/utsav/services/api/internal/repository/publicrepo"
	authservice "github.com/bhune/utsav/services/api/internal/service/auth"
	billingservice "github.com/bhune/utsav/services/api/internal/service/billing"
	broadcastservice "github.com/bhune/utsav/services/api/internal/service/broadcast"
	eventservice "github.com/bhune/utsav/services/api/internal/service/event"
	galleryservice "github.com/bhune/utsav/services/api/internal/service/gallery"
	guestservice "github.com/bhune/utsav/services/api/internal/service/guest"
	memorybookservice "github.com/bhune/utsav/services/api/internal/service/memorybook"
	organiserservice "github.com/bhune/utsav/services/api/internal/service/organiser"
	publicservice "github.com/bhune/utsav/services/api/internal/service/public"
	rsvpservice "github.com/bhune/utsav/services/api/internal/service/rsvp"
	shagunservice "github.com/bhune/utsav/services/api/internal/service/shagun"
	vendorservice "github.com/bhune/utsav/services/api/internal/service/vendor"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	if cfg.RunMigrations {
		if err := migrate.Up(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
			log.Fatalf("migrate: %v", err)
		}
	}

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	if strings.TrimSpace(cfg.SentryDSN) != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			Environment:      cfg.Env,
			EnableTracing:    false,
			AttachStacktrace: true,
		}); err != nil {
			log.Printf("sentry init failed: %v", err)
		} else {
			r.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
		}
	}
	r.Use(middleware.RecoverJSON(), middleware.RequestID(), middleware.Logger(), middleware.CORS(cfg.CORSOrigins))

	srv := &httpserver.Server{
		Pool:         pool,
		Config:       cfg,
		MediaSigner:  media.URLSigner{BaseURL: cfg.ObjectStorePublicBaseURL},
	}
	window := time.Duration(cfg.RateLimitWindowSec) * time.Second
	isProd := strings.EqualFold(strings.TrimSpace(cfg.Env), "production")
	if isProd && strings.TrimSpace(cfg.DevOTPCode) != "" {
		log.Fatal("DEV_OTP_CODE must be empty in production")
	}
	if isProd && (strings.TrimSpace(cfg.UpstashRESTURL) == "" || strings.TrimSpace(cfg.UpstashRESTToken) == "") {
		log.Fatal("Upstash Redis must be configured in production for distributed rate limiting")
	}
	newLimiter := func(max int) ratelimit.Limiter {
		if strings.TrimSpace(cfg.UpstashRESTURL) != "" && strings.TrimSpace(cfg.UpstashRESTToken) != "" {
			return ratelimit.NewUpstashRESTLimiter(cfg.UpstashRESTURL, cfg.UpstashRESTToken, max, window)
		}
		return ratelimit.NewInMemoryLimiter(max, window)
	}
	var otpSender otp.Sender
	if strings.EqualFold(strings.TrimSpace(cfg.OTPProvider), "msg91") {
		otpSender = otp.NewMSG91Sender(cfg.OTPAPIKey, cfg.OTPSenderID, "")
	}
	if isProd && otpSender == nil {
		log.Fatal("OTP provider must be configured in production")
	}
	srv.AuthService = authservice.NewService(
		authrepo.NewPGRepository(pool),
		newLimiter(cfg.AuthOTPRequestLimit),
		newLimiter(cfg.AuthOTPVerifyLimit),
		cfg.DevOTPCode,
		cfg.JWTSecret,
		cfg.Env,
		otpSender,
		cfg.OTPMaxAttempts,
	)
	srv.BillingService = billingservice.NewService(billingrepo.NewPGRepository(pool))
	srv.BroadcastService = broadcastservice.NewService(broadcastrepo.NewPGRepository(pool))
	srv.EventService = eventservice.NewService(eventrepo.NewPGRepository(pool))
	srv.GalleryService = galleryservice.NewService(galleryrepo.NewPGRepository(pool), srv.MediaSigner)
	srv.OrganiserService = organiserservice.NewService(organiserrepo.NewPGRepository(pool))
	srv.MemoryBookService = memorybookservice.NewService(memorybookrepo.NewPGRepository(pool))
	srv.PublicService = publicservice.NewService(publicrepo.NewPGRepository(pool), srv.MediaSigner)
	srv.RSVPService = rsvpservice.NewService(
		rsvprepo.NewPGRepository(pool),
		newLimiter(cfg.RSVPOTPRequestLimit),
		newLimiter(cfg.RSVPOTPVerifyLimit),
		newLimiter(cfg.PublicRSVPSubmitLimit),
		cfg.DevOTPCode,
		cfg.JWTSecret,
		cfg.Env,
		otpSender,
		cfg.OTPMaxAttempts,
	)
	srv.GuestService = guestservice.NewService(guestrepo.NewPGRepository(pool))
	srv.ShagunService = shagunservice.NewService(shagunrepo.NewPGRepository(pool))
	srv.VendorService = vendorservice.NewService(vendorrepo.NewPGRepository(pool))
	srv.Mount(r)

	if strings.TrimSpace(cfg.BetterstackHeartbeatURL) != "" {
		go startHeartbeat(cfg.BetterstackHeartbeatURL)
	}

	addr := ":" + cfg.HTTPPort
	log.Printf("utsav api listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func startHeartbeat(url string) {
	client := &http.Client{Timeout: 5 * time.Second}
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		req, err := http.NewRequest(http.MethodGet, strings.TrimSpace(url), nil)
		if err != nil {
			log.Printf("heartbeat request build failed: %v", err)
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("heartbeat send failed: %v", err)
			continue
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode >= 300 {
			log.Printf("heartbeat responded non-2xx: %d", resp.StatusCode)
		}
	}
}
