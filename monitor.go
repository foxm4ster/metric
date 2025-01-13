package metric

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Option func(*Monitor)

type Metric struct {
	Name       string
	Collector  prometheus.Collector
	Middleware func(http.Handler) http.Handler
}

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

func (m Monitor) Middlewares() []func(http.Handler) http.Handler {
	return m.middlewares
}

func (m Monitor) Expose() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func attachItems(m *Monitor, items ...Metric) {
	for _, i := range items {
		if _, ok := m.collectors[i.Name]; !ok {
			m.collectors[i.Name] = i.Collector
			m.middlewares = append(m.middlewares, i.Middleware)
		}
	}
}
