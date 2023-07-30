package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	AliveClients prometheus.Gauge
}

func NewMetrics() *Metrics {
	return &Metrics{
		AliveClients: promauto.NewGauge(prometheus.GaugeOpts{
			Subsystem: "server",
			Name:      "alive_clients",
			Help:      "Number of alive clients",
		}),
	}
}
