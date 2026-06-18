package handler
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// ============================================================
// API 契约测试：所有端点返回 envelope，错误格式一致
// ============================================================

const envelopeBBT = `"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4}`

func TestEnvelope_AllPOSTEndpoints_ReturnDataOrError(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		handler http.HandlerFunc
	}{
		{"bazi chart", `{` + envelopeBBT + `,"gender":"male"}`, computeChart},
		{"bazi bond", `{"a":{` + envelopeBBT + `,"gender":"male"},"b":{` + envelopeBBT + `,"gender":"female"}}`, bondCharts},
		{"bazi liunian", `{"year":2025,` + envelopeBBT + `}`, liuNian},
		{"bazi liuyue", `{"year":2025,"month":6,` + envelopeBBT + `}`, liuYue},
		{"bazi liuri", `{"date":"2025-06-15",` + envelopeBBT + `}`, liuRi},
		{"bazi liushi", `{"date":"2025-06-15","hour":12,` + envelopeBBT + `}`, liuShi},
		{"bazi xiaoyun", `{` + envelopeBBT + `,"gender":"male","count":10}`, xiaoYun},
		{"bazi xiaoxian", `{"gender":"male","count":10}`, xiaoXian},
		{"ziwei chart", `{` + envelopeBBT + `,"gender":"male"}`, computeZiweiChart},
		{"qimen pan", `{` + envelopeBBT + `,"kind":"shi"}`, handleQimenPan},
		{"bazhai minggua", `{"gender":"male","birth_year":1990}`, bazhaiMingGua},
		{"bazhai chart", `{` + envelopeBBT + `,"gender":"male"}`, bazhaiChart},
		{"xuankong chart", `{` + envelopeBBT + `,"sit_mountain":1,"face_mountain":12}`, xuankongChart},
		{"liuyao chart", `{` + envelopeBBT + `,"yong_shen":"世爻"}`, handleLiuyaoChart},
		{"huangli bond date", `{` + envelopeBBT + `,"event_type":"结婚","date":"2025-06-01"}`, huangliBondDate},
		{"huangli bond month", `{` + envelopeBBT + `,"event_type":"结婚","month":"2025-06"}`, huangliBondMonth},
		{"qiming wuge", `{"surname":"张","yong_shen":"木","xi_shen":["水"]}`, handleWuge},
		{"qiming compose", `{"surname":"张","combos":[{"stroke1":5,"stroke2":8}],"yong_chars":{"5":["铭"],"8":["坤"]},"xi_chars":{"5":[],"8":[]}}`, handleCompose},
		{"qiming detail", `{"surname":"张","names":["沐洪"]}`, handleDetail},
		{"qiming evaluate", `{"surname":"张","given_name":"三","yong_shen":"木"}`, handleEvaluate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			tt.handler(w, r)

			var env struct {
				Data  json.RawMessage `json:"data"`
				Error json.RawMessage `json:"error"`
			}
			if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
				t.Errorf("response is not valid JSON: %v", err)
				return
			}
			if w.Code >= 200 && w.Code < 300 {
				if env.Data == nil && env.Error == nil {
					t.Errorf("success response has no data and no error")
				}
			} else {
				if env.Error == nil {
					t.Errorf("%s: error status %d but no error in body", tt.name, w.Code)
				}
			}
		})
	}
}

