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
		gateway := apiv1.Group("/gateway")
		gateway.Get("/:bagid", h.getBag)
		gateway.Get("/:bagid/*", h.getPath)

		gateway.Get("/health", h.health)
		gateway.Get("/metrics", h.authorizationMiddleware, h.metrics)
	}

	{
		reports := apiv1.Group("/reports")

		// admins only
		reports.Get("", h.authorizationMiddleware, h.getAllReports)
		reports.Get("/:bagid", h.authorizationMiddleware, h.getReportsByBagID)
		reports.Get("/:bagid/ban", h.authorizationMiddleware, h.getBan)
		reports.Put("", h.authorizationMiddleware, h.updateBanStatus)

		// anyone
		reports.Put("/:bagid", h.addReport)
	}
}
