package files

import (
	"context"
	"time"

	"mytonstorage-gateway/pkg/cache"
	"mytonstorage-gateway/pkg/models/private"
)

type cacheMiddleware struct {
	svc   Files
	cache *cache.SimpleCache
}

func (c *cacheMiddleware) GetPathInfo(ctx context.Context, bagID, path string) (info private.FolderInfo, err error) {
	info, err = c.svc.GetPathInfo(ctx, bagID, path)
	if err != nil {
		return
	}

	return
}

func NewCacheMiddleware(
	svc Files,
) Files {
	return &cacheMiddleware{
		svc:   svc,
		cache: cache.NewSimpleCache(1 * time.Minute),
	}
}
