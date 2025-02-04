package metric

import (
	"net/http"
	"slices"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func requestDuration(skipPaths []string, buckets []float64) Metric {

	if len(buckets) == 0 {
		buckets = []float64{0.1, 0.3, 1.2, 5, 10}
	}

	name := "request_duration"
	vec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    "The time server took to handle the request.",
			Buckets: buckets,
		},
		[]string{"uri"},
	)
	mw := func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			if slices.Contains(skipPaths, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			next.ServeHTTP(w, r)

			vec.WithLabelValues(r.URL.Path).Observe(time.Since(start).Seconds())
		}
		return http.HandlerFunc(fn)
	}

	return Metric{
		Name:       name,
		Collector:  vec,
		Middleware: mw,
	}
}
