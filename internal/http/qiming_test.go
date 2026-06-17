package handler

import (
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
	body := `{"surname":"张","combos":[{"stroke1":5,"stroke2":8}],"yong_chars":{"5":["铭","钧"],"8":["坤","坪"]},"xi_chars":{"5":[],"8":[]}}`
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
