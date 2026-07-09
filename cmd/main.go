package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/ishanwardhono/community-waste/pkg/config"
	"github.com/ishanwardhono/community-waste/pkg/logger"
)

func main() {
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	logger.Init(cfg.LogLevel)
	ctx := context.Background()

	app, err := NewApp(cfg)
	if err != nil {
		logger.Fatalf(ctx, "init app: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Infof(ctx, "listening on :%s", cfg.AppPort)
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-stop:
		logger.Infof(ctx, "shutdown signal received")
	case err := <-errCh:
		logger.Errorf(ctx, "server error: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
	defer cancel()
	if err := app.Server.Shutdown(shutdownCtx); err != nil {
		logger.Errorf(ctx, "server shutdown: %v", err)
	}
	if err := app.DB.Close(); err != nil {
		logger.Errorf(ctx, "db close: %v", err)
	}
	logger.Infof(ctx, "stopped")
}
