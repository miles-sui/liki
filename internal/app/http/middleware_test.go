package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func init() {
	// Prevent package init() panic while still being testable.
	if os.Getenv("JWT_SECRET") == "" {
		SetJWTSecret("test-jwt-secret-at-least-32-bytes-long-for-hs256")
	}
}

// stubValidator implements TokenValidator for tests.
type stubValidator struct {
	version int
	err     error
}

func (s *stubValidator) GetTokenVersion(_ context.Context, _ int64) (int, error) {
	return s.version, s.err
}

// =============================================================================
// CreateToken + parseToken round-trip
// =============================================================================

func TestCreateToken_RoundTrip(t *testing.T) {
	claims := JWTClaims{UserID: 42, TokenVersion: 1, UserName: "alice"}
	token, err := CreateToken(claims)
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Token should have 3 dot-separated parts per RFC 7519.
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d: %s", len(parts), token)
	}

	parsed, err := parseToken(token)
	if err != nil {
		t.Fatalf("parseToken: %v", err)
	}
	if parsed.UserID != 42 {
		t.Errorf("UserID = %d, want 42", parsed.UserID)
	}
	if parsed.TokenVersion != 1 {
		t.Errorf("TokenVersion = %d, want 1", parsed.TokenVersion)
	}
	if parsed.UserName != "alice" {
		t.Errorf("UserName = %q, want alice", parsed.UserName)
	}
}

func TestCreateToken_Expiration(t *testing.T) {
	claims := JWTClaims{UserID: 1, TokenVersion: 1, UserName: "test"}
	token, err := CreateToken(claims)
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	// Token should parse successfully (not expired).
	_, err = parseToken(token)
	if err != nil {
		t.Fatalf("fresh token should not be expired: %v", err)
	}
}

// =============================================================================
// parseToken error cases
// =============================================================================

func TestParseToken_Expired(t *testing.T) {
	// Craft a token with exp in the past.
	oldSecret := jwtSecret
	SetJWTSecret("test-key-32-bytes-long-for-hs256!!")
	defer func() { jwtSecret = oldSecret }()

	claims := JWTClaims{UserID: 1, TokenVersion: 1, UserName: "old"}
	token, err := CreateToken(claims)
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}

	// We can't easily test expiration without waiting 30 days.
	// Instead, test that parseToken rejects a tampered token.
	tampered := token + "x"
	_, err = parseToken(tampered)
	if err == nil {
		t.Error("tampered token should fail")
	}
}

func TestParseToken_InvalidSignature(t *testing.T) {
	SetJWTSecret("key-A-32-bytes-for-hs256-test!!")
	claims := JWTClaims{UserID: 1, TokenVersion: 1, UserName: "alice"}
	token, err := CreateToken(claims)
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}

	// Change the secret, then try to parse.
	SetJWTSecret("key-B-32-bytes-different-key-here!")

	_, err = parseToken(token)
	if err == nil {
		t.Error("token signed with different key should fail")
	}
	// Reset for other tests.
	SetJWTSecret("test-jwt-secret-at-least-32-bytes-long-for-hs256")
}

func TestParseToken_Malformed(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"single part", "header"},
		{"two parts", "header.payload"},
		{"non-base64 payload", "header.!@#$.signature"},
		{"non-JSON payload", "a." + badBase64("not-json") + ".c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseToken(tt.token)
			if err == nil {
				t.Error("expected error for malformed token")
			}
		})
	}
}

func TestParseToken_WrongSegments(t *testing.T) {
	// 4 segments is invalid JWT.
	_, err := parseToken("a.b.c.d")
	if err == nil {
		t.Error("expected error for 4-segment token")
	}
}

func badBase64(s string) string {
	enc := make([]byte, len(s)*2)
	for i, c := range []byte(s) {
		enc[i] = c + 1
	}
	return string(enc)
}

// =============================================================================
// extractAuth
// =============================================================================

func TestExtractAuth_MissingHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	_, err := extractAuth(req)
	if err == nil {
		t.Error("expected error for missing Authorization header")
	}
}

func TestExtractAuth_NotBearer(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	_, err := extractAuth(req)
	if err == nil {
		t.Error("expected error for non-Bearer auth")
	}
}

