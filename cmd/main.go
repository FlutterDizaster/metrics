package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/FlutterDizaster/file-server/pkg/configloader"
	"github.com/Plasmat1x/metrics/internal/application"
)

func main() {
	os.Exit(mainWithCode())
}

func mainWithCode() int {
	slog.SetDefault(
		slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
	)

	slog.Debug("hello metrics world")

	ctx, cancle := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancle()

	settings := application.Settings{}

	err := configloader.LoadConfig(&settings)
	if err != nil {
		slog.Error("failed to load config", slog.Any("error", err))
		return 1
	}

	app, err := application.New(ctx, settings)
	if err != nil {
		slog.Error("failed to create application", slog.Any("error", err))
		return 1
	}

	err = app.Start(ctx)
	if err != nil {
		slog.Error("failed to start application", slog.Any("error", err))
		return 1
	}

	return 0
}
