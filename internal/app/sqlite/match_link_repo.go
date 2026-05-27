package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/25types/25types/internal/app/application/matchlink"
	"github.com/25types/25types/internal/app/domain"
)

// MatchLinkRepo implements matchlink.MatchLinkRepository.
type MatchLinkRepo struct {
	db *sql.DB
}

// NewMatchLinkRepo creates a MatchLinkRepo.
func NewMatchLinkRepo(db *sql.DB) *MatchLinkRepo {
	return &MatchLinkRepo{db: db}
}

func (r *MatchLinkRepo) Create(ctx context.Context, userID int64, token string, linkType string) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO match_links (user_id, token, type, created_at) VALUES (?, ?, ?, ?)`,
		userID, token, linkType, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return 0, fmt.Errorf("Create match link: %w", err)
	}
	return res.LastInsertId()
}

func (r *MatchLinkRepo) FindByToken(ctx context.Context, token string) (*domain.MatchLink, error) {
	var ml domain.MatchLink
	var createdAt sql.NullString
	var deletedAt sql.NullString
	var linkType sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, token, type, created_at, deleted_at FROM match_links WHERE token = ?`, token,
	).Scan(&ml.ID, &ml.UserID, &ml.Token, &linkType, &createdAt, &deletedAt)
	if err != nil {
		return nil, domain.ErrMatchLinkNotFound
	}
	if deletedAt.Valid {
		return nil, domain.ErrMatchLinkNotFound
	}
	ml.Type = linkType.String
	if t := parseNullTime(createdAt); t != nil {
		ml.CreatedAt = *t
	}
	return &ml, nil
}

func (r *MatchLinkRepo) ListByUser(ctx context.Context, userID int64, linkType string) ([]matchlink.MatchLinkItem, error) {
	query := `SELECT ml.id, ml.token, ml.type, ml.created_at,
		COALESCE(be.cnt, 0) AS bond_count,
		COALESCE(bme.cnt, 0) AS match_count
	 FROM match_links ml
	 LEFT JOIN (SELECT link_id, COUNT(*) AS cnt FROM bond_events GROUP BY link_id) be ON ml.id = be.link_id
	 LEFT JOIN (SELECT link_id, COUNT(*) AS cnt FROM mingli_match_events GROUP BY link_id) bme ON ml.id = bme.link_id
	 WHERE ml.user_id = ? AND ml.deleted_at IS NULL`
	args := []interface{}{userID}
	if linkType != "" {
		query += " AND ml.type = ?"
		args = append(args, linkType)
	}
	query += " ORDER BY ml.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ListByUser: %w", err)
	}
	defer rows.Close()

	var items []matchlink.MatchLinkItem
	for rows.Next() {
		var item matchlink.MatchLinkItem
		var ca sql.NullString
		var lt sql.NullString
		if err := rows.Scan(&item.ID, &item.Token, &lt, &ca, &item.BondCount, &item.MatchCount); err != nil {
			return nil, fmt.Errorf("ListByUser scan: %w", err)
		}
		item.Type = lt.String
		item.CreatedAt = ca.String
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListByUser: %w", err)
	}
	return items, nil
}

func (r *MatchLinkRepo) SoftDelete(ctx context.Context, id int64, userID int64) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE match_links SET deleted_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = ? AND user_id = ? AND deleted_at IS NULL`, id, userID)
	if err != nil {
		return false, fmt.Errorf("SoftDelete: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// InsertMingliMatchEvent stores a BaZi match result.
func (r *MatchLinkRepo) InsertMingliMatchEvent(ctx context.Context, params matchlink.InsertMingliMatchEventParams) error {
	var otherID interface{}
	if params.OtherUserID != nil {
		otherID = *params.OtherUserID
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO mingli_match_events (link_id, initiator_user_id, other_user_id, other_name, chart_a_json, chart_b_json, match_json, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		params.LinkID, params.InitiatorUserID, otherID, params.OtherName,
		params.ChartAJSON, params.ChartBJSON, params.MatchJSON,
		time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("InsertMingliMatchEvent: %w", err)
	}
	return nil
}

// Ensure MatchLinkRepo satisfies the interface.
var _ matchlink.MatchLinkRepository = (*MatchLinkRepo)(nil)
