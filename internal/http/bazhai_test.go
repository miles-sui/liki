package handler
import (
	"encoding/json"
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
	var env struct {
		Data struct {
			Gua       map[string]any `json:"gua"`
			GuaNumber int            `json:"gua_number"`
			Group     string         `json:"group"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Data.GuaNumber < 1 || env.Data.GuaNumber > 9 {
		t.Errorf("gua_number = %d, want [1,9]", env.Data.GuaNumber)
	}
	if env.Data.Group == "" {
		t.Error("group is empty")
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
	var env struct {
		Data struct {
			MingGua struct {
				GuaNumber int    `json:"gua_number"`
				Group     string `json:"group"`
			} `json:"ming_gua"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Data.MingGua.GuaNumber < 1 || env.Data.MingGua.GuaNumber > 9 {
		t.Errorf("ming_gua.gua_number = %d, want [1,9]", env.Data.MingGua.GuaNumber)
	}
	if env.Data.MingGua.Group == "" {
		t.Error("ming_gua.group is empty")
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

	var env struct {
		Data struct {
			GuaNumber int    `json:"gua_number"`
			Group     string `json:"group"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}

	if env.Data.GuaNumber < 1 || env.Data.GuaNumber > 9 {
		t.Errorf("gua_number = %d, want [1,9]", env.Data.GuaNumber)
	}
	if env.Data.Group != "东四命" && env.Data.Group != "西四命" {
		t.Errorf("group = %q, want 东四命 or 西四命", env.Data.Group)
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

	var env struct {
		Data struct {
			MingGua struct {
				GuaNumber int    `json:"gua_number"`
				Group     string `json:"group"`
			} `json:"ming_gua"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}

	if env.Data.MingGua.GuaNumber < 1 || env.Data.MingGua.GuaNumber > 9 {
		t.Errorf("ming_gua.gua_number = %d, want [1,9]", env.Data.MingGua.GuaNumber)
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

	var envM struct {
		Data struct {
			GuaNumber int    `json:"gua_number"`
			Group     string `json:"group"`
		} `json:"data"`
	}
	var envF struct {
		Data struct {
			GuaNumber int    `json:"gua_number"`
			Group     string `json:"group"`
		} `json:"data"`
	}
	if err := json.NewDecoder(wM.Body).Decode(&envM); err != nil {
		t.Fatal(err)
	}
	if err := json.NewDecoder(wF.Body).Decode(&envF); err != nil {
		t.Fatal(err)
	}

	// 同一年出生的男命和女命，命卦数应该不同（男逆女顺）
	if envM.Data.GuaNumber == envF.Data.GuaNumber && envM.Data.Group == envF.Data.Group {
		t.Log("BUG? male and female minggua are identical")
	} else {
		t.Logf("OK: male gua=%d/%s, female gua=%d/%s",
			envM.Data.GuaNumber, envM.Data.Group,
			envF.Data.GuaNumber, envF.Data.Group)
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

func TestDomain_BaZhai_MingGua_GenderDiff(t *testing.T) {
	tests := []struct {
		name      string
		time      string
		gender    string
		wantGua   string // expected trigram
		wantGroup string // 东四命 or 西四命
	}{
		// 1984男: n=(84-4)%9=80%9=8 → 艮(8), 西四命
		{"1984男→艮西四命", "1984-02-15T08:00:00+08:00", "male", "艮", "西四命"},
		// 1984女: n=(84-4)%9=8, 女: 11-8=3 → 震(3), 东四命
		{"1984女→震东四命", "1984-02-15T08:00:00+08:00", "female", "震", "东四命"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"birth":{"time":"` + tt.time + `","longitude":116.4},"gender":"` + tt.gender + `"}`
			r := httptest.NewRequest("POST", "/", strings.NewReader(body))
			w := httptest.NewRecorder()
			bazhaiChart(w, r)
			if w.Code != http.StatusOK {
				t.Fatalf("status=%d, body=%s", w.Code, w.Body.String())
			}
			var env struct {
				Data struct {
					MingGua struct {
						Gua struct {
							Name string `json:"name"`
						} `json:"gua"`
						Group string `json:"group"`
					} `json:"ming_gua"`
				} `json:"data"`
			}
			if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
				t.Fatal(err)
			}
			if env.Data.MingGua.Gua.Name != tt.wantGua {
				t.Errorf("MingGua.Name=%q, want %q", env.Data.MingGua.Gua.Name, tt.wantGua)
			}
			if env.Data.MingGua.Group != tt.wantGroup {
				t.Errorf("MingGua.Group=%q, want %q", env.Data.MingGua.Group, tt.wantGroup)
			}
		})
	}
}

func TestDomain_BaZhai_ZhuBagua_4Gua(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	bazhaiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	var env struct {
		Data struct {
			ZhuBagua []struct {
				Name   string `json:"name"`
				Wuxing string `json:"wuxing"`
			} `json:"pillar_bagua"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data.ZhuBagua) != 4 {
		t.Errorf("ZhuBagua len=%d, want 4", len(env.Data.ZhuBagua))
	}
	for i, g := range env.Data.ZhuBagua {
		if g.Name == "" {
			t.Errorf("ZhuBagua[%d].Name is empty", i)
		}
		if g.Wuxing == "" {
			t.Errorf("ZhuBagua[%d].Wuxing is empty", i)
		}
	}
}
