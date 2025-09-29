package httpServer

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	totalRequests *prometheus.CounterVec
	durationSec   *prometheus.HistogramVec
	inFlight      prometheus.Gauge
}

func (m *metrics) metricsMiddleware(ctx *fiber.Ctx) (err error) {
	m.inFlight.Inc()
	s := time.Now()
	defer func() {
		m.inFlight.Dec()
	}()

	err = ctx.Next()

	routeLabel := "<unmatched>"
	if r := ctx.Route(); r != nil && r.Path != "" {
		routeLabel = r.Path
	}

	labels := []string{
		routeLabel,
		string(ctx.Context().Method()),
		strconv.Itoa(ctx.Response().StatusCode()),
	}

	m.totalRequests.WithLabelValues(labels...).Inc()
	m.durationSec.WithLabelValues(labels...).Observe(time.Since(s).Seconds())

	return
}

func newMetrics(namespace, subsystem string) *metrics {
	labels := []string{"route", "method", "code"}

	t := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "requests_total",
		Help:      "Total number of requests",
	}, labels)
	d := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "requests_duration",
		Help:      "Duration of requests",
		Buckets: []float64{
			0.01, 0.025, 0.05, 0.1,
			0.25, 0.5, 0.75, 1,
			1.5, 2, 3, 5, 8, 12,
		},
	}, labels)
	inF := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "requests_in_flight",
		Help:      "Current number of in-flight HTTP requests",
	})

	prometheus.MustRegister(t, d, inF)

	return &metrics{
		totalRequests: t,
		durationSec:   d,
		inFlight:      inF,
	}
}
