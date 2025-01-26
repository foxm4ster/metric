package metric

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metric struct {
	Name       string
	Collector  prometheus.Collector
	Middleware func(http.Handler) http.Handler
}

type Options struct {
	skipPaths []string
	metrics   []Metric
}

type OptionFunc func(*Options)

type Monitor struct {
	registry    *prometheus.Registry
	collectors  map[string]prometheus.Collector
	middlewares []func(http.Handler) http.Handler
}

func NewMonitor(opts ...OptionFunc) (*Monitor, error) {

	options := &Options{}

	for _, opt := range opts {
		opt(options)
	}

	var (
		middlewares []func(http.Handler) http.Handler
		collectors  = make(map[string]prometheus.Collector)
		registry    = prometheus.NewRegistry()
	)

	for _, mtr := range options.metrics {
		if _, ok := collectors[mtr.Name]; ok {
			continue
		}

		if err := registry.Register(mtr.Collector); err != nil {
			return nil, fmt.Errorf("register '%s' metric: %w", mtr.Name, err)
		}

		collectors[mtr.Name] = mtr.Collector

		if mtr.Middleware != nil {
			middlewares = append(middlewares, mtr.Middleware)
		}
	}

	return &Monitor{
		registry:    registry,
		collectors:  collectors,
		middlewares: middlewares,
	}, nil
}

func (m *Monitor) Middlewares() []func(http.Handler) http.Handler {
	return m.middlewares
}

func (m *Monitor) Expose() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *Monitor) Register(collectors ...prometheus.Collector) error {
	for _, coll := range collectors {
		err := m.registry.Register(coll)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	return nil
}

func (m *Monitor) CounterVec(name string) *prometheus.CounterVec {
	coll, ok := m.collectors[name]
	if !ok {
		return nil
	}

	vec, ok := coll.(*prometheus.CounterVec)
	if !ok {
		return nil
	}

	return vec
}

func (m *Monitor) HistogramVec(name string) *prometheus.HistogramVec {
	coll, ok := m.collectors[name]
	if !ok {
		return nil
	}

	vec, ok := coll.(*prometheus.HistogramVec)
	if !ok {
		return nil
	}

	return vec
}

func (m *Monitor) GaugeVec(name string) *prometheus.GaugeVec {
	coll, ok := m.collectors[name]
	if !ok {
		return nil
	}

	vec, ok := coll.(*prometheus.GaugeVec)
	if !ok {
		return nil
	}

	return vec
}

func (m *Monitor) SummaryVec(name string) *prometheus.SummaryVec {
	coll, ok := m.collectors[name]
	if !ok {
		return nil
	}

	vec, ok := coll.(*prometheus.SummaryVec)
	if !ok {
		return nil
	}

	return vec
}
