package http

import (
	"testing"
)

// Tests in this file cover user_handler.go:
//   Register, Login, LoginByEmail, Logout, ChangePassword,
//   ForgotPassword, ResetPassword, VerifyEmail, ResendVerification,
//   GetMe, UpdateMe, DeactivateMe, ExportMe.

// =============================================================================
// Register
// =============================================================================

func TestRegister(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	t.Run("OK", func(t *testing.T) {
		b := postBody(t, srv.URL+"/api/auth/register",
			`{"name":"reg-user","email":"reg@test.com","password":"secret1234"}`)
		data := envelopeOk(t, b)
		if data["token"] == nil || data["token"].(string) == "" {
			t.Error("expected token")
		}
	})

	t.Run("Duplicate", func(t *testing.T) {
		postBody(t, srv.URL+"/api/auth/register",
			`{"name":"dup-user","email":"dup@test.com","password":"secret1234"}`)
		code, body := doReq(t, "POST", srv.URL+"/api/auth/register",
			`{"name":"dup-user","email":"dup2@test.com","password":"secret1234"}`, "")
		if code != 409 {
			t.Errorf("status = %d, want 409", code)
		}
		_ = envelopeErr(t, body)
	})

	t.Run("ShortPassword", func(t *testing.T) {
		code, _ := doReq(t, "POST", srv.URL+"/api/auth/register",
			`{"name":"shortpw","email":"short@test.com","password":"123"}`, "")
		if code != 400 {
			t.Errorf("status = %d, want 400", code)
		}
	})

	t.Run("MissingEmail", func(t *testing.T) {
		code, _ := doReq(t, "POST", srv.URL+"/api/auth/register",
			`{"name":"noemail","password":"secret1234"}`, "")
		if code != 400 {
			t.Errorf("status = %d, want 400", code)
		}
	})

	t.Run("InvalidEmail", func(t *testing.T) {
		code, _ := doReq(t, "POST", srv.URL+"/api/auth/register",
			`{"name":"bademail","email":"not-an-email","password":"secret1234"}`, "")
		if code != 400 {
			t.Errorf("status = %d, want 400", code)
		}
	})

	t.Run("WithAnonymousToken", func(t *testing.T) {
		b := postBody(t, srv.URL+"/api/auth/register",
			`{"name":"anon-claim","email":"anonclaim@test.com","password":"secret1234","anonymous_token":"tok-12345"}`)
		data := envelopeOk(t, b)
		if data["token"] == nil || data["token"].(string) == "" {
			t.Error("expected token")
		}
	})
}

// =============================================================================
// Login
// =============================================================================

func TestLogin(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "login-user", "secret1234")

	t.Run("OK", func(t *testing.T) {
		if token == "" {
			t.Error("expected non-empty token")
		}
	})

	t.Run("WrongPassword", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/auth/login",
			`{"name":"login-user","password":"wrong-pass-123"}`, "")
		if code != 401 {
			t.Errorf("status = %d, want 401", code)
		}
		if c := envelopeErr(t, body); c != "unauthorized" {
			t.Errorf("error = %q, want unauthorized", c)
		}
	})

	t.Run("NonexistentUser", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/auth/login",
			`{"name":"no-such-user","password":"secret1234"}`, "")
		if code != 401 {
			t.Errorf("status = %d, want 401", code)
		}
		_ = envelopeErr(t, body)
	})
}

func TestLoginByEmail(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	postBody(t, srv.URL+"/api/auth/register",
		`{"name":"email-login","email":"emaillogin@test.com","password":"secret1234"}`)
	code, body := doReq(t, "POST", srv.URL+"/api/auth/login",
		`{"name":"emaillogin@test.com","password":"secret1234"}`, "")
	if code != 200 {
		t.Fatalf("login by email status = %d, want 200", code)
	}
	data := envelopeOk(t, body)
	if data["token"] == nil || data["token"].(string) == "" {
		t.Error("expected token")
	}
}

// =============================================================================
// Logout
// =============================================================================

