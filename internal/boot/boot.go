// Package boot provides application initialization and bootstrapping.
package boot

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/internal-transfers-service/internal/config"
	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/interceptors"
	"github.com/internal-transfers-service/internal/logger"
	"github.com/internal-transfers-service/internal/modules/account"
	"github.com/internal-transfers-service/internal/modules/health"
	"github.com/internal-transfers-service/internal/modules/idempotency"
	"github.com/internal-transfers-service/internal/modules/transaction"
	"github.com/internal-transfers-service/internal/tracing"
	"github.com/internal-transfers-service/pkg/database"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Default cleanup interval for idempotency keys
const idempotencyCleanupInterval = 1 * time.Hour

// App holds all application dependencies
type App struct {
	Config     *config.Config
	Database   *database.Database
	Tracer     *tracing.Tracer
	Modules    *Modules
	MainServer *http.Server
	OpsServer  *http.Server
	MainRouter chi.Router
	OpsRouter  chi.Router
}

// Modules holds all application modules
type Modules struct {
	Account     account.IModule
	Transaction transaction.IModule
	Health      health.IModule
	Idempotency idempotency.IModule
}

// Initialize creates and initializes all application dependencies.
func Initialize(ctx context.Context) (*App, error) {
	app := &App{}

	if err := app.loadConfig(); err != nil {
		return nil, err
	}

	if err := app.initLogger(); err != nil {
		return nil, err
	}

	app.logStartup()

	if err := app.initTracer(ctx); err != nil {
		return nil, err
	}

	if err := app.initDatabase(ctx); err != nil {
		return nil, err
	}

	app.initModules(ctx)
	app.setupRouters()
	app.createServers()

	return app, nil
}

// loadConfig loads the application configuration
func (a *App) loadConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	a.Config = cfg
	return nil
}

// initLogger initializes the structured logger
func (a *App) initLogger() error {
	return logger.Initialize(a.Config.Logging.Level, a.Config.Logging.Format)
}

// logStartup logs service startup information
func (a *App) logStartup() {
	logger.Info(constants.LogMsgStartingService,
		constants.LogFieldName, a.Config.App.Name,
		constants.LogFieldEnv, a.Config.App.Env,
		constants.LogFieldPort, a.Config.App.Port,
		constants.LogFieldOpsPort, a.Config.App.OpsPort,
	)
}

// initTracer initializes the OpenTelemetry tracer
func (a *App) initTracer(ctx context.Context) error {
	tracer, err := tracing.Initialize(ctx, &a.Config.Tracing, a.Config.App.Name)
	if err != nil {
		logger.Error(constants.LogMsgTracerInitFailed, constants.LogKeyError, err)
		return err
	}
	a.Tracer = tracer
	return nil
}

// initDatabase initializes the database connection with retry support
func (a *App) initDatabase(ctx context.Context) error {
	db, err := database.InitializeWithRetry(ctx, &a.Config.Database)
	if err != nil {
		logger.Fatal(constants.LogMsgFailedToInitDB, constants.LogKeyError, err)
		return err
	}
	a.Database = db
	return nil
}

// initModules initializes all application modules
func (a *App) initModules(ctx context.Context) {
	accountModule := account.NewModule(ctx, a.Database.GetPool())
	transactionModule := transaction.NewModule(ctx, a.Database.GetPool(), accountModule.GetRepository())
	healthModule := health.NewModule(ctx, a.Database)
	idempotencyModule := idempotency.NewModule(ctx, a.Database.GetPool())

	// Start idempotency cleanup worker
	ttl := a.getIdempotencyTTL()
	idempotencyModule.StartCleanupWorker(ctx, ttl, idempotencyCleanupInterval)

	a.Modules = &Modules{
		Account:     accountModule,
		Transaction: transactionModule,
		Health:      healthModule,
		Idempotency: idempotencyModule,
	}
}

// getIdempotencyTTL parses the idempotency TTL from configuration
func (a *App) getIdempotencyTTL() time.Duration {
	ttl, err := time.ParseDuration(a.Config.Idempotency.TTL)
	if err != nil {
		return 24 * time.Hour // Default to 24 hours
	}
	return ttl
}

// setupRouters creates and configures the HTTP routers
func (a *App) setupRouters() {
	a.MainRouter = a.createMainRouter()
	a.OpsRouter = a.createOpsRouter()
}

// createMainRouter creates the main API router with middleware
func (a *App) createMainRouter() chi.Router {
	router := chi.NewRouter()

	// Apply middleware chain from interceptors package (with idempotency, rate limiting, and tracing)
	middlewareCfg := interceptors.MiddlewareConfig{
		RateLimit:   a.Config.RateLimit,
		Tracing:     a.Config.Tracing,
		Idempotency: a.Modules.Idempotency.GetRepository(),
	}
	for _, mw := range interceptors.GetChiMiddlewareWithFullConfig(middlewareCfg) {
		router.Use(mw)
	}

	// Register versioned API routes under /v1
	router.Route(constants.APIVersionPrefix, func(r chi.Router) {
		a.Modules.Account.GetHandler().RegisterRoutes(r)
		a.Modules.Transaction.GetHandler().RegisterRoutes(r)
	})

	return router
}

