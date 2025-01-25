package metric

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

func requestTotal(skipPaths []string) Metric {
	name := "request_total"
	vec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: "All the server received request num with every uri.",
		},
		[]string{"uri", "method", "status_code"},
	)

	mw := func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			if slices.Contains(skipPaths, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			srw := &statusResponseWriter{ResponseWriter: w}

			next.ServeHTTP(srw, r)

			vec.WithLabelValues(r.URL.Path, r.Method, strconv.Itoa(srw.code)).Inc()
		}
		return http.HandlerFunc(fn)
	}

	return Metric{
		Name:       name,
		Collector:  vec,
		Middleware: mw,
	}
}
