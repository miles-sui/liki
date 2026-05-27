package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/25types/25types/internal/25types"
	"github.com/25types/25types/internal/app/application/reviewlink"
	"github.com/25types/25types/internal/app/domain"
)

// ReviewLinkRepo implements reviewlink.ReviewLinkRepository.
type ReviewLinkRepo struct {
	db *sql.DB
}

// NewReviewLinkRepo creates a new ReviewLinkRepo.
func NewReviewLinkRepo(db *sql.DB) *ReviewLinkRepo {
	return &ReviewLinkRepo{db: db}
}

func (r *ReviewLinkRepo) CreateLink(ctx context.Context, subjectUserID int64, token, expiresAt string) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO review_links (subject_user_id, token, expires_at) VALUES (?, ?, ?)`,
		subjectUserID, token, expiresAt)
	if err != nil {
		return 0, fmt.Errorf("CreateLink: %w", err)
	}
	return res.LastInsertId()
}

func (r *ReviewLinkRepo) FindByToken(ctx context.Context, token string) (*domain.ReviewLink, error) {
	return r.scanReviewLink(r.db.QueryRowContext(ctx,
		`SELECT id, subject_user_id, token, expires_at, created_at, deleted_at
		 FROM review_links WHERE token = ? AND deleted_at IS NULL`, token), "FindByToken")
}

func (r *ReviewLinkRepo) FindLinkByID(ctx context.Context, id int64) (*domain.ReviewLink, error) {
	return r.scanReviewLink(r.db.QueryRowContext(ctx,
		`SELECT id, subject_user_id, token, expires_at, created_at, deleted_at
		 FROM review_links WHERE id = ? AND deleted_at IS NULL`, id), "FindLinkByID")
}

func (r *ReviewLinkRepo) ListBySubject(ctx context.Context, subjectUserID int64) ([]reviewlink.ReviewLinkItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT rl.id, rl.token, rl.expires_at, rl.created_at,
		 (SELECT COUNT(*) FROM assessments WHERE review_link_id = rl.id) as submission_count
		 FROM review_links rl WHERE rl.subject_user_id = ? AND rl.deleted_at IS NULL ORDER BY rl.created_at DESC`, subjectUserID)
	if err != nil {
		return nil, fmt.Errorf("ListBySubject: %w", err)
	}
	defer rows.Close()

	var items []reviewlink.ReviewLinkItem
	for rows.Next() {
		var item reviewlink.ReviewLinkItem
		var exp sql.NullString
		rows.Scan(&item.ID, &item.Token, &exp, &item.CreatedAt, &item.SubmissionCount)
		if exp.Valid {
			item.ExpiresAt = exp.String
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListBySubject: %w", err)
	}
	return items, nil
}

func (r *ReviewLinkRepo) SoftDelete(ctx context.Context, id int64, subjectUserID int64) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE review_links SET deleted_at = ? WHERE id = ? AND subject_user_id = ? AND deleted_at IS NULL`,
		time.Now().UTC().Format(time.RFC3339), id, subjectUserID)
	if err != nil {
		return false, fmt.Errorf("SoftDelete: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *ReviewLinkRepo) Renew(ctx context.Context, id int64, subjectUserID int64, newExpires string) (string, bool, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE review_links SET expires_at = ? WHERE id = ? AND subject_user_id = ? AND deleted_at IS NULL`,
		newExpires, id, subjectUserID)
	if err != nil {
		return "", false, fmt.Errorf("Renew update: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return "", false, nil
	}
	var token string
	if err := r.db.QueryRowContext(ctx, `SELECT token FROM review_links WHERE id = ?`, id).Scan(&token); err != nil {
		return "", false, fmt.Errorf("Renew readback: %w", err)
	}
	return token, true, nil
}

