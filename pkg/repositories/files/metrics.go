package files

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"mytonstorage-gateway/pkg/models/db"
)

type metricsMiddleware struct {
	reqCount    *prometheus.CounterVec
	reqDuration *prometheus.HistogramVec
	repo        Repository
}

func (m *metricsMiddleware) HasBan(ctx context.Context, bagID string) (ok bool, err error) {
	defer func(s time.Time) {
		labels := []string{
			"HasBan", strconv.FormatBool(err != nil),
		}
		m.reqCount.WithLabelValues(labels...).Add(1)
		m.reqDuration.WithLabelValues(labels...).Observe(time.Since(s).Seconds())
	}(time.Now())
	return m.repo.HasBan(ctx, bagID)
}

func (m *metricsMiddleware) GetReports(ctx context.Context, limit int, offset int) (reports []db.Report, err error) {
	defer func(s time.Time) {
		labels := []string{
			"GetReports", strconv.FormatBool(err != nil),
		}
		m.reqCount.WithLabelValues(labels...).Add(1)
		m.reqDuration.WithLabelValues(labels...).Observe(time.Since(s).Seconds())
	}(time.Now())
	return m.repo.GetReports(ctx, limit, offset)
}

func (m *metricsMiddleware) GetReportsByBagID(ctx context.Context, bagID string) (reports []db.Report, err error) {
	defer func(s time.Time) {
		labels := []string{
			"GetReportsByBagID", strconv.FormatBool(err != nil),
		}
		m.reqCount.WithLabelValues(labels...).Add(1)
		m.reqDuration.WithLabelValues(labels...).Observe(time.Since(s).Seconds())
	}(time.Now())
	return m.repo.GetReportsByBagID(ctx, bagID)
}

func (m *metricsMiddleware) GetBan(ctx context.Context, bagID string) (status *db.BanStatus, err error) {
	defer func(s time.Time) {
		labels := []string{
			"GetBan", strconv.FormatBool(err != nil),
		}
		m.reqCount.WithLabelValues(labels...).Add(1)
		m.reqDuration.WithLabelValues(labels...).Observe(time.Since(s).Seconds())
	}(time.Now())
	return m.repo.GetBan(ctx, bagID)
}

func (m *metricsMiddleware) GetAllBans(ctx context.Context, limit int, offset int) (bans []db.BanStatus, err error) {
	defer func(s time.Time) {
		labels := []string{
			"GetAllBans", strconv.FormatBool(err != nil),
		}
		m.reqCount.WithLabelValues(labels...).Add(1)
		m.reqDuration.WithLabelValues(labels...).Observe(time.Since(s).Seconds())
	}(time.Now())
	return m.repo.GetAllBans(ctx, limit, offset)
}

func (m *metricsMiddleware) AddReport(ctx context.Context, report db.Report) (err error) {
	defer func(s time.Time) {
		labels := []string{
			"AddReport", strconv.FormatBool(err != nil),
		}
		m.reqCount.WithLabelValues(labels...).Add(1)
		m.reqDuration.WithLabelValues(labels...).Observe(time.Since(s).Seconds())
	}(time.Now())
	return m.repo.AddReport(ctx, report)
}

func (m *metricsMiddleware) UpdateBanStatus(ctx context.Context, statuses []db.BanStatus) (err error) {
	defer func(s time.Time) {
		labels := []string{
			"UpdateBanStatus", strconv.FormatBool(err != nil),
		}
		m.reqCount.WithLabelValues(labels...).Add(1)
		m.reqDuration.WithLabelValues(labels...).Observe(time.Since(s).Seconds())
	}(time.Now())
	return m.repo.UpdateBanStatus(ctx, statuses)
}

func NewMetrics(reqCount *prometheus.CounterVec, reqDuration *prometheus.HistogramVec, repo Repository) Repository {
	return &metricsMiddleware{
		reqCount:    reqCount,
		reqDuration: reqDuration,
		repo:        repo,
	}
}
