package metric

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type statusResponseWriter struct {
	http.ResponseWriter
	code int
}

func (w statusResponseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

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

func requestDuration(skipPaths []string, buckets []float64) Metric {
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

func slowRequestTotal(skipPaths []string, slowTime time.Duration) Metric {
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
