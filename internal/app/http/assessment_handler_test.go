package http

import (
	"fmt"
	"math"
	"testing"
)

func TestSubmitAssessment(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "assess-user", "secret1234")

	t.Run("Round1", func(t *testing.T) {
		b := postAuthBody(t, srv.URL+"/api/assessments", token,
			`{"answers":[{"qid":"Q01","selections":["W","F"]},{"qid":"Q02","selections":["W","M"]}]}`)
		data := envelopeOk(t, b)
		if data["profile"] == nil {
			t.Error("profile missing in round 1 response")
		}
		if data["complete"] != false {
			t.Errorf("complete = %v, want false after round 1", data["complete"])
		}
	})

	t.Run("All30", func(t *testing.T) {
		b := postAuthBody(t, srv.URL+"/api/assessments", token,
			`{"answers":`+fullAnswersJSON()+`}`)
		data := envelopeOk(t, b)
		if data["complete"] != true {
			t.Errorf("complete = %v, want true after 30 answers", data["complete"])
		}
		profile := data["profile"].(map[string]interface{})
		d := profile["d"].(map[string]interface{})
		assertDeviationSumZero(t, d, "profile.d")
		// p sum ≈ 1
		p := profile["p"].(map[string]interface{})
		pSum := 0.0
		for _, v := range p {
			pSum += v.(float64)
		}
		if math.Abs(pSum-1.0) > 1e-10 {
			t.Errorf("p Σ = %v, want 1.0", pSum)
		}
	})

	t.Run("NoAnswers", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/assessments", `{}`, token)
		if code != 400 {
			t.Errorf("status = %d, want 400", code)
		}
		if c := envelopeErr(t, body); c != "invalid_request" {
			t.Errorf("error code = %q, want invalid_request", c)
		}
	})

	t.Run("Anonymous", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/assessments",
			`{"answers":[{"qid":"Q01","selections":["W","F"]}],"anonymous_token":"anon-1"}`, "")
		if code != 201 {
			t.Errorf("status = %d, want 201", code)
		}
		data := envelopeOk(t, body)
		if data["anonymous_token"] != "anon-1" {
			t.Errorf("anonymous_token = %v, want anon-1", data["anonymous_token"])
		}
	})
}

func TestGetQuestions(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	t.Run("en", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/assessments/questions?locale=en", "", "")
		if code != 200 {
			t.Errorf("status = %d, want 200", code)
		}
		data := envelopeOk(t, body)
		rounds := data["rounds"].([]interface{})
		var total int
		for _, r := range rounds {
			rm := r.(map[string]interface{})
			qs := rm["questions"].([]interface{})
			total += len(qs)
		}
		if total != 30 {
			t.Errorf("got %d questions, want 30", total)
		}
	})

	t.Run("zh", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/assessments/questions?locale=zh-CN", "", "")
		if code != 200 {
			t.Errorf("status = %d, want 200", code)
		}
		envelopeOk(t, body)
	})
}

func TestListAssessments(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "list-assess", "secret1234")

	t.Run("OK", func(t *testing.T) {
		b := getAuthBody(t, srv.URL+"/api/assessments", token)
		data := envelopeOk(t, b)
		// items is null when list is empty (Go nil slice marshals to null)
		if _, ok := data["items"]; !ok {
			t.Error("assessment list missing items key")
		}
		if data["total"] == nil {
			t.Error("assessment list missing total")
		}
	})

	t.Run("NoAuth", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/assessments", "", "")
		if code != 401 {
			t.Errorf("status = %d, want 401", code)
		}
		if c := envelopeErr(t, body); c != "unauthorized" {
			t.Errorf("error code = %q, want unauthorized", c)
		}
	})
}

func TestGetAssessment(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "get-assess", "secret1234")
	assData := submitFullAssessment(t, srv, token)
	aid := fmt.Sprintf("%.0f", assData["id"].(float64))

	t.Run("OK", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/assessments/"+aid, "", "")
		if code != 200 {
			t.Fatalf("status = %d, want 200", code)
		}
		data := envelopeOk(t, body)
		if fmt.Sprintf("%.0f", data["id"].(float64)) != aid {
			t.Errorf("id = %v, want %s", data["id"], aid)
		}
		profile := data["profile"].(map[string]interface{})
		d := profile["d"].(map[string]interface{})
		assertDeviationSumZero(t, d, "profile.d")
	})

	t.Run("NotFound", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/assessments/99999", "", "")
		if code != 404 {
			t.Errorf("status = %d, want 404", code)
		}
		if c := envelopeErr(t, body); c != "not_found" {
			t.Errorf("error code = %q, want not_found", c)
		}
	})
}

func TestPeers(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "peer-subject", "secret1234")
	submitFullAssessment(t, srv, token)

	t.Run("NoPeers", func(t *testing.T) {
		b := getAuthBody(t, srv.URL+"/api/assessments/peers", token)
		data := envelopeOk(t, b)
		if data["self"] == nil {
			t.Error("peers response missing self")
		}
		pc, _ := data["peer_count"].(float64)
		if pc != 0 {
			t.Errorf("peer_count = %v, want 0", pc)
		}
	})
}

func TestClaim(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	// Submit anonymous assessment
	postBody(t, srv.URL+"/api/assessments",
		`{"answers":[{"qid":"Q01","selections":["W","F"]}],"anonymous_token":"claim-tok"}`)
	token, _ := registerAndLogin(t, srv, "claimer", "secret1234")

	t.Run("OK", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/assessments/claim",
			`{"anonymous_token":"claim-tok"}`, token)
		if code != 200 {
			t.Fatalf("claim status = %d, want 200", code)
		}
		data := envelopeOk(t, body)
		claimed := data["claimed"].(float64)
		if claimed <= 0 {
			t.Errorf("claimed = %.0f, want > 0", claimed)
		}
		// Verify the assessment now appears in the user's list
		listBody := getAuthBody(t, srv.URL+"/api/assessments", token)
		listData := envelopeOk(t, listBody)
		items := listData["items"].([]interface{})
		if len(items) == 0 {
			t.Error("assessment list empty after claim")
		}
	})

	t.Run("NonexistentToken", func(t *testing.T) {
		code, body := doReq(t, "POST", srv.URL+"/api/assessments/claim",
			`{"anonymous_token":"nonexistent"}`, token)
		if code != 200 {
			t.Errorf("claim nonexistent token status = %d, want 200", code)
		}
		data := envelopeOk(t, body)
		if v, _ := data["claimed"].(float64); v != 0 {
			t.Errorf("claimed = %.0f for nonexistent token, want 0", v)
		}
	})
}
