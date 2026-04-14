package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bhune/utsav/services/api/internal/config"
	"github.com/bhune/utsav/services/api/internal/db"
	"github.com/bhune/utsav/services/api/internal/httpserver"
	"github.com/bhune/utsav/services/api/internal/media"
	"github.com/bhune/utsav/services/api/internal/middleware"
	"github.com/bhune/utsav/services/api/internal/migrate"
	"github.com/bhune/utsav/services/api/internal/ratelimit"
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
	r.Use(middleware.RecoverJSON(), middleware.RequestID(), middleware.Logger(), middleware.CORS(cfg.CORSOrigins))

	srv := &httpserver.Server{
		Pool:         pool,
		Config:       cfg,
		AuthOTPLimit: ratelimit.New(5, 15*time.Minute),
		RSVPOTPLimit: ratelimit.New(10, 15*time.Minute),
		MediaSigner:  media.URLSigner{BaseURL: cfg.ObjectStorePublicBaseURL},
	}
	srv.Mount(r)

	addr := ":" + cfg.HTTPPort
	log.Printf("utsav api listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