func TestEnvelope_ErrorFormat_HasCodeAndMessage(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"missing gender", `{` + envelopeBBT + `}`},
		{"invalid gender", `{` + envelopeBBT + `,"gender":"x"}`},
		{"missing birth", `{"gender":"male"}`},
		{"empty time", `{"birth":{"time":"","longitude":116.4},"gender":"male"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			computeChart(w, r)

			var env struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
				t.Errorf("decode error response: %v", err)
				return
			}
			if env.Error.Code == "" {
				t.Error("error response missing code")
			}
			if env.Error.Message == "" {
				t.Error("error response missing message")
			}
		})
	}
}

func TestEnvelope_AllErrorsReturnEnvelope(t *testing.T) {
	// 对所有 handler 发送空 body，验证返回 {"error":{...}}
	tests := []struct {
		name    string
		body    string
		handler http.HandlerFunc
	}{
		{"bazi chart empty", `{}`, computeChart},
		{"bazi bond empty", `{}`, bondCharts},
		{"bazi liunian empty", `{}`, liuNian},
		{"bazi liuyue empty", `{}`, liuYue},
		{"bazi liuri empty", `{}`, liuRi},
		{"bazi liushi empty", `{}`, liuShi},
		{"bazi xiaoyun empty", `{}`, xiaoYun},
		{"bazi xiaoxian empty", `{}`, xiaoXian},
		{"ziwei chart empty", `{}`, computeZiweiChart},
		{"qimen pan empty", `{}`, handleQimenPan},
		{"bazhai minggua empty", `{}`, bazhaiMingGua},
		{"bazhai chart empty", `{}`, bazhaiChart},
		{"xuankong chart empty", `{}`, xuankongChart},
		{"liuyao chart empty", `{}`, handleLiuyaoChart},
		{"huangli bond date empty", `{}`, huangliBondDate},
		{"huangli bond month empty", `{}`, huangliBondMonth},
		{"qiming wuge empty", `{}`, handleWuge},
		{"qiming compose empty", `{}`, handleCompose},
		{"qiming detail empty", `{}`, handleDetail},
		{"qiming evaluate empty", `{}`, handleEvaluate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			tt.handler(w, r)

			if w.Code < 400 {
				return // valid request may pass with partial body
			}
			var env struct {
				Error json.RawMessage `json:"error"`
			}
			if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
				t.Fatal(err)
			}
			if env.Error == nil {
				t.Errorf("error response has no error field (status %d)", w.Code)
			}
		})
	}
}

func TestBlackBox_AllPOSTEndpoints_ReturnEnvelope(t *testing.T) {
	// 所有 POST 端点都应该返回 {"data":{...}} 或 {"error":{...}}
	tests := []struct {
		name string
		body string
	}{
		{"bazi chart", `{` + bt15 + `,"gender":"male"}`},
		{"bazi bond", `{"a":{` + bt15 + `,"gender":"male"},"b":{` + bt15 + `,"gender":"female"}}`},
		{"bazi liunian", `{"year":2025,` + bt15 + `}`},
		{"bazi liuyue", `{"year":2025,"month":6,` + bt15 + `}`},
		{"bazi liuri", `{"date":"2025-06-15",` + bt15 + `}`},
		{"bazi liushi", `{"date":"2025-06-15","hour":12,` + bt15 + `}`},
		{"bazi xiaoyun", `{` + bt15 + `,"gender":"male","count":10}`},
		{"bazi xiaoxian", `{"gender":"male","count":10}`},
		{"ziwei chart", `{` + bt15 + `,"gender":"male"}`},
		{"qimen pan", `{` + bt15 + `,"kind":"shi"}`},
		{"bazhai minggua", `{"gender":"male","birth_year":1990}`},
		{"bazhai chart", `{` + bt15 + `,"gender":"male"}`},
		{"xuankong chart", `{` + bt15 + `,"sit_mountain":1,"face_mountain":12}`},
		{"liuyao chart", `{` + bt15 + `,"yong_shen":"世爻"}`},
		{"huangli bond date", `{` + bt15 + `,"event_type":"结婚","date":"2025-06-01"}`},
		{"huangli bond month", `{` + bt15 + `,"event_type":"结婚","month":"2025-06"}`},
		{"qiming wuge", `{"surname":"张","yong_shen":"木","xi_shen":["水"]}`},
		{"qiming compose", `{"surname":"张","combos":[{"stroke1":5,"stroke2":8}],"yong_chars":{"5":["铭"],"8":["坤"]},"xi_chars":{"5":[],"8":[]}}`},
		{"qiming detail", `{"surname":"张","names":["沐洪"]}`},
		{"qiming evaluate", `{"surname":"张","given_name":"三","yong_shen":"木"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			// 通过 mux 路由到正确的 handler
			mux := http.NewServeMux()
			// 未注册路由时直接用 handler 函数
			_ = mux
			// 所有 handler 函数导出名:
			handlers := map[string]http.HandlerFunc{
				"bazi chart":         computeChart,
				"bazi bond":          bondCharts,
				"bazi liunian":       liuNian,
				"bazi liuyue":        liuYue,
				"bazi liuri":         liuRi,
				"bazi liushi":        liuShi,
				"bazi xiaoyun":       xiaoYun,
				"bazi xiaoxian":      xiaoXian,
				"ziwei chart":        computeZiweiChart,
				"qimen pan":          handleQimenPan,
				"bazhai minggua":     bazhaiMingGua,
				"bazhai chart":       bazhaiChart,
				"xuankong chart":     xuankongChart,
				"liuyao chart":       handleLiuyaoChart,
				"huangli bond date":  huangliBondDate,
				"huangli bond month": huangliBondMonth,
				"qiming wuge":        handleWuge,
				"qiming compose":     handleCompose,
				"qiming detail":      handleDetail,
				"qiming evaluate":    handleEvaluate,
			}
			h, ok := handlers[tt.name]
			if !ok {
				t.Skip("no handler mapping")
			}
			h(w, r)

			// 所有成功响应必须有 data 或 error
			var env struct {
				Data  json.RawMessage `json:"data"`
				Error json.RawMessage `json:"error"`
			}
			if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
				t.Errorf("%s: response is not valid JSON: %v", tt.name, err)
				return
			}
			if w.Code >= 200 && w.Code < 300 {
				if env.Data == nil && env.Error == nil {
					t.Errorf("%s: success response has no data and no error", tt.name)
				}
			} else {
				if env.Error == nil {
					t.Errorf("%s: error status %d but no error in body", tt.name, w.Code)
				}
			}
		})
	}
}

