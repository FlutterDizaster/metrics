package controller

import (
	"context"
	"log/slog"

	"github.com/Plasmat1x/metrics/internal/models"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollector interface {
	Collect(ctx context.Context) ([]models.Metrics, error)
}

type Controller struct {
	collector MetricsCollector

	onlineCPUs    *prometheus.GaugeVec
	totalCPUUsage *prometheus.GaugeVec
	perCPUUsage   *prometheus.GaugeVec
	memUsage      *prometheus.GaugeVec
	maxMemUsage   *prometheus.GaugeVec
	memLimit      *prometheus.GaugeVec
}

func New(collector MetricsCollector) *Controller {
	slog.Debug("creating metrics controller")
	labels := []string{
		"id",
		"name",
		"status",
		"state",
	}

	c := &Controller{
		collector: collector,
		onlineCPUs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "online_cpus",
				Help: "Online CPUs",
			},
			labels,
		),
		totalCPUUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "total_cpu_usage",
				Help: "Total CPU usage",
			},
			labels,
		),
		perCPUUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "per_cpu_usage",
				Help: "Per CPU usage",
			},
			append(labels, "cpu"),
		),
		memUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mem_usage",
				Help: "Memory usage",
			},
			labels,
		),
		maxMemUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "max_mem_usage",
				Help: "Maximum memory usage",
			},
			labels,
		),
		memLimit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mem_limit",
				Help: "Memory limit",
			},
			labels,
		),
	}

	slog.Debug("registering metrics")

	prometheus.MustRegister(c.onlineCPUs)
	prometheus.MustRegister(c.totalCPUUsage)
	prometheus.MustRegister(c.perCPUUsage)
	prometheus.MustRegister(c.memUsage)
	prometheus.MustRegister(c.maxMemUsage)
	prometheus.MustRegister(c.memLimit)

	return c
}

func (c *Controller) Collect(ctx context.Context) error {
	slog.Debug("collecting metrics")
	metrics, err := c.collector.Collect(ctx)
	if err != nil {
		slog.Error("failed to collect metrics", slog.Any("error", err))
		return err
	}

	for i := range metrics {
		slog.Debug("Updating metrics for container", slog.String("id", metrics[i].ID))

		values := []string{
			metrics[i].ID,
			metrics[i].Name,
			metrics[i].Status,
			metrics[i].State,
		}

		// CPU metrics
		c.onlineCPUs.WithLabelValues(values...).Set(float64(metrics[i].OnlineCPUs))
		c.totalCPUUsage.WithLabelValues(values...).Set(float64(metrics[i].TotalCPUUsage))

		for j, usage := range metrics[i].PerCPUUsage {
			cpuValues := append(values, string(rune(j)))
			c.perCPUUsage.WithLabelValues(cpuValues...).Set(float64(usage))
		}

		// Memory metrics
		c.memUsage.WithLabelValues(values...).Set(float64(metrics[i].MemUsage))
		c.maxMemUsage.WithLabelValues(values...).Set(float64(metrics[i].MaxMemUsage))
		c.memLimit.WithLabelValues(values...).Set(float64(metrics[i].MemLimit))
	}

	slog.Debug("metrics collected")

	return nil
}
