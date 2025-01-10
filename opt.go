package metric

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type statusResponseWriter struct {
	http.ResponseWriter
	code int
}

func (w statusResponseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

// WithSkipPaths should be called first to work for all middlewares
func WithSkipPaths(paths []string) func(*Metric) {
	return func(m *Metric) {
		m.skipPaths = paths
	}
}

func WithCustom(item Item) func(*Metric) {
	return func(m *Metric) {
		attachItems(m, item)
	}
}

func WithBasic() func(*Metric) {
	slowTime := time.Second * 5
	buckets := []float64{0.1, 0.3, 1.2, 5, 10}

	return func(m *Metric) {
		attachItems(
			m,
			slowRequestTotal(m.skipPaths, slowTime),
			requestDuration(m.skipPaths, buckets),
			requestTotal(m.skipPaths),
		)
	}
}

func WithGoRuntime() func(*Metric) {
	return func(m *Metric) {
		m.collectors["go"] = collectors.NewGoCollector()
	}
}

func WithProcessMetrics() func(*Metric) {
	return func(m *Metric) {
		m.collectors["process"] = collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})
	}
}

func WithRequestTotal() func(*Metric) {
	return func(m *Metric) {
		attachItems(m, requestTotal(m.skipPaths))
	}
}

func WithRequestDuration(buckets []float64) func(*Metric) {
	return func(m *Metric) {
		attachItems(m, requestDuration(m.skipPaths, buckets))
	}
}

func WithSlowRequest(slowTime time.Duration) func(*Metric) {
	if slowTime <= 0 {
		slowTime = 5 * time.Second
	}

	return func(m *Metric) {
		attachItems(m, slowRequestTotal(m.skipPaths, slowTime))
	}
}

func requestTotal(skipPaths []string) Item {
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

	return Item{
		Name:       name,
		Collector:  vec,
		Middleware: mw,
	}
}

func requestDuration(skipPaths []string, buckets []float64) Item {
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

	return Item{
		Name:       name,
		Collector:  vec,
		Middleware: mw,
	}
}

func slowRequestTotal(skipPaths []string, slowTime time.Duration) Item {
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

	return Item{
		Name:       name,
		Collector:  vec,
		Middleware: mw,
	}
}
