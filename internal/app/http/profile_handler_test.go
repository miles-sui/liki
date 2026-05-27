package http

import (
	"fmt"
	"testing"
)

func TestGetProfile(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	tokenA, _ := registerAndLogin(t, srv, "prof-pub", "secret1234")
	submitFullAssessment(t, srv, tokenA)

	// Make public.
	patchAuthBody(t, srv.URL+"/api/users/me", tokenA, `{"is_public":true}`)

	t.Run("Public", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/profiles/prof-pub", "", "")
		if code != 200 {
			t.Fatalf("status = %d, want 200", code)
		}
		data := envelopeOk(t, body)
		if data["user"] == nil {
			t.Error("user missing")
		}
		if data["profile"] == nil {
			t.Error("profile missing (user completed assessment)")
		}
		if data["flow_month"] == nil {
			t.Error("flow_month missing")
		}
		if data["is_public"] != true {
			t.Errorf("is_public = %v, want true for public profile", data["is_public"])
		}
		if data["is_owner"] != false {
			t.Errorf("is_owner = %v, want false for anonymous viewer", data["is_owner"])
		}
		// peers uses omitempty — may be absent when no peer reviews exist
		if _, ok := data["peers"]; !ok {
			t.Log("peers field absent (expected when no peer reviews)")
		}
	})

	t.Run("PrivateOther", func(t *testing.T) {
		// Set private.
		patchAuthBody(t, srv.URL+"/api/users/me", tokenA, `{"is_public":false}`)

		otherToken, _ := registerAndLogin(t, srv, "prof-other", "secret1234")
		code, body := doReq(t, "GET", srv.URL+"/api/profiles/prof-pub", "", otherToken)
		if code != 404 {
			t.Errorf("status = %d, want 404", code)
		}
		if c := envelopeErr(t, body); c != "not_found" {
			t.Errorf("error code = %q, want not_found", c)
		}
	})

	t.Run("PrivateSelf", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/profiles/prof-pub", "", tokenA)
		if code != 200 {
			t.Fatalf("status = %d, want 200", code)
		}
		data := envelopeOk(t, body)
		if data["is_owner"] != true {
			t.Errorf("is_owner = %v, want true for self", data["is_owner"])
		}
		if data["user"] == nil {
			t.Error("user missing (should see own private profile)")
		}
		if data["profile"] == nil {
			t.Error("profile missing (should see own assessment)")
		}
	})

	t.Run("NonExistent", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/profiles/no-such-user", "", "")
		if code != 404 {
			t.Errorf("status = %d, want 404", code)
		}
		if c := envelopeErr(t, body); c != "not_found" {
			t.Errorf("error code = %q, want not_found", c)
		}
	})
}

func TestComputeBond(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	tokenA, _ := registerAndLogin(t, srv, "bond-comp-a", "secret1234")
	submitFullAssessment(t, srv, tokenA)

	tokenB, idB := registerAndLogin(t, srv, "bond-comp-b", "secret1234")
	submitFullAssessment(t, srv, tokenB)

	t.Run("WithUserID", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/bond",
			fmt.Sprintf(`{"with_user_id":%.0f}`, idB), tokenA)
		if code != 200 {
			t.Fatalf("status = %d, want 200 (body: %s)", code, body)
		}
		data := envelopeOk(t, body)
		if data["self"] == nil {
			t.Error("self missing in bond response")
		}
		if data["other"] == nil {
			t.Error("other missing in bond response")
		}
		if data["delta_a"] == nil {
			t.Error("delta_a missing in bond response")
		}
		if data["delta_b"] == nil {
			t.Error("delta_b missing in bond response")
		}
	})

	t.Run("WithName", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/bond",
			`{"with_name":"bond-comp-b"}`, tokenA)
		if code != 200 {
			t.Fatalf("status = %d, want 200 (body: %s)", code, body)
		}
		envelopeOk(t, body)
	})

	t.Run("Self", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/bond",
			fmt.Sprintf(`{"with_user_id":%.0f}`, idB), tokenB) // B vs B = self
		_ = tokenA // suppress unused
		if code != 400 {
			t.Errorf("status = %d, want 400", code)
		}
		if c := envelopeErr(t, body); c != "invalid_request" {
			t.Errorf("error code = %q, want invalid_request", c)
		}
	})

	t.Run("NoProfile", func(t *testing.T) {
		noProfToken, _ := registerAndLogin(t, srv, "bond-noprofile", "secret1234")
		code, body := doReq(t, "POST", srv.URL+"/api/bond",
			fmt.Sprintf(`{"with_user_id":%.0f}`, idB), noProfToken)
		if code != 404 {
			t.Errorf("status = %d, want 404", code)
		}
		if c := envelopeErr(t, body); c != "not_found" {
			t.Errorf("error code = %q, want not_found", c)
		}
	})
}

