package files

import (
	"context"
	"strings"
	"time"

	"mytonstorage-gateway/pkg/cache"
	"mytonstorage-gateway/pkg/models/db"
)

type cacheMiddleware struct {
	repo  Repository
	cache *cache.SimpleCache
}

var (
	hasBanKey = "hb:"
)

func (c *cacheMiddleware) HasBan(ctx context.Context, bagID string) (bool, error) {
	key := hasBanKey + strings.ToLower(bagID)
	if banned, ok := c.cache.Get(key); ok {
		if banned.(bool) {
			return true, nil
		}
	}

	result, err := c.repo.HasBan(ctx, bagID)
	if err == nil {
		if result {
			c.cache.Set(key, result)
		}
	}

	return result, err
}

func (c *cacheMiddleware) GetBan(ctx context.Context, bagID string) (status *db.BanStatus, err error) {
	return c.repo.GetBan(ctx, bagID)
}

func (c *cacheMiddleware) GetReports(ctx context.Context, limit int, offset int) (reports []db.Report, err error) {
	return c.repo.GetReports(ctx, limit, offset)
}

func (c *cacheMiddleware) GetReportsByBagID(ctx context.Context, bagID string) (reports []db.Report, err error) {
	return c.repo.GetReportsByBagID(ctx, bagID)
}

func (c *cacheMiddleware) AddReport(ctx context.Context, report db.Report) (err error) {
	return c.repo.AddReport(ctx, report)
}

func (c *cacheMiddleware) UpdateBanStatus(ctx context.Context, statuses []db.BanStatus) (err error) {
	err = c.repo.UpdateBanStatus(ctx, statuses)
	if err != nil {
		return
	}

	for _, status := range statuses {
		key := hasBanKey + strings.ToLower(status.BagID)
		if !status.Status {
			c.cache.Release(key)
		} else {
			c.cache.Set(key, true)
		}
	}

	return
}

func NewCache(repo Repository) Repository {
	return &cacheMiddleware{
		repo:  repo,
		cache: cache.NewSimpleCache(1 * time.Hour),
	}
}
