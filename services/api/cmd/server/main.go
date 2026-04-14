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
	r.Use(middleware.RecoverJSON(), middleware.RequestID(), middleware.Logger(), middleware.CORS(cfg.CORSOrigins))

	srv := &httpserver.Server{
		Pool:         pool,
		Config:       cfg,
		AuthOTPLimit: ratelimit.New(5, 15*time.Minute),
		RSVPOTPLimit: ratelimit.New(10, 15*time.Minute),
		MediaSigner:  media.URLSigner{BaseURL: cfg.ObjectStorePublicBaseURL},
	}
	srv.AuthService = authservice.NewService(
		authrepo.NewPGRepository(pool),
		srv.AuthOTPLimit,
		cfg.DevOTPCode,
		cfg.JWTSecret,
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
		srv.RSVPOTPLimit,
		cfg.DevOTPCode,
		cfg.JWTSecret,
	)
	srv.GuestService = guestservice.NewService(guestrepo.NewPGRepository(pool))
	srv.ShagunService = shagunservice.NewService(shagunrepo.NewPGRepository(pool))
	srv.VendorService = vendorservice.NewService(vendorrepo.NewPGRepository(pool))
	srv.Mount(r)

	addr := ":" + cfg.HTTPPort
	log.Printf("utsav api listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
