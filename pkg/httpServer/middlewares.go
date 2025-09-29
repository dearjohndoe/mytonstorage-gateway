package httpServer

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// requirePermission создает middleware для проверки конкретного разрешения
func (h *handler) requirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accessToken := c.Get("Authorization")
		if accessToken == "" {
			return errorHandler(c, fiber.NewError(fiber.StatusUnauthorized, "unauthorized"))
		}

		if strings.HasPrefix(strings.ToLower(accessToken), "bearer ") {
			accessToken = accessToken[7:]
		}

		hash := md5.Sum([]byte(accessToken))
		tokenHash := fmt.Sprintf("%x", hash[:])

		tokenPermissions, exists := h.accessTokens[tokenHash]
		if !exists {
			return errorHandler(c, fiber.NewError(fiber.StatusForbidden, "forbidden"))
		}

		hasPermission := false
		switch permission {
		case "bans":
			hasPermission = tokenPermissions.Bans
		case "reports":
			hasPermission = tokenPermissions.Reports
		case "metrics":
			hasPermission = tokenPermissions.Metrics
		}

		if !hasPermission {
			return errorHandler(c, fiber.NewError(fiber.StatusForbidden, "forbidden"))
		}

		return c.Next()
	}
}

func (h *handler) requireBans() fiber.Handler {
	return h.requirePermission("bans")
}

func (h *handler) requireReports() fiber.Handler {
	return h.requirePermission("reports")
}

func (h *handler) requireMetrics() fiber.Handler {
	return h.requirePermission("metrics")
}

func (h *handler) loggerMiddleware(c *fiber.Ctx) error {
	headers := c.GetReqHeaders()
	delete(headers, "Authorization")
	delete(headers, "Cookie")

	if _, ok := headers["Cookie"]; ok {
		headers["Cookie"] = []string{"REDACTED"}
	}

	res := c.Next()

	h.logger.Debug(
		"request",
		"status_code", c.Response().StatusCode(),
		"method", c.Method(),
		"url", c.OriginalURL(),
		"headers", headers,
		"body_length", len(c.Body()),
	)

	return res
}

// securityHeadersMiddleware adds security headers for gateway routes
func (h *handler) securityHeadersMiddleware(c *fiber.Ctx) error {
	// Content-Security-Policy with sandbox directives to restrict capabilities
	// allow-scripts - allows script execution
	// allow-forms - allows form submission
	// NOT including allow-same-origin - this prevents access to localStorage, cookies, etc.
	c.Set("Content-Security-Policy", "sandbox allow-scripts allow-forms allow-downloads")

	// X-Frame-Options to prevent clickjacking
	c.Set("X-Frame-Options", "DENY")

	// X-Content-Type-Options to prevent MIME-sniffing
	c.Set("X-Content-Type-Options", "nosniff")

	// Referrer-Policy to control referrer sending
	c.Set("Referrer-Policy", "no-referrer")

	return c.Next()
}
