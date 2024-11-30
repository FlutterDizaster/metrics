package application

import (
	"context"
	"time"

	"github.com/Plasmat1x/metrics/internal/collector"
	"github.com/Plasmat1x/metrics/internal/controller"
	"github.com/Plasmat1x/metrics/internal/server"
	"github.com/Plasmat1x/metrics/internal/server/handler"
)

const (
	serverShutdownMaxTime = 5 * time.Second
)

type Settings struct {
	Addr string `desc:"server address" env:"HTTP_ADDR" default:"localhost" name:"http-addr" short:"a"`
	Port string `desc:"server port"    env:"HTTP_PORT" default:"8080"      name:"http-port" short:"p"`
}

type Service interface {
	Start(ctx context.Context) error
}

func New(ctx context.Context, settings Settings) (Service, error) {
	// Setup collector
	metricsCollector, err := collector.New(ctx)
	if err != nil {
		return nil, err
	}

	// Setup controller
	controller := controller.New(metricsCollector)

	// Setup handler
	h := handler.New(controller)

	// Setup server
	serverSettings := server.Settings{
		Addr:            settings.Addr,
		Port:            settings.Port,
		Handler:         h,
		ShutdownMaxTime: serverShutdownMaxTime,
	}

	server := server.New(serverSettings)

	return server, nil
}