func TestLogout(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "logout-user", "secret1234")

	t.Run("OK", func(t *testing.T) {
		code, _ := doReq(t, "POST", srv.URL+"/api/auth/logout", "", token)
		if code != 200 {
			t.Errorf("logout status = %d, want 200", code)
		}
	})

	t.Run("OldTokenInvalid", func(t *testing.T) {
		code, _ := doReq(t, "GET", srv.URL+"/api/users/me", "", token)
		if code != 401 {
			t.Errorf("after logout status = %d, want 401", code)
		}
	})
}

// =============================================================================
// Change Password
// =============================================================================

func TestChangePassword(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	t.Run("OK", func(t *testing.T) {
		token, _ := registerAndLogin(t, srv, "chpwd-ok", "secret1234")
		code, body := doReq(t, "PUT", srv.URL+"/api/auth/password",
			`{"current_password":"secret1234","new_password":"new-secret-456"}`, token)
		if code != 200 {
			t.Fatalf("change password status = %d, want 200", code)
		}
		data := envelopeOk(t, body)
		if data["token"] == nil || data["token"].(string) == "" {
			t.Error("expected new token after password change")
		}
	})

	t.Run("WrongCurrent", func(t *testing.T) {
		token, _ := registerAndLogin(t, srv, "chpwd-wrong", "secret1234")
		code, body := doReq(t, "PUT", srv.URL+"/api/auth/password",
			`{"current_password":"wrong-pass","new_password":"new-secret-789"}`, token)
		if code != 401 {
			t.Errorf("status = %d, want 401", code)
		}
		if c := envelopeErr(t, body); c != "incorrect_password" {
			t.Errorf("error = %q, want incorrect_password", c)
		}
	})

	t.Run("ShortNew", func(t *testing.T) {
		token, _ := registerAndLogin(t, srv, "chpwd-short", "secret1234")
		code, _ := doReq(t, "PUT", srv.URL+"/api/auth/password",
			`{"current_password":"secret1234","new_password":"sh"}`, token)
		if code != 400 {
			t.Errorf("status = %d, want 400", code)
		}
	})
}

// =============================================================================
// Forgot / Reset Password
// =============================================================================

func TestForgotPassword(t *testing.T) {
	stub := &stubEmailSender{}
	srv := newTestServerWithEmail(t, stub)
	defer srv.Close()

	registerAndLogin(t, srv, "forgot-user", "secret1234")

	t.Run("SendsResetEmail", func(t *testing.T) {
		code, _ := doReq(t, "POST", srv.URL+"/api/auth/forgot-password",
			`{"email":"forgot-user@test.com"}`, "")
		if code != 200 {
			t.Errorf("forgot password status = %d, want 200", code)
		}
		found := false
		for _, s := range stub.sent {
			if s == "reset:forgot-user@test.com" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected reset email sent to forgot-user@test.com, got %v", stub.sent)
		}
	})

	t.Run("UnknownEmailAlsoOk", func(t *testing.T) {
		code, _ := doReq(t, "POST", srv.URL+"/api/auth/forgot-password",
			`{"email":"unknown@test.com"}`, "")
		if code != 200 {
			t.Errorf("forgot password status = %d, want 200", code)
		}
	})
}

func TestResetPassword(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	t.Run("InvalidToken", func(t *testing.T) {
		code, _ := doReq(t, "POST", srv.URL+"/api/auth/reset-password",
			`{"token":"bad-token","new_password":"new-secret-456"}`, "")
		if code != 400 {
			t.Errorf("status = %d, want 400", code)
		}
	})

	t.Run("ShortPassword", func(t *testing.T) {
		code, _ := doReq(t, "POST", srv.URL+"/api/auth/reset-password",
			`{"token":"any-token","new_password":"sh"}`, "")
		if code != 400 {
			t.Errorf("status = %d, want 400", code)
		}
	})
}

// =============================================================================
// Verify Email / Resend Verification
// =============================================================================

