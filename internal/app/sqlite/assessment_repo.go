package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/25types/25types/internal/25types"
	"github.com/25types/25types/internal/app/application/assessment"
	"github.com/25types/25types/internal/app/domain"
)

// AssessmentRepo implements assessment.AssessmentRepository and domain.ProfileLoader.
type AssessmentRepo struct {
	db *sql.DB
}

// NewAssessmentRepo creates a new AssessmentRepo.
func NewAssessmentRepo(db *sql.DB) *AssessmentRepo {
	return &AssessmentRepo{db: db}
}

// --- AssessmentRepository ---

func (r *AssessmentRepo) LoadProfile(ctx context.Context, userID int64) (*domain.PersonalityProfile, error) {
	prof, err := r.FindLatestProfile(ctx, userID)
	if errors.Is(err, domain.ErrNoProfile) {
		return nil, domain.ErrUserNotFound
	}
	return prof, err
}

func (r *AssessmentRepo) CreateSelf(ctx context.Context, userID int64, profile domain.PersonalityProfile, answers []persona.Answer) (int64, error) {
	pj, aj := assessmentJSON(profile, answers)
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO assessments (user_id, assessment_type, identity_id, answers_json, profile_json)
		 VALUES (?, '`+string(domain.AssessSelf)+`', ?, ?, ?)`,
		userID, profile.Identity.ID, aj, pj)
	if err != nil {
		return 0, fmt.Errorf("CreateSelf: %w", err)
	}
	return res.LastInsertId()
}

func (r *AssessmentRepo) CreateAnonymous(ctx context.Context, profile domain.PersonalityProfile, answers []persona.Answer, token string) (int64, error) {
	pj, aj := assessmentJSON(profile, answers)
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO assessments (assessment_type, identity_id, answers_json, profile_json, legacy_user_token)
		 VALUES ('`+string(domain.AssessSelf)+`', ?, ?, ?, ?)`,
		profile.Identity.ID, aj, pj, token)
	if err != nil {
		return 0, fmt.Errorf("CreateAnonymous: %w", err)
	}
	return res.LastInsertId()
}

// assessmentJSON marshals profile and answers to JSON strings for storage.
func assessmentJSON(profile domain.PersonalityProfile, answers []persona.Answer) (profileJSON, answersJSON string) {
	return marshalJSON(map[string]interface{}{"d": profile.D, "p": profile.P}), marshalJSON(answers)
}

