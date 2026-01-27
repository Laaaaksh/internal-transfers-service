// Package main is the entry point for the internal transfers service.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/internal-transfers-service/internal/boot"
	"github.com/internal-transfers-service/internal/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := boot.Initialize(ctx)
	if err != nil {
		panic("failed to initialize application: " + err.Error())
	}
	defer logger.Sync()

	app.Start()

	waitForShutdown(ctx, app, cancel)
}

// waitForShutdown waits for OS signals and performs graceful shutdown.
func waitForShutdown(ctx context.Context, app *boot.App, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	app.Shutdown(ctx)
	cancel()
}
