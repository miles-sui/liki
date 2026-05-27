package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/25types/25types/internal/app/application/profile"
	"github.com/25types/25types/internal/app/domain"
)

// ProfileRepo implements profile.ProfileRepository and profile.BondStore.
type ProfileRepo struct {
	*UserRepo
	*AssessmentRepo
	db *sql.DB
}

// NewProfileRepo creates a ProfileRepo backed by existing repos plus bond_event access.
func NewProfileRepo(userRepo *UserRepo, assRepo *AssessmentRepo) *ProfileRepo {
	return &ProfileRepo{
		UserRepo:       userRepo,
		AssessmentRepo: assRepo,
		db:             userRepo.db,
	}
}

// InsertBondEvent stores a bond computation result with bond_json snapshot.
// otherID of 0 means anonymous (null in DB).
func (r *ProfileRepo) InsertBondEvent(ctx context.Context, params profile.InsertBondParams) error {
	otherID := params.OtherID
	var other interface{} = otherID
	if otherID == 0 {
		other = nil
	}

	bondJSON := "{}"
	if params.Bond != nil {
		b, err := json.Marshal(params.Bond)
		if err == nil {
			bondJSON = string(b)
		}
	}

	// Dedup: one bond per pair. Run DELETE + INSERT in a transaction to prevent
	// a concurrent insert from defeating dedup and creating a duplicate row.
	if otherID != 0 {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("InsertBondEvent begin tx: %w", err)
		}
		defer tx.Rollback()

		_, err = tx.ExecContext(ctx,
			`DELETE FROM bond_events WHERE (initiator_user_id = ? AND other_user_id = ?) OR (initiator_user_id = ? AND other_user_id = ?)`,
			params.InitiatorID, otherID, otherID, params.InitiatorID)
		if err != nil {
			return fmt.Errorf("InsertBondEvent dedup: %w", err)
		}

		_, err = tx.ExecContext(ctx,
			`INSERT INTO bond_events (link_id, initiator_user_id, other_user_id, assessment_id, bond_json, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
			params.LinkID, params.InitiatorID, other, params.AssessmentID, bondJSON, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			return fmt.Errorf("InsertBondEvent: %w", err)
		}

		return tx.Commit()
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO bond_events (link_id, initiator_user_id, other_user_id, assessment_id, bond_json, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		params.LinkID, params.InitiatorID, other, params.AssessmentID, bondJSON, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("InsertBondEvent: %w", err)
	}
	return nil
}

// ListBondEvents returns all bond events involving the user, with bond_json included.
func (r *ProfileRepo) ListBondEvents(ctx context.Context, userID int64) ([]domain.BondEvent, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT be.id, be.link_id, be.initiator_user_id, be.other_user_id,
		 be.other_name, be.bond_json, be.created_at, ou.name
		 FROM bond_events be
		 LEFT JOIN users ou ON ou.id = CASE WHEN be.initiator_user_id = ? THEN be.other_user_id ELSE be.initiator_user_id END
		 WHERE be.initiator_user_id = ? OR be.other_user_id = ?
		 ORDER BY be.created_at DESC`, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("ListBondEvents: %w", err)
	}
	defer rows.Close()

	var events []domain.BondEvent
	for rows.Next() {
		var e domain.BondEvent
		var linkID sql.NullInt64
		var otherUserID sql.NullInt64
		var otherName sql.NullString
		var bondJSON sql.NullString
		var createdAt sql.NullString
		var joinName sql.NullString
		if err := rows.Scan(&e.ID, &linkID, &e.InitiatorUserID, &otherUserID, &otherName, &bondJSON, &createdAt, &joinName); err != nil {
			return nil, fmt.Errorf("ListBondEvents scan: %w", err)
		}
		if otherUserID.Valid {
			v := otherUserID.Int64
			e.OtherUserID = &v
		}
		if joinName.Valid {
			e.OtherName = joinName.String
		} else if otherName.Valid {
			e.OtherName = otherName.String
		}
		if bondJSON.Valid {
			e.BondJSON = bondJSON.String
		}
		if linkID.Valid {
			e.LinkID = &linkID.Int64
		}
		if t := parseNullTime(createdAt); t != nil {
			e.CreatedAt = *t
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListBondEvents: %w", err)
	}
	return events, nil
}

// FindActiveReviewLink returns the token of the first active review link for a user.
func (r *ProfileRepo) FindActiveReviewLink(ctx context.Context, userID int64) (string, bool) {
	var token string
	err := r.db.QueryRowContext(ctx,
		`SELECT token FROM review_links
		 WHERE subject_user_id = ? AND (expires_at IS NULL OR expires_at > datetime('now'))
		 AND deleted_at IS NULL
		 ORDER BY created_at DESC LIMIT 1`, userID).Scan(&token)
	if err != nil {
		return "", false
	}
	return token, true
}

// Ensure ProfileRepo satisfies the interfaces.
var (
	_ domain.ProfileLoader   = (*ProfileRepo)(nil)
	_ profile.ProfilePageRepo = (*ProfileRepo)(nil)
	_ profile.BondStore      = (*ProfileRepo)(nil)
	_ profile.UserLookup     = (*ProfileRepo)(nil)
)
