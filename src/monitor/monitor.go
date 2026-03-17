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
			}, []string{"path", "method"}),
	}
	return m
}