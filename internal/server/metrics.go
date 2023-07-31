package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	AliveClients prometheus.Gauge
}

func NewMetrics(namespace string) *Metrics {
	return &Metrics{
		AliveClients: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "server",
			Name:      "alive_clients",
			Help:      "Number of alive clients",
		}),
	}
}
