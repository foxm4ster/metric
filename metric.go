package metric

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Option func(*Metric)

type Item struct {
	Name       string
	Collector  prometheus.Collector
	Middleware func(http.Handler) http.Handler
}

type Metric struct {
	skipPaths   []string
	collectors  map[string]prometheus.Collector
	middlewares []func(http.Handler) http.Handler
	registry    *prometheus.Registry
}

func New(opts ...Option) (*Metric, error) {

	m := &Metric{
		registry:   prometheus.NewRegistry(),
		collectors: make(map[string]prometheus.Collector),
	}

	for _, opt := range opts {
		opt(m)
	}

	for _, collector := range m.collectors {
		if err := m.registry.Register(collector); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m Metric) Middlewares() []func(http.Handler) http.Handler {
	return m.middlewares
}

func (m Metric) Expose() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func attachItems(m *Metric, items ...Item) {
	for _, i := range items {
		if _, ok := m.collectors[i.Name]; !ok {
			m.collectors[i.Name] = i.Collector
			m.middlewares = append(m.middlewares, i.Middleware)
		}
	}
}