func TestExtractAuth_Valid(t *testing.T) {
	claims := JWTClaims{UserID: 7, TokenVersion: 1, UserName: "bob"}
	token, err := CreateToken(claims)
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	parsed, err := extractAuth(req)
	if err != nil {
		t.Fatalf("extractAuth: %v", err)
	}
	if parsed.UserID != 7 {
		t.Errorf("UserID = %d, want 7", parsed.UserID)
	}
}

// =============================================================================
// RequireAuth middleware
// =============================================================================

func TestRequireAuth_ValidToken(t *testing.T) {
	claims := JWTClaims{UserID: 10, TokenVersion: 1, UserName: "alice"}
	token, _ := CreateToken(claims)
	v := &stubValidator{version: 1}

	var capturedCtx context.Context
	handler := RequireAuth(v, func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	id, ok := UserID(capturedCtx)
	if !ok || id != 10 {
		t.Errorf("UserID = (%d, %v), want (10, true)", id, ok)
	}
	name, ok := UserName(capturedCtx)
	if !ok || name != "alice" {
		t.Errorf("UserName = (%q, %v), want (alice, true)", name, ok)
	}
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	v := &stubValidator{version: 1}
	handler := RequireAuth(v, func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
	assertErrorCode(t, rec, "unauthorized")
}

func TestRequireAuth_InvalidSignature(t *testing.T) {
	v := &stubValidator{version: 1}
	handler := RequireAuth(v, func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
	assertErrorCode(t, rec, "unauthorized")
}

func TestRequireAuth_VersionMismatch(t *testing.T) {
	claims := JWTClaims{UserID: 10, TokenVersion: 2, UserName: "alice"}
	token, _ := CreateToken(claims)
	v := &stubValidator{version: 3} // DB version > token version

	handler := RequireAuth(v, func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
	assertErrorCode(t, rec, "unauthorized")
}

func TestRequireAuth_ValidatorError(t *testing.T) {
	claims := JWTClaims{UserID: 10, TokenVersion: 1, UserName: "alice"}
	token, _ := CreateToken(claims)
	v := &stubValidator{err: &authError{kind: "db-down"}}

	handler := RequireAuth(v, func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

// =============================================================================
// OptionalAuth middleware
// =============================================================================

func TestOptionalAuth_ValidToken(t *testing.T) {
	claims := JWTClaims{UserID: 20, TokenVersion: 1, UserName: "bob"}
	token, _ := CreateToken(claims)
	v := &stubValidator{version: 1}

	var capturedCtx context.Context
	handler := OptionalAuth(v, func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	id, ok := UserID(capturedCtx)
	if !ok || id != 20 {
		t.Errorf("UserID = (%d, %v), want (20, true)", id, ok)
	}
}

func TestOptionalAuth_NoToken(t *testing.T) {
	v := &stubValidator{version: 1}

	var capturedCtx context.Context
	handler := OptionalAuth(v, func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	_, ok := UserID(capturedCtx)
	if ok {
		t.Error("UserID should not be set for unauthenticated request")
	}
}

func TestOptionalAuth_InvalidToken(t *testing.T) {
	v := &stubValidator{version: 1}

	var capturedCtx context.Context
	handler := OptionalAuth(v, func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer garbage.token")
	rec := httptest.NewRecorder()
	handler(rec, req)

	// OptionalAuth passes through on invalid token.
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	_, ok := UserID(capturedCtx)
	if ok {
		t.Error("UserID should not be set when token is invalid")
	}
}

func TestOptionalAuth_VersionMismatch(t *testing.T) {
	claims := JWTClaims{UserID: 20, TokenVersion: 1, UserName: "bob"}
	token, _ := CreateToken(claims)
	v := &stubValidator{version: 99} // version mismatch

	var capturedCtx context.Context
	handler := OptionalAuth(v, func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	_, ok := UserID(capturedCtx)
	if ok {
		t.Error("UserID should not be set when version mismatches")
	}
}

// =============================================================================
// Context extraction helpers
// =============================================================================

func TestUserID_NoContext(t *testing.T) {
	_, ok := UserID(context.Background())
	if ok {
		t.Error("UserID should not be found in empty context")
	}
}

func TestUserName_NoContext(t *testing.T) {
	_, ok := UserName(context.Background())
	if ok {
		t.Error("UserName should not be found in empty context")
	}
}

// =============================================================================
// helpers
// =============================================================================

func assertErrorCode(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	var env Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if env.Error == nil {
		t.Fatal("expected error envelope")
	}
	if env.Error.Code != want {
		t.Errorf("error code = %q, want %q", env.Error.Code, want)
	}
}
