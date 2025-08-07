package httpServer

import (
	"github.com/gofiber/fiber/v2"

	"mytonstorage-gateway/pkg/models"
)

func okHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

func errorHandler(c *fiber.Ctx, err error) error {
	if e, ok := err.(*fiber.Error); ok {
		return c.Status(e.Code).JSON(fiber.Map{
			"error": e.Message,
		})
	}

	if appErr, ok := err.(*models.AppError); ok {
		return c.Status(appErr.Code).JSON(fiber.Map{
			"error": appErr.Message,
		})
	}

	errorResponse := errorResponse{
		Error: err.Error(),
	}

	return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
}

func validateBagID(bagid string) bool {
	if len(bagid) != 64 {
		return false
	}

	for i := range 64 {
		c := bagid[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}

	return true
}

func setFileContentType(c *fiber.Ctx, ext, filename string) {
	switch ext {
	case ".jpg", ".jpeg":
		c.Set("Content-Type", "image/jpeg")
	case ".png":
		c.Set("Content-Type", "image/png")
	case ".gif":
		c.Set("Content-Type", "image/gif")
	case ".webp":
		c.Set("Content-Type", "image/webp")
	case ".svg":
		c.Set("Content-Type", "image/svg+xml")
	case ".pdf":
		c.Set("Content-Type", "application/pdf")
	case ".txt":
		c.Set("Content-Type", "text/plain; charset=utf-8")
	case ".html", ".htm":
		c.Set("Content-Type", "text/html; charset=utf-8")
	case ".css":
		c.Set("Content-Type", "text/css; charset=utf-8")
	case ".js":
		c.Set("Content-Type", "application/javascript; charset=utf-8")
	case ".json":
		c.Set("Content-Type", "application/json; charset=utf-8")
	case ".xml":
		c.Set("Content-Type", "application/xml; charset=utf-8")
	case ".mp4":
		c.Set("Content-Type", "video/mp4")
	case ".mp3":
		c.Set("Content-Type", "audio/mpeg")
	case ".wav":
		c.Set("Content-Type", "audio/wav")
	default:
		c.Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	}
}
