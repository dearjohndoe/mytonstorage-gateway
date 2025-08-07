package httpServer

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"mytonstorage-gateway/pkg/models/private"
)

type files interface {
	GetPathInfo(ctx context.Context, bagID, path string) (private.FolderInfo, error)
}

type templatesSvc interface {
	ContentType(ext, filename string) (string, string)
	ErrorTemplate(err error) (string, error)
	HtmlFilesListWithTemplate(f private.FolderInfo, path string) (string, error)
}

type errorResponse struct {
	Error string `json:"error"`
}

type handler struct {
	server       *fiber.App
	logger       *slog.Logger
	files        files
	templates    templatesSvc
	namespace    string
	subsystem    string
	accessTokens map[string]struct{}
}

func New(
	server *fiber.App,
	files files,
	templates templatesSvc,
	accessTokens []string,
	namespace string,
	subsystem string,
	logger *slog.Logger,
) *handler {
	accessTokensMap := make(map[string]struct{})
	for _, token := range accessTokens {
		accessTokensMap[token] = struct{}{}
	}

	h := &handler{
		server:       server,
		files:        files,
		templates:    templates,
		namespace:    namespace,
		subsystem:    subsystem,
		accessTokens: accessTokensMap,
		logger:       logger,
	}

	return h
}
