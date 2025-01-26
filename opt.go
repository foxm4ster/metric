package metric

import (
	"time"

	"github.com/prometheus/client_golang/prometheus/collectors"
)

// WithSkipPaths should be called first to work for all middlewares
func WithSkipPaths(paths ...string) func(*MonitorOptions) {
	return func(o *MonitorOptions) {
		o.skipPaths = paths
	}
}

func WithCustom(fn func() Metric) func(*MonitorOptions) {
	return func(o *MonitorOptions) {
		o.metrics = append(o.metrics, fn())
	}
}

func WithGoRuntime() func(*MonitorOptions) {
	return func(o *MonitorOptions) {
		o.metrics = append(o.metrics, Metric{
			Name:      "go",
			Collector: collectors.NewGoCollector(),
		})
	}
}

func WithProcess() func(*MonitorOptions) {
	return func(o *MonitorOptions) {
		o.metrics = append(o.metrics, Metric{
			Name:      "process",
			Collector: collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		})
	}
}

func WithRequestTotal() func(*MonitorOptions) {
	return func(o *MonitorOptions) {
		o.metrics = append(o.metrics, requestTotal(o.skipPaths))
	}
}

func WithRequestDuration(buckets []float64) func(*MonitorOptions) {
	return func(o *MonitorOptions) {
		o.metrics = append(o.metrics, requestDuration(o.skipPaths, buckets))
	}
}

func WithSlowRequest(slowTime time.Duration) func(*MonitorOptions) {
	return func(o *MonitorOptions) {
		o.metrics = append(o.metrics, slowRequestTotal(o.skipPaths, slowTime))
	}
}
