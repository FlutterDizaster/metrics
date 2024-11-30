package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsController interface {
	Collect(ctx context.Context) error
}

type Handler struct {
	promHandler http.Handler
	router      *http.ServeMux
	ctrl        MetricsController
}

func New(controller MetricsController) *Handler {
	h := &Handler{
		promHandler: promhttp.Handler(),
		ctrl:        controller,
	}

	h.setupRoutes()

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *Handler) setupRoutes() {
	router := http.NewServeMux()

	router.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("new incoming request", slog.String("path", r.URL.Path))
		startTime := time.Now()
		err := h.ctrl.Collect(r.Context())
		if err != nil {
			slog.Error("failed to collect metrics", slog.Any("error", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		h.promHandler.ServeHTTP(w, r)

		slog.Info(
			"request processed",
			slog.String("path", r.URL.Path),
			slog.String("duration", time.Since(startTime).String()),
		)
	})

	h.router = router
}
