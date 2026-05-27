package http

import (
	"fmt"
	"strings"
	"testing"
)

func TestMatchLinkCreate(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "ml-create", "secret1234")

	t.Run("OK", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/match-links", "", token)
		if code != 201 {
			t.Fatalf("status = %d, want 201", code)
		}
		data := envelopeOk(t, body)
		if data["token"] == nil || data["token"].(string) == "" {
			t.Error("token missing or empty")
		}
		if data["id"] == nil || data["id"].(float64) <= 0 {
			t.Error("id missing or <= 0")
		}
	})

	t.Run("NoAuth", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/match-links", "", "")
		if code != 401 {
			t.Errorf("status = %d, want 401", code)
		}
		if c := envelopeErr(t, body); c != "unauthorized" {
			t.Errorf("error code = %q, want unauthorized", c)
		}
	})
}

func TestMatchLinkList(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "ml-list", "secret1234")

	// Create 2 links.
	postAuthBody(t, srv.URL+"/api/match-links", token, "")
	postAuthBody(t, srv.URL+"/api/match-links", token, "")

	t.Run("OK", func(t *testing.T) {
		b := getAuthBody(t, srv.URL+"/api/match-links", token)
		data := envelopeOk(t, b)
		items := data["items"].([]interface{})
		if len(items) != 2 {
			t.Errorf("items length = %d, want 2", len(items))
		}
	})

	t.Run("NoAuth", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/match-links", "", "")
		if code != 401 {
			t.Errorf("status = %d, want 401", code)
		}
		if c := envelopeErr(t, body); c != "unauthorized" {
			t.Errorf("error code = %q, want unauthorized", c)
		}
	})
}

func TestMatchLinkDelete(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	creatorToken, _ := registerAndLogin(t, srv, "ml-del-creator", "secret1234")
	otherToken, _ := registerAndLogin(t, srv, "ml-del-other", "secret1234")

	b := postAuthBody(t, srv.URL+"/api/match-links", creatorToken, "")
	data := envelopeOk(t, b)
	linkID := fmt.Sprintf("%.0f", data["id"].(float64))
	linkToken := data["token"].(string)

	t.Run("NonCreatorDelete", func(t *testing.T) {
		code, body := deleteAuth(t, srv.URL+"/api/match-links/"+linkID, otherToken)
		if code != 404 {
			t.Errorf("status = %d, want 404", code)
		}
		if c := envelopeErr(t, body); c != "not_found" {
			t.Errorf("error code = %q, want not_found", c)
		}
	})

	t.Run("CreatorDelete", func(t *testing.T) {
		code, _ := deleteAuth(t, srv.URL+"/api/match-links/"+linkID, creatorToken)
		if code != 200 {
			t.Errorf("status = %d, want 200", code)
		}
	})

	t.Run("DeletedLinkGet", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/m/"+linkToken, "", "")
		if code != 404 {
			t.Errorf("status = %d, want 404", code)
		}
		if c := envelopeErr(t, body); c != "not_found" {
			t.Errorf("error code = %q, want not_found", c)
		}
	})
}

// ── Match Link Landing ──

func TestGetMatchLinkInfo(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	creatorToken, _ := registerAndLogin(t, srv, "ml-info-creator", "secret1234")
	b := postAuthBody(t, srv.URL+"/api/match-links", creatorToken, "")
	data := envelopeOk(t, b)
	linkToken := data["token"].(string)
	linkID := fmt.Sprintf("%.0f", data["id"].(float64))

	t.Run("Valid", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/m/"+linkToken, "", "")
		if code != 200 {
			t.Fatalf("status = %d, want 200", code)
		}
		data := envelopeOk(t, body)
		if data["token"] == nil || data["token"].(string) == "" {
			t.Error("token missing or empty")
		}
		if data["creator_name"] == nil || data["creator_name"].(string) == "" {
			t.Error("creator_name missing or empty (handler should look up user)")
		}
		if data["valid"] != true {
			t.Errorf("valid = %v, want true", data["valid"])
		}
	})

	t.Run("InvalidToken", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/m/nonexistent", "", "")
		if code != 404 {
			t.Errorf("status = %d, want 404", code)
		}
		if c := envelopeErr(t, body); c != "not_found" {
			t.Errorf("error code = %q, want not_found", c)
		}
	})

	t.Run("Deleted", func(t *testing.T) {
		deleteAuth(t, srv.URL+"/api/match-links/"+linkID, creatorToken)
		code, body := doReq(t, "GET", srv.URL+"/api/m/"+linkToken, "", "")
		if code != 404 {
			t.Errorf("status = %d, want 404", code)
		}
		if c := envelopeErr(t, body); c != "not_found" {
			t.Errorf("error code = %q, want not_found", c)
		}
	})
}