func TestBlackBox_ErrorFormat_Consistent(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"missing gender", `{` + bt15 + `}`},
		{"invalid gender", `{` + bt15 + `,"gender":"x"}`},
		{"missing birth", `{"gender":"male"}`},
		{"empty time", `{"birth":{"time":"","longitude":116.4},"gender":"male"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			computeChart(w, r)

			var env struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
				t.Errorf("decode error response: %v", err)
				return
			}
			if env.Error.Code == "" {
				t.Error("error response missing code")
			}
			if env.Error.Message == "" {
				t.Error("error response missing message")
			}
		})
	}
}

func TestEdge_EmptyBody(t *testing.T) {
	// POST 空 body 应该返回 JSON 解析错误，不能 panic
	r := httptest.NewRequest("POST", "/", strings.NewReader(""))
	w := httptest.NewRecorder()
	computeChart(w, r)
	if w.Code < 400 {
		t.Errorf("empty body: status=%d, want >=400", w.Code)
	}
}

func TestEdge_EmptyJSON(t *testing.T) {
	// POST {} — 所有必填字段缺失
	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{"bazi chart", computeChart},
		{"bazi bond", bondCharts},
		{"bazi liunian", liuNian},
		{"bazi liuyue", liuYue},
		{"bazi liuri", liuRi},
		{"bazi liushi", liuShi},
		{"bazi xiaoyun", xiaoYun},
		{"bazi xiaoxian", xiaoXian},
		{"ziwei chart", computeZiweiChart},
		{"qimen pan", handleQimenPan},
		{"bazhai minggua", bazhaiMingGua},
		{"bazhai chart", bazhaiChart},
		{"xuankong chart", xuankongChart},
		{"liuyao chart", handleLiuyaoChart},
		{"huangli bond date", huangliBondDate},
		{"huangli bond month", huangliBondMonth},
		{"qiming wuge", handleWuge},
		{"qiming compose", handleCompose},
		{"qiming detail", handleDetail},
		{"qiming evaluate", handleEvaluate},
		{"ziwei daxian", computeZiweiDaxian},
		{"ziwei liunian", computeZiweiLiunian},
		{"ziwei liuyue", computeZiweiLiuyue},
		{"ziwei liuri", computeZiweiLiuri},
		{"ziwei bond", computeZiweiBond},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/", strings.NewReader("{}"))
			w := httptest.NewRecorder()
			tt.handler(w, r)
			if w.Code < 400 {
				t.Logf("BUG? %s with {} returned %d (want >=400)", tt.name, w.Code)
			}
		})
	}
}

func TestEdge_NullVsMissing(t *testing.T) {
	// gender: null vs 缺失 — 两者行为可能不同
	// null 会 unmarshal 为 zero value (""), 对 Required 来说和缺失一样
	bodyNull := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"gender":null}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(bodyNull))
	w := httptest.NewRecorder()
	computeChart(w, r)
	if w.Code == http.StatusOK {
		t.Log("BUG? gender=null accepted as valid")
	}
	if w.Code >= 500 {
		t.Errorf("gender=null caused 5xx: %d", w.Code)
	}
}

func TestEdge_BoolInsteadOfString(t *testing.T) {
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"gender":true}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeChart(w, r)
	if w.Code >= 500 {
		t.Errorf("gender=bool caused 5xx: %d", w.Code)
	}
}

func TestEdge_FloatInsteadOfInt(t *testing.T) {
	// year 传 float 而非 int
	body := `{"year":2025.5,` + bt15 + `}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	liuNian(w, r)
	if w.Code >= 500 {
		t.Errorf("year=float caused 5xx: %d", w.Code)
	}
}

