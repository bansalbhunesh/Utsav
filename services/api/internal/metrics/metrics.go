package metrics

import (
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	webhookOnce sync.Once
	otpOnce     sync.Once
	dbOnce      sync.Once

	otpVerifyTotal *prometheus.CounterVec

	webhookRetryRowsProcessed prometheus.Counter
	webhookRetryIterations    prometheus.Counter
	webhookRetryErrors        prometheus.Counter
)

func registerWebhookMetrics() {
	webhookOnce.Do(func() {
		webhookRetryRowsProcessed = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "utsav_webhook_retry_rows_processed_total",
			Help: "Cumulative webhook delivery rows replayed successfully by the worker",
		})
		webhookRetryIterations = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "utsav_webhook_retry_iterations_total",
			Help: "Cumulative webhook retry worker ticks (each tick may process 0+ rows)",
		})
		webhookRetryErrors = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "utsav_webhook_retry_errors_total",
			Help: "Cumulative webhook retry worker failures (service error)",
		})
		prometheus.MustRegister(webhookRetryRowsProcessed, webhookRetryIterations, webhookRetryErrors)
	})
}

func registerOTPMetrics() {
	otpOnce.Do(func() {
		otpVerifyTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "utsav_otp_verify_total",
				Help: "OTP verify attempts by channel and result (failure rate: 1 - sum(rate(...{result=\"success\"})) / sum(rate(...)))",
			},
			[]string{"channel", "result"},
		)
		prometheus.MustRegister(otpVerifyTotal)
	})
}

// RegisterAPI registers Prometheus collectors used by the HTTP API (DB pool gauges + OTP verify metrics).
func RegisterAPI(writer, reader *pgxpool.Pool) {
	registerOTPMetrics()

	trackReader := reader != nil && reader != writer

	dbOnce.Do(func() {
		reg := []prometheus.Collector{
			newPoolGaugeFunc("utsav_db_writer_pool_acquired_conns", "Writer pool connections checked out", writer, func(s *pgxpool.Stat) float64 { return float64(s.AcquiredConns()) }),
			newPoolGaugeFunc("utsav_db_writer_pool_idle_conns", "Writer pool idle connections", writer, func(s *pgxpool.Stat) float64 { return float64(s.IdleConns()) }),
			newPoolGaugeFunc("utsav_db_writer_pool_max_conns", "Writer pool max connections", writer, func(s *pgxpool.Stat) float64 { return float64(s.MaxConns()) }),
		}
		if trackReader {
			reg = append(reg,
				newPoolGaugeFunc("utsav_db_reader_pool_acquired_conns", "Reader pool connections checked out", reader, func(s *pgxpool.Stat) float64 { return float64(s.AcquiredConns()) }),
				newPoolGaugeFunc("utsav_db_reader_pool_idle_conns", "Reader pool idle connections", reader, func(s *pgxpool.Stat) float64 { return float64(s.IdleConns()) }),
				newPoolGaugeFunc("utsav_db_reader_pool_max_conns", "Reader pool max connections", reader, func(s *pgxpool.Stat) float64 { return float64(s.MaxConns()) }),
			)
		}
		prometheus.MustRegister(reg...)
	})
}

func newPoolGaugeFunc(name, help string, pool *pgxpool.Pool, pick func(*pgxpool.Stat) float64) prometheus.Collector {
	return prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{Name: name, Help: help},
		func() float64 {
			if pool == nil {
				return 0
			}
			return pick(pool.Stat())
		},
	)
}

// RegisterWorker registers webhook retry counters for the worker process only.
func RegisterWorker() {
	registerWebhookMetrics()
}

// ResultLabel normalizes a service error code for the utsav_otp_verify_total label.
func ResultLabel(code string) string {
	v := strings.TrimSpace(code)
	if v == "" {
		return "unknown"
	}
	return v
}

// OTPVerify records one OTP verify attempt. channel is "auth" or "rsvp". result is "success" or a stable error code.
func OTPVerify(channel, result string) {
	if otpVerifyTotal == nil {
		return
	}
	otpVerifyTotal.WithLabelValues(channel, result).Inc()
}

// WebhookRetryTick records one worker iteration (before processing outcome).
func WebhookRetryTick() {
	if webhookRetryIterations == nil {
		return
	}
	webhookRetryIterations.Inc()
}

// WebhookRetryRows adds successfully processed rows in this tick.
func WebhookRetryRows(n int) {
	if webhookRetryRowsProcessed == nil || n <= 0 {
		return
	}
	webhookRetryRowsProcessed.Add(float64(n))
}

// WebhookRetryError records a failed retry tick (service error).
func WebhookRetryError() {
	if webhookRetryErrors == nil {
		return
	}
	webhookRetryErrors.Inc()
}
