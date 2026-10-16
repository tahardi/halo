package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	_ "time/tzdata"
)

var errUnknownCommand = errors.New("unknown command")

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	err := dispatch(ctx, os.Args[1:], logger)
	stop()
	if err != nil {
		logger.Error("running halo", "error", err)
		os.Exit(1)
	}
}

func dispatch(ctx context.Context, args []string, logger *slog.Logger) error {
	if len(args) == 0 {
		args = []string{"serve"}
	}
	switch args[0] {
	case "serve":
		return serve(ctx, logger)
	default:
		return fmt.Errorf("%w: %s", errUnknownCommand, args[0])
	}
}
