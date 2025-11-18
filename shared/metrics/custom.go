package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// NewCounter creates a new Counter metric and registers it
func (m *Metrics) NewCounter(name, help string, labels []string) *prometheus.CounterVec {
	factory := promauto.With(m.registry)
	return factory.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
}

// NewGauge creates a new Gauge metric and registers it
func (m *Metrics) NewGauge(name, help string, labels []string) *prometheus.GaugeVec {
	factory := promauto.With(m.registry)
	return factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
}

// NewHistogram creates a new Histogram metric and registers it
func (m *Metrics) NewHistogram(name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	factory := promauto.With(m.registry)

	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	return factory.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)
}

// NewSummary creates a new Summary metric and registers it
func (m *Metrics) NewSummary(name, help string, labels []string) *prometheus.SummaryVec {
	factory := promauto.With(m.registry)
	return factory.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
}
