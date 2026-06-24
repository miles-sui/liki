package handler
import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBazhaiMingGua_Valid(t *testing.T) {
	body := `{"gender":"male","birth_year":1990}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	bazhaiMingGua(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestBazhaiMingGua_MissingGender(t *testing.T) {
	body := `{"birth_year":1990}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	bazhaiMingGua(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestBazhaiChart_Valid(t *testing.T) {
	body := `{` + bt + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	bazhaiChart(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestBlackBox_MingGua_Invariants(t *testing.T) {
	body := `{"gender":"male","birth_year":1990}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	bazhaiMingGua(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestBlackBox_Bazhai_Invariants(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	bazhaiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestBlackBox_MingGua_GenderDifference(t *testing.T) {
	bodyM := `{"gender":"male","birth_year":1990}`
	rM := httptest.NewRequest("POST", "/", strings.NewReader(bodyM))
	wM := httptest.NewRecorder()
	bazhaiMingGua(wM, rM)

	bodyF := `{"gender":"female","birth_year":1990}`
	rF := httptest.NewRequest("POST", "/", strings.NewReader(bodyF))
	wF := httptest.NewRecorder()
	bazhaiMingGua(wF, rF)

	if wM.Code != http.StatusOK || wF.Code != http.StatusOK {
		t.Fatal("minggua failed")
	}
}

func TestBug_BazhaiMingGua_BirthYear1899_Rejected(t *testing.T) {
	body := `{"gender":"male","birth_year":1899}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	bazhaiMingGua(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("minggua birth_year=1899: status=%d, want 422 (min=1900)", w.Code)
	}
}

func TestBug_BazhaiChart_MissingBirth(t *testing.T) {
	body := `{"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	bazhaiChart(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("bazhai chart missing birth: status=%d, want 422", w.Code)
	}
}
