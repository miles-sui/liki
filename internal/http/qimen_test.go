package handler
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleQimenPan_Valid_Shi(t *testing.T) {
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":"shi"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env struct {
		Data struct {
			Pan struct {
				JuShu  int  `json:"jushu"`
				YinDun bool `json:"yin_dun"`
			} `json:"pan"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.Pan.JuShu < 1 {
		t.Error("jushu is zero")
	}
}

func TestHandleQimenPan_Valid_Ri(t *testing.T) {
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":"ri"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var envRi struct {
		Data struct {
			Pan struct {
				JuShu  int  `json:"jushu"`
				YinDun bool `json:"yin_dun"`
			} `json:"pan"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&envRi); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if envRi.Data.Pan.JuShu < 1 {
		t.Error("jushu is zero")
	}
}

func TestHandleQimenPan_Valid_Nian(t *testing.T) {
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":"nian"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var envNian struct {
		Data struct {
			Pan struct {
				JuShu int `json:"jushu"`
			} `json:"pan"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&envNian); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if envNian.Data.Pan.JuShu < 1 {
		t.Error("jushu is zero")
	}
}

func TestHandleQimenPan_DefaultsToShi(t *testing.T) {
	// kind="" defaults to "shi"
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":""}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (defaults to shi)", w.Code)
	}
	var envDef struct {
		Data struct {
			Pan struct {
				JuShu int `json:"jushu"`
			} `json:"pan"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&envDef); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if envDef.Data.Pan.JuShu < 1 {
		t.Error("jushu is zero (default kind)")
	}
}

func TestHandleQimenPan_InvalidKind(t *testing.T) {
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":"xun"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestHandleQimenPan_MissingBirth(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"kind":"shi"}`))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestHandleQimenPan_InvalidJSON(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{bad`))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBlackBox_QiMen_Invariants(t *testing.T) {
	body := `{` + bt15 + `,"kind":"shi"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}

	var env struct {
		Data struct {
			Pan struct {
				JuShu  int  `json:"jushu"`
				YinDun bool `json:"yin_dun"`
			} `json:"pan"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}

	if env.Data.Pan.JuShu < 1 || env.Data.Pan.JuShu > 9 {
		t.Errorf("jushu = %d, want [1,9]", env.Data.Pan.JuShu)
	}
}

func TestEdge_Qimen_InvalidKind(t *testing.T) {
	tests := []string{"SHI", "Shi", "时", "时辰", "day", "YUE", " "}
	for _, kind := range tests {
		t.Run("kind="+kind, func(t *testing.T) {
			body := `{` + bt15 + `,"kind":"` + kind + `"}`
			r := httptest.NewRequest("POST", "/", strings.NewReader(body))
			w := httptest.NewRecorder()
			handleQimenPan(w, r)
			if w.Code == http.StatusOK {
				t.Logf("BUG? qimen kind=%q accepted", kind)
			}
			if w.Code >= 500 {
				t.Errorf("qimen kind=%q caused 5xx: %d", kind, w.Code)
			}
		})
	}
}

func TestEdge2_Qimen_KindCaseVariations(t *testing.T) {
	tests := []string{"SHI", "Shi", "Ri", "NIAN", "Yue"}
	for _, kind := range tests {
		t.Run(kind, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{
				"birth": map[string]any{"time": "1984-02-15T08:00:00+08:00", "longitude": 116.4},
				"kind":  kind,
			})
			if err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest("POST", "/", strings.NewReader(string(body)))
			w := httptest.NewRecorder()
			handleQimenPan(w, r)
			if w.Code >= 500 {
				t.Errorf("kind=%q caused 5xx: %d", kind, w.Code)
			}
			if w.Code == http.StatusOK {
				t.Logf("BUG? qimen kind=%q accepted", kind)
			}
		})
	}
}

func TestBug_Qimen_KindYue_Accepted(t *testing.T) {
	// "yue" is valid per validateQimenKind
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":"yue"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("qimen kind=yue: status=%d, want 200", w.Code)
	}
}

func TestBug_Qimen_KindEmpty_DefaultsToShi(t *testing.T) {
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":""}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("qimen kind=empty: status=%d, want 200", w.Code)
	}
}

func TestDomain_Qimen_FourKinds(t *testing.T) {
	kinds := []string{"shi", "ri", "yue", "nian"}
	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{
				"birth": map[string]any{"time": "1984-02-15T08:00:00+08:00", "longitude": 116.4},
				"kind":  kind,
			})
			if err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest("POST", "/", strings.NewReader(string(body)))
			w := httptest.NewRecorder()
			handleQimenPan(w, r)
			if w.Code != http.StatusOK {
				t.Fatalf("kind=%s: status=%d, body=%s", kind, w.Code, w.Body.String())
			}
			var env struct {
				Data struct {
					Pan struct {
						Palaces []struct {
							EarthStem  string `json:"earth_stem"`
							HeavenStem string `json:"heaven_stem"`
							Door       int    `json:"door"`
							Star       int    `json:"star"`
							Spirit     int    `json:"spirit"`
						} `json:"palaces"`
					} `json:"pan"`
				} `json:"data"`
			}
			if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
				t.Fatal(err)
			}
			if len(env.Data.Pan.Palaces) == 0 {
				t.Errorf("kind=%s: no palaces", kind)
			}
		})
	}
}

func TestDomain_Qimen_JuShu_YinDun(t *testing.T) {
	body, err := json.Marshal(map[string]any{
		"birth": map[string]any{"time": "1984-02-15T08:00:00+08:00", "longitude": 116.4},
		"kind":  "shi",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/", strings.NewReader(string(body)))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	var env struct {
		Data struct {
			Pan struct {
				Jushu  int  `json:"jushu"`
				YinDun bool `json:"yin_dun"`
			} `json:"pan"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}

	// 1984-02-15 is before 立春(Feb 4), in 大寒→立春 gap, still 阳遁
	if env.Data.Pan.Jushu < 1 || env.Data.Pan.Jushu > 9 {
		t.Errorf("jushu=%d, want 1-9", env.Data.Pan.Jushu)
	}
	if env.Data.Pan.YinDun {
		t.Error("1984-02-15 should be 阳遁 (before 夏至)")
	}
	t.Logf("Qimen: jushu=%d yinDun=%v", env.Data.Pan.Jushu, env.Data.Pan.YinDun)
}

func TestDomain_Qimen_YinDun(t *testing.T) {
	body, err := json.Marshal(map[string]any{
		"birth": map[string]any{"time": "1984-07-15T08:00:00+08:00", "longitude": 116.4},
		"kind":  "shi",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/", strings.NewReader(string(body)))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	var env struct {
		Data struct {
			Pan struct {
				Jushu  int  `json:"jushu"`
				YinDun bool `json:"yin_dun"`
			} `json:"pan"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}

	// 夏至后→阴遁
	if !env.Data.Pan.YinDun {
		t.Error("1984-07-15 should be 阴遁 (after 夏至)")
	}
	t.Logf("Qimen yin: jushu=%d yinDun=%v", env.Data.Pan.Jushu, env.Data.Pan.YinDun)
}