func (r *AssessmentRepo) FindLatestProfile(ctx context.Context, userID int64) (*domain.PersonalityProfile, error) {
	var identityID sql.NullString
	var profileJSON sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT identity_id, profile_json FROM assessments
		 WHERE user_id = ? AND assessment_type = '`+string(domain.AssessSelf)+`' ORDER BY id DESC LIMIT 1`, userID,
	).Scan(&identityID, &profileJSON)
	if err != nil || !profileJSON.Valid {
		return nil, domain.ErrNoProfile
	}
	return loadProfile(identityID, profileJSON)
}

func (r *AssessmentRepo) ListSelf(ctx context.Context, userID int64, offset, limit int) ([]domain.Assessment, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM assessments WHERE user_id = ? AND assessment_type = '`+string(domain.AssessSelf)+`'`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("ListSelf count: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, identity_id, profile_json, created_at FROM assessments
		 WHERE user_id = ? AND assessment_type = '`+string(domain.AssessSelf)+`'
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("ListSelf: %w", err)
	}
	defer rows.Close()

	var items []domain.Assessment
	for rows.Next() {
		var a domain.Assessment
		var createdAt sql.NullString
		if err := rows.Scan(&a.ID, &a.IdentityID, &a.ProfileJSON, &createdAt); err != nil {
			return nil, 0, fmt.Errorf("ListSelf scan: %w", err)
		}
		if t := parseNullTime(createdAt); t != nil {
			a.CreatedAt = *t
		}
		items = append(items, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("ListSelf: %w", err)
	}
	return items, total, nil
}

func (r *AssessmentRepo) FindAssessmentByID(ctx context.Context, id int64) (*domain.Assessment, error) {
	var a domain.Assessment
	var uid sql.NullInt64
	var rlid sql.NullInt64
	var createdAt sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, assessment_type, identity_id, answers_json, profile_json,
		 review_link_id, reviewer_name, legacy_user_token, created_at
		 FROM assessments WHERE id = ?`, id,
	).Scan(&a.ID, &uid, &a.AssessmentType, &a.IdentityID, &a.AnswersJSON, &a.ProfileJSON,
		&rlid, &a.ReviewerName, &a.AnonymousToken, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("FindAssessmentByID: %w", err)
	}
	if uid.Valid {
		a.UserID = &uid.Int64
	}
	if rlid.Valid {
		v := rlid.Int64
		a.ReviewLinkID = &v
	}
	if t := parseNullTime(createdAt); t != nil {
		a.CreatedAt = *t
	}
	return &a, nil
}

// CROSS-AGGREGATE: queries users table (name) for display convenience.
func (r *AssessmentRepo) FindAssessmentByIDWithUser(ctx context.Context, id int64) (*domain.Assessment, *string, error) {
	a, err := r.FindAssessmentByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	var userName sql.NullString
	r.db.QueryRowContext(ctx,
		`SELECT u.name FROM assessments a LEFT JOIN users u ON a.user_id = u.id WHERE a.id = ?`, id,
	).Scan(&userName)
	if userName.Valid {
		return a, &userName.String, nil
	}
	return a, nil, nil
}

func (r *AssessmentRepo) ClaimAnonymous(ctx context.Context, userID int64, token string) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("ClaimAnonymous: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE assessments SET user_id = ? WHERE legacy_user_token = ? AND user_id IS NULL`,
		userID, token)
	if err != nil {
		return 0, fmt.Errorf("ClaimAnonymous: %w", err)
	}
	n, _ := res.RowsAffected()

	// Clear the fingerprinting token now that assessments are linked to a real identity.
	if n > 0 {
		if _, err := tx.ExecContext(ctx,
			`UPDATE assessments SET legacy_user_token = '' WHERE legacy_user_token = ? AND user_id = ?`,
			token, userID); err != nil {
			return 0, fmt.Errorf("ClaimAnonymous clear token: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("ClaimAnonymous: %w", err)
	}
	return n, nil
}

func (r *AssessmentRepo) FindSelfAnswers(ctx context.Context, userID int64) ([]persona.Answer, error) {
	var aj sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT answers_json FROM assessments WHERE user_id = ? AND assessment_type = '`+string(domain.AssessSelf)+`' ORDER BY id DESC LIMIT 1`, userID,
	).Scan(&aj)
	if err != nil || !aj.Valid {
		return nil, fmt.Errorf("FindSelfAnswers: %w", err)
	}
	var ans []persona.Answer
	if err := json.Unmarshal([]byte(aj.String), &ans); err != nil {
		return nil, fmt.Errorf("FindSelfAnswers unmarshal: %w", err)
	}
	return ans, nil
}

// ListPeerAnswersForUser returns all peer answers for a user from non-deleted links.
func (r *AssessmentRepo) ListPeerAnswersForUser(ctx context.Context, userID int64) ([]persona.Answer, int, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT a.answers_json FROM assessments a
		 JOIN review_links r ON a.review_link_id = r.id
		 WHERE r.subject_user_id = ? AND a.assessment_type = '`+string(domain.AssessPeer)+`' AND r.deleted_at IS NULL`, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("ListPeerAnswersForUser: %w", err)
	}
	defer rows.Close()

	var all []persona.Answer
	count := 0
	for rows.Next() {
		var aj string
		if err := rows.Scan(&aj); err != nil {
			continue
		}
		var ans []persona.Answer
		if json.Unmarshal([]byte(aj), &ans) == nil {
			all = append(all, ans...)
			count++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("ListPeerAnswersForUser: %w", err)
	}
	return all, count, nil
}

// loadProfile deserializes a PersonalityProfile from DB columns.
func loadProfile(identityID sql.NullString, profileJSON sql.NullString) (*domain.PersonalityProfile, error) {
	var raw struct {
		D persona.Deviation  `json:"d"`
		P persona.Proportion `json:"p"`
	}
	if err := json.Unmarshal([]byte(profileJSON.String), &raw); err != nil {
		return nil, err
	}

	idStr := ""
	if identityID.Valid {
		idStr = identityID.String
	}
	identity := persona.Identity{
		ID:       idStr,
		Label:    idStr,
		Category: persona.DeriveCategory(idStr),
	}

	prof := domain.NewProfile(raw.D, raw.P, identity)
	return &prof, nil
}

// Compile-time interface checks.
var (
	_ domain.ProfileLoader            = (*AssessmentRepo)(nil)
	_ assessment.AssessmentRepository = (*AssessmentRepo)(nil)
)
