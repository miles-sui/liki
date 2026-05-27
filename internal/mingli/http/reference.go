package minglihttp

import (
	"net/http"

	"github.com/25types/25types/internal/ganzhi"
	"github.com/25types/25types/internal/httputil"
	"github.com/25types/25types/internal/mingli/bazi"
	"github.com/25types/25types/internal/mingli/fengshui"
	"github.com/25types/25types/internal/mingli/huangli"
	"github.com/25types/25types/internal/tianwen"
)

// ReferenceHandler serves reference/lookup data — public, no auth.
type ReferenceHandler struct{}

// GET /api/reference/stems — all 10 heavenly stems
func (h *ReferenceHandler) Stems(w http.ResponseWriter, r *http.Request) {
	type stemEntry struct {
		Index   int    `json:"index"`
		Name    string `json:"name"`
		Element string `json:"element"`
		YinYang string `json:"yin_yang"`
	}
	items := make([]stemEntry, 10)
	for i := 1; i <= 10; i++ {
		s := ganzhi.Stem(i)
		yy := "阴"
		if ganzhi.StemYinYang(s) {
			yy = "阳"
		}
		items[i-1] = stemEntry{
			Index:   i,
			Name:    ganzhi.StemName(s),
			Element: ganzhi.StemElement(s).String(),
			YinYang: yy,
		}
	}
	httputil.RespondList(w, items, len(items))
}

// GET /api/reference/branches — all 12 earthly branches
func (h *ReferenceHandler) Branches(w http.ResponseWriter, r *http.Request) {
	type hiddenStem struct {
		Stem     int    `json:"stem"`
		StemName string `json:"stem_name"`
		Role     string `json:"role"` // "main"/"mid"/"minor"
	}

	type branchEntry struct {
		Index        int          `json:"index"`
		Name         string       `json:"name"`
		Element      string       `json:"element"`
		ZodiacAnimal string       `json:"zodiac_animal"`
		HourRange    string       `json:"hour_range"`
		HiddenStems  []hiddenStem `json:"hidden_stems"`
	}

	items := make([]branchEntry, 12)
	for i := 1; i <= 12; i++ {
		b := ganzhi.Branch(i)
		entry := branchEntry{
			Index:        i,
			Name:         ganzhi.BranchName(b),
			Element:      ganzhi.BranchElement(b).String(),
			ZodiacAnimal: ganzhi.Zodiac(b),
			HourRange:    ganzhi.BranchHourRange(b),
		}

		if hs, ok := bazi.HiddenStemsTable[i]; ok {
			if hs.Main != nil {
				entry.HiddenStems = append(entry.HiddenStems, hiddenStem{
					Stem: *hs.Main, StemName: ganzhi.StemName(ganzhi.Stem(*hs.Main)), Role: "main",
				})
			}
			if hs.Mid != nil {
				entry.HiddenStems = append(entry.HiddenStems, hiddenStem{
					Stem: *hs.Mid, StemName: ganzhi.StemName(ganzhi.Stem(*hs.Mid)), Role: "mid",
				})
			}
			if hs.Minor != nil {
				entry.HiddenStems = append(entry.HiddenStems, hiddenStem{
					Stem: *hs.Minor, StemName: ganzhi.StemName(ganzhi.Stem(*hs.Minor)), Role: "minor",
				})
			}
		}

		items[i-1] = entry
	}
	httputil.RespondList(w, items, len(items))
}

// GET /api/reference/nayin — all 60 JiaZi NaYin combinations
func (h *ReferenceHandler) Nayin(w http.ResponseWriter, r *http.Request) {
	type nayinEntry struct {
		Index      int    `json:"index"`
		Stem       int    `json:"stem"`
		Branch     int    `json:"branch"`
		StemName   string `json:"stem_name"`
		BranchName string `json:"branch_name"`
		Nayin      string `json:"nayin"`
	}

	items := make([]nayinEntry, 60)
	for idx := 0; idx < 60; idx++ {
		stem := ganzhi.Stem((idx % 10) + 1)
		branch := ganzhi.Branch((idx % 12) + 1)
		nayin := bazi.NayinTable[idx]
		items[idx] = nayinEntry{
			Index:      idx,
			Stem:       int(stem),
			Branch:     int(branch),
			StemName:   ganzhi.StemName(stem),
			BranchName: ganzhi.BranchName(branch),
			Nayin:      nayin,
		}
	}
	httputil.RespondList(w, items, len(items))
}

// GET /api/reference/shensha — shensha (神煞) rules overview
func (h *ReferenceHandler) ShenSha(w http.ResponseWriter, r *http.Request) {
	type ruleEntry struct {
		Name        string              `json:"name"`
		Category    string              `json:"category"`
		Description string              `json:"description"`
		Rules       map[string][]string `json:"rules"`
	}

	var items []ruleEntry

	ss := huangli.JianChuConfig.ShenSha

	if tianyi, ok := ss["tianyi"]; ok {
		desc := "天乙贵人，命中吉神，主逢凶化吉，遇难有救"
		items = append(items, ruleEntry{
			Name: "天乙贵人", Category: "吉", Description: desc, Rules: tianyi,
		})
	}
	if yuede, ok := ss["yuede"]; ok {
		items = append(items, ruleEntry{
			Name: "月德", Category: "吉", Description: "月德贵人，与月令相合之吉神",
			Rules: yuede,
		})
	}
	if tiande, ok := ss["tian_de"]; ok {
		items = append(items, ruleEntry{
			Name: "天德", Category: "吉", Description: "天德贵人，与天道相合之大吉神",
			Rules: tiande,
		})
	}

	httputil.RespondList(w, items, len(items))
}