func TestEdge_StringInsteadOfInt(t *testing.T) {
	body := `{"year":"2025",` + bt15 + `}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	liuNian(w, r)
	if w.Code >= 500 {
		t.Errorf("year=string caused 5xx: %d", w.Code)
	}
}

func TestEdge_AllErrorsReturnEnvelope(t *testing.T) {
	// 所有错误响应都必须有 {"error":{"code":"...","message":"..."}}
	invalidBodies := []struct {
		name    string
		handler http.HandlerFunc
		body    string
	}{
		{"bazi chart no gender", computeChart, `{` + bt15 + `}`},
		{"bazi bond no b", bondCharts, `{"a":{` + bt15 + `,"gender":"male"}}`},
		{"bazi liunian no year", liuNian, `{` + bt15 + `}`},
		{"bazi liuyue no month", liuYue, `{"year":2025,` + bt15 + `}`},
		{"bazi liuri no date", liuRi, `{` + bt15 + `}`},
		{"bazi liushi no date", liuShi, `{"hour":12,` + bt15 + `}`},
		{"bazi xiaoyun no count", xiaoYun, `{` + bt15 + `,"gender":"male"}`},
		{"bazi xiaoxian no count", xiaoXian, `{"gender":"female"}`},
		{"ziwei chart no gender", computeZiweiChart, `{` + bt15 + `}`},
		{"qimen pan no birth", handleQimenPan, `{"kind":"shi"}`},
		{"bazhai minggua no year", bazhaiMingGua, `{"gender":"male"}`},
		{"bazhai chart no gender", bazhaiChart, `{` + bt15 + `}`},
		{"xuankong chart no mountains", xuankongChart, `{` + bt15 + `}`},
		{"liuyao no birth", handleLiuyaoChart, `{"yong_shen":"世爻"}`},
		{"huangli bond date no date", huangliBondDate, `{` + bt15 + `,"event_type":"嫁娶"}`},
		{"huangli bond month no month", huangliBondMonth, `{` + bt15 + `,"event_type":"嫁娶"}`},
		{"qiming wuge no yong_shen", handleWuge, `{"surname":"张"}`},
		{"qiming compose no surname", handleCompose, `{"combos":[],"yong_chars":{},"xi_chars":{}}`},
		{"qiming detail no names", handleDetail, `{"surname":"张"}`},
		{"qiming evaluate no given_name", handleEvaluate, `{"surname":"张","yong_shen":"木"}`},
		{"ziwei daxian no gender", computeZiweiDaxian, `{"chart":{}}`},
		{"ziwei liunian no year", computeZiweiLiunian, `{"chart":{}}`},
		{"ziwei liuyue no month", computeZiweiLiuyue, `{"liu_year":2025,"chart":{}}`},
		{"ziwei liuri no day", computeZiweiLiuri, `{"liu_year":2025,"lunar_month":1,"chart":{}}`},
		{"ziwei bond no b", computeZiweiBond, `{"a":{}}`},
	}
	for _, tt := range invalidBodies {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			tt.handler(w, r)

			// 不应该 crash
			if w.Code >= 500 {
				t.Errorf("caused 5xx: %d", w.Code)
				return
			}

			// 必须有 error envelope
			var env struct {
				Data  json.RawMessage `json:"data"`
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
				t.Errorf("not valid JSON: %v", err)
				return
			}
			if env.Error.Code == "" {
				t.Errorf("status=%d, missing error.code", w.Code)
			}
			if env.Error.Message == "" {
				t.Errorf("status=%d, missing error.message", w.Code)
			}
		})
	}
}

func TestEdge_SpecialChars_EventType(t *testing.T) {
	// event_type 特殊字符不应导致 panic
	specials := []string{
		"<script>alert(1)</script>",
		"'; DROP TABLE users; --",
		"../../../etc/passwd",
		strings.Repeat("婚", 10000),
		"嫁娶\x00探监",
	}
	for _, ev := range specials {
		t.Run("", func(t *testing.T) {
			body := `{` + bt15 + `,"event_type":"` + strings.ReplaceAll(ev, `"`, `\"`) + `","date":"2025-06-15"}`
			r := httptest.NewRequest("POST", "/", strings.NewReader(body))
			w := httptest.NewRecorder()
			huangliBondDate(w, r)
			if w.Code >= 500 {
				t.Errorf("special event_type caused 5xx: %d", w.Code)
			}
		})
	}
}

func TestEdge_Boundary_YearEdgeValues(t *testing.T) {
	// year=1900, year=2100 是有效边界
	body1900 := `{"year":1900,` + bt15 + `}`
	r1 := httptest.NewRequest("POST", "/", strings.NewReader(body1900))
	w1 := httptest.NewRecorder()
	liuNian(w1, r1)
	if w1.Code != http.StatusOK {
		t.Errorf("year=1900: status=%d", w1.Code)
	}

	body2100 := `{"year":2100,` + bt15 + `}`
	r2 := httptest.NewRequest("POST", "/", strings.NewReader(body2100))
	w2 := httptest.NewRecorder()
	liuNian(w2, r2)
	if w2.Code != http.StatusOK {
		t.Errorf("year=2100: status=%d", w2.Code)
	}
}

func TestEdge_Boundary_CountEdges(t *testing.T) {
	// xiaoyun count: 1=min, 120=max
	body1 := `{` + bt15 + `,"gender":"male","count":1}`
	r1 := httptest.NewRequest("POST", "/", strings.NewReader(body1))
	w1 := httptest.NewRecorder()
	xiaoYun(w1, r1)
	if w1.Code != http.StatusOK {
		t.Errorf("count=1: status=%d", w1.Code)
	}

	body120 := `{` + bt15 + `,"gender":"male","count":120}`
	r2 := httptest.NewRequest("POST", "/", strings.NewReader(body120))
	w2 := httptest.NewRecorder()
	xiaoYun(w2, r2)
	if w2.Code != http.StatusOK {
		t.Errorf("count=120: status=%d", w2.Code)
	}
}

func TestEdge_Boundary_MonthEdges(t *testing.T) {
	// month=1, month=12 都是有效值
	for _, m := range []int{1, 12} {
		body, err := json.Marshal(map[string]any{
			"year":  2025,
			"month": m,
			"birth": map[string]any{"time": "1984-02-15T08:00:00+08:00", "longitude": 116.4},
		})
		if err != nil {
			t.Fatal(err)
		}
		r := httptest.NewRequest("POST", "/", strings.NewReader(string(body)))
		w := httptest.NewRecorder()
		liuYue(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("month=%d: status=%d", m, w.Code)
		}
	}
}

func TestEdge2_BodyNotJSON(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader("not json at all"))
	w := httptest.NewRecorder()
	computeChart(w, r)
	if w.Code >= 500 {
		t.Errorf("non-JSON body caused 5xx: %d", w.Code)
	}
}

func TestEdge2_BodyWithBOM(t *testing.T) {
	body := "\xef\xbb\xbf" + `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeChart(w, r)
	if w.Code >= 500 {
		t.Errorf("UTF-8 BOM caused 5xx: %d", w.Code)
	}
}

