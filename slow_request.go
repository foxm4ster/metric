package metric

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func slowRequestTotal(skipPaths []string, slowTime time.Duration) Metric {

	if slowTime <= 0 {
		slowTime = 5 * time.Second
	}

	name := "slow_request_total"
	vec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: fmt.Sprintf("The server handled slow requests counter, t=%d.", slowTime),
		},
		[]string{"uri", "method", "code"},
	)
	mw := func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			if slices.Contains(skipPaths, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			srw := &statusResponseWriter{ResponseWriter: w}

			next.ServeHTTP(srw, r)

			if time.Since(start) > slowTime {
				vec.WithLabelValues(r.URL.Path, r.Method, strconv.Itoa(srw.code)).Inc()
			}
		}
		return http.HandlerFunc(fn)
	}

	return Metric{
		Name:       name,
		Collector:  vec,
		Middleware: mw,
	}
}
