//go:build !debug
// +build !debug

package httpServer

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/limiter"
)

const (
	MaxRequests     = 100
	RateLimitWindow = 60 * time.Second
)

func (h *handler) RegisterRoutes() {
	h.logger.Info("Registering routes")

	m := newMetrics(h.namespace, h.subsystem)

	h.server.Use(m.metricsMiddleware)

	h.server.Use(limiter.New(limiter.Config{
		Max:               MaxRequests,
		Expiration:        RateLimitWindow,
		LimitReached:      h.limitReached,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))

	apiv1 := h.server.Group("/api/v1", h.loggerMiddleware)
	{
		gateway := apiv1.Group("/gateway")

		gateway.Get("/:bagid", h.getBag)
		gateway.Get("/:bagid/*", h.getPath)

		gateway.Get("/health", h.health)
		gateway.Get("/metrics", h.requireMetrics(), h.metrics)
	}

	{
		reports := apiv1.Group("/reports")

		reports.Get("", h.requireBans(), h.getAllReports)
		reports.Get("/:bagid", h.requireBans(), h.getReportsByBagID)
		reports.Get("/:bagid/ban", h.requireBans(), h.getBan)
		reports.Put("", h.requireBans(), h.updateBanStatus)

		reports.Put("/:bagid", h.requireReports(), h.addReport)
	}
}
