package worker

import (
	"context"
	"log"
	"time"

	"github.com/bhune/utsav/services/api/internal/metrics"
	billingservice "github.com/bhune/utsav/services/api/internal/service/billing"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StartIdempotencyKeyCleanup periodically deletes expired idempotency keys.
func StartIdempotencyKeyCleanup(ctx context.Context, pool *pgxpool.Pool) {
	go func() {
		t := time.NewTicker(1 * time.Hour)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				cctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				_, err := pool.Exec(cctx, `DELETE FROM idempotency_keys WHERE expires_at < now()`)
				cancel()
				if err != nil {
					log.Printf("WARN: idempotency_keys cleanup: %v", err)
				}
			}
		}
	}()
}

// StartWebhookDeliveriesCleanup deletes old webhook delivery audit rows.
func StartWebhookDeliveriesCleanup(ctx context.Context, pool *pgxpool.Pool) {
	go func() {
		t := time.NewTicker(24 * time.Hour)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				cctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				_, err := pool.Exec(cctx, `DELETE FROM webhook_deliveries WHERE created_at < now() - interval '90 days'`)
				cancel()
				if err != nil {
					log.Printf("WARN: webhook_deliveries cleanup: %v", err)
				}
			}
		}
	}()
}

// StartWebhookRetry runs the Razorpay webhook replay loop (uses SKIP LOCKED in repo).
func StartWebhookRetry(ctx context.Context, billing *billingservice.Service) {
	if billing == nil {
		return
	}
	go func() {
		t := time.NewTicker(1 * time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				metrics.WebhookRetryTick()
				cctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
				processed, svcErr := billing.RetryPendingWebhookDeliveries(cctx, "razorpay", 25)
				cancel()
				if svcErr != nil {
					metrics.WebhookRetryError()
					log.Printf("WARN: webhook retry worker: %s", svcErr.Message)
					continue
				}
				metrics.WebhookRetryRows(processed)
				if processed > 0 {
					log.Printf("webhook retry worker processed=%d", processed)
				}
			}
		}
	}()
}
