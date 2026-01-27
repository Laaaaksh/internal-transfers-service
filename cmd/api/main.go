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

	cfg := loadConfiguration()
	initializeLogger(cfg)
	defer logger.Sync()

	logServiceStartup(cfg)

	db := initializeDatabase(ctx, cfg)
	defer db.Close()

	modules := initializeModules(ctx, db)
	mainServer, opsServer := setupAndStartServers(cfg, modules)

	waitForShutdownSignal(ctx, cfg, modules.health, mainServer, opsServer, cancel)
}

// loadConfiguration loads the application configuration
func loadConfiguration() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load configuration: " + err.Error())
	}
	return cfg
}

// initializeLogger sets up the structured logger
func initializeLogger(cfg *config.Config) {
	if err := logger.Initialize(cfg.Logging.Level, cfg.Logging.Format); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
}

// logServiceStartup logs service startup information
func logServiceStartup(cfg *config.Config) {
	logger.Info(constants.LogMsgStartingService,
		constants.LogFieldName, cfg.App.Name,
		constants.LogFieldEnv, cfg.App.Env,
		constants.LogFieldPort, cfg.App.Port,
		constants.LogFieldOpsPort, cfg.App.OpsPort,
	)
}

// initializeDatabase creates and returns a database connection
func initializeDatabase(ctx context.Context, cfg *config.Config) *database.Database {
	db, err := database.Initialize(ctx, &cfg.Database)
	if err != nil {
		logger.Fatal(constants.LogMsgFailedToInitDB, constants.LogKeyError, err)
	}
	return db
}

// modules holds all initialized application modules
type modules struct {
	account     account.IModule
	transaction transaction.IModule
	health      health.IModule
}

// initializeModules creates all application modules
func initializeModules(ctx context.Context, db *database.Database) *modules {
	accountModule := account.NewModule(ctx, db.GetPool())
	transactionModule := transaction.NewModule(ctx, db.GetPool(), accountModule.GetRepository())
	healthModule := health.NewModule(ctx, db)

	return &modules{
		account:     accountModule,
		transaction: transactionModule,
		health:      healthModule,
	}
}

// setupAndStartServers configures routers and starts HTTP servers
func setupAndStartServers(cfg *config.Config, m *modules) (*http.Server, *http.Server) {
	mainRouter := setupMainRouter(m)
	opsRouter := setupOpsRouter(m)

	mainServer := createServer(cfg.App.Port, mainRouter)
	opsServer := createServer(cfg.App.OpsPort, opsRouter)

	startServer(mainServer, constants.LogMsgMainServerStarting, constants.LogMsgMainServerFailed)
	startServer(opsServer, constants.LogMsgOpsServerStarting, constants.LogMsgOpsServerFailed)

	return mainServer, opsServer
}

// setupMainRouter creates the main API router with middleware
func setupMainRouter(m *modules) chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(requestLogger)

	m.account.GetHandler().RegisterRoutes(router)
	m.transaction.GetHandler().RegisterRoutes(router)

	return router
}

// setupOpsRouter creates the operations router for health and metrics
func setupOpsRouter(m *modules) chi.Router {
	router := chi.NewRouter()
	healthHandler := health.NewHTTPHandler(m.health.GetCore())
	healthHandler.RegisterRoutes(router)
	router.Handle("/metrics", promhttp.Handler())
	return router
}

// createServer creates an HTTP server with the given address and handler
func createServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}

// startServer starts an HTTP server in a goroutine
func startServer(server *http.Server, startMsg, failMsg string) {
	go func() {
		logger.Info(startMsg, constants.LogFieldAddr, server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(failMsg, constants.LogKeyError, err)
		}
	}()
}

// waitForShutdownSignal waits for OS signals and performs graceful shutdown
func waitForShutdownSignal(ctx context.Context, cfg *config.Config, healthModule health.IModule, mainServer, opsServer *http.Server, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	logger.Info(constants.LogMsgShutdownSignalReceived)

	performGracefulShutdown(ctx, cfg, healthModule, mainServer, opsServer)

	cancel()
	logger.Info(constants.LogMsgServiceStopped)
}

// performGracefulShutdown handles the graceful shutdown sequence
func performGracefulShutdown(ctx context.Context, cfg *config.Config, healthModule health.IModule, mainServer, opsServer *http.Server) {
	healthModule.GetCore().MarkUnhealthy()

	waitForConnectionDrain(cfg.App.ShutdownDelay)

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, time.Duration(cfg.App.ShutdownTimeout)*time.Second)
	defer shutdownCancel()

	waitForServerShutdown(shutdownCtx, mainServer, opsServer)
}

// waitForConnectionDrain waits for load balancer to drain connections
func waitForConnectionDrain(delaySeconds int) {
	logger.Info(constants.LogMsgWaitingForShutdownDelay, constants.LogFieldDelaySeconds, delaySeconds)
	time.Sleep(time.Duration(delaySeconds) * time.Second)
}

// waitForServerShutdown shuts down servers and waits for completion
func waitForServerShutdown(ctx context.Context, mainServer, opsServer *http.Server) {
	shutdownComplete := make(chan struct{})
	go func() {
		defer close(shutdownComplete)
		shutdownServers(ctx, mainServer, opsServer)
	}()

	select {
	case <-shutdownComplete:
		logger.Info(constants.LogMsgGracefulShutdownComplete)
	case <-ctx.Done():
		logger.Warn(constants.LogMsgShutdownTimeoutExceeded)
	}
}

// shutdownServers gracefully shuts down the main and ops servers
func shutdownServers(ctx context.Context, mainServer, opsServer *http.Server) {
	if err := mainServer.Shutdown(ctx); err != nil {
		logger.Error(constants.LogMsgMainServerShutdownErr, constants.LogKeyError, err)
		return
	}
	logger.Info(constants.LogMsgMainServerShutdownDone)

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
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		logHTTPRequest(r, ww, time.Since(start))
	})
}

// logHTTPRequest logs the completed HTTP request details
func logHTTPRequest(r *http.Request, ww middleware.WrapResponseWriter, duration time.Duration) {
	logger.Ctx(r.Context()).Infow(constants.LogMsgHTTPRequestCompleted,
		constants.LogKeyMethod, r.Method,
		constants.LogKeyPath, r.URL.Path,
		constants.LogKeyStatusCode, ww.Status(),
		constants.LogKeyDuration, duration.Milliseconds(),
		constants.LogFieldBytesWritten, ww.BytesWritten(),
	)
}
