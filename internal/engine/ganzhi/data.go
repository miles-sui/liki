package ganzhi

import (
	_ "embed"
	"encoding/json"
	"log"
)

//go:embed data/he_hua.json
var heHuaJSON []byte

//go:embed data/chong_xing_hai.json
var chongXingHaiJSON []byte

//go:embed data/nayin.json
var nayinJSON []byte

//go:embed data/hidden_stems.json
var hiddenStemsJSON []byte

//go:embed data/chang_sheng.json
var lifeStagesJSON []byte

//go:embed data/ren_yuan.json
var renYuanJSON []byte

// NayinTable maps sexagenary index (1-60) to nayin name.
var NayinTable map[int]string

var (
	GanHes        []GanHe
	ZhiHes        []ZhiHe
	TripleHeList  []SanHeHui
	TripleHuiList []SanHeHui
	ChongPairs    []BranchPair
	XingGroups    []Xing
	HaiPairs      []BranchPair
)

// HiddenStems holds the hidden (藏干) stems for a branch.
type HiddenStems struct {
	Main  *Gan
	Mid   *Gan
	Minor *Gan
}

// Slice returns the three hidden stems as a [3]*Gan for indexed access.
func (h HiddenStems) Slice() [3]*Gan {
	return [3]*Gan{h.Main, h.Mid, h.Minor}
}

// HiddenStemsTable maps branch to its hidden stems.
var HiddenStemsTable map[Zhi]HiddenStems

// ChangShengTable maps stem to the 12 branch positions for 十二长生.
var ChangShengTable map[Gan][]Zhi

// StageNamesZH is the Chinese names for the 12 life stages.
var StageNamesZH = [12]string{
	"长生", "沐浴", "冠带", "临官", "帝旺",
	"衰", "病", "死", "墓", "绝", "胎", "养",
}

// RenYuanPhase is one phase in the 人元司令分野 table.
type RenYuanPhase struct {
	Gan     Gan    `json:"gan"`
	GanName string `json:"gan_name"`
	Days    int    `json:"days"`
}

// RenYuanTable maps month branch to its governing stem phases (人元司令分野).
var renYuanTable map[Zhi][]RenYuanPhase

// GanHe describes a 天干五合 pair and its resulting element.
type GanHe struct {
	A, B   Gan
	Result Wuxing
}

// ZhiHe describes a 地支六合 pair and its resulting element.
type ZhiHe struct {
	A, B   Zhi
	Result Wuxing
}

// BranchPair describes a pair of branches (used for 六冲, 六害, 暗合, 破).
type BranchPair struct {
	A, B Zhi
}

// SanHeHui describes a triple-branch configuration (三合 or 三会).
type SanHeHui struct {
	Branches []Zhi
	Element  Wuxing
}

// Xing describes a 相刑 group.
type Xing struct {
	Type     string
	Branches []Zhi
}

func init() {
	if err := loadHeHua(); err != nil {
		log.Fatalf("ganzhi: load he_hua: %v", err)
	}
	if err := loadChongXingHai(); err != nil {
		log.Fatalf("ganzhi: load chong_xing_hai: %v", err)
	}
	if err := loadNayin(); err != nil {
		log.Fatalf("ganzhi: load nayin: %v", err)
	}
	if err := loadHiddenStems(); err != nil {
		log.Fatalf("ganzhi: load hidden_stems: %v", err)
	}
	if err := loadLifeStages(); err != nil {
		log.Fatalf("ganzhi: load life_stages: %v", err)
	}
	if err := loadRenYuan(); err != nil {
		log.Fatalf("ganzhi: load ren_yuan: %v", err)
	}
}

func parseGan(s string) Gan {
	g, _ := ParseGan(s)
	return g
}
func parseZhi(s string) Zhi {
	z, _ := ParseZhi(s)
	return z
}
func parseWuxing(s string) Wuxing {
	w, _ := ParseWuxing(s)
	return w
}

func loadHeHua() error {
	var cfg struct {
		StemHe struct {
			Pairs [][3]string `json:"pairs"`
		} `json:"stem_he"`
		BranchHe struct {
			Pairs [][3]string `json:"pairs"`
		} `json:"branch_he"`
		TripleHe []struct {
			Branches []string `json:"branches"`
			Element  string   `json:"element"`
		} `json:"triple_he"`
		TripleHui []struct {
			Branches []string `json:"branches"`
			Element  string   `json:"element"`
		} `json:"triple_hui"`
	}
	if err := json.Unmarshal(heHuaJSON, &cfg); err != nil {
		return err
	}
	GanHes = make([]GanHe, len(cfg.StemHe.Pairs))
	for i, p := range cfg.StemHe.Pairs {
		GanHes[i] = GanHe{A: parseGan(p[0]), B: parseGan(p[1]), Result: parseWuxing(p[2])}
	}
	ZhiHes = make([]ZhiHe, len(cfg.BranchHe.Pairs))
	for i, p := range cfg.BranchHe.Pairs {
		ZhiHes[i] = ZhiHe{A: parseZhi(p[0]), B: parseZhi(p[1]), Result: parseWuxing(p[2])}
	}
	for _, th := range cfg.TripleHe {
		branches := make([]Zhi, len(th.Branches))
		for i, s := range th.Branches {
			branches[i] = parseZhi(s)
		}
		TripleHeList = append(TripleHeList, SanHeHui{Branches: branches, Element: parseWuxing(th.Element)})
	}
	for _, th := range cfg.TripleHui {
		branches := make([]Zhi, len(th.Branches))
		for i, s := range th.Branches {
			branches[i] = parseZhi(s)
		}
		TripleHuiList = append(TripleHuiList, SanHeHui{Branches: branches, Element: parseWuxing(th.Element)})
	}
	return nil
}

