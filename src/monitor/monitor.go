package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	RegisterCounter prometheus.Counter
	TweetCounter prometheus.Counter
	FollowCounter prometheus.Counter
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		RegisterCounter: promauto.With(reg).NewCounter(
			prometheus.CounterOpts{
				Name: "register_request_count",
				Help: "Number of requests to register endpoint",
			}),
		TweetCounter: promauto.With(reg).NewCounter(
			prometheus.CounterOpts{
				Name: "tweet_request_count",
				Help: "Number of requests to tweet endpoint",
			}),
		FollowCounter: promauto.With(reg).NewCounter(
			prometheus.CounterOpts{
				Name: "follow_request_count",
				Help: "Number of requests to follow endpoint. (include both follow and unfollow)",
			}),
	}
	return m
}