func TestVerifyEmail(t *testing.T) {
	stub := &stubEmailSender{}
	srv := newTestServerWithEmail(t, stub)
	defer srv.Close()

	registerAndLogin(t, srv, "verify-me", "secret1234")

	t.Run("OK", func(t *testing.T) {
		if len(stub.verifyTokens) == 0 {
			t.Fatal("no verification token captured")
		}
		tok := stub.verifyTokens[0]
		code, body := doReq(t, "GET", srv.URL+"/api/auth/verify-email?token="+tok, "", "")
		if code != 200 {
			t.Fatalf("verify email status = %d, want 200 (body: %s)", code, body)
		}
		data := envelopeOk(t, body)
		if data["email_verified"] != true {
			t.Errorf("email_verified = %v, want true", data["email_verified"])
		}
	})

	t.Run("InvalidToken", func(t *testing.T) {
		code, _ := doReq(t, "GET", srv.URL+"/api/auth/verify-email?token=bad-token", "", "")
		if code != 400 {
			t.Errorf("verify email status = %d, want 400", code)
		}
	})

	t.Run("MissingToken", func(t *testing.T) {
		code, _ := doReq(t, "GET", srv.URL+"/api/auth/verify-email", "", "")
		if code != 400 {
			t.Errorf("verify email status = %d, want 400", code)
		}
	})
}

func TestResendVerification(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "resend-user", "secret1234")

	t.Run("OK", func(t *testing.T) {
		code, _ := doReq(t, "POST", srv.URL+"/api/auth/resend-verification", "", token)
		if code != 200 {
			t.Errorf("resend verification status = %d, want 200", code)
		}
	})

	t.Run("NoAuth", func(t *testing.T) {
		code, _ := doReq(t, "POST", srv.URL+"/api/auth/resend-verification", "", "")
		if code != 401 {
			t.Errorf("status = %d, want 401", code)
		}
	})
}

// =============================================================================
// Get / Update / Deactivate / Export Me
// =============================================================================

func TestGetMe(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "me-user", "secret1234")

	t.Run("OK", func(t *testing.T) {
		b := getAuthBody(t, srv.URL+"/api/users/me", token)
		data := envelopeOk(t, b)
		if data["name"] != "me-user" {
			t.Errorf("name = %v, want me-user", data["name"])
		}
	})

	t.Run("NoAuth", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/users/me", "", "")
		if code != 401 {
			t.Errorf("status = %d, want 401", code)
		}
		if c := envelopeErr(t, body); c != "unauthorized" {
			t.Errorf("error code = %q, want unauthorized", c)
		}
	})
}

func TestUpdateMe(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "update-me", "secret1234")

	t.Run("Rename", func(t *testing.T) {
		b := patchAuthBody(t, srv.URL+"/api/users/me", token, `{"name":"renamed"}`)
		data := envelopeOk(t, b)
		if data["name"] != "renamed" {
			t.Errorf("name = %v, want renamed", data["name"])
		}
	})

	t.Run("SetPublic", func(t *testing.T) {
		b := patchAuthBody(t, srv.URL+"/api/users/me", token, `{"is_public":true}`)
		data := envelopeOk(t, b)
		if data["is_public"] != true {
			t.Errorf("is_public = %v, want true", data["is_public"])
		}
	})

	t.Run("EmptyBody", func(t *testing.T) {
		code, body := doReq(t, "PATCH", srv.URL+"/api/users/me", `{}`, token)
		if code != 400 {
			t.Errorf("status = %d, want 400", code)
		}
		if c := envelopeErr(t, body); c != "invalid_request" {
			t.Errorf("error code = %q, want invalid_request", c)
		}
	})
}

func TestDeactivateMe(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "delete-me", "secret1234")

	t.Run("OK", func(t *testing.T) {
		code, body := doReq(t, "DELETE", srv.URL+"/api/users/me", "", token)
		if code != 200 {
			t.Fatalf("status = %d, want 200", code)
		}
		data := envelopeOk(t, body)
		if data["reactivate_by"] == nil || data["reactivate_by"].(string) == "" {
			t.Error("reactivate_by missing")
		}
	})

	t.Run("AfterDelete", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/users/me", "", token)
		if code != 401 {
			t.Errorf("status = %d, want 401", code)
		}
		if c := envelopeErr(t, body); c != "unauthorized" {
			t.Errorf("error code = %q, want unauthorized", c)
		}
	})
}

func TestExportMe(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "export-me", "secret1234")

	b := getAuthBody(t, srv.URL+"/api/users/me/export", token)
	data := envelopeOk(t, b)
	if data["user"] == nil {
		t.Error("export missing user field")
	}
}
