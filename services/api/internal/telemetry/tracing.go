// Package telemetry configures OpenTelemetry for the API process.
package telemetry

import (
	"context"
	"log"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// SetupTracer configures the global TracerProvider used by otelgin and other instrumentation.
// Default (unset / none): noop provider (no overhead, no spans exported).
// OTEL_TRACES_EXPORTER=stdout: batch export spans to stdout (dev only).
func SetupTracer(ctx context.Context) (shutdown func(context.Context) error) {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv("OTEL_TRACES_EXPORTER")))
	switch raw {
	case "", "none", "noop":
		otel.SetTracerProvider(noop.NewTracerProvider())
		return func(context.Context) error { return nil }

	case "stdout":
		exp, err := stdouttrace.New()
		if err != nil {
			log.Printf("WARN: otel stdout exporter: %v; using noop tracer", err)
			otel.SetTracerProvider(noop.NewTracerProvider())
			return func(context.Context) error { return nil }
		}
		tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp))
		otel.SetTracerProvider(tp)
		return tp.Shutdown

	default:
		log.Printf("WARN: unknown OTEL_TRACES_EXPORTER=%q; using noop tracer", os.Getenv("OTEL_TRACES_EXPORTER"))
		otel.SetTracerProvider(noop.NewTracerProvider())
		return func(context.Context) error { return nil }
	}
}
