package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/bootstrap"
	"backend/internal/infrastructure/config"
	"backend/internal/infrastructure/logger"
)

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Error("load config", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app, err := bootstrap.Wire(ctx, log, cfg)
	if err != nil {
		log.Error("wire application", "err", err)
		os.Exit(1)
	}

	go app.Worker.Start(ctx)

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           app.Handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Info("starting api", "addr", server.Addr)
	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "err", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown server", "err", err)
	}
	if err := app.Close(shutdownCtx); err != nil {
		log.Error("shutdown dependencies", "err", err)
	}
	log.Info("api stopped")
}
