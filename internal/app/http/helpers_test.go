package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/25types/25types/internal/app/application/user"
	"github.com/25types/25types/internal/app/db"
	"github.com/25types/25types/internal/app/sqlite"
)

func init() {
	SetJWTSecret("test-secret-32-bytes-long-for-hs256!")
}
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return newTestServerWithEmail(t, nil)
}

// ── HTTP helpers ──

func doReq(t *testing.T, method, url, body, token string) (int, string) {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func getBody(t *testing.T, url string) (int, string) { return doReq(t, "GET", url, "", "") }
func postBody(t *testing.T, url, body string) string {
	_, b := doReq(t, "POST", url, body, "")
	return b
}
func getAuthBody(t *testing.T, url, token string) string {
	_, b := doReq(t, "GET", url, "", token)
	return b
}
func postAuthBody(t *testing.T, url, token, body string) string {
	_, b := doReq(t, "POST", url, body, token)
	return b
}
func patchAuthBody(t *testing.T, url, token, body string) string {
	_, b := doReq(t, "PATCH", url, body, token)
	return b
}
func deleteAuth(t *testing.T, url, token string) (int, string) {
	return doReq(t, "DELETE", url, "", token)
}

// envelopeOk extracts data field, failing on any error envelope.
func envelopeOk(t *testing.T, body string) map[string]interface{} {
	t.Helper()
	var env struct {
		Data  json.RawMessage `json:"data"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &env); err != nil {
		t.Fatalf("parse envelope: %v\nbody: %s", err, body)
	}
	if env.Error != nil {
		t.Fatalf("unexpected API error: %s — %s", env.Error.Code, env.Error.Message)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(env.Data, &m); err != nil {
		t.Fatalf("parse data: %v", err)
	}
	return m
}

// envelopeErr extracts the error code from an error response.
func envelopeErr(t *testing.T, body string) string {
	t.Helper()
	var env struct {
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &env); err != nil {
		t.Fatalf("parse error envelope: %v\nbody: %s", err, body)
	}
	if env.Error == nil {
		t.Fatalf("expected error envelope, got success\nbody: %s", body)
	}
	return env.Error.Code
}

// registerAndLogin creates a user and returns token + userID.
func registerAndLogin(t *testing.T, srv *httptest.Server, name, password string) (token string, userID float64) {
	t.Helper()
	b := postBody(t, srv.URL+"/api/auth/register",
		fmt.Sprintf(`{"name":%q,"email":%q,"password":%q}`, name, name+"@test.com", password))
	data := envelopeOk(t, b)
	return data["token"].(string), data["user"].(map[string]interface{})["id"].(float64)
}

// fullAnswersJSON returns 30 answers covering all 5 elements.
func fullAnswersJSON() string {
	pairs := [][3]string{
		{"Q01", "W", "F"}, {"Q02", "W", "M"}, {"Q03", "F", "E"}, {"Q04", "W", "F"}, {"Q05", "E", "M"},
		{"Q06", "W", "F"}, {"Q07", "W", "E"}, {"Q08", "F", "E"}, {"Q09", "W", "E"}, {"Q10", "F", "M"},
		{"Q11", "W", "E"}, {"Q12", "F", "E"}, {"Q13", "W", "R"}, {"Q14", "W", "F"}, {"Q15", "E", "M"},
		{"Q16", "W", "M"}, {"Q17", "F", "E"}, {"Q18", "W", "F"}, {"Q19", "F", "R"}, {"Q20", "W", "E"},
		{"Q21", "W", "E"}, {"Q22", "F", "R"}, {"Q23", "W", "F"}, {"Q24", "E", "M"}, {"Q25", "W", "R"},
		{"Q26", "W", "F"}, {"Q27", "W", "E"}, {"Q28", "W", "R"}, {"Q29", "F", "E"}, {"Q30", "F", "M"},
	}
	var parts []string
	for _, p := range pairs {
		parts = append(parts, `{"qid":"`+p[0]+`","selections":["`+p[1]+`","`+p[2]+`"]}`)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// submitFullAssessment submits all 30 answers for an authenticated user.
func submitFullAssessment(t *testing.T, srv *httptest.Server, token string) map[string]interface{} {
	t.Helper()
	b := postAuthBody(t, srv.URL+"/api/assessments", token,
		`{"answers":`+fullAnswersJSON()+`}`)
	return envelopeOk(t, b)
}
// stubEmailSender records sent emails for testing.
type stubEmailSender struct {
	sent  []string
	verifyTokens []string // captured verification tokens
	resetTokens  []string // captured reset tokens
}

func (s *stubEmailSender) SendVerificationEmail(_ context.Context, to, token, locale string) error {
	s.sent = append(s.sent, "verify:"+to)
	s.verifyTokens = append(s.verifyTokens, token)
	return nil
}
func (s *stubEmailSender) SendPasswordResetEmail(_ context.Context, to, token, locale string) error {
	s.sent = append(s.sent, "reset:"+to)
	s.resetTokens = append(s.resetTokens, token)
	return nil
}
func (s *stubEmailSender) SendBondNotification(_ context.Context, to, otherName, creatorName, locale string) error {
	s.sent = append(s.sent, "bond:"+to+":"+otherName)
	return nil
}

// newTestServerWithEmail is like newTestServer but injects a stub email sender.
func newTestServerWithEmail(t *testing.T, sender user.EmailSender) *httptest.Server {
	t.Helper()

	database, err := db.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	userRepo := sqlite.NewUserRepo(database)
	assRepo := sqlite.NewAssessmentRepo(database)
	rlRepo := sqlite.NewReviewLinkRepo(database)
	profileRepo := sqlite.NewProfileRepo(userRepo, assRepo)
	matchLinkRepo := sqlite.NewMatchLinkRepo(database)
	cfg := ServerConfig{
		UserRepo:   userRepo,
		UserHasher: sqlite.PasswordHasher{},
		AssRepo:    assRepo,
		LinkRepo:   rlRepo,
		SubRepo:    rlRepo,
		Profiles:          assRepo,
		ProfileRepo:       profileRepo,
		ProfileUsers:      profileRepo,
		BondStore:         profileRepo,
		MatchLinkRepo:     matchLinkRepo,
		UserEmailSender: sender,
		TokenValidator:  userRepo,
		UserLookup:      userRepo,
		ExportRepo:      userRepo,
		DB:              database,
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, cfg)
	return httptest.NewServer(mux)
}

// ── Shared assertions ──

var elementKeys = []string{"wood", "fire", "earth", "metal", "water"}

// assertDeviationSumZero checks a deviation map has all 5 elements and sums to ≈0.
func assertDeviationSumZero(t *testing.T, d map[string]interface{}, label string) {
	t.Helper()
	for _, k := range elementKeys {
		if _, ok := d[k]; !ok {
			t.Errorf("%s missing key %q", label, k)
		}
	}
	sum := 0.0
	for _, v := range d {
		sum += v.(float64)
	}
	if sum > 1e-6 || sum < -1e-6 {
		t.Errorf("%s Σ = %v, want ≈0", label, sum)
	}
}

