package http

import (
	"testing"
)

func TestGetFlow(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "flow-user", "secret1234")
	submitFullAssessment(t, srv, token)

	t.Run("OK", func(t *testing.T) {
		b := getAuthBody(t, srv.URL+"/api/flow", token)
		data := envelopeOk(t, b)
		if data["month_id"] == nil || data["month_id"].(string) == "" {
			t.Error("flow response missing or empty month_id")
		}
		if data["month_en"] == nil || data["month_en"].(string) == "" {
			t.Error("flow response missing or empty month_en")
		}
		if g, ok := data["generates"].(float64); !ok || (int(g) < 0 || int(g) > 4) {
			t.Error("flow response missing or invalid generates")
		}
		if r, ok := data["restrains"].(float64); !ok || (int(r) < 0 || int(r) > 4) {
			t.Error("flow response missing or invalid restrains")
		}
	})

	t.Run("NoAuth", func(t *testing.T) {
		code, body := doReq(t, "GET", srv.URL+"/api/flow", "", "")
		if code != 401 {
			t.Errorf("status = %d, want 401", code)
		}
		if c := envelopeErr(t, body); c != "unauthorized" {
			t.Errorf("error code = %q, want unauthorized", c)
		}
	})
}

func TestGetFlowYearly(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "flow-yearly", "secret1234")
	submitFullAssessment(t, srv, token)

	b := getAuthBody(t, srv.URL+"/api/flow/yearly", token)
	data := envelopeOk(t, b)
	months := data["months"].([]interface{})
	if len(months) != 12 {
		t.Errorf("flow yearly months = %d, want 12", len(months))
	}
	if data["current"] == nil || data["current"].(string) == "" {
		t.Error("flow yearly missing or empty current")
	}
	for i, m := range months {
		mo := m.(map[string]interface{})
		if mo["month_id"] == nil || mo["month_id"].(string) == "" {
			t.Errorf("month[%d] missing month_id", i)
		}
		if mo["month_en"] == nil || mo["month_en"].(string) == "" {
			t.Errorf("month[%d] missing month_en", i)
		}
		if mo["generates"] == nil {
			t.Errorf("month[%d] missing generates", i)
		}
		if mo["restrains"] == nil {
			t.Errorf("month[%d] missing restrains", i)
		}
	}
}

func TestGetSolarTerms(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	code, body := doReq(t, "GET", srv.URL+"/api/solar-terms", "", "")
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
	data := envelopeOk(t, body)
	months := data["months"].([]interface{})
	if len(months) != 12 {
		t.Errorf("solar terms months = %d, want 12", len(months))
	}
	if data["year"] == nil {
		t.Error("solar terms missing year")
	}
	if data["current"] == nil {
		t.Error("solar terms missing current")
	}
	for i, m := range months {
		mo := m.(map[string]interface{})
		if mo["id"] == nil || mo["id"].(string) == "" {
			t.Errorf("month[%d] missing id", i)
		}
		if mo["name_en"] == nil || mo["name_en"].(string) == "" {
			t.Errorf("month[%d] missing name_en", i)
		}
		if mo["start"] == nil || mo["start"].(string) == "" {
			t.Errorf("month[%d] missing start", i)
		}
		if mo["end"] == nil || mo["end"].(string) == "" {
			t.Errorf("month[%d] missing end", i)
		}
	}
}
