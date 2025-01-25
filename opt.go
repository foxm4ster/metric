package metric

import (
	"time"

	"github.com/prometheus/client_golang/prometheus/collectors"
)

// WithSkipPaths should be called first to work for all middlewares
func WithSkipPaths(paths []string) func(*Monitor) {
	return func(m *Monitor) {
		m.skipPaths = paths
	}
}

func WithCustom(fn func() Metric) func(*Monitor) {
	return func(m *Monitor) {
		m.attachMetrics(fn())
	}
}

func WithBasic() func(*Monitor) {
	return func(m *Monitor) {
		m.attachMetrics(
			slowRequestTotal(m.skipPaths, 0),
			requestDuration(m.skipPaths, nil),
			requestTotal(m.skipPaths),
		)
	}
}

func WithGoRuntime() func(*Monitor) {
	const name = "go"

	return func(m *Monitor) {
		if _, ok := m.collectors[name]; !ok {
			m.collectors[name] = collectors.NewGoCollector()
		}
	}
}

func WithProcess() func(*Monitor) {
	const name = "process"

	return func(m *Monitor) {
		if _, ok := m.collectors[name]; !ok {
			m.collectors[name] = collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})
		}
	}
}

func WithRequestTotal() func(*Monitor) {
	return func(m *Monitor) {
		m.attachMetrics(requestTotal(m.skipPaths))
	}
}

func WithRequestDuration(buckets []float64) func(*Monitor) {
	return func(m *Monitor) {
		m.attachMetrics(requestDuration(m.skipPaths, buckets))
	}
}

func WithSlowRequest(slowTime time.Duration) func(*Monitor) {
	return func(m *Monitor) {
		m.attachMetrics(slowRequestTotal(m.skipPaths, slowTime))
	}
}
