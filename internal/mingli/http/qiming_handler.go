package minglihttp

import (
	"encoding/json"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"net/http"
	"strconv"

	"github.com/25types/25types/internal/ganzhi"
	"github.com/25types/25types/internal/httputil"
	"github.com/25types/25types/internal/mingli/qiming"
)

// QimingHandler serves name generation, evaluation, and character lookup endpoints.
type QimingHandler struct{}

// POST /api/qiming/generate
func (h *QimingHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Surname  string   `json:"surname"`
		YongShen string   `json:"yong_shen"`
		XiShen   []string `json:"xi_shen"`
		Zodiac   int      `json:"zodiac"` // year branch 1-12
		Gender   string   `json:"gender"`
		Limit    int      `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Surname, validation.Required, validation.RuneLength(1, 2)),
		validation.Field(&req.YongShen, validation.Required),
		validation.Field(&req.Zodiac, validation.Min(0), validation.Max(12)),
	); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	surnameElem := qiming.LookupSurnameElement(req.Surname)

	zodiac := qiming.ZodiacHint{}
	if req.Zodiac >= 1 && req.Zodiac <= 12 {
		zodiac = qiming.ZodiacFromYearBranch(ganzhi.Branch(req.Zodiac))
	}

	analysis := qiming.NamingAnalysis{
		Surname:    req.Surname,
		YongShen:   req.YongShen,
		XiShen:     req.XiShen,
		ZodiacHint: zodiac,
	}

	candidates := qiming.GenerateCandidates(req.Surname, analysis, req.Limit)
	if candidates == nil {
		candidates = []qiming.NameCandidate{}
	}

	httputil.RespondJSON(w, http.StatusOK, struct {
		Surname        string              `json:"surname"`
		SurnameElement string              `json:"surname_element"`
		YongShen       string              `json:"yong_shen"`
		XiShen         []string            `json:"xi_shen"`
		ZodiacHint     qiming.ZodiacHint   `json:"zodiac_hint"`
		Candidates     []qiming.NameCandidate `json:"candidates"`
	}{
		Surname: req.Surname, SurnameElement: surnameElem.String(),
		YongShen: req.YongShen, XiShen: req.XiShen,
		ZodiacHint: zodiac, Candidates: candidates,
	})
}

// POST /api/qiming/evaluate
func (h *QimingHandler) Evaluate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Surname   string `json:"surname"`
		GivenName string `json:"given_name"`
		YongShen  string `json:"yong_shen"`
		Zodiac    int    `json:"zodiac"` // year branch 1-12
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Surname, validation.Required),
		validation.Field(&req.GivenName, validation.Required, validation.RuneLength(1, 2).Error("given_name must be 1-2 Chinese characters")),
		validation.Field(&req.Zodiac, validation.Min(0), validation.Max(12)),
	); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	zodiac := ganzhi.Branch(req.Zodiac)
	result := qiming.EvaluateName(req.Surname, req.GivenName, req.YongShen, zodiac)

	httputil.RespondJSON(w, http.StatusOK, struct {
		Surname     string                  `json:"surname"`
		GivenName   string                  `json:"given_name"`
		Characters  []qiming.CharacterEntry `json:"characters"`
		WuGe        qiming.WuGe           `json:"wu_ge"`
		SanCai      qiming.SanCai         `json:"san_cai"`
		Phonetic    qiming.PhoneticMark   `json:"phonetic"`
		WuxingMatch bool                    `json:"wuxing_match"`
		ZodiacNotes []string                `json:"zodiac_notes"`
	}{
		Surname: result.Surname, GivenName: result.GivenName,
		Characters: result.Characters, WuGe: result.WuGe,
		SanCai: result.SanCai, Phonetic: result.Phonetic,
		WuxingMatch: result.WuxingMatch, ZodiacNotes: result.ZodiacNotes,
	})
}

// GET /api/qiming/characters?element=wood&stroke_min=0&stroke_max=0&limit=50
func (h *QimingHandler) GetCharacters(w http.ResponseWriter, r *http.Request) {
	elementStr := r.URL.Query().Get("element")
	var elem ganzhi.Element
	switch elementStr {
	case "wood", "木":
		elem = ganzhi.ElemWood
	case "fire", "火":
		elem = ganzhi.ElemFire
	case "earth", "土":
		elem = ganzhi.ElemEarth
	case "metal", "金":
		elem = ganzhi.ElemMetal
	case "water", "水":
		elem = ganzhi.ElemWater
	default:
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "element must be wood/火/土/金/水")
		return
	}

	strokeMin, _ := strconv.Atoi(r.URL.Query().Get("stroke_min"))
	strokeMax, _ := strconv.Atoi(r.URL.Query().Get("stroke_max"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	chars := qiming.GetCharactersByElement(elem, strokeMin, strokeMax, limit)
	httputil.RespondList(w, chars, len(chars))
}
