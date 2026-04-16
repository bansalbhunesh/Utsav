package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bhune/utsav/services/api/internal/config"
	"github.com/bhune/utsav/services/api/internal/metrics"
	"github.com/bhune/utsav/services/api/internal/repository/billingrepo"
	billingservice "github.com/bhune/utsav/services/api/internal/service/billing"
	"github.com/bhune/utsav/services/api/internal/worker"
	"github.com/bhune/utsav/services/api/pkg/db"
	"github.com/bhune/utsav/services/api/pkg/migrate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.LoadWorker()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
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

	if cfg.RunMigrations {
		if err := migrate.Up(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
			log.Fatalf("migrate: %v", err)
		}
	}

	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()

	metrics.RegisterWorker()

	worker.StartIdempotencyKeyCleanup(cleanupCtx, pool)
	worker.StartWebhookDeliveriesCleanup(cleanupCtx, pool)
	billing := billingservice.NewService(billingrepo.NewPGRepository(pool))
	worker.StartWebhookRetry(cleanupCtx, billing)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/metrics", promhttp.Handler())
	addr := ":" + cfg.HTTPPort
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		log.Printf("utsav worker listening on %s (/health, /metrics)", addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("worker http: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Print("worker shutdown signal received")
	cleanupCancel()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}
