package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	sentry "github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/bhune/utsav/services/api/pkg/cache"
	"github.com/bhune/utsav/services/api/internal/config"
	"github.com/bhune/utsav/services/api/internal/metrics"
	"github.com/bhune/utsav/services/api/internal/telemetry"
	"github.com/bhune/utsav/services/api/pkg/db"
	"github.com/bhune/utsav/services/api/internal/handler"
	"github.com/bhune/utsav/services/api/pkg/media"
	"github.com/bhune/utsav/services/api/internal/middleware"
	"github.com/bhune/utsav/services/api/pkg/migrate"
	"github.com/bhune/utsav/services/api/pkg/otp"
	"github.com/bhune/utsav/services/api/pkg/ratelimit"
	"github.com/bhune/utsav/services/api/internal/repository/authrepo"
	"github.com/bhune/utsav/services/api/internal/repository/billingrepo"
	"github.com/bhune/utsav/services/api/internal/repository/broadcastrepo"
	"github.com/bhune/utsav/services/api/internal/repository/eventrepo"
	"github.com/bhune/utsav/services/api/internal/repository/galleryrepo"
	"github.com/bhune/utsav/services/api/internal/repository/guestrepo"
	"github.com/bhune/utsav/services/api/internal/repository/memorybookrepo"
	"github.com/bhune/utsav/services/api/internal/repository/organiserrepo"
	"github.com/bhune/utsav/services/api/internal/repository/publicrepo"
	"github.com/bhune/utsav/services/api/internal/repository/rsvprepo"
	"github.com/bhune/utsav/services/api/internal/repository/shagunrepo"
	"github.com/bhune/utsav/services/api/internal/repository/vendorrepo"
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
	tracerShutdown := telemetry.SetupTracer(ctx)
	defer func() {
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracerShutdown(sctx); err != nil {
			log.Printf("otel shutdown: %v", err)
		}
	}()

	isProd := strings.EqualFold(strings.TrimSpace(cfg.Env), "production")
	pool, err := db.Connect(ctx, db.PoolConfig{
		DatabaseURL:           cfg.DatabaseURL,
		PingConnBeforeAcquire: !isProd,
		MaxConns:              int32(cfg.DBMaxConns),
		MinConns:              int32(cfg.DBMinConns),
		StatementTimeoutMs:    cfg.DBStatementTimeoutMs,
	})
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	readPool := pool
	if strings.TrimSpace(cfg.DatabaseReadURL) != "" {
		rp, rerr := db.Connect(ctx, db.PoolConfig{
			DatabaseURL:           cfg.DatabaseReadURL,
			PingConnBeforeAcquire: !isProd,
			MaxConns:              int32(cfg.DBMaxConns),
			MinConns:              int32(cfg.DBMinConns),
			StatementTimeoutMs:    cfg.DBStatementTimeoutMs,
		})
		if rerr != nil {
			log.Fatalf("db read replica: %v", rerr)
		}
		readPool = rp
		defer rp.Close()
	}

	metrics.RegisterAPI(pool, readPool)

	if cfg.RunMigrations {
		if err := migrate.Up(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
			log.Fatalf("migrate: %v", err)
		}
	}

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.MaxMultipartMemory = 8 << 20
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
	r.Use(
		func(c *gin.Context) {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2*1024*1024)
			c.Next()
		},
		middleware.RecoverJSON(),
		middleware.RequestID(),
		otelgin.Middleware("utsav-api"),
		middleware.Metrics(),
		middleware.Logger(),
		middleware.CORS(cfg.CORSOrigins),
	)

	dbHealthCtx, dbHealthCancel := context.WithCancel(context.Background())
	defer dbHealthCancel()

	srv := &httpserver.Server{
		Pool:          pool,
		ReadPool:      readPool,
		Config:        cfg,
		MediaSigner:   media.URLSigner{BaseURL: cfg.ObjectStorePublicBaseURL},
		DBHealthCtx:   dbHealthCtx,
	}
	window := time.Duration(cfg.RateLimitWindowSec) * time.Second
	if isProd && strings.TrimSpace(cfg.DevOTPCode) != "" {
		log.Fatal("DEV_OTP_CODE must be empty in production")
	}
	if isProd && (strings.TrimSpace(cfg.UpstashRESTURL) == "" || strings.TrimSpace(cfg.UpstashRESTToken) == "") {
		log.Fatal("Upstash Redis must be configured in production for distributed rate limiting")
	}
	if isProd && strings.TrimSpace(cfg.RazorpayWebhookSecret) == "" {
		log.Fatal("RAZORPAY_WEBHOOK_SECRET must be set in production")
	}
	if isProd && strings.EqualFold(strings.TrimSpace(cfg.OTPProvider), "msg91") {
		if strings.TrimSpace(cfg.OTPAPIKey) == "" {
			log.Fatal("OTP_API_KEY must be set when OTP_PROVIDER=msg91 in production")
		}
		if strings.TrimSpace(cfg.OTPSenderID) == "" {
			log.Fatal("OTP_SENDER_ID must be set when OTP_PROVIDER=msg91 in production")
		}
	}
	newLimiter := func(max int) ratelimit.Limiter {
		if strings.TrimSpace(cfg.UpstashRESTURL) != "" && strings.TrimSpace(cfg.UpstashRESTToken) != "" {
			return ratelimit.NewUpstashRESTLimiter(cfg.UpstashRESTURL, cfg.UpstashRESTToken, max, window)
		}
		return ratelimit.NewInMemoryLimiter(max, window)
	}
	var otpSender otp.Sender
	var asynqServer *asynq.Server
	if strings.EqualFold(strings.TrimSpace(cfg.OTPProvider), "msg91") {
		otpSender = otp.NewResilientSender(otp.NewMSG91Sender(cfg.OTPAPIKey, cfg.OTPSenderID, ""))
	}
	if isProd && otpSender == nil {
		log.Fatal("OTP provider must be configured in production")
	}
	var otpDispatcher otp.Dispatcher = &otp.DirectDispatcher{Sender: otpSender}
	if strings.TrimSpace(cfg.RedisURL) != "" && otpSender != nil {
		redisOpt, err := otp.RedisClientOptFromURL(cfg.RedisURL)
		if err != nil {
			log.Fatalf("invalid REDIS_URL for OTP queue: %v", err)
		}
		asynqClient := asynq.NewClient(redisOpt)
		otpDispatcher = otp.NewQueueDispatcher(asynqClient)
		asynqServer = asynq.NewServer(redisOpt, asynq.Config{
			Concurrency: 20,
			Queues: map[string]int{
				"critical": 10,
				"default":  5,
			},
		})
		go func() {
			if runErr := asynqServer.Run(otp.NewOTPTaskHandler(otpSender)); runErr != nil {
				log.Printf("WARN: otp queue server stopped: %v", runErr)
			}
		}()
	}
	srv.AuthService = authservice.NewService(
		authrepo.NewPGRepository(pool),
		newLimiter(cfg.AuthOTPRequestLimit),
		newLimiter(cfg.AuthOTPVerifyLimit),
		cfg.DevOTPCode,
		cfg.JWTSecret,
		cfg.OTPSecret,
		cfg.Env,
		otpDispatcher,
		cfg.OTPMaxAttempts,
	)
	srv.BillingService = billingservice.NewService(billingrepo.NewPGRepository(pool))
	srv.BroadcastService = broadcastservice.NewService(broadcastrepo.NewPGRepository(pool))
	srv.EventService = eventservice.NewService(eventrepo.NewPGRepository(pool))
	srv.GalleryService = galleryservice.NewService(galleryrepo.NewPGRepository(pool), srv.MediaSigner)
	srv.OrganiserService = organiserservice.NewService(organiserrepo.NewPGRepository(pool))
	srv.MemoryBookService = memorybookservice.NewService(memorybookrepo.NewPGRepository(pool))
	var publicCache cache.Cache
	if strings.TrimSpace(cfg.RedisURL) != "" {
		if redisCache, cacheErr := cache.NewRedisCache(cfg.RedisURL); cacheErr != nil {
			log.Printf("WARN: redis cache disabled due to config error: %v", cacheErr)
		} else {
			publicCache = redisCache
		}
	}
	srv.Cache = publicCache
	srv.PublicService = publicservice.NewService(publicrepo.NewPGRepository(pool, readPool), srv.MediaSigner, publicCache)
	srv.RSVPService = rsvpservice.NewService(
		rsvprepo.NewPGRepository(pool),
		newLimiter(cfg.RSVPOTPRequestLimit),
		newLimiter(cfg.RSVPOTPVerifyLimit),
		newLimiter(cfg.PublicRSVPSubmitLimit),
		cfg.DevOTPCode,
		cfg.JWTSecret,
		cfg.OTPSecret,
		cfg.Env,
		otpDispatcher,
		cfg.OTPMaxAttempts,
	)
	srv.GuestService = guestservice.NewService(guestrepo.NewPGRepository(pool, readPool), publicCache)
	srv.ShagunService = shagunservice.NewService(shagunrepo.NewPGRepository(pool))
	srv.VendorService = vendorservice.NewService(vendorrepo.NewPGRepository(pool))
	srv.Mount(r)

	if strings.TrimSpace(cfg.BetterstackHeartbeatURL) != "" {
		go startHeartbeat(cfg.BetterstackHeartbeatURL)
	}

	addr := ":" + cfg.HTTPPort
	log.Printf("utsav api listening on %s", addr)
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       90 * time.Second,
	}
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Print("shutdown signal received")
	dbHealthCancel()
	if asynqServer != nil {
		asynqServer.Shutdown()
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown failed: %v", err)
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
