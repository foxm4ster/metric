package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"metric"
)

func main() {
	mtr, err := metric.NewMonitor(
		metric.WithSkipPaths("/health"), // should be called first to enable for all metrics
		metric.WithRequestTotal(),
		metric.WithSlowRequest(time.Second*10),
		metric.WithRequestDuration(nil),
		metric.WithGoRuntime(),
		metric.WithProcess(),
		metric.WithCustom(customMetric),
	)
	if err != nil {
		panic(err)
	}

	var handler http.Handler = http.HandlerFunc(example)

	for _, mw := range mtr.Middlewares() {
		handler = mw(handler)
	}

	http.Handle("/metrics", mtr.Expose())
	http.Handle("/", handler)

	http.ListenAndServe(":8080", nil)
}

func example(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func customMetric() metric.Metric {
	name := "custom_func"
	vec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: "Custom metric desc.",
		},
		[]string{"uri", "method"},
	)

	mw := func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			vec.WithLabelValues(r.URL.Path, r.Method).Inc()
		}
		return http.HandlerFunc(fn)
	}

	return metric.Metric{
		Name:       name,
		Collector:  vec,
		Middleware: mw,
	}
}