func TestSubmitMatchLinkAnswers(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	creatorToken, _ := registerAndLogin(t, srv, "ml-sub-creator", "secret1234")
	submitFullAssessment(t, srv, creatorToken)

	b := postAuthBody(t, srv.URL+"/api/match-links", creatorToken, "")
	data := envelopeOk(t, b)
	linkToken := data["token"].(string)

	t.Run("AnonymousSubmit", func(t *testing.T) {
		body := `{"answers":` + fullAnswersJSON() + `,"other_name":"Anonymous Friend"}`
		code, respBody := doReq(t, "POST", srv.URL+"/api/m/"+linkToken, body, "")
		if code != 201 {
			t.Fatalf("status = %d, want 201 (body: %s)", code, respBody)
		}
		data := envelopeOk(t, respBody)
		if data["profile"] == nil {
			t.Error("profile missing")
		}
		if data["bond"] == nil {
			t.Error("bond missing when creator has profile")
		}
		if data["assessment_id"] == nil {
			t.Error("assessment_id missing")
		}
	})

	t.Run("LoggedInSubmit", func(t *testing.T) {
		otherToken, _ := registerAndLogin(t, srv, "ml-sub-other", "secret1234")
		body := `{"answers":` + fullAnswersJSON() + `}`
		code, respBody := doReq(t, "POST", srv.URL+"/api/m/"+linkToken, body, otherToken)
		if code != 201 {
			t.Fatalf("status = %d, want 201 (body: %s)", code, respBody)
		}
		data := envelopeOk(t, respBody)
		if data["bond"] == nil {
			t.Error("bond missing for logged-in submit")
		}
		if data["assessment_id"] == nil {
			t.Error("assessment_id missing for logged-in submit")
		}
	})

	t.Run("EmptyAnswers", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/m/"+linkToken, `{"answers":[]}`, "")
		if code != 400 {
			t.Errorf("status = %d, want 400", code)
		}
		if c := envelopeErr(t, body); c != "invalid_request" {
			t.Errorf("error code = %q, want invalid_request", c)
		}
	})

	t.Run("InvalidToken", func(t *testing.T) {
		body := `{"answers":[{"qid":"Q01","selections":["W","F"]}]}`
		code, respBody := doReq(t, "POST", srv.URL+"/api/m/bad-token", body, "")
		if code != 404 {
			t.Errorf("status = %d, want 404", code)
		}
		if c := envelopeErr(t, respBody); c != "not_found" {
			t.Errorf("error code = %q, want not_found", c)
		}
	})
}

func TestSubmitMatchLinkUseExisting(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	// Creator with profile.
	creatorToken, creatorID := registerAndLogin(t, srv, "ml-use-creator", "secret1234")
	submitFullAssessment(t, srv, creatorToken)

	b := postAuthBody(t, srv.URL+"/api/match-links", creatorToken, "")
	data := envelopeOk(t, b)
	linkToken := data["token"].(string)

	t.Run("WithProfile", func(t *testing.T) {
		otherToken, _ := registerAndLogin(t, srv, "ml-use-other", "secret1234")
		submitFullAssessment(t, srv, otherToken)

		body := `{"use_existing":true}`
		code, respBody := doReq(t, "POST", srv.URL+"/api/m/"+linkToken, body, otherToken)
		if code != 201 {
			t.Fatalf("status = %d, want 201 (body: %s)", code, respBody)
		}
		data := envelopeOk(t, respBody)
		if data["bond"] == nil {
			t.Error("bond missing for use_existing")
		}
	})

	t.Run("NoProfile", func(t *testing.T) {
		noProfToken, _ := registerAndLogin(t, srv, "ml-use-noprofile", "secret1234")
		// Don't submit assessment.

		code, body := doReq(t, "POST", srv.URL+"/api/m/"+linkToken, `{"use_existing":true}`, noProfToken)
		if code != 404 {
			t.Errorf("status = %d, want 404", code)
		}
		if c := envelopeErr(t, body); c != "not_found" {
			t.Errorf("error code = %q, want not_found", c)
		}
	})

	t.Run("NoAuth", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/m/"+linkToken, `{"use_existing":true}`, "")
		if code != 401 {
			t.Errorf("status = %d, want 401", code)
		}
		if c := envelopeErr(t, body); c != "unauthorized" {
			t.Errorf("error code = %q, want unauthorized", c)
		}
	})

	_ = creatorID // used to suppress unused var warning
}

