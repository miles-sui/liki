package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/25types/25types/internal/app/application/user"
	"github.com/25types/25types/internal/app/domain"
)

// UserRepo implements user.UserRepository and user.TokenValidator.
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

const userQuery = `SELECT u.id, u.name, u.password_hash, u.token_version, u.email, u.email_verified_at, u.pending_email, u.is_public, u.deactivated_at, u.birth_info_json, u.created_at, u.updated_at, ds.first_donation FROM users u LEFT JOIN (SELECT user_id, MIN(created_at) AS first_donation FROM donations GROUP BY user_id) ds ON u.id = ds.user_id`

// --- UserRepository ---

func (r *UserRepo) FindByName(ctx context.Context, name string) (*domain.User, error) {
	var u domain.User
	err := r.scanUser(r.db.QueryRowContext(ctx,
		userQuery+` WHERE u.name = ?`, name), &u)
	if err != nil {
		return nil, fmt.Errorf("FindByName: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.scanUser(r.db.QueryRowContext(ctx,
		userQuery+` WHERE u.email = ? AND u.email_verified_at IS NOT NULL`, email), &u)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("FindByEmail: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) FindByPendingEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.scanUser(r.db.QueryRowContext(ctx,
		userQuery+` WHERE u.pending_email = ? AND u.pending_email != ''`, email), &u)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("FindByPendingEmail: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	var u domain.User
	err := r.scanUser(r.db.QueryRowContext(ctx,
		userQuery+` WHERE u.id = ?`, id), &u)
	if err != nil {
		return nil, fmt.Errorf("FindByID: %w", err)
	}
	return &u, nil
}

// IsPublicByID returns whether a user's profile is public (for visibility checks).
func (r *UserRepo) IsPublicByID(ctx context.Context, userID int64) (bool, error) {
	var isPublic bool
	err := r.db.QueryRowContext(ctx,
		`SELECT is_public FROM users WHERE id = ?`, userID).Scan(&isPublic)
	if err != nil {
		return false, err
	}
	return isPublic, nil
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO users (name, password_hash) VALUES (?, ?)`,
		u.Name, u.PasswordHash)
	if err != nil {
		if de := isConstraintError(err, domain.ErrUsernameTaken); de != nil {
			return 0, de
		}
		return 0, fmt.Errorf("Create: %w", err)
	}
	return result.LastInsertId()
}

func (r *UserRepo) UpdateTokenVersion(ctx context.Context, id int64) (int, error) {
	return r.updateAndFetchTokenVersion(ctx,
		`UPDATE users SET token_version = token_version + 1 WHERE id = ?`, id)
}

func (r *UserRepo) UpdatePasswordHash(ctx context.Context, id int64, hash string) (int, error) {
	return r.updateAndFetchTokenVersion(ctx,
		`UPDATE users SET password_hash = ?, token_version = token_version + 1 WHERE id = ?`,
		hash, id)
}

func (r *UserRepo) UpdateFields(ctx context.Context, id int64, fields user.UpdateUserFields) error {
	cols := []string{}
	args := []interface{}{}
	if fields.Name != nil {
		cols = append(cols, "name = ?")
		args = append(args, *fields.Name)
	}
	if fields.Email != nil {
		cols = append(cols, "pending_email = ?")
		args = append(args, *fields.Email)
	}
	if fields.IsPublic != nil {
		cols = append(cols, "is_public = ?")
		args = append(args, *fields.IsPublic)
	}
		if fields.BirthInfo != nil {
		cols = append(cols, "birth_info_json = ?")
		args = append(args, *fields.BirthInfo)
		}
	if len(cols) == 0 {
		return nil
	}
	cols = append(cols, "updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')")
	query := "UPDATE users SET " + strings.Join(cols, ", ") + " WHERE id = ?"
	args = append(args, id)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		if de := isConstraintError(err, domain.ErrUsernameTaken); de != nil {
			return de
		}
		return fmt.Errorf("UpdateFields: %w", err)
	}
	if fields.EmailVerToken != nil {
		_, err = r.db.ExecContext(ctx,
			`INSERT OR REPLACE INTO user_tokens (user_id, token_type, token, expires_at)
			 VALUES (?, 'email_verify', ?, ?)`,
			id, *fields.EmailVerToken, time.Now().UTC().Add(24*time.Hour).Format(time.RFC3339))
		if err != nil {
			return fmt.Errorf("UpdateFields token: %w", err)
		}
	}
	return nil
}

func (r *UserRepo) SetDeactivated(ctx context.Context, id int64, at time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET token_version = token_version + 1, deactivated_at = ? WHERE id = ?`,
		at.Format(time.RFC3339), id)
	if err != nil {
		return fmt.Errorf("SetDeactivated: %w", err)
	}
	return nil
}

