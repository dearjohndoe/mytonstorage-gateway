package httpServer

import (
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (h *handler) limitReached(c *fiber.Ctx) error {
	log := h.logger.With(
		slog.String("method", "limitReached"),
		slog.String("method", c.Method()),
		slog.String("url", c.OriginalURL()),
		slog.Any("headers", c.GetReqHeaders()),
	)

	log.Warn("rate limit reached for request")
	return fiber.NewError(fiber.StatusTooManyRequests, "too many requests, please try again later")
}

func (h *handler) getBag(c *fiber.Ctx) (err error) {
	bagid := c.Params("bagid")

	log := h.logger.With(
		slog.String("func", "getBag"),
		slog.String("bagid", bagid),
		slog.String("method", c.Method()),
		slog.String("url", c.OriginalURL()),
		slog.Any("headers", c.GetReqHeaders()),
	)

	return h.getBagInfoResponse(c, bagid, "", log)
}

func (h *handler) getPath(c *fiber.Ctx) (err error) {
	bagid := c.Params("bagid")
	rawPath := c.Params("*")

	log := h.logger.With(
		slog.String("func", "getPath"),
		slog.String("bagid", bagid),
		slog.String("raw_path", rawPath),
		slog.String("method", c.Method()),
		slog.String("url", c.OriginalURL()),
		slog.Any("headers", c.GetReqHeaders()),
	)

	decodedPath, err := url.QueryUnescape(rawPath)
	if err != nil {
		log.Error("failed to decode path", slog.String("error", err.Error()))
		err = fiber.NewError(fiber.StatusBadRequest, "invalid path encoding")
		return errorHandler(c, err)
	}

	return h.getBagInfoResponse(c, bagid, decodedPath, log)
}

func (h *handler) getBagInfoResponse(c *fiber.Ctx, bagid, path string, log *slog.Logger) (err error) {
	if !validateBagID(strings.ToLower(bagid)) {
		log.Error("invalid bagid format")
		err = fiber.NewError(fiber.StatusBadRequest, "invalid bagid")
		return errorHandler(c, err)
	}

	bagInfo, err := h.files.GetPathInfo(c.Context(), bagid, path)
	if err != nil {
		log.Error("failed to get bag path", slog.String("error", err.Error()))
		return errorHandler(c, err)
	}

	if !bagInfo.IsValid {
		log.Error("bag not found", slog.String("bagid", bagid))
		err = fiber.NewError(fiber.StatusNotFound, "bag not found")
		return errorHandler(c, err)
	}

	// Serve single file
	if len(bagInfo.Files) == 1 && !bagInfo.Files[0].IsFolder && strings.HasSuffix(path, bagInfo.Files[0].Name) {
		filename := bagInfo.Files[0].Name
		filePath := filepath.Join(bagInfo.DiskPath, path)

		ext := strings.ToLower(filepath.Ext(filename))
		header, value := h.templates.ContentType(ext, filename)
		c.Set(header, value)

		return c.SendFile(filePath)
	}

	// Serve directory
	html, err := h.templates.HtmlFilesListWithTemplate(bagInfo, path)
	if err != nil {
		var tErr error
		html, tErr = h.templates.ErrorTemplate(err)
		if tErr != nil {
			log.Error("failed to render error template", slog.String("error", tErr.Error()))
			return errorHandler(c, fiber.NewError(fiber.StatusInternalServerError, ""))
		}
	}

	return c.Type("html").SendString(html)
}

func (h *handler) health(c *fiber.Ctx) error {
	return c.JSON(okHandler(c))
}

func (h *handler) metrics(c *fiber.Ctx) error {
	m := promhttp.Handler()

	return adaptor.HTTPHandler(m)(c)
}
