// !ONLY FOR DEBUG PURPOSES
//
//go:build debug
// +build debug

package httpServer

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

const (
	MaxRequests     = 100
	RateLimitWindow = 60 * time.Second
)

func (h *handler) RegisterRoutes() {
	h.logger.Info("Registering debug routes")

	// On server side nginx or other reverse proxy should handle CORS
	// and OPTIONS requests, but for debug purposes we handle it here.
	h.server.Use(func(c *fiber.Ctx) error {
		// Always set CORS headers
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Set("Access-Control-Allow-Headers", "*")

		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusOK)
		}
		return c.Next()
	})

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
		gateway := apiv1.Group("/gateway", h.securityHeadersMiddleware)

		gateway.Get("/:bagid", h.getBag)
		gateway.Get("/:bagid/*", h.getPath)

		gateway.Get("/health", h.health)
		gateway.Get("/metrics", h.requireMetrics(), h.metrics)
	}

	{
		reports := apiv1.Group("/reports")

		reports.Get("", h.requireReports(), h.getReports)
		reports.Post("", h.requireReports(), h.addReport)
		reports.Get("/:bagid", h.requireReports(), h.getReportsByBagID)
	}

	{
		bans := apiv1.Group("/bans")

		bans.Get("", h.requireBans(), h.getAllBans)
		bans.Put("", h.requireBans(), h.updateBanStatus)
		bans.Get("/:bagid", h.requireBans(), h.getBan)
	}
}