// ReactivateUser clears deactivated_at and increments token_version.
// Called after successful login within the 7-day grace period.
func (r *UserRepo) ReactivateUser(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET token_version = token_version + 1, deactivated_at = NULL WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("ReactivateUser: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) VerifyEmailByToken(ctx context.Context, token string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("VerifyEmailByToken begin: %w", err)
	}
	defer tx.Rollback()

	// Look up token in user_tokens and delete it atomically.
	var userID int64
	err = tx.QueryRowContext(ctx,
		`SELECT user_id FROM user_tokens
		 WHERE token = ? AND token_type = 'email_verify' AND expires_at > ?`,
		token, time.Now().UTC().Format(time.RFC3339)).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrTokenExpired
		}
		return fmt.Errorf("VerifyEmailByToken lookup: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM user_tokens WHERE user_id = ? AND token_type = 'email_verify'`, userID); err != nil {
		return fmt.Errorf("VerifyEmailByToken delete: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET
			email = CASE WHEN pending_email != '' THEN pending_email ELSE email END,
			pending_email = '',
			email_verified_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		 WHERE id = ?`, userID); err != nil {
		return fmt.Errorf("VerifyEmailByToken update: %w", err)
	}

	return tx.Commit()
}

func (r *UserRepo) SetPasswordResetToken(ctx context.Context, email, token string, exp time.Time) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO user_tokens (user_id, token_type, token, expires_at)
		 SELECT id, 'password_reset', ?, ? FROM users
		 WHERE (email = ? AND email_verified_at IS NOT NULL) OR pending_email = ?`,
		token, exp.UTC().Format(time.RFC3339), email, email)
	if err != nil {
		return fmt.Errorf("SetPasswordResetToken: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) FindByPasswordResetToken(ctx context.Context, token string) (int64, error) {
	var uid int64
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id FROM user_tokens
		 WHERE token = ? AND token_type = 'password_reset' AND expires_at > ?`,
		token, time.Now().UTC().Format(time.RFC3339)).Scan(&uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrTokenExpired
		}
		return 0, fmt.Errorf("FindByPasswordResetToken: %w", err)
	}
	return uid, nil
}

func (r *UserRepo) ResetPassword(ctx context.Context, id int64, hash string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ResetPassword begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET password_hash = ?, token_version = token_version + 1 WHERE id = ?`,
		hash, id); err != nil {
		return fmt.Errorf("ResetPassword update: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM user_tokens WHERE user_id = ? AND token_type = 'password_reset'`, id); err != nil {
		return fmt.Errorf("ResetPassword delete token: %w", err)
	}
	return tx.Commit()
}

func (r *UserRepo) DeleteUser(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("DeleteUser begin: %w", err)
	}
	defer tx.Rollback()

	// Anonymize assessments: unlink from user, wipe fingerprinting token.
	if _, err := tx.ExecContext(ctx,
		`UPDATE assessments SET user_id = NULL, legacy_user_token = '' WHERE user_id = ?`, id); err != nil {
		return fmt.Errorf("DeleteUser anonymize assessments: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM assessments WHERE review_link_id IN (SELECT id FROM review_links WHERE subject_user_id = ?)`, id); err != nil {
		return fmt.Errorf("DeleteUser peer assessments: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM review_links WHERE subject_user_id = ?`, id); err != nil {
		return fmt.Errorf("DeleteUser review links: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM bond_events WHERE initiator_user_id = ? OR other_user_id = ?`, id, id); err != nil {
		return fmt.Errorf("DeleteUser bond events: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM match_links WHERE user_id = ?`, id); err != nil {
		return fmt.Errorf("DeleteUser match links: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM user_tokens WHERE user_id = ?`, id); err != nil {
		return fmt.Errorf("DeleteUser tokens: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM users WHERE id = ?`, id); err != nil {
		return fmt.Errorf("DeleteUser user row: %w", err)
	}

	return tx.Commit()
}

// --- TokenValidator ---

func (r *UserRepo) GetTokenVersion(ctx context.Context, userID int64) (int, error) {
	var dbVersion int
	if err := r.db.QueryRowContext(ctx,
		`SELECT token_version FROM users WHERE id = ? AND deactivated_at IS NULL`,
		userID).Scan(&dbVersion); err != nil {
		return 0, fmt.Errorf("GetTokenVersion: %w", err)
	}
	return dbVersion, nil
}

// --- Export helpers ---

// GetExportAssessments returns all assessments for a user.
func (r *UserRepo) GetExportAssessments(ctx context.Context, userID int64) ([]user.ExportAssessment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, assessment_type, identity_id, profile_json, answers_json,
		 created_at, review_link_id, reviewer_name
		 FROM assessments WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("GetExportAssessments: %w", err)
	}
	defer rows.Close()

	var as []user.ExportAssessment
	for rows.Next() {
		var a user.ExportAssessment
		var rlid sql.NullInt64
		var rname string
		if err := rows.Scan(&a.ID, &a.Type, &a.IdentityID, &a.ProfileJSON, &a.AnswersJSON,
			&a.CreatedAt, &rlid, &rname); err != nil {
			return nil, fmt.Errorf("GetExportAssessments scan: %w", err)
		}
		if rlid.Valid {
			v := rlid.Int64
			a.ReviewLinkID = &v
		}
		if rname != "" {
			a.ReviewerName = rname
		}
		as = append(as, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetExportAssessments: %w", err)
	}
	return as, nil
}

// GetExportReviewLinks returns all non-deleted review links for a user.
func (r *UserRepo) GetExportReviewLinks(ctx context.Context, userID int64) ([]user.ExportReviewLink, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, token, expires_at, created_at FROM review_links
		 WHERE subject_user_id = ? AND deleted_at IS NULL`, userID)
	if err != nil {
		return nil, fmt.Errorf("GetExportReviewLinks: %w", err)
	}
	defer rows.Close()

	var rl []user.ExportReviewLink
	for rows.Next() {
		var l user.ExportReviewLink
		if err := rows.Scan(&l.ID, &l.Token, &l.ExpiresAt, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("GetExportReviewLinks scan: %w", err)
		}
		rl = append(rl, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetExportReviewLinks: %w", err)
	}
	return rl, nil
}

// --- Helpers ---

// scanUser scans a user row from a query into a domain.User.
func (r *UserRepo) scanUser(row *sql.Row, u *domain.User) error {
	var emailVerified, pendingEmail, deactivated sql.NullString
	var createdAt, updatedAt, supporterSince sql.NullString
	var birthInfoJSON string
	err := row.Scan(&u.ID, &u.Name, &u.PasswordHash, &u.TokenVersion, &u.Email, &emailVerified,
		&pendingEmail, &u.IsPublic, &deactivated, &birthInfoJSON, &createdAt, &updatedAt, &supporterSince)
	if err != nil {
		return err
	}
	u.EmailVerifiedAt = parseNullTime(emailVerified)
	u.PendingEmail = nullStrToStr(pendingEmail)
	u.DeactivatedAt = parseNullTime(deactivated)
	u.SupporterSince = parseNullTime(supporterSince)
	u.SetBirthInfoJSON(birthInfoJSON)
	if t := parseNullTime(createdAt); t != nil {
		u.CreatedAt = *t
	}
	if t := parseNullTime(updatedAt); t != nil {
		u.UpdatedAt = *t
	}
	return nil
}

// updateAndFetchTokenVersion executes an update and returns the new token_version.
// Uses a transaction so the returned version matches the post-update state.
func (r *UserRepo) updateAndFetchTokenVersion(ctx context.Context, query string, args ...interface{}) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("updateAndFetchTokenVersion begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return 0, fmt.Errorf("updateAndFetchTokenVersion update: %w", err)
	}
	var tv int
	if err := tx.QueryRowContext(ctx,
		`SELECT token_version FROM users WHERE id = ?`, args[len(args)-1]).Scan(&tv); err != nil {
		return 0, fmt.Errorf("updateAndFetchTokenVersion select: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("updateAndFetchTokenVersion commit: %w", err)
	}
	return tv, nil
}

func parseNullTime(ns sql.NullString) *time.Time {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, ns.String)
	if err != nil {
		return nil
	}
	return &t
}

func nullStrToStr(ns sql.NullString) *string {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	return &ns.String
}

// Ensure UserRepo satisfies the interfaces.
var (
	_ user.UserRepository   = (*UserRepo)(nil)
	_ user.TokenValidator   = (*UserRepo)(nil)
	_ user.ExportRepository = (*UserRepo)(nil)
)