func TestGetBonds(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	tokenA, _ := registerAndLogin(t, srv, "bonds-a", "secret1234")
	submitFullAssessment(t, srv, tokenA)

	tokenB, idB := registerAndLogin(t, srv, "bonds-b", "secret1234")
	submitFullAssessment(t, srv, tokenB)

	// Create bond: A compares with B.
	postAuthBody(t, srv.URL+"/api/bond", tokenA,
		fmt.Sprintf(`{"with_user_id":%.0f}`, idB))

	t.Run("HasBond", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/profiles/bonds-a/bonds", "", tokenA)
		if code != 200 {
			t.Fatalf("status = %d, want 200", code)
		}
		data := envelopeOk(t, body)
		items := data["items"].([]interface{})
		if len(items) == 0 {
			t.Fatal("expected at least 1 bond item")
		}
		item := items[0].(map[string]interface{})
		if item["bond"] == nil {
			t.Error("bond missing in item")
		}
		if item["other_name"] == nil {
			t.Error("other_name missing in item")
		}
		if item["source"] == nil {
			t.Error("source missing in item")
		}
		if item["source"] != "instant" {
			t.Errorf("source = %q, want instant", item["source"])
		}
	})

	t.Run("PerspectiveSwap", func(t *testing.T) {
		// B should see bonds with self= B's profile.
		code, body := doReq(t, "GET", srv.URL+"/api/profiles/bonds-b/bonds", "", tokenB)
		if code != 200 {
			t.Fatalf("status = %d, want 200", code)
		}
		data := envelopeOk(t, body)
		items := data["items"].([]interface{})
		if len(items) == 0 {
			t.Fatal("B should see the bond A created")
		}
		item := items[0].(map[string]interface{})
		if item["other_name"].(string) != "bonds-a" {
			t.Errorf("B's other_name = %q, want bonds-a", item["other_name"])
		}
	})

	t.Run("OtherBonds", func(t *testing.T) {
		otherToken, _ := registerAndLogin(t, srv, "bonds-intruder", "secret1234")
		code, body := doReq(t, "GET", srv.URL+"/api/profiles/bonds-a/bonds", "", otherToken)
		if code != 404 {
			t.Errorf("status = %d, want 404", code)
		}
		if c := envelopeErr(t, body); c != "not_found" {
			t.Errorf("error code = %q, want not_found", c)
		}
	})

	t.Run("NoAuth", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/profiles/bonds-a/bonds", "", "")
		if code != 401 {
			t.Errorf("status = %d, want 401", code)
		}
		if c := envelopeErr(t, body); c != "unauthorized" {
			t.Errorf("error code = %q, want unauthorized", c)
		}
	})
}
func TestBondDedup(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	// Create two users with profiles.
	tokenA, _ := registerAndLogin(t, srv, "dedup-a", "secret1234")
	submitFullAssessment(t, srv, tokenA)

	tokenB, idB := registerAndLogin(t, srv, "dedup-b", "secret1234")
	submitFullAssessment(t, srv, tokenB)

	// Create first bond.
	postAuthBody(t, srv.URL+"/api/bond", tokenA,
		fmt.Sprintf(`{"with_user_id":%.0f}`, idB))

	// Create second bond (same pair).
	postAuthBody(t, srv.URL+"/api/bond", tokenA,
		fmt.Sprintf(`{"with_user_id":%.0f}`, idB))

	// Check: ListBondEvents should return only 1 record (the latest).
	b := getAuthBody(t, srv.URL+"/api/profiles/dedup-a/bonds", tokenA)
	data := envelopeOk(t, b)
	items := data["items"].([]interface{})
	if len(items) != 1 {
		t.Fatalf("expected 1 bond item after dedup, got %d", len(items))
	}
}

// =============================================================================
// Frontend Error Collection
// =============================================================================

