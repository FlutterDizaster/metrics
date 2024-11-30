package collector

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/Plasmat1x/metrics/internal/models"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

var (
	ErrContainersListNotReceived = errors.New("containers list is not received")
)

type MetricsCollector struct {
	dClient *client.Client
}

func New(ctx context.Context) (*MetricsCollector, error) {
	slog.Debug("creating metrics controller")

	ctrl := &MetricsCollector{}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	cli.NegotiateAPIVersion(ctx)

	ctrl.dClient = cli

	return ctrl, nil
}

func (c *MetricsCollector) Collect(ctx context.Context) ([]models.Metrics, error) {
	slog.Debug("collecting metrics")
	// Get containers list
	list, err := c.dClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, errors.Join(ErrContainersListNotReceived, err)
	}

	slog.Debug("containers list received", slog.Int("count", len(list)), func() slog.Attr {
		var attrs []any
		for i := range list {
			attrs = append(attrs, slog.Group(
				list[i].ID,
				slog.String("name", list[i].Names[0]),
				slog.String("status", list[i].Status),
				slog.String("state", list[i].State),
			))
		}
		return slog.Group("containers", attrs...)
	}())

	metrics := make([]models.Metrics, 0, len(list))
	mx := sync.Mutex{}
	wg := sync.WaitGroup{}

	// Get metrics for each container
	for i := range list {
		wg.Add(1)
		go func(cont types.Container) {
			containerMetrics, mErr := c.collectContainerMetrics(ctx, cont)
			if mErr != nil {
				slog.Error("failed to collect metrics for container", slog.Any("error", mErr))
				return
			}

			slog.Debug(
				"metrics collected",
				slog.String("container-id", cont.ID),
			)

			mx.Lock()
			metrics = append(metrics, containerMetrics)
			mx.Unlock()
			wg.Done()
		}(list[i])
	}

	wg.Wait()

	slog.Debug("metrics collection completed")
	return metrics, nil
}

func (c *MetricsCollector) collectContainerMetrics(
	ctx context.Context,
	cont types.Container,
) (models.Metrics, error) {
	m := models.Metrics{
		ID:     cont.ID,
		Name:   cont.Names[0],
		Status: cont.Status,
		State:  cont.State,
	}

	if cont.State != "running" {
		m.Time = time.Now()
		return m, nil
	}

	rawStats, err := c.dClient.ContainerStats(ctx, cont.ID, false)
	if err != nil {
		slog.Error("failed to get container stats", slog.Any("error", err))
		return models.Metrics{}, err
	}
	defer rawStats.Body.Close()

	body, err := io.ReadAll(rawStats.Body)
	if err != nil {
		slog.Error("failed to read container stats", slog.Any("error", err))
		return models.Metrics{}, err
	}

	stats := container.Stats{}
	err = json.Unmarshal(body, &stats)
	if err != nil {
		slog.Error("failed to unmarshal container stats", slog.Any("error", err))
		return models.Metrics{}, err
	}

	// Data mapping
	m.Time = stats.Read

	m.OnlineCPUs = stats.CPUStats.OnlineCPUs
	m.TotalCPUUsage = stats.CPUStats.CPUUsage.TotalUsage
	m.PerCPUUsage = stats.CPUStats.CPUUsage.PercpuUsage

	m.MemUsage = stats.MemoryStats.Usage
	m.MaxMemUsage = stats.MemoryStats.MaxUsage
	m.MemLimit = stats.MemoryStats.Limit

	return m, nil
}
