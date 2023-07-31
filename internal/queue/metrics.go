package queue

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	ExploitInstancesRunning *prometheus.GaugeVec
	ExploitsFinished        *prometheus.CounterVec
	ExploitsFailed          *prometheus.CounterVec
	ExploitRunTime          *prometheus.HistogramVec

	MaxJobs prometheus.Gauge
}

func NewMetrics(namespace, id string, queueType Type) *Metrics {
	const subsystem = "exploit_queue"
	exploitLabels := []string{"exploit_id", "exploit_version", "exploit_type"}
	constLabels := prometheus.Labels{
		"queue_type": queueType.String(),
		"queue_id":   id,
	}

	return &Metrics{
		ExploitInstancesRunning: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "exploit_instances_running",
				Help:        "Number of exploit instances currently running",
				ConstLabels: constLabels,
			},
			exploitLabels,
		),
		ExploitsFinished: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "exploits_finished_total",
				Help:        "Number of exploits finished",
				ConstLabels: constLabels,
			},
			exploitLabels,
		),
		ExploitsFailed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "exploits_failed_total",
				Help:        "Number of exploits failed",
				ConstLabels: constLabels,
			},
			exploitLabels,
		),
		ExploitRunTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "exploit_run_time_seconds",
				Help:        "Time it took to run an exploit",
				ConstLabels: constLabels,
				Buckets:     prometheus.ExponentialBucketsRange(0.001, 300, 30),
			},
			exploitLabels,
		),

		MaxJobs: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "max_jobs",
				Help:        "Maximum number of jobs for the current runner",
				ConstLabels: constLabels,
			},
		),
	}
}
