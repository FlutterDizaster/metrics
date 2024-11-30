package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

type MetricsController interface {
}

type Settings struct {
	Addr string
	Port string

	Handler http.Handler

	ShutdownMaxTime time.Duration
}

type Server struct {
	server          *http.Server
	addr            string
	port            string
	shutdownMaxTime time.Duration
}

func New(settings Settings) *Server {
	s := &Server{
		addr:            settings.Addr,
		port:            settings.Port,
		shutdownMaxTime: settings.ShutdownMaxTime,
	}

	s.server = &http.Server{
		ReadHeaderTimeout: time.Second,
		Addr:              fmt.Sprintf(":%s", settings.Port),
		Handler:           settings.Handler,
	}

	return s
}

func (s *Server) Start(ctx context.Context) error {
	slog.Info("starting server")
	eg, egCtx := errgroup.WithContext(ctx)

	errorOnStart := false

	eg.Go(func() error {
		<-egCtx.Done()

		if errorOnStart {
			return nil
		}

		slog.Info("graceful shutdown", slog.String("addr", s.addr), slog.String("port", s.port))
		shutdownCtx, cancle := context.WithTimeout(context.Background(), s.shutdownMaxTime)
		defer cancle()

		return s.server.Shutdown(shutdownCtx)
	})

	eg.Go(func() error {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start server", slog.Any("error", err))
			errorOnStart = true
			return err
		}
		return nil
	})

	slog.Info(
		"server started",
		slog.String("listening on address", fmt.Sprintf("%s:%s", s.addr, s.port)),
	)

	return eg.Wait()
}
