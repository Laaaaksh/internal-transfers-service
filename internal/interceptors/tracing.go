// Package interceptors provides HTTP middleware for tracing.
package interceptors

import (
	"context"
	"net/http"

	"github.com/internal-transfers-service/internal/config"
	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/constants/contextkeys"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware creates an OpenTelemetry HTTP tracing middleware
func TracingMiddleware(cfg config.TracingConfig) func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return createPassThroughMiddleware()
	}
	return createTracingHandler()
}

// createTracingHandler creates the actual tracing middleware using otelhttp
func createTracingHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := otelhttp.NewHandler(next, constants.TracingSpanHTTPRequest,
			otelhttp.WithSpanNameFormatter(formatSpanName),
		)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
		})
	}
}

// TraceContextMiddleware extracts trace/span IDs and adds them to context
func TraceContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = addTraceIDsToContext(ctx)
		addCustomSpanAttributes(r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// addTraceIDsToContext adds trace and span IDs to context for logging
func addTraceIDsToContext(ctx context.Context) context.Context {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ctx
	}

	if span.SpanContext().HasTraceID() {
		ctx = context.WithValue(ctx, contextkeys.TraceID, span.SpanContext().TraceID().String())
	}
	if span.SpanContext().HasSpanID() {
		ctx = context.WithValue(ctx, contextkeys.SpanID, span.SpanContext().SpanID().String())
	}
	return ctx
}

// formatSpanName formats the span name based on method and path
func formatSpanName(_ string, r *http.Request) string {
	return r.Method + " " + r.URL.Path
}

// addCustomSpanAttributes adds custom attributes to the current span
func addCustomSpanAttributes(r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	if !span.IsRecording() {
		return
	}

	// Add request ID if present
	if requestID, ok := r.Context().Value(contextkeys.RequestID).(string); ok {
		span.SetAttributes(attribute.String(constants.TracingAttrHTTPRequestID, requestID))
	}

	// Add user agent
	if userAgent := r.UserAgent(); userAgent != "" {
		span.SetAttributes(attribute.String(constants.TracingAttrHTTPUserAgent, userAgent))
	}

	// Add client IP
	if clientIP := r.RemoteAddr; clientIP != "" {
		span.SetAttributes(attribute.String(constants.TracingAttrHTTPClientIP, clientIP))
	}
}
