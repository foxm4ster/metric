package metric

import (
	"time"

	"github.com/prometheus/client_golang/prometheus/collectors"
)

// WithSkipPaths should be called first to work for all middlewares
func WithSkipPaths(paths []string) func(*Options) {
	return func(o *Options) {
		o.skipPaths = paths
	}
}

func WithCustom(fn func() Metric) func(*Options) {
	return func(o *Options) {
		o.metrics = append(o.metrics, fn())
	}
}

func WithBasic() func(*Options) {
	return func(o *Options) {
		o.metrics = append(o.metrics,
			slowRequestTotal(o.skipPaths, 0),
			requestDuration(o.skipPaths, nil),
			requestTotal(o.skipPaths),
		)
	}
}

func WithGoRuntime() func(*Options) {
	return func(o *Options) {
		o.metrics = append(o.metrics, Metric{
			Name:      "go",
			Collector: collectors.NewGoCollector(),
		})
	}
}

func WithProcess() func(*Options) {
	return func(o *Options) {
		o.metrics = append(o.metrics, Metric{
			Name:      "process",
			Collector: collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		})
	}
}

func WithRequestTotal() func(*Options) {
	return func(o *Options) {
		o.metrics = append(o.metrics, requestTotal(o.skipPaths))
	}
}

func WithRequestDuration(buckets []float64) func(*Options) {
	return func(o *Options) {
		o.metrics = append(o.metrics, requestDuration(o.skipPaths, buckets))
	}
}

func WithSlowRequest(slowTime time.Duration) func(*Options) {
	return func(o *Options) {
		o.metrics = append(o.metrics, slowRequestTotal(o.skipPaths, slowTime))
	}
}
