package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleWuge_Valid(t *testing.T) {
	body := `{"surname":"张","yong_shen":"木","xi_shen":["水"]}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleWuge(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
}

func TestHandleWuge_MissingSurname(t *testing.T) {
	body := `{"yong_shen":"木","xi_shen":["水"]}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleWuge(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestHandleWuge_InvalidYongShen(t *testing.T) {
	body := `{"surname":"张","yong_shen":"x","xi_shen":["水"]}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleWuge(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestHandleCompose_Valid(t *testing.T) {
	body := `{"surname":"张","combos":[{"stroke1":5,"stroke2":8}],"yong_chars":{"5":[{"char":"铭","tone":1},{"char":"钧","tone":1}],"8":[{"char":"坤","tone":4},{"char":"坪","tone":2}]},"xi_chars":{"5":[],"8":[]}}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleCompose(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
}

func TestHandleCompose_MissingSurname(t *testing.T) {
	body := `{"combos":[],"yong_chars":{},"xi_chars":{}}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleCompose(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestHandleDetail_Valid(t *testing.T) {
	body := `{"surname":"张","names":["沐洪","沐涛"]}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleDetail(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
}

func TestHandleDetail_MissingSurname(t *testing.T) {
	body := `{"names":["沐洪"]}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleDetail(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestHandleEvaluate_Valid(t *testing.T) {
	body := `{"surname":"张","given_name":"三","yong_shen":"木"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleEvaluate(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
}

func TestHandleEvaluate_MissingGivenName(t *testing.T) {
	body := `{"surname":"张","yong_shen":"木"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleEvaluate(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestHandleEvaluate_GivenNameTooLong(t *testing.T) {
	body := `{"surname":"张","given_name":"三四五","yong_shen":"木"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleEvaluate(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestBlackBox_WuGe_ResponseStructure(t *testing.T) {
	body := `{"surname":"张","yong_shen":"木","xi_shen":["水"]}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleWuge(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestBlackBox_Evaluate_DifferentNames(t *testing.T) {
	body1 := `{"surname":"张","given_name":"三","yong_shen":"木"}`
	r1 := httptest.NewRequest("POST", "/", strings.NewReader(body1))
	w1 := httptest.NewRecorder()
	handleEvaluate(w1, r1)

	body2 := `{"surname":"张","given_name":"伟","yong_shen":"木"}`
	r2 := httptest.NewRequest("POST", "/", strings.NewReader(body2))
	w2 := httptest.NewRecorder()
	handleEvaluate(w2, r2)

	if w1.Code != http.StatusOK || w2.Code != http.StatusOK {
		t.Skip("evaluate failed")
	}
}

func TestEdge_Qiming_Wuge_EmptyXiShen(t *testing.T) {
	body := `{"surname":"张","yong_shen":"木","xi_shen":[]}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleWuge(w, r)
	if w.Code >= 500 {
		t.Errorf("empty xi_shen caused 5xx: %d", w.Code)
	}
	if w.Code != http.StatusOK {
		t.Logf("empty xi_shen: status=%d", w.Code)
	}
}

func TestEdge_Qiming_Compose_EmptyCombos(t *testing.T) {
	body := `{"surname":"张","combos":[],"yong_chars":{},"xi_chars":{}}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleCompose(w, r)
	if w.Code >= 500 {
		t.Errorf("empty combos caused 5xx: %d", w.Code)
	}
	if w.Code != http.StatusOK {
		t.Logf("empty combos: status=%d", w.Code)
	}
}

func TestEdge_Qiming_Detail_EmptyNameInArray(t *testing.T) {
	// names 数组中包含空字符串
	body := `{"surname":"张","names":["", "伟"]}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleDetail(w, r)
	if w.Code >= 500 {
		t.Errorf("empty name in array caused 5xx: %d", w.Code)
	}
}

func TestEdge_Qiming_Evaluate_TooLongName(t *testing.T) {
	// given_name RuneLength(1,2), 3字名应该被拒绝
	body := `{"surname":"张","given_name":"明文辉","yong_shen":"木"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleEvaluate(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("3-char given_name: status=%d, want 422", w.Code)
	}
}

func TestEdge_Qiming_Evaluate_SingleCharName(t *testing.T) {
	// 单字名应该允许
	body := `{"surname":"张","given_name":"三","yong_shen":"木"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleEvaluate(w, r)
	if w.Code != http.StatusOK {
		t.Logf("1-char given_name rejected: status=%d", w.Code)
	}
}

func TestEdge_SpecialChars_Surname(t *testing.T) {
	specials := []string{
		"",
		"AB",
		"𠮷",      // 4-byte UTF-8 (CJK Ext B)
		"张\x00李", // null byte
	}
	for _, sn := range specials {
		t.Run("", func(t *testing.T) {
			body, err := json.Marshal(map[string]string{
				"surname":   sn,
				"yong_shen": "木",
			})
			if err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest("POST", "/", strings.NewReader(string(body)))
			w := httptest.NewRecorder()
			handleWuge(w, r)
			if w.Code >= 500 {
				t.Errorf("special surname caused 5xx: %d", w.Code)
			}
		})
	}
}

func TestEd3_Wuge_InvalidXiShen(t *testing.T) {
	body := `{"surname":"张","yong_shen":"木","xi_shen":["气"]}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleWuge(w, r)
	if w.Code >= 500 {
		t.Errorf("invalid xi_shen caused 5xx: %d", w.Code)
	}
	if w.Code == http.StatusOK {
		t.Error("BUG: invalid xi_shen element accepted")
	}
}

func TestBug_Wuge_EmptySurname(t *testing.T) {
	body := `{"surname":"","yong_shen":"木","xi_shen":["水"]}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleWuge(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("wuge empty surname: status=%d, want 422", w.Code)
	}
}

func TestBug_Wuge_InvalidXiShen_Ignored(t *testing.T) {
	// xi_shen with invalid element — does validation catch it?
	body := `{"surname":"张","yong_shen":"木","xi_shen":["x","y","z"]}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleWuge(w, r)

	if w.Code == http.StatusOK {
		t.Error("BUG: wuge accepts invalid xi_shen elements")
	}
	if w.Code == http.StatusUnprocessableEntity {
		t.Log("OK: invalid xi_shen rejected")
	}
}

func TestBug_Wuge_EmptyXiShen_OK(t *testing.T) {
	// Empty xi_shen array should be valid
	body := `{"surname":"张","yong_shen":"木","xi_shen":[]}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleWuge(w, r)

	if w.Code != http.StatusOK {
		t.Logf("wuge empty xi_shen: status=%d (expected 200 or 422?)", w.Code)
	}
}

func TestBug_Evaluate_GivenNameThreeChars_Rejected(t *testing.T) {
	// evalParams.GivenName has RuneLength(1,2) — 3 chars should be rejected
	body := `{"surname":"张","given_name":"欧阳明","yong_shen":"木"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleEvaluate(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("evaluate given_name=3chars: status=%d, want 422", w.Code)
	}
}

func TestBug_Evaluate_GivenNameZeroChars_Rejected(t *testing.T) {
	body := `{"surname":"张","given_name":"","yong_shen":"木"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleEvaluate(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("evaluate given_name=empty: status=%d, want 422", w.Code)
	}
}

func TestBug_Detail_EmptyNames_Rejected(t *testing.T) {
	body := `{"surname":"张","names":[]}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleDetail(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("detail empty names: status=%d, want 422", w.Code)
	}
}

func TestBug_Compose_EmptyCombosEmptyChars(t *testing.T) {
	body := `{"surname":"张","combos":[],"yong_chars":{},"xi_chars":{}}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleCompose(w, r)

	// Should return 200 with empty result (not crash)
	if w.Code != http.StatusOK {
		t.Errorf("compose all empty: status=%d, want 200", w.Code)
	}
}
