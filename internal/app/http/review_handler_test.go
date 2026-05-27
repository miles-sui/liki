package http

import (
	"fmt"
	"testing"
)

func TestReviewCreateLink(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "reviewer-a", "secret1234")
	submitFullAssessment(t, srv, token)

	t.Run("OK", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/reviews", "", token)
		if code != 201 {
			t.Fatalf("status = %d, want 201", code)
		}
		data := envelopeOk(t, body)
		if data["token"] == nil || data["token"].(string) == "" {
			t.Error("review link token missing")
		}
		if data["url"] == nil || data["url"].(string) == "" {
			t.Error("review link url missing")
		}
		if data["expires_at"] == nil {
			t.Error("review link expires_at missing")
		}
	})
}

func TestReviewCreateLinkNoAuth(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	code, body := doReq(t, "POST", srv.URL+"/api/reviews", "", "")
	if code != 401 {
		t.Errorf("status = %d, want 401", code)
	}
	if c := envelopeErr(t, body); c != "unauthorized" {
		t.Errorf("error code = %q, want unauthorized", c)
	}
}

func TestReviewListGetDelete(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "review-crud", "secret1234")
	submitFullAssessment(t, srv, token)

	// Create
	crBody := postAuthBody(t, srv.URL+"/api/reviews", token, "")
	rData := envelopeOk(t, crBody)
	linkID := fmt.Sprintf("%.0f", rData["id"].(float64))

	t.Run("List", func(t *testing.T) {
		b := getAuthBody(t, srv.URL+"/api/reviews", token)
		data := envelopeOk(t, b)
		if data["items"] == nil {
			t.Error("review list missing items")
		}
	})

	t.Run("Get", func(t *testing.T) {
		b := getAuthBody(t, srv.URL+"/api/reviews/"+linkID, token)
		data := envelopeOk(t, b)
		if data["id"] == nil {
			t.Error("review detail missing id")
		}
		if data["token"] == nil {
			t.Error("review detail missing token")
		}
		if data["url"] == nil {
			t.Error("review detail missing url")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		code, _ := deleteAuth(t, srv.URL+"/api/reviews/"+linkID, token)
		if code != 200 {
			t.Errorf("status = %d, want 200", code)
		}
	})

	t.Run("DeleteNotFound", func(t *testing.T) {
		code, body := deleteAuth(t, srv.URL+"/api/reviews/99999", token)
		if code != 404 {
			t.Errorf("status = %d, want 404", code)
		}
		if c := envelopeErr(t, body); c != "not_found" {
			t.Errorf("error code = %q, want not_found", c)
		}
	})
}

func TestReviewRenew(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "renew-test", "secret1234")
	submitFullAssessment(t, srv, token)

	crBody := postAuthBody(t, srv.URL+"/api/reviews", token, "")
	rData := envelopeOk(t, crBody)
	linkID := fmt.Sprintf("%.0f", rData["id"].(float64))

	t.Run("OK", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/reviews/"+linkID+"/renew", "", token)
		if code != 200 {
			t.Errorf("status = %d, want 200", code)
		}
		data := envelopeOk(t, body)
		if data["id"] == nil {
			t.Error("renew response missing id")
		}
		if data["token"] == nil {
			t.Error("renew response missing token")
		}
		if data["expires_at"] == nil {
			t.Error("renew response missing expires_at")
		}
	})
}

func TestReviewLinkInfo(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "linkinfo-subject", "secret1234")
	submitFullAssessment(t, srv, token)

	crBody := postAuthBody(t, srv.URL+"/api/reviews", token, "")
	rData := envelopeOk(t, crBody)
	reviewToken := rData["token"].(string)
	linkID := fmt.Sprintf("%.0f", rData["id"].(float64))

	t.Run("OK", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/r/"+reviewToken, "", "")
		if code != 200 {
			t.Errorf("status = %d, want 200", code)
		}
		data := envelopeOk(t, body)
		if data["subject_name"] == nil {
			t.Error("link info missing subject_name")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/r/nonexistent-token", "", "")
		if code != 404 {
			t.Errorf("status = %d, want 404", code)
		}
		if c := envelopeErr(t, body); c != "not_found" {
			t.Errorf("error code = %q, want not_found", c)
		}
	})

	t.Run("Deleted", func(t *testing.T) {
		deleteAuth(t, srv.URL+"/api/reviews/"+linkID, token)
		code, body := doReq(t, "GET", srv.URL+"/api/r/"+reviewToken, "", "")
		if code != 404 {
			t.Errorf("status = %d, want 404 (deleted link)", code)
		}
		if c := envelopeErr(t, body); c != "not_found" {
			t.Errorf("error code = %q, want not_found", c)
		}
	})
}

func TestReviewSubmitPeer(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "peer-subject-2", "secret1234")
	submitFullAssessment(t, srv, token)

	crBody := postAuthBody(t, srv.URL+"/api/reviews", token, "")
	rData := envelopeOk(t, crBody)
	reviewToken := rData["token"].(string)

	peerAnswers := `{"reviewer_name":"Peer Reviewer","answers":[
		{"qid":"Q01","selections":["M","R"]},
		{"qid":"Q02","selections":["R","W"]},
		{"qid":"Q03","selections":["W","F"]},
		{"qid":"Q04","selections":["F","E"]},
		{"qid":"Q05","selections":["E","M"]}
	]}`

	t.Run("OK", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/r/"+reviewToken, peerAnswers, "")
		if code != 201 {
			t.Fatalf("status = %d, want 201 (body: %s)", code, body)
		}
		data := envelopeOk(t, body)
		if data["subject_identity"] == nil {
			t.Error("peer review response missing subject_identity")
		}
	})

	t.Run("NoReviewerName", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/r/"+reviewToken,
			`{"answers":[{"qid":"Q01","selections":["W","F"]}]}`, "")
		if code != 400 {
			t.Errorf("status = %d, want 400", code)
		}
		if c := envelopeErr(t, body); c != "invalid_request" {
			t.Errorf("error code = %q, want invalid_request", c)
		}
	})

	t.Run("NoAnswers", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/r/"+reviewToken,
			`{"reviewer_name":"x"}`, "")
		if code != 400 {
			t.Errorf("status = %d, want 400", code)
		}
		if c := envelopeErr(t, body); c != "invalid_request" {
			t.Errorf("error code = %q, want invalid_request", c)
		}
	})
}

func TestReviewsGiven(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	t.Run("Auth", func(t *testing.T) {
		token, _ := registerAndLogin(t, srv, "rg-auth", "secret1234")
		b := getAuthBody(t, srv.URL+"/api/reviews/given", token)
		data := envelopeOk(t, b)
		if data["items"] == nil {
			t.Error("reviews given missing items")
		}
		if data["total"] == nil {
			t.Error("reviews given missing total")
		}
	})

	t.Run("AnonToken", func(t *testing.T) {
		code, _ := doReq(t, "GET", srv.URL+"/api/reviews/given?anonymous_token=peer-xyz", "", "")
		if code != 200 {
			t.Errorf("status = %d, want 200", code)
		}
	})
}
