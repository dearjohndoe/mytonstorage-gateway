package files

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mytonstorage-gateway/pkg/models/db"
)

type repository struct {
	db *pgxpool.Pool
}

type Repository interface {
	HasBan(ctx context.Context, bagID string) (bool, error)
	GetBan(ctx context.Context, bagID string) (*db.BanStatus, error)
	GetAllBans(ctx context.Context, limit int, offset int) ([]db.BanStatus, error)
	GetReports(ctx context.Context, limit int, offset int) ([]db.Report, error)
	GetReportsByBagID(ctx context.Context, bagID string) ([]db.Report, error)
	AddReport(ctx context.Context, report db.Report) error
	UpdateBanStatus(ctx context.Context, statuses []db.BanStatus) error
}

func (r *repository) HasBan(ctx context.Context, bagID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM files.blacklist WHERE bagid = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, bagID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *repository) GetBan(ctx context.Context, bagID string) (*db.BanStatus, error) {
	query := `
		SELECT bagid, admin, reason, comment, true as status, created_at
		FROM files.blacklist
		WHERE bagid = $1
		LIMIT 1`

	row := r.db.QueryRow(ctx, query, bagID)
	var b db.BanStatus
	if err := row.Scan(&b.BagID, &b.Admin, &b.Reason, &b.Comment, &b.Status, &b.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &b, nil
}

func (r *repository) GetAllBans(ctx context.Context, limit int, offset int) (bans []db.BanStatus, err error) {
	query := `
		SELECT bagid, admin, reason, comment, true as status, created_at 
		FROM files.blacklist 
		ORDER BY created_at DESC 
		LIMIT $1 
		OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b db.BanStatus
		if err := rows.Scan(&b.BagID, &b.Admin, &b.Reason, &b.Comment, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}

		bans = append(bans, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return
}

func (r *repository) GetReports(ctx context.Context, limit int, offset int) (reports []db.Report, err error) {
	query := `
		SELECT bagid, reason, sender, comment, created_at 
		FROM files.reports 
		ORDER BY created_at DESC 
		LIMIT $1 
		OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r db.Report
		if err := rows.Scan(&r.BagID, &r.Reason, &r.Sender, &r.Comment, &r.CreatedAt); err != nil {
			return nil, err
		}

		reports = append(reports, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return
}

func (r *repository) GetReportsByBagID(ctx context.Context, bagID string) (reports []db.Report, err error) {
	query := `		
		SELECT bagid, reason, sender, comment, created_at 
		FROM files.reports 
		WHERE bagid = $1`

	rows, err := r.db.Query(ctx, query, bagID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r db.Report
		if err := rows.Scan(&r.BagID, &r.Reason, &r.Sender, &r.Comment, &r.CreatedAt); err != nil {
			return nil, err
		}

		reports = append(reports, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return
}

func (r *repository) AddReport(ctx context.Context, report db.Report) (err error) {
	query := `
		INSERT INTO files.reports (bagid, reason, sender, comment)
		VALUES ($1, $2, $3, $4)`
	_, err = r.db.Exec(ctx, query, report.BagID, report.Reason, report.Sender, report.Comment)
	return
}

func (r *repository) UpdateBanStatus(ctx context.Context, statuses []db.BanStatus) (err error) {
	query := `
		WITH cte AS (
		    SELECT 
				c->> 'bag_id' AS bagid,
				c->> 'admin' AS admin,
				(c->>'status')::boolean AS is_banned,
				(c->>'reason') AS reason,
				c->> 'comment' AS comment
			FROM jsonb_array_elements($1::jsonb) AS c
		),
		update AS (
			INSERT INTO files.blacklist (bagid, admin, reason, comment)
			SELECT bagid, admin, reason, comment
			FROM cte c
			WHERE c.is_banned
			ON CONFLICT (bagid) DO UPDATE
			SET
				admin = EXCLUDED.admin,
				reason = EXCLUDED.reason,
				comment = EXCLUDED.comment
		)
		DELETE FROM files.blacklist
		WHERE bagid IN (SELECT bagid FROM cte WHERE NOT is_banned)
	`

	_, err = r.db.Exec(ctx, query, statuses)

	return
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}