// =============================================================================
// Bond Notification (on match link submission)
// =============================================================================

func TestMatchLinkBondNotification(t *testing.T) {
	stub := &stubEmailSender{}
	srv := newTestServerWithEmail(t, stub)
	defer srv.Close()

	creatorToken, creatorID := registerAndLogin(t, srv, "bn-creator", "secret1234")
	submitFullAssessment(t, srv, creatorToken)
	patchAuthBody(t, srv.URL+"/api/users/me", creatorToken, `{"email":"creator@test.com"}`)
	_ = creatorID

	clearSent := func() { stub.sent = nil }

	t.Run("UseExisting", func(t *testing.T) {
		clearSent()
		b := postAuthBody(t, srv.URL+"/api/match-links", creatorToken, "")
		data := envelopeOk(t, b)
		linkToken := data["token"].(string)

		otherToken, _ := registerAndLogin(t, srv, "bn-useexist", "secret1234")
		submitFullAssessment(t, srv, otherToken)

		doReq(t, "POST", srv.URL+"/api/m/"+linkToken, `{"use_existing":true}`, otherToken)

		found := false
		for _, s := range stub.sent {
			if strings.HasPrefix(s, "bond:") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected bond notification, got %v", stub.sent)
		}
	})

	t.Run("AnswersLoggedIn", func(t *testing.T) {
		clearSent()
		b := postAuthBody(t, srv.URL+"/api/match-links", creatorToken, "")
		data := envelopeOk(t, b)
		linkToken := data["token"].(string)

		otherToken, _ := registerAndLogin(t, srv, "bn-loggedin", "secret1234")
		body := `{"answers":` + fullAnswersJSON() + `}`
		doReq(t, "POST", srv.URL+"/api/m/"+linkToken, body, otherToken)

		found := false
		for _, s := range stub.sent {
			if strings.HasPrefix(s, "bond:") && strings.Contains(s, "bn-loggedin") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected bond notification with username, got %v", stub.sent)
		}
	})

	t.Run("AnswersAnonymousWithName", func(t *testing.T) {
		clearSent()
		b := postAuthBody(t, srv.URL+"/api/match-links", creatorToken, "")
		data := envelopeOk(t, b)
		linkToken := data["token"].(string)

		body := `{"answers":` + fullAnswersJSON() + `,"other_name":"Alice"}`
		doReq(t, "POST", srv.URL+"/api/m/"+linkToken, body, "")

		found := false
		for _, s := range stub.sent {
			if strings.HasPrefix(s, "bond:") && strings.Contains(s, "Alice") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected bond notification with other_name=Alice, got %v", stub.sent)
		}
	})

	t.Run("AnswersAnonymousNoName", func(t *testing.T) {
		clearSent()
		b := postAuthBody(t, srv.URL+"/api/match-links", creatorToken, "")
		data := envelopeOk(t, b)
		linkToken := data["token"].(string)

		body := `{"answers":` + fullAnswersJSON() + `}`
		doReq(t, "POST", srv.URL+"/api/m/"+linkToken, body, "")

		found := false
		for _, s := range stub.sent {
			if strings.HasPrefix(s, "bond:") && strings.Contains(s, "Anonymous") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected bond notification with Anonymous, got %v", stub.sent)
		}
	})

	t.Run("NilSender", func(t *testing.T) {
		srv2 := newTestServer(t)
		defer srv2.Close()

		ct, _ := registerAndLogin(t, srv2, "bn-nil-creator", "secret1234")
		submitFullAssessment(t, srv2, ct)

		b := postAuthBody(t, srv2.URL+"/api/match-links", ct, "")
		data := envelopeOk(t, b)
		linkToken := data["token"].(string)

		body := `{"answers":` + fullAnswersJSON() + `,"other_name":"Test"}`
		code, _ := doReq(t, "POST", srv2.URL+"/api/m/"+linkToken, body, "")
		if code != 201 {
			t.Errorf("status = %d, want 201 (should not panic with nil EmailSender)", code)
		}
	})
}