// GET /api/reference/zodiac — zodiac compatibility tables
func (h *ReferenceHandler) Zodiac(w http.ResponseWriter, r *http.Request) {

	type pairEntry struct {
		A     int    `json:"a"`
		B     int    `json:"b"`
		AName string `json:"a_name"`
		BName string `json:"b_name"`
	}
	type tripleEntry struct {
		Branches []int    `json:"branches"`
		Names    []string `json:"names"`
		Element  string   `json:"element"`
	}

	var sixHe, sixChong, sixHai []pairEntry
	var tripleHe, tripleHui []tripleEntry

	for _, p := range bazi.BranchHePairs {
		sixHe = append(sixHe, pairEntry{
			A: int(p.A), B: int(p.B),
			AName: ganzhi.BranchName(p.A),
			BName: ganzhi.BranchName(p.B),
		})
	}
	for _, t := range bazi.TripleHeList {
		names := make([]string, len(t.Branches))
		for i, b := range t.Branches {
			names[i] = ganzhi.BranchName(ganzhi.Branch(b))
		}
		tripleHe = append(tripleHe, tripleEntry{
			Branches: t.Branches, Names: names,
			Element: ganzhi.Element(t.Element).String(),
		})
	}
	for _, t := range bazi.TripleHuiList {
		names := make([]string, len(t.Branches))
		for i, b := range t.Branches {
			names[i] = ganzhi.BranchName(ganzhi.Branch(b))
		}
		tripleHui = append(tripleHui, tripleEntry{
			Branches: t.Branches, Names: names,
			Element: ganzhi.Element(t.Element).String(),
		})
	}
	for _, p := range bazi.ChongPairs {
		sixChong = append(sixChong, pairEntry{
			A: int(p.A), B: int(p.B),
			AName: ganzhi.BranchName(p.A),
			BName: ganzhi.BranchName(p.B),
		})
	}
	for _, p := range bazi.HaiPairs {
		sixHai = append(sixHai, pairEntry{
			A: int(p.A), B: int(p.B),
			AName: ganzhi.BranchName(p.A),
			BName: ganzhi.BranchName(p.B),
		})
	}

	type xingEntry struct {
		Type     string   `json:"type"`
		Branches []int    `json:"branches"`
		Names    []string `json:"names"`
	}
	var xing []xingEntry
	for _, x := range bazi.XingGroups {
		names := make([]string, len(x.Branches))
		for i, b := range x.Branches {
			names[i] = ganzhi.BranchName(ganzhi.Branch(b))
		}
		xing = append(xing, xingEntry{
			Type: x.Type, Branches: x.Branches, Names: names,
		})
	}

	httputil.RespondJSON(w, http.StatusOK, struct {
		SixHe     []pairEntry   `json:"six_he"`
		TripleHe  []tripleEntry `json:"triple_he"`
		TripleHui []tripleEntry `json:"triple_hui"`
		SixChong  []pairEntry   `json:"six_chong"`
		SixHai    []pairEntry   `json:"six_hai"`
		Xing      []xingEntry   `json:"xing"`
	}{
		SixHe:     sixHe,
		TripleHe:  tripleHe,
		TripleHui: tripleHui,
		SixChong:  sixChong,
		SixHai:    sixHai,
		Xing:      xing,
	})
}

// GET /api/reference/mansions — all 28 lunar mansions
func (h *ReferenceHandler) Mansions(w http.ResponseWriter, r *http.Request) {
	mansions := huangli.AllMansions()
	httputil.RespondList(w, mansions[:], 28)
}

// GET /api/reference/trigrams — all 8 Bagua trigrams
func (h *ReferenceHandler) Trigrams(w http.ResponseWriter, r *http.Request) {
	trigrams := fengshui.AllTrigrams()
	httputil.RespondList(w, trigrams[1:], 8)
}

// GET /api/reference/huangdao — all 12 yellow/black path stars
func (h *ReferenceHandler) HuangDao(w http.ResponseWriter, r *http.Request) {
	stars := huangli.AllHuangDaoStars()
	httputil.RespondList(w, stars[:], 12)
}

// GET /api/reference/24-shan — all 24 mountain directions
func (h *ReferenceHandler) TwentyFourShan(w http.ResponseWriter, r *http.Request) {
	httputil.RespondJSON(w, http.StatusOK, struct {
		Mountains [24]fengshui.Mountain24 `json:"mountains"`
	}{Mountains: fengshui.All24Mountains()})
}

// GET /api/reference/cities?q=北京 — city search
func (h *ReferenceHandler) Cities(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	var results []tianwen.City
	if q != "" {
		results = tianwen.SearchCities(q)
	} else {
		results = tianwen.LoadedCities()
	}
	if results == nil {
		results = []tianwen.City{}
	}
	httputil.RespondList(w, results, len(results))
}
