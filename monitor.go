package metric

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Option func(*Monitor)

type Monitor struct {
	skipPaths   []string
	collectors  map[string]prometheus.Collector
	middlewares []func(http.Handler) http.Handler
	registry    *prometheus.Registry
}

func NewMonitor(opts ...Option) (*Monitor, error) {

	m := &Monitor{
		registry:   prometheus.NewRegistry(),
		collectors: make(map[string]prometheus.Collector),
	}

	for _, opt := range opts {
		opt(m)
	}

	for name, collector := range m.collectors {
		if err := m.registry.Register(collector); err != nil {
			return nil, fmt.Errorf("register metric %s: %w", name, err)
		}
	}

	return m, nil
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

type Metric struct {
	Name       string
	Collector  prometheus.Collector
	Middleware func(http.Handler) http.Handler
}

func (m *Monitor) attach(metrics ...Metric) {
	for _, metric := range metrics {

		if _, ok := m.collectors[metric.Name]; ok {
			continue
		}

		if metric.Middleware == nil || metric.Collector == nil {
			continue
		}

		m.collectors[metric.Name] = metric.Collector
		m.middlewares = append(m.middlewares, metric.Middleware)
	}
}
