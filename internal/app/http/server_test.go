package http

import (
	"testing"
)

// Tests in this file are cross-cutting (health, route policy, paywall guards, frontend errors).
// Domain-specific endpoint tests live in handler-aligned files:
//   user_handler_test.go         — register, login, password, verification
//   assessment_handler_test.go   — submit, questions, list, peers, claim
//   review_handler_test.go       — review link CRUD, peer submit
//   flow_handler_test.go         — flow, flow yearly, solar terms
//   match_link_handler_test.go   — match link CRUD, submit, bond notification
//   profile_handler_test.go      — profile, compute bond, bonds, dedup
//   middleware_test.go            — auth middleware (20 cases)
//   response_test.go             — envelope formatting
//   payment_handler_test.go      — checkout, webhook
// Shared helpers are in helpers_test.go.

// ── Health ──

func TestHealth(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	code, body := getBody(t, srv.URL+"/api/health")
	if code != 200 {
		t.Fatalf("GET /api/health status = %d, want 200", code)
	}
	data := envelopeOk(t, body)
	if data["status"] != "ok" {
		t.Errorf("status = %v, want ok", data["status"])
	}
}

// ── No-paywalls guard ──

func TestNoPaywalls(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "np-user", "test-pass-123")
	submitFullAssessment(t, srv, token)

	token2, _ := registerAndLogin(t, srv, "np-other", "test-pass-123")
	submitFullAssessment(t, srv, token2)

	// Bond — should work without payment
	b := postAuthBody(t, srv.URL+"/api/bond", token, `{"with_name":"np-other"}`)
	bondData := envelopeOk(t, b)
	if bondData["self"] == nil {
		t.Error("bond response missing self")
	}
	if bondData["other"] == nil {
		t.Error("bond response missing other")
	}
	if bondData["concord"] == nil {
		t.Error("bond response missing concord")
	}

	// Flow yearly — should work without payment
	code, body := doReq(t, "GET", srv.URL+"/api/flow/yearly", "", token)
	if code != 200 {
		t.Fatalf("flow yearly status = %d, want 200", code)
	}
	flowData := envelopeOk(t, body)
	months := flowData["months"].([]interface{})
	if len(months) == 0 {
		t.Error("flow yearly has no months")
	}

	// Export — should work without payment
	code, body = doReq(t, "GET", srv.URL+"/api/users/me/export", "", token)
	if code != 200 {
		t.Fatalf("export status = %d, want 200", code)
	}
	exportData := envelopeOk(t, body)
	if exportData["user"] == nil {
		t.Error("export missing user field")
	}
}

// ── Deleted routes ──

func TestDeletedRoutes(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	routes := []struct{ method, path string }{
		{"GET", "/api/nonexistent-route"},
		{"POST", "/api/nonexistent-route"},
		{"GET", "/api/deleted-endpoint"},
		{"DELETE", "/api/deleted-endpoint"},
	}
	for _, r := range routes {
		code, _ := doReq(t, r.method, srv.URL+r.path, "", "")
		if code != 404 {
			t.Errorf("%s %s: %d, want 404", r.method, r.path, code)
		}
	}
}

// =============================================================================
// Frontend Error Collection
// =============================================================================

func TestCollectFrontendError_Normal(t *testing.T) {
	srv := newTestServer(t)

	body := `{"message":"test error","filename":"js/app.js","lineno":42,"colno":5,"stack":"Error: boom","url":"/en/assess"}`
	resp := postAuthBody(t, srv.URL+"/api/errors/frontend", "", body)
	data := envelopeOk(t, resp)
	if data["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", data["status"])
	}
}

func TestCollectFrontendError_Trimmed(t *testing.T) {
	srv := newTestServer(t)

	body := `{"message":"  padded  ","filename":"  /js/app.js  ","lineno":1,"colno":0,"stack":"","url":"/"}`
	resp := postAuthBody(t, srv.URL+"/api/errors/frontend", "", body)
	data := envelopeOk(t, resp)
	if data["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", data["status"])
	}
}

func TestCollectFrontendError_CapReached(t *testing.T) {
	srv := newTestServer(t)

	body := `{"message":"cap test","filename":"x.js","lineno":1,"colno":0,"stack":"","url":"/"}`
	for i := 0; i < 1000; i++ {
		resp := postAuthBody(t, srv.URL+"/api/errors/frontend", "", body)
		d := envelopeOk(t, resp)
		if d["status"] != "ok" {
			t.Fatalf("insert %d: expected 'ok', got %q", i, d["status"])
		}
	}

	resp := postAuthBody(t, srv.URL+"/api/errors/frontend", "", body)
	data := envelopeOk(t, resp)
	if data["status"] != "dropped" {
		t.Errorf("expected status 'dropped' at cap, got %q", data["status"])
	}
}

func TestCollectFrontendError_InvalidJSON(t *testing.T) {
	srv := newTestServer(t)

	resp := postAuthBody(t, srv.URL+"/api/errors/frontend", "", `not json`)
	data := envelopeOk(t, resp)
	if data["status"] != "dropped" {
		t.Errorf("expected status 'dropped' for invalid JSON, got %q", data["status"])
	}
}
