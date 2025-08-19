package files

import (
	"context"
	"fmt"
	"time"

	"mytonstorage-gateway/pkg/cache"
	"mytonstorage-gateway/pkg/models/private"
)

type cacheMiddleware struct {
	svc   Files
	cache *cache.SimpleCache
}

func (c *cacheMiddleware) GetPathInfo(ctx context.Context, bagID, path string) (info private.FolderInfo, err error) {
	cacheKey := fmt.Sprintf("%s:%s", bagID, path)

	if cachedInfo, ok := c.cache.Get(cacheKey); ok {
		return cachedInfo.(private.FolderInfo), nil
	}

	info, err = c.svc.GetPathInfo(ctx, bagID, path)
	if err != nil {
		return
	}

	if info.StreamFile != nil {
		return
	}

	c.cache.Set(cacheKey, info)

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