func loadChongXingHai() error {
	var cfg struct {
		Chong struct {
			Pairs [][2]string `json:"pairs"`
		} `json:"chong"`
		Xing []struct {
			Type     string   `json:"type"`
			Branches []string `json:"branches"`
		} `json:"xing"`
		Hai struct {
			Pairs [][2]string `json:"pairs"`
		} `json:"hai"`
	}
	if err := json.Unmarshal(chongXingHaiJSON, &cfg); err != nil {
		return err
	}
	ChongPairs = make([]BranchPair, len(cfg.Chong.Pairs))
	for i, p := range cfg.Chong.Pairs {
		ChongPairs[i] = BranchPair{A: parseZhi(p[0]), B: parseZhi(p[1])}
	}
	for _, x := range cfg.Xing {
		branches := make([]Zhi, len(x.Branches))
		for i, s := range x.Branches {
			branches[i] = parseZhi(s)
		}
		XingGroups = append(XingGroups, Xing{Type: x.Type, Branches: branches})
	}
	HaiPairs = make([]BranchPair, len(cfg.Hai.Pairs))
	for i, p := range cfg.Hai.Pairs {
		HaiPairs[i] = BranchPair{A: parseZhi(p[0]), B: parseZhi(p[1])}
	}
	return nil
}

// --- query functions ---

// -- nayin --

func loadNayin() error {
	var cfg struct {
		Nayin map[int]string `json:"nayin"`
	}
	if err := json.Unmarshal(nayinJSON, &cfg); err != nil {
		return err
	}
	NayinTable = cfg.Nayin
	return nil
}

// NaYinLabel returns the NaYin name for a stem-branch combination.
func NaYinLabel(s Gan, b Zhi) string {
	idx := SixtyCycleName(s, b)
	if idx < 60 && NayinTable != nil {
		if name, ok := NayinTable[idx]; ok {
			return name
		}
	}
	return "未知"
}

// NaYinWuxing extracts the five-element from a nayin name by its last character.
func NaYinWuxing(nayin string) Wuxing {
	rs := []rune(nayin)
	if len(rs) == 0 {
		return 0
	}
	wx, err := ParseWuxing(string(rs[len(rs)-1]))
	if err != nil {
		return 0
	}
	return wx
}

// -- hidden stems --

func loadHiddenStems() error {
	var cfg struct {
		Branches map[string]struct {
			Main  string  `json:"main"`
			Mid   *string `json:"mid"`
			Minor *string `json:"minor"`
		} `json:"branches"`
	}
	if err := json.Unmarshal(hiddenStemsJSON, &cfg); err != nil {
		return err
	}
	HiddenStemsTable = make(map[Zhi]HiddenStems, len(cfg.Branches))
	for k, v := range cfg.Branches {
		z := parseZhi(k)
		mainGan := parseGan(v.Main)
		hs := HiddenStems{Main: &mainGan}
		if v.Mid != nil {
			mg := parseGan(*v.Mid)
			hs.Mid = &mg
		}
		if v.Minor != nil {
			mg := parseGan(*v.Minor)
			hs.Minor = &mg
		}
		HiddenStemsTable[z] = hs
	}
	return nil
}

func loadLifeStages() error {
	var cfg struct {
		Stems map[string]struct {
			Stages []string `json:"stages"`
		} `json:"stems"`
	}
	if err := json.Unmarshal(lifeStagesJSON, &cfg); err != nil {
		return err
	}
	ChangShengTable = make(map[Gan][]Zhi, len(cfg.Stems))
	for k, v := range cfg.Stems {
		g := parseGan(k)
		stages := make([]Zhi, len(v.Stages))
		for i, s := range v.Stages {
			stages[i] = parseZhi(s)
		}
		ChangShengTable[g] = stages
	}
	return nil
}

// HiddenStemsForBranch returns the hidden stems (藏干) for a branch.
func HiddenStemsForBranch(b Zhi) HiddenStems {
	if hs, ok := HiddenStemsTable[b]; ok {
		return hs
	}
	return HiddenStems{}
}

func loadRenYuan() error {
	var cfg struct {
		Months map[string][]struct {
			Gan  string `json:"gan"`
			Days int    `json:"days"`
		} `json:"months"`
	}
	if err := json.Unmarshal(renYuanJSON, &cfg); err != nil {
		return err
	}
	renYuanTable = make(map[Zhi][]RenYuanPhase, len(cfg.Months))
	for k, v := range cfg.Months {
		z := parseZhi(k)
		phases := make([]RenYuanPhase, len(v))
		for i, p := range v {
			g := parseGan(p.Gan)
			phases[i] = RenYuanPhase{Gan: g, GanName: GanName(g), Days: p.Days}
		}
		renYuanTable[z] = phases
	}
	return nil
}

// RenYuanPhasesForBranch returns the 人元司令分野 phases for a month branch.
func RenYuanPhasesForBranch(branch Zhi) []RenYuanPhase {
	if phases, ok := renYuanTable[branch]; ok {
		return phases
	}
	return nil
}
