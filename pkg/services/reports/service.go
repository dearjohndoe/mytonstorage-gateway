package reports

import (
	"context"
	"log/slog"
	"strings"

	"mytonstorage-gateway/pkg/models"
	v1 "mytonstorage-gateway/pkg/models/api/v1"
	"mytonstorage-gateway/pkg/models/db"
)

type filesDb interface {
	GetReports(ctx context.Context, limit int, offset int) ([]db.Report, error)
	GetReportsByBagID(ctx context.Context, bagID string) ([]db.Report, error)
	AddReport(ctx context.Context, report db.Report) error
	UpdateBanStatus(ctx context.Context, statuses []db.BanStatus) error
	GetBan(ctx context.Context, bagID string) (*db.BanStatus, error)
	GetAllBans(ctx context.Context, limit int, offset int) ([]db.BanStatus, error)
}

type service struct {
	files  filesDb
	logger *slog.Logger
}

type Reports interface {
	GetReports(ctx context.Context, limit int, offset int) ([]v1.Report, error)
	GetReportsByBagID(ctx context.Context, bagID string) ([]v1.Report, error)
	GetBan(ctx context.Context, bagID string) (*v1.BanInfo, error)
	GetAllBans(ctx context.Context, limit int, offset int) ([]v1.BanInfo, error)
	AddReport(ctx context.Context, report v1.Report) error
	UpdateBanStatus(ctx context.Context, statuses []v1.BanStatus) error
}

func (s *service) GetReports(ctx context.Context, limit int, offset int) (reports []v1.Report, err error) {
	log := s.logger.With(slog.String("method", "GetReports"))

	dbReports, err := s.files.GetReports(ctx, limit, offset)
	if err != nil {
		log.Error("failed to get reports", slog.String("error", err.Error()))
		return nil, models.NewAppError(models.InternalServerErrorCode, "")
	}

	for _, r := range dbReports {
		var createdAt uint64
		if r.CreatedAt != nil {
			createdAt = uint64(r.CreatedAt.Unix())
		}

		reports = append(reports, v1.Report{
			BagID:     strings.ToUpper(r.BagID),
			Reason:    r.Reason,
			Sender:    r.Sender,
			Comment:   r.Comment,
			CreatedAt: createdAt,
		})
	}

	return reports, nil
}

func (s *service) GetReportsByBagID(ctx context.Context, bagID string) ([]v1.Report, error) {
	log := s.logger.With(
		slog.String("method", "GetReport"),
		slog.String("bagID", bagID),
	)

	dbReports, err := s.files.GetReportsByBagID(ctx, bagID)
	if err != nil {
		log.Error("failed to get report", slog.String("error", err.Error()))
		return nil, models.NewAppError(models.InternalServerErrorCode, "")
	}

	if dbReports == nil {
		return nil, nil
	}

	var resp []v1.Report
	for _, dbReport := range dbReports {
		var createdAt uint64
		if dbReport.CreatedAt != nil {
			createdAt = uint64(dbReport.CreatedAt.Unix())
		}

		resp = append(resp, v1.Report{
			BagID:     strings.ToUpper(dbReport.BagID),
			Reason:    dbReport.Reason,
			Sender:    dbReport.Sender,
			Comment:   dbReport.Comment,
			CreatedAt: createdAt,
		})
	}

	return resp, nil
}

func (s *service) GetBan(ctx context.Context, bagID string) (*v1.BanInfo, error) {
	log := s.logger.With(
		slog.String("method", "GetBan"),
		slog.String("bagID", bagID),
	)

	status, err := s.files.GetBan(ctx, bagID)
	if err != nil {
		log.Error("failed to get ban info", slog.String("error", err.Error()))
		return nil, models.NewAppError(models.InternalServerErrorCode, "")
	}

	if status == nil {
		return nil, nil
	}

	return &v1.BanInfo{
		BagID:   strings.ToUpper(status.BagID),
		Admin:   status.Admin,
		Reason:  status.Reason,
		Comment: status.Comment,
	}, nil
}

func (s *service) GetAllBans(ctx context.Context, limit int, offset int) (bans []v1.BanInfo, err error) {
	log := s.logger.With(slog.String("method", "GetAllBans"))

	dbBans, err := s.files.GetAllBans(ctx, limit, offset)
	if err != nil {
		log.Error("failed to get bans", slog.String("error", err.Error()))
		return nil, models.NewAppError(models.InternalServerErrorCode, "")
	}

	for _, b := range dbBans {
		bans = append(bans, v1.BanInfo{
			BagID:   strings.ToUpper(b.BagID),
			Admin:   b.Admin,
			Reason:  b.Reason,
			Comment: b.Comment,
		})
	}

	return bans, nil
}

func (s *service) AddReport(ctx context.Context, report v1.Report) error {
	log := s.logger.With(
		slog.String("method", "AddReport"),
		slog.String("bagID", report.BagID),
	)

	if len(report.BagID) != 64 {
		return models.NewAppError(models.BadRequestErrorCode, "invalid bag ID")
	}

	dbReport := db.Report{
		BagID:   report.BagID,
		Reason:  report.Reason,
		Sender:  report.Sender,
		Comment: report.Comment,
	}

	if err := s.files.AddReport(ctx, dbReport); err != nil {
		log.Error("failed to add report", slog.String("error", err.Error()))
		return models.NewAppError(models.InternalServerErrorCode, "")
	}

	return nil
}

func (s *service) UpdateBanStatus(ctx context.Context, statuses []v1.BanStatus) error {
	log := s.logger.With(slog.String("method", "UpdateBanStatus"))

	if len(statuses) == 0 {
		return nil
	}

	dbStatuses := make([]db.BanStatus, len(statuses))
	for i, s := range statuses {
		if len(s.BagID) != 64 {
			return models.NewAppError(models.BadRequestErrorCode, "invalid bag ID")
		}

		dbStatuses[i] = db.BanStatus{
			BagID:   strings.ToLower(s.BagID),
			Admin:   s.Admin,
			Reason:  s.Reason,
			Comment: s.Comment,
			Status:  s.Status,
		}
	}

	if err := s.files.UpdateBanStatus(ctx, dbStatuses); err != nil {
		log.Error("failed to update ban status", slog.String("error", err.Error()))
		return models.NewAppError(models.InternalServerErrorCode, "")
	}

	return nil
}

func NewService(files filesDb, logger *slog.Logger) Reports {
	return &service{
		files:  files,
		logger: logger,
	}
}
