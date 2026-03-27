package monitor

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

func MetricsMiddleware(metrics *Metrics) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := mux.CurrentRoute(r)
			path := r.URL.Path
			if route != nil {
				tmpl, err := route.GetPathTemplate()
				if err != nil {
					log.Printf("MetricsMiddleware: Error getting getting path template: %v\n", err)
				}
				path = tmpl
			}
			metrics.RequestCounter.With(prometheus.Labels{"path": path, "method": r.Method}).Inc()

			start := time.Now()
			next.ServeHTTP(w, r)

			metrics.RequestDuration.
				With(prometheus.Labels{"path": path, "method": r.Method}).
				Observe(time.Since(start).Seconds())
		})
	}
}
