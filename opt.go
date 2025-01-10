package metric

import (
	"time"

	"github.com/prometheus/client_golang/prometheus/collectors"
)

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