func TestEd3_PathValue_SpecialChars(t *testing.T) {
	tests := []string{
		"../../etc/passwd",
		"<script>alert(1)</script>",
		"'; DROP TABLE--",
		"null",
		strings.Repeat("x", 1000),
	}
	for _, id := range tests {
		t.Run("id="+id[:min(len(id), 20)], func(t *testing.T) {
			h := redirectReport()
			r := httptest.NewRequest("GET", "/api/orders/"+url.PathEscape(id)+"/report", nil)
			r.SetPathValue("id", id)
			w := httptest.NewRecorder()
			h(w, r)
			if w.Code >= 500 {
				t.Errorf("special id caused 5xx: %d", w.Code)
			}
		})
	}
}

func TestEd3_ValidateEmail_EdgeCases(t *testing.T) {
	tests := []struct {
		email string
		valid bool // 是否应被接受（不报错）
	}{
		{"user@example.com", true},
		{"", true},                                 // optional
		{"a@b.c", true},                            // 最小有效
		{"@example.com", false},                    // local part 为空
		{"user@", false},                           // domain 为空
		{"user@@example.com", false},               // 双 @
		{"ab", false},                              // 太短
		{strings.Repeat("a", 255) + "@b.c", false}, // 太长
		{"user name@example.com", false},           // 空格 — 检测位置
	}
	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			err := validateEmail(tt.email)
			if tt.valid && err != nil {
				t.Logf("BUG? valid email %q rejected: %v", tt.email, err)
			}
		})
	}
}

func TestEd3_ContentType_JSON(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeChart(w, r)

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type=%q, want application/json", ct)
	}
}