// createOpsRouter creates the operations router for health and metrics
func (a *App) createOpsRouter() chi.Router {
	router := chi.NewRouter()

	healthHandler := health.NewHTTPHandler(a.Modules.Health.GetCore())
	healthHandler.RegisterRoutes(router)

	router.Handle(constants.RouteMetrics, promhttp.Handler())

	return router
}

// createServers creates HTTP servers
func (a *App) createServers() {
	a.MainServer = &http.Server{
		Addr:    a.Config.App.Port,
		Handler: a.MainRouter,
	}

	a.OpsServer = &http.Server{
		Addr:    a.Config.App.OpsPort,
		Handler: a.OpsRouter,
	}
}

// Start starts all HTTP servers
func (a *App) Start() {
	go a.startMainServer()
	go a.startOpsServer()
}

// startMainServer starts the main API server
func (a *App) startMainServer() {
	logger.Info(constants.LogMsgMainServerStarting, constants.LogFieldAddr, a.MainServer.Addr)
	if err := a.MainServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(constants.LogMsgMainServerFailed, constants.LogKeyError, err)
	}
}

// startOpsServer starts the ops server
func (a *App) startOpsServer() {
	logger.Info(constants.LogMsgOpsServerStarting, constants.LogFieldAddr, a.OpsServer.Addr)
	if err := a.OpsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(constants.LogMsgOpsServerFailed, constants.LogKeyError, err)
	}
}

// Shutdown performs graceful shutdown
func (a *App) Shutdown(ctx context.Context) {
	logger.Info(constants.LogMsgShutdownSignalReceived)

	// Mark service as unhealthy
	a.Modules.Health.GetCore().MarkUnhealthy()

	// Stop idempotency cleanup worker
	a.Modules.Idempotency.StopCleanupWorker()

	// Wait for load balancer to drain connections
	a.waitForConnectionDrain()

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, time.Duration(a.Config.App.ShutdownTimeout)*time.Second)
	defer cancel()

	// Shutdown servers
	a.shutdownServers(shutdownCtx)

	// Close database
	a.Database.Close()

	// Shutdown tracer
	a.Tracer.Shutdown(shutdownCtx)

	logger.Info(constants.LogMsgServiceStopped)
}

// waitForConnectionDrain waits for load balancer to drain connections
func (a *App) waitForConnectionDrain() {
	delay := a.Config.App.ShutdownDelay
	logger.Info(constants.LogMsgWaitingForShutdownDelay, constants.LogFieldDelaySeconds, delay)
	time.Sleep(time.Duration(delay) * time.Second)
}

// shutdownServers gracefully shuts down all servers
func (a *App) shutdownServers(ctx context.Context) {
	done := make(chan struct{})

	go func() {
		defer close(done)
		a.shutdownServer(ctx, a.MainServer, constants.ServerNameMain)
		a.shutdownServer(ctx, a.OpsServer, constants.ServerNameOps)
	}()

	select {
	case <-done:
		logger.Info(constants.LogMsgGracefulShutdownComplete)
	case <-ctx.Done():
		logger.Warn(constants.LogMsgShutdownTimeoutExceeded)
	}
}

// shutdownServer shuts down a single server
func (a *App) shutdownServer(ctx context.Context, server *http.Server, name string) {
	if err := server.Shutdown(ctx); err != nil {
		logShutdownError(name, err)
		return
	}
	logShutdownComplete(name)
}

// logShutdownError logs server shutdown error based on server name
func logShutdownError(name string, err error) {
	errMsg := getShutdownErrorMessage(name)
	logger.Error(errMsg, constants.LogKeyError, err)
}

// getShutdownErrorMessage returns the appropriate error message for the server
func getShutdownErrorMessage(name string) string {
	if name == constants.ServerNameMain {
		return constants.LogMsgMainServerShutdownErr
	}
	return constants.LogMsgOpsServerShutdownErr
}

// logShutdownComplete logs server shutdown completion based on server name
func logShutdownComplete(name string) {
	doneMsg := getShutdownDoneMessage(name)
	logger.Info(doneMsg)
}

// getShutdownDoneMessage returns the appropriate done message for the server
func getShutdownDoneMessage(name string) string {
	if name == constants.ServerNameMain {
		return constants.LogMsgMainServerShutdownDone
	}
	return constants.LogMsgOpsServerShutdownDone
}

// GetEnv returns the current environment
func GetEnv() string {
	env := os.Getenv(constants.EnvKeyAppEnv)
	if env == "" {
		env = constants.EnvDefaultDev
	}
	return env
}
