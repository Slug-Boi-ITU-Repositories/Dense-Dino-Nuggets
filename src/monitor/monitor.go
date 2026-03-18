package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	RequestCounter *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		RequestCounter: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "requests_count",
				Help: "Number of requests made to Minitwit",
			}, []string{"path", "method"}),
		RequestDuration: promauto.With(reg).NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "request_duration",
				Help: "Duration of processing time for requests to Minitwit in seconds",
				Buckets: []float64{0.001, 0.01, 0.05, 0.1, 0.2, 0.3, 0.5, 1},
			}, []string{"path", "method"}),
	}
	return m
}