func (r *ReviewLinkRepo) CreatePeerSubmission(ctx context.Context, sub *reviewlink.PeerSubmission) error {
	var userID interface{} = nil
	if sub.UserID != nil {
		userID = *sub.UserID
	}
	profileJSON := marshalJSON(map[string]interface{}{
		"d": sub.Profile.D, "p": sub.Profile.P,
	})
	answersJSON := marshalJSON(sub.Answers)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO assessments (user_id, assessment_type, identity_id, answers_json, profile_json, review_link_id, reviewer_name, legacy_user_token)
		 VALUES (?, '`+string(domain.AssessPeer)+`', ?, ?, ?, ?, ?, ?)`,
		userID, sub.Profile.Identity.ID, answersJSON, profileJSON,
		sub.ReviewLinkID, sub.ReviewerName, sub.AnonymousToken)
	if err != nil {
		return fmt.Errorf("CreatePeerSubmission: %w", err)
	}
	return nil
}

func (r *ReviewLinkRepo) GetPeerQIDStats(ctx context.Context, linkID int64) (map[string]domain.QIDStat, error) {
	stats := map[string]domain.QIDStat{}
	rows, err := r.db.QueryContext(ctx,
		`SELECT answers_json FROM assessments WHERE review_link_id = ? AND assessment_type = '`+string(domain.AssessPeer)+`'`, linkID)
	if err != nil {
		return nil, fmt.Errorf("GetPeerQIDStats query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var aj string
		if rows.Scan(&aj) != nil {
			continue
		}
		var answers []persona.Answer
		if json.Unmarshal([]byte(aj), &answers) != nil {
			continue
		}
		for _, a := range answers {
			if a.QID == "" {
				continue
			}
			s := stats[a.QID]
			s.Count++
			stats[a.QID] = s
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetPeerQIDStats: %w", err)
	}
	return stats, nil
}

// CROSS-AGGREGATE: reads users.name for display convenience.
func (r *ReviewLinkRepo) GetSubjectName(ctx context.Context, subjectUserID int64) string {
	var name string
	r.db.QueryRowContext(ctx, `SELECT name FROM users WHERE id = ?`, subjectUserID).Scan(&name)
	return name
}

func (r *ReviewLinkRepo) GetReviewSubmissions(ctx context.Context, linkID int64) ([]reviewlink.ReviewSubmissionItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT reviewer_name, answers_json, created_at FROM assessments WHERE review_link_id = ? ORDER BY created_at DESC`, linkID)
	if err != nil {
		return nil, fmt.Errorf("GetReviewSubmissions: %w", err)
	}
	defer rows.Close()

	var subs []reviewlink.ReviewSubmissionItem
	for rows.Next() {
		var s reviewlink.ReviewSubmissionItem
		var aj string
		if err := rows.Scan(&s.ReviewerName, &aj, &s.LastSubmittedAt); err != nil {
			continue
		}
		var ans []persona.Answer
		if json.Unmarshal([]byte(aj), &ans) != nil {
			continue
		}
		s.AnsweredCount = len(ans)
		subs = append(subs, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetReviewSubmissions: %w", err)
	}
	return subs, nil
}

func (r *ReviewLinkRepo) ListReviewsGivenByUser(ctx context.Context, userID int64) ([]reviewlink.ReviewsGivenItem, error) {
	return r.listReviewsGiven(ctx,
		`SELECT a.reviewer_name, a.answers_json, a.created_at, u.name
		 FROM assessments a
		 JOIN review_links r ON a.review_link_id = r.id
		 JOIN users u ON r.subject_user_id = u.id
		 WHERE a.user_id = ? AND a.assessment_type = '`+string(domain.AssessPeer)+`'
		 ORDER BY a.created_at DESC`, userID, "ListReviewsGivenByUser")
}

func (r *ReviewLinkRepo) ListReviewsGivenByToken(ctx context.Context, anonToken string) ([]reviewlink.ReviewsGivenItem, error) {
	return r.listReviewsGiven(ctx,
		`SELECT a.reviewer_name, a.answers_json, a.created_at, u.name
		 FROM assessments a
		 JOIN review_links r ON a.review_link_id = r.id
		 JOIN users u ON r.subject_user_id = u.id
		 WHERE a.legacy_user_token = ? AND a.assessment_type = '`+string(domain.AssessPeer)+`'
		 ORDER BY a.created_at DESC`, anonToken, "ListReviewsGivenByToken")
}

func (r *ReviewLinkRepo) listReviewsGiven(ctx context.Context, query string, arg interface{}, caller string) ([]reviewlink.ReviewsGivenItem, error) {
	rows, err := r.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", caller, err)
	}
	defer rows.Close()

	var items []reviewlink.ReviewsGivenItem
	for rows.Next() {
		var item reviewlink.ReviewsGivenItem
		var aj, subjName, reviewerName string
		if err := rows.Scan(&reviewerName, &aj, &item.CreatedAt, &subjName); err != nil {
			continue
		}
		item.SubjectName = subjName
		var ans []persona.Answer
		if json.Unmarshal([]byte(aj), &ans) != nil {
			continue
		}
		item.AnsweredCount = len(ans)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", caller, err)
	}
	return items, nil
}

func (r *ReviewLinkRepo) scanReviewLink(row *sql.Row, caller string) (*domain.ReviewLink, error) {
	var link domain.ReviewLink
	var subUID sql.NullInt64
	var exp, deleted, created sql.NullString
	err := row.Scan(&link.ID, &subUID, &link.Token, &exp, &created, &deleted)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", caller, err)
	}
	if subUID.Valid {
		link.SubjectUserID = subUID.Int64
	}
	if t := parseNullTime(exp); t != nil {
		link.ExpiresAt = *t
	}
	if t := parseNullTime(created); t != nil {
		link.CreatedAt = *t
	}
	link.DeletedAt = parseNullTime(deleted)
	return &link, nil
}

// marshalJSON serializes v to a JSON string. Panics on error (programming error).
func marshalJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("marshalJSON: %v", err))
	}
	return string(b)
}

// Compile-time interface checks.
var (
	_ reviewlink.ReviewLinkRepository       = (*ReviewLinkRepo)(nil)
	_ reviewlink.ReviewSubmissionRepository = (*ReviewLinkRepo)(nil)
)
