package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/25types/25types/internal/app/application/reports"
	"github.com/25types/25types/internal/app/domain"
)

// ReportsRepo implements reports.Repository backed by SQLite.
type ReportsRepo struct {
	db *sql.DB
}

// NewReportsRepo creates a new ReportsRepo.
func NewReportsRepo(db *sql.DB) *ReportsRepo {
	return &ReportsRepo{db: db}
}

var _ reports.Repository = (*ReportsRepo)(nil)

func (r *ReportsRepo) Create(ctx context.Context, report *domain.Report) (int64, error) {
	engineStr := string(report.EngineData)
	if engineStr == "" {
		engineStr = "{}"
	}
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO reports (user_id, scene, sub_scene, question, engine_data, content, locale, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		report.UserID, string(report.Scene), report.SubScene, report.Question,
		engineStr, report.Content, report.Locale, now,
	)
	if err != nil {
		return 0, fmt.Errorf("reports_repo: create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("reports_repo: last_insert_id: %w", err)
	}
	report.ID = id
	report.CreatedAt, _ = time.Parse(time.RFC3339, now)
	return id, nil
}

func (r *ReportsRepo) FindByID(ctx context.Context, id int64, userID int64) (*domain.Report, error) {
	var report domain.Report
	var engineStr, createdAt string
	var scene, subScene, question, locale string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, scene, sub_scene, question, engine_data, content, locale, created_at
		 FROM reports WHERE id = ? AND user_id = ? AND deleted_at IS NULL`, id, userID,
	).Scan(&report.ID, &report.UserID, &scene, &subScene, &question, &engineStr, &report.Content, &locale, &createdAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrReportNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("reports_repo: find: %w", err)
	}
	report.Scene = domain.Scene(scene)
	report.SubScene = subScene
	report.Question = question
	report.EngineData = json.RawMessage(engineStr)
	report.Locale = locale
	report.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &report, nil
}

func (r *ReportsRepo) ListByUser(ctx context.Context, userID int64, scene string, limit, offset int) ([]domain.ReportItem, int, error) {
	var rows *sql.Rows
	var err error
	if scene != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, scene, sub_scene, question, created_at
			 FROM reports WHERE user_id = ? AND scene = ? AND deleted_at IS NULL
			 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
			userID, scene, limit, offset,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, scene, sub_scene, question, created_at
			 FROM reports WHERE user_id = ? AND deleted_at IS NULL
			 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
			userID, limit, offset,
		)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("reports_repo: list: %w", err)
	}
	defer rows.Close()

	var items []domain.ReportItem
	for rows.Next() {
		var item domain.ReportItem
		var s, sub, q, ca string
		if err := rows.Scan(&item.ID, &s, &sub, &q, &ca); err != nil {
			return nil, 0, fmt.Errorf("reports_repo: scan: %w", err)
		}
		item.Scene = domain.Scene(s)
		item.SubScene = sub
		item.Question = q
		item.CreatedAt, _ = time.Parse(time.RFC3339, ca)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("reports_repo: rows: %w", err)
	}

	// Count query.
	var total int
	countQuery := `SELECT COUNT(*) FROM reports WHERE user_id = ? AND deleted_at IS NULL`
	if scene != "" {
		countQuery = `SELECT COUNT(*) FROM reports WHERE user_id = ? AND scene = ? AND deleted_at IS NULL`
		if err := r.db.QueryRowContext(ctx, countQuery, userID, scene).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("reports_repo: count: %w", err)
		}
	} else {
		if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("reports_repo: count: %w", err)
		}
	}

	return items, total, nil
}

func (r *ReportsRepo) SoftDelete(ctx context.Context, id int64, userID int64) (bool, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(ctx,
		`UPDATE reports SET deleted_at = ? WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		now, id, userID,
	)
	if err != nil {
		return false, fmt.Errorf("reports_repo: soft_delete: %w", err)
	}
	n, err := result.RowsAffected()
	return n > 0, err
}

func (r *ReportsRepo) CreateShare(ctx context.Context, reportID int64, token string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO report_shares (report_id, token, created_at) VALUES (?, ?, ?)`,
		reportID, token, now,
	)
	if err != nil {
		return fmt.Errorf("reports_repo: create_share: %w", err)
	}
	return nil
}

func (r *ReportsRepo) FindShareByToken(ctx context.Context, token string) (*domain.ReportShare, *domain.Report, error) {
	var share domain.ReportShare
	var report domain.Report
	var expiresAt sql.NullString
	var engineStr, createdAt string
	var scene, subScene, question, locale string

	err := r.db.QueryRowContext(ctx,
		`SELECT rs.token, rs.report_id, rs.created_at, rs.expires_at,
		        r.id, r.user_id, r.scene, r.sub_scene, r.question, r.engine_data, r.content, r.locale, r.created_at
		 FROM report_shares rs JOIN reports r ON rs.report_id = r.id
		 WHERE rs.token = ? AND r.deleted_at IS NULL`, token,
	).Scan(&share.Token, &share.ReportID, &expiresAt,
		&report.ID, &report.UserID, &scene, &subScene, &question, &engineStr, &report.Content, &locale, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil, domain.ErrReportNotFound
	}
	if err != nil {
		return nil, nil, fmt.Errorf("reports_repo: find_share: %w", err)
	}

	// Check expiry.
	if expiresAt.Valid && expiresAt.String != "" {
		exp, err := time.Parse(time.RFC3339, expiresAt.String)
		if err == nil && time.Now().After(exp) {
			return nil, nil, domain.ErrReportNotFound
		}
	}

	report.Scene = domain.Scene(scene)
	report.SubScene = subScene
	report.Question = question
	report.EngineData = json.RawMessage(engineStr)
	report.Locale = locale
	report.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	if expiresAt.Valid {
		t, _ := time.Parse(time.RFC3339, expiresAt.String)
		share.ExpiresAt = &t
	}

	return &share, &report, nil
}

func (r *ReportsRepo) RevokeShare(ctx context.Context, reportID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM report_shares WHERE report_id = ?`, reportID,
	)
	if err != nil {
		return fmt.Errorf("reports_repo: revoke_share: %w", err)
	}
	return nil
}
