package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/tahardi/halo/backend/internal/config"
	"github.com/tahardi/halo/backend/internal/server"
)

const (
	readHeaderTimeout = 10 * time.Second
	writeTimeout      = 3 * time.Minute
	shutdownTimeout   = 10 * time.Second
)

func serve(ctx context.Context, logger *slog.Logger) error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}
	if err := config.CheckPaths(cfg); err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           server.NewServer(cfg.Key, server.Deps{}, logger),
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	logger.Info("listening", "addr", cfg.ListenAddr)

	select {
	case err = <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
