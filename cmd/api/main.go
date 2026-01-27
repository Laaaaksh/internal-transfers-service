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
	"github.com/internal-transfers-service/internal/constants"
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

	logger.Info(constants.LogMsgStartingService,
		constants.LogFieldName, cfg.App.Name,
		constants.LogFieldEnv, cfg.App.Env,
		constants.LogFieldPort, cfg.App.Port,
		constants.LogFieldOpsPort, cfg.App.OpsPort,
	)

	// Initialize database
	db, err := database.Initialize(ctx, &cfg.Database)
	if err != nil {
		logger.Fatal(constants.LogMsgFailedToInitDB, constants.LogKeyError, err)
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
		logger.Info(constants.LogMsgMainServerStarting, constants.LogFieldAddr, cfg.App.Port)
		if err := mainServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(constants.LogMsgMainServerFailed, constants.LogKeyError, err)
		}
	}()

	// Start ops server in goroutine
	go func() {
		logger.Info(constants.LogMsgOpsServerStarting, constants.LogFieldAddr, cfg.App.OpsPort)
		if err := opsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(constants.LogMsgOpsServerFailed, constants.LogKeyError, err)
		}
	}()

	// Wait for shutdown signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	<-c

	logger.Info(constants.LogMsgShutdownSignalReceived)

	// Mark service as unhealthy
	healthModule.GetCore().MarkUnhealthy()

	// Wait for shutdown delay to allow load balancer to drain connections
	logger.Info(constants.LogMsgWaitingForShutdownDelay, constants.LogFieldDelaySeconds, cfg.App.ShutdownDelay)
	time.Sleep(time.Duration(cfg.App.ShutdownDelay) * time.Second)

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, time.Duration(cfg.App.ShutdownTimeout)*time.Second)
	defer shutdownCancel()

	// Shutdown servers concurrently
	shutdownComplete := make(chan struct{})
	go func() {
		defer close(shutdownComplete)
		shutdownServers(shutdownCtx, mainServer, opsServer)
	}()

	// Wait for shutdown to complete
	select {
	case <-shutdownComplete:
		logger.Info(constants.LogMsgGracefulShutdownComplete)
	case <-shutdownCtx.Done():
		logger.Warn(constants.LogMsgShutdownTimeoutExceeded)
	}

	// Cancel context to cleanup any remaining resources
	cancel()

	logger.Info(constants.LogMsgServiceStopped)
}

// shutdownServers gracefully shuts down the main and ops servers
func shutdownServers(ctx context.Context, mainServer, opsServer *http.Server) {
	// Shutdown main server
	if err := mainServer.Shutdown(ctx); err != nil {
		logger.Error(constants.LogMsgMainServerShutdownErr, constants.LogKeyError, err)
		return
	}
	logger.Info(constants.LogMsgMainServerShutdownDone)

	// Shutdown ops server
	if err := opsServer.Shutdown(ctx); err != nil {
		logger.Error(constants.LogMsgOpsServerShutdownErr, constants.LogKeyError, err)
		return
	}
	logger.Info(constants.LogMsgOpsServerShutdownDone)
}

// requestLogger is a middleware that logs HTTP requests
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		duration := time.Since(start)

		logger.Ctx(r.Context()).Infow(constants.LogMsgHTTPRequestCompleted,
			constants.LogKeyMethod, r.Method,
			constants.LogKeyPath, r.URL.Path,
			constants.LogKeyStatusCode, ww.Status(),
			constants.LogKeyDuration, duration.Milliseconds(),
			constants.LogFieldBytesWritten, ww.BytesWritten(),
		)
	})
}
