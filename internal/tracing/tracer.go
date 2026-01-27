// Package tracing provides distributed tracing using OpenTelemetry.
package tracing

import (
	"context"
	"time"

	"github.com/internal-transfers-service/internal/config"
	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Tracer holds the OpenTelemetry tracer provider
type Tracer struct {
	provider *sdktrace.TracerProvider
	enabled  bool
}

// Initialize creates and configures the OpenTelemetry tracer
func Initialize(ctx context.Context, cfg *config.TracingConfig, serviceName string) (*Tracer, error) {
	if !cfg.Enabled {
		logger.Info(constants.LogMsgTracerDisabled)
		return &Tracer{enabled: false}, nil
	}

	exporter, err := createExporter(ctx, cfg)
	if err != nil {
		return nil, err
	}

	res, err := createResource(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	provider := createTracerProvider(exporter, res, cfg)
	configureGlobalTracer(provider)

	logger.Info(constants.LogMsgTracerInitialized,
		constants.LogFieldEndpoint, cfg.Endpoint,
		constants.LogFieldSampleRate, cfg.SampleRate,
	)

	return &Tracer{provider: provider, enabled: true}, nil
}

// createExporter creates the OTLP gRPC exporter
func createExporter(ctx context.Context, cfg *config.TracingConfig) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
	}

	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	}

	return otlptracegrpc.New(ctx, opts...)
}

// createResource creates the OpenTelemetry resource with service information
func createResource(ctx context.Context, serviceName string) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
}

// createTracerProvider creates the tracer provider with sampling and batching
func createTracerProvider(exporter sdktrace.SpanExporter, res *resource.Resource, cfg *config.TracingConfig) *sdktrace.TracerProvider {
	sampler := createSampler(cfg.SampleRate)
	batchTimeout := parseBatchTimeout(cfg.BatchTimeout)

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(batchTimeout)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)
}

// createSampler creates a sampler based on the sample rate
func createSampler(sampleRate float64) sdktrace.Sampler {
	if sampleRate >= 1.0 {
		return sdktrace.AlwaysSample()
	}
	if sampleRate <= 0.0 {
		return sdktrace.NeverSample()
	}
	return sdktrace.TraceIDRatioBased(sampleRate)
}

// parseBatchTimeout parses the batch timeout string
func parseBatchTimeout(timeout string) time.Duration {
	d, err := time.ParseDuration(timeout)
	if err != nil {
		return time.Duration(constants.DefaultTracingBatchTimeoutSecond) * time.Second
	}
	return d
}

// configureGlobalTracer sets the global tracer provider and propagator
func configureGlobalTracer(provider *sdktrace.TracerProvider) {
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
}

// Shutdown gracefully shuts down the tracer provider
func (t *Tracer) Shutdown(ctx context.Context) {
	if !t.enabled || t.provider == nil {
		return
	}

	if err := t.provider.Shutdown(ctx); err != nil {
		logger.Error(constants.LogMsgTracerShutdownFailed, constants.LogKeyError, err)
		return
	}
	logger.Info(constants.LogMsgTracerShutdown)
}

// IsEnabled returns whether tracing is enabled
func (t *Tracer) IsEnabled() bool {
	return t.enabled
}

// GetTracer returns an OpenTelemetry tracer for the given name
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// SpanFromContext returns the current span from context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// TraceIDFromContext extracts the trace ID from context
func TraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().HasTraceID() {
		return ""
	}
	return span.SpanContext().TraceID().String()
}

// SpanIDFromContext extracts the span ID from context
func SpanIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().HasSpanID() {
		return ""
	}
	return span.SpanContext().SpanID().String()
}
