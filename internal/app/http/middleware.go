package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/25types/25types/internal/app/application/user"
)


type contextKey string

const (
	ctxUserID   contextKey = "user_id"
	ctxTokenVer contextKey = "token_version"
	ctxUserName contextKey = "user_name"
)

var jwtSecret []byte

func SetJWTSecret(s string) {
	jwtSecret = []byte(s)
}

// JWTClaims is the JWT payload.
type JWTClaims struct {
	UserID       int64  `json:"uid"`
	TokenVersion int    `json:"tv"`
	UserName     string `json:"name"`
}

// CreateToken signs a new JWT token valid for 30 days.
func CreateToken(claims JWTClaims) (string, error) {
	if len(jwtSecret) == 0 {
		return "", errors.New("JWT_SECRET environment variable is required")
	}
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadBytes, err := json.Marshal(map[string]interface{}{
		"uid":  claims.UserID,
		"tv":   claims.TokenVersion,
		"name": claims.UserName,
		"exp":  time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	})
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := header + "." + payload

	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + signature, nil
}

// parseToken validates and extracts claims from a JWT token string.
func parseToken(tokenStr string) (*JWTClaims, error) {
	if len(jwtSecret) == 0 {
		return nil, errInvalidToken
	}
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, errInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(signingInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, errInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errInvalidToken
	}

	var raw struct {
		UID  int64  `json:"uid"`
		TV   int    `json:"tv"`
		Name string `json:"name"`
		Exp  int64  `json:"exp"`
	}
	if err := json.Unmarshal(payloadBytes, &raw); err != nil {
		return nil, errInvalidToken
	}
	if time.Unix(raw.Exp, 0).Before(time.Now()) {
		return nil, errTokenExpired
	}

	return &JWTClaims{
		UserID:       raw.UID,
		TokenVersion: raw.TV,
		UserName:     raw.Name,
	}, nil
}

var (
	errInvalidToken = &authError{"invalid"}
	errTokenExpired = &authError{"expired"}
)

type authError struct{ kind string }

func (e *authError) Error() string { return "auth error: " + e.kind }

// RequireAuth is middleware that requires a valid JWT token.
func RequireAuth(v user.TokenValidator, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := extractAuth(r)
		if err != nil {
			if e, ok := err.(*authError); ok && e.kind == "expired" {
				respondError(w, http.StatusUnauthorized, "token_expired", "Token expired")
				return
			}
			respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
			return
		}

		dbVersion, err := v.GetTokenVersion(r.Context(), claims.UserID)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
			return
		}
		if dbVersion != claims.TokenVersion {
			respondError(w, http.StatusUnauthorized, "unauthorized", "Token invalidated")
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
		ctx = context.WithValue(ctx, ctxTokenVer, claims.TokenVersion)
		ctx = context.WithValue(ctx, ctxUserName, claims.UserName)
		next(w, r.WithContext(ctx))
	}
}

// OptionalAuth extracts JWT if present but does not require it.
func OptionalAuth(v user.TokenValidator, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := extractAuth(r)
		if err != nil {
			next(w, r)
			return
		}

		dbVersion, err := v.GetTokenVersion(r.Context(), claims.UserID)
		if err != nil {
			next(w, r)
			return
		}
		if dbVersion != claims.TokenVersion {
			next(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
		ctx = context.WithValue(ctx, ctxTokenVer, claims.TokenVersion)
		ctx = context.WithValue(ctx, ctxUserName, claims.UserName)
		next(w, r.WithContext(ctx))
	}
}

func extractAuth(r *http.Request) (*JWTClaims, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return nil, errInvalidToken
	}
	tokenStr := strings.TrimPrefix(header, "Bearer ")
	if tokenStr == header {
		return nil, errInvalidToken
	}
	return parseToken(tokenStr)
}

// UserID extracts the authenticated user ID from context.
func UserID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(ctxUserID).(int64)
	return id, ok
}

// UserName extracts the authenticated user name from context.
func UserName(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(ctxUserName).(string)
	return name, ok
}

var adminUsers []string

func SetAdminUsers(users []string) {
	adminUsers = users
}

// RequireAdmin is middleware that requires the user to be an admin.
func RequireAdmin(v user.TokenValidator, next http.HandlerFunc) http.HandlerFunc {
	return RequireAuth(v, func(w http.ResponseWriter, r *http.Request) {
		name, _ := UserName(r.Context())
		if len(adminUsers) > 0 && !slices.Contains(adminUsers, name) {
			respondError(w, http.StatusForbidden, "forbidden", "Admin access required")
			return
		}
		next(w, r)
	})
}
