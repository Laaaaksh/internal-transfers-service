// Package main is the entry point for the internal transfers service.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/internal-transfers-service/internal/config"
	"github.com/internal-transfers-service/internal/logger"
	"github.com/internal-transfers-service/internal/modules/account"
	"github.com/internal-transfers-service/internal/modules/health"
	"github.com/internal-transfers-service/internal/modules/transaction"
	"github.com/internal-transfers-service/pkg/database"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration based on APP_ENV environment variable
	// Loads default.toml first, then merges environment-specific overrides
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load configuration: " + err.Error())
	}

	// Initialize logger
	if err := logger.Initialize(cfg.Logging.Level, cfg.Logging.Format); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("Starting service",
		"name", cfg.App.Name,
		"env", cfg.App.Env,
		"port", cfg.App.Port,
		"ops_port", cfg.App.OpsPort,
	)

	// Initialize database
	db, err := database.Initialize(ctx, &cfg.Database)
	if err != nil {
		logger.Fatal("Failed to initialize database", "error", err)
	}
	defer db.Close()

	// Initialize modules
	accountModule := account.NewModule(ctx, db.GetPool())
	transactionModule := transaction.NewModule(ctx, db.GetPool(), accountModule.GetRepository())
	healthModule := health.NewModule(ctx, db)

	// Create main HTTP router
	mainRouter := chi.NewRouter()
	mainRouter.Use(middleware.RequestID)
	mainRouter.Use(middleware.RealIP)
	mainRouter.Use(middleware.Recoverer)
	mainRouter.Use(requestLogger)

	// Register routes
	accountModule.GetHandler().RegisterRoutes(mainRouter)
	transactionModule.GetHandler().RegisterRoutes(mainRouter)

	// Create ops router for health checks and metrics
	opsRouter := chi.NewRouter()
	healthHandler := health.NewHTTPHandler(healthModule.GetCore())
	healthHandler.RegisterRoutes(opsRouter)
	opsRouter.Handle("/metrics", promhttp.Handler())

	// Start servers
	mainServer := &http.Server{
		Addr:    cfg.App.Port,
		Handler: mainRouter,
	}

	opsServer := &http.Server{
		Addr:    cfg.App.OpsPort,
		Handler: opsRouter,
	}

	// Start main server in goroutine
	go func() {
		logger.Info("Main HTTP server starting", "addr", cfg.App.Port)
		if err := mainServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Main server failed", "error", err)
		}
	}()

	// Start ops server in goroutine
	go func() {
		logger.Info("Ops HTTP server starting", "addr", cfg.App.OpsPort)
		if err := opsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Ops server failed", "error", err)
		}
	}()

	// Wait for shutdown signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	<-c

	logger.Info("Shutdown signal received, initiating graceful shutdown")

	// Mark service as unhealthy
	healthModule.GetCore().MarkUnhealthy()

	// Wait for shutdown delay to allow load balancer to drain connections
	logger.Info("Waiting for shutdown delay", "delay_seconds", cfg.App.ShutdownDelay)
	time.Sleep(time.Duration(cfg.App.ShutdownDelay) * time.Second)

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, time.Duration(cfg.App.ShutdownTimeout)*time.Second)
	defer shutdownCancel()

	// Shutdown servers concurrently
	shutdownComplete := make(chan struct{})
	go func() {
		defer close(shutdownComplete)

		// Shutdown main server
		if err := mainServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("Main server shutdown error", "error", err)
		} else {
			logger.Info("Main server shutdown complete")
		}

		// Shutdown ops server
		if err := opsServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("Ops server shutdown error", "error", err)
		} else {
			logger.Info("Ops server shutdown complete")
		}
	}()

	// Wait for shutdown to complete
	select {
	case <-shutdownComplete:
		logger.Info("Graceful shutdown complete")
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout exceeded, forcing exit")
	}

	// Cancel context to cleanup any remaining resources
	cancel()

	logger.Info("Service stopped")
}

// requestLogger is a middleware that logs HTTP requests
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		duration := time.Since(start)

		logger.Ctx(r.Context()).Infow("HTTP request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status_code", ww.Status(),
			"duration_ms", duration.Milliseconds(),
			"bytes_written", ww.BytesWritten(),
		)
	})
}
