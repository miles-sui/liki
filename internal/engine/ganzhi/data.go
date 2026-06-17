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

//go:embed data/life_stages.json
var lifeStagesJSON []byte

//go:embed data/ren_yuan.json
var renYuanJSON []byte

// NayinTable maps sexagenary index (1-60) to nayin name.
var NayinTable map[int]string

var (
	GanHes    []GanHe
	ZhiHes    []ZhiHe
	TripleHeList  []SanHeHui
	TripleHuiList []SanHeHui
	ChongPairs    []BranchPair
	XingGroups    []Xing
	HaiPairs      []BranchPair
)

// anHePairs lists 地支暗合 pairs (寅丑, 卯申, 午亥, 子戌).
var anHePairs = []BranchPair{
	{A: 3, B: 2},  // 寅丑
	{A: 4, B: 9},  // 卯申
	{A: 7, B: 12}, // 午亥
	{A: 1, B: 11}, // 子戌
}

// poPairs lists 地支相破 pairs (子酉, 寅亥, 辰丑, 午卯, 申巳, 戌未).
var poPairs = []BranchPair{
	{A: 1, B: 10}, // 子酉
	{A: 3, B: 12}, // 寅亥
	{A: 5, B: 2},  // 辰丑
	{A: 7, B: 4},  // 午卯
	{A: 9, B: 6},  // 申巳
	{A: 11, B: 8}, // 戌未
}

// HiddenStems holds the hidden (藏干) stems for a branch.
type HiddenStems struct {
	Main  *int
	Mid   *int
	Minor *int
}

// Slice returns the three hidden stems as a [3]*int for indexed access.
func (h HiddenStems) Slice() [3]*int {
	return [3]*int{h.Main, h.Mid, h.Minor}
}

// HiddenStemsTable maps branch index to its hidden stems.
var HiddenStemsTable map[int]HiddenStems

// LifeStagesTable maps stem index to the 12 branch positions for 十二长生.
var LifeStagesTable map[int][]int

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
var renYuanTable map[int][]RenYuanPhase

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
	Branches []int
	Element  Wuxing
}

// Xing describes a 相刑 group.
type Xing struct {
	Type     string
	Branches []int
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

func loadHeHua() error {
	var cfg struct {
		StemHe struct {
			Pairs [][3]int `json:"pairs"`
		} `json:"stem_he"`
		BranchHe struct {
			Pairs [][3]int `json:"pairs"`
		} `json:"branch_he"`
		TripleHe []struct {
			Branches []int `json:"branches"`
			Element  int   `json:"element"`
		} `json:"triple_he"`
		TripleHui []struct {
			Branches []int `json:"branches"`
			Element  int   `json:"element"`
		} `json:"triple_hui"`
	}
	if err := json.Unmarshal(heHuaJSON, &cfg); err != nil {
		return err
	}
	GanHes = make([]GanHe, len(cfg.StemHe.Pairs))
	for i, p := range cfg.StemHe.Pairs {
		GanHes[i] = GanHe{A: Gan(p[0]), B: Gan(p[1]), Result: Wuxing(p[2])}
	}
	ZhiHes = make([]ZhiHe, len(cfg.BranchHe.Pairs))
	for i, p := range cfg.BranchHe.Pairs {
		ZhiHes[i] = ZhiHe{A: Zhi(p[0]), B: Zhi(p[1]), Result: Wuxing(p[2])}
	}
	TripleHeList = make([]SanHeHui, 0, len(cfg.TripleHe))
	for _, th := range cfg.TripleHe {
		TripleHeList = append(TripleHeList, SanHeHui{th.Branches, Wuxing(th.Element)})
	}
	TripleHuiList = make([]SanHeHui, 0, len(cfg.TripleHui))
	for _, th := range cfg.TripleHui {
		TripleHuiList = append(TripleHuiList, SanHeHui{th.Branches, Wuxing(th.Element)})
	}
	return nil
}

func loadChongXingHai() error {
	var cfg struct {
		Chong struct {
			Pairs [][2]int `json:"pairs"`
		} `json:"chong"`
		Xing []struct {
			Type     string `json:"type"`
			Branches []int  `json:"branches"`
		} `json:"xing"`
		Hai struct {
			Pairs [][2]int `json:"pairs"`
		} `json:"hai"`
	}
	if err := json.Unmarshal(chongXingHaiJSON, &cfg); err != nil {
		return err
	}
	ChongPairs = make([]BranchPair, len(cfg.Chong.Pairs))
	for i, p := range cfg.Chong.Pairs {
		ChongPairs[i] = BranchPair{A: Zhi(p[0]), B: Zhi(p[1])}
	}
	for _, x := range cfg.Xing {
		XingGroups = append(XingGroups, Xing{x.Type, x.Branches})
	}
	HaiPairs = make([]BranchPair, len(cfg.Hai.Pairs))
	for i, p := range cfg.Hai.Pairs {
		HaiPairs[i] = BranchPair{A: Zhi(p[0]), B: Zhi(p[1])}
	}
	return nil
}

// --- query functions ---

func inBranchList(branches []int, b Zhi) bool {
	v := int(b)
	for _, x := range branches {
		if x == v {
			return true
		}
	}
	return false
}

// IsGanHe returns true if the two stems form a 天干五合 pair.
func IsGanHe(a, b Gan) bool {
	for _, p := range GanHes {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			return true
		}
	}
	return false
}

// IsZhiHe returns true if the two branches form a 地支六合 pair.
func IsZhiHe(a, b Zhi) bool {
	for _, p := range ZhiHes {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			return true
		}
	}
	return false
}

// IsTripleHe returns true if the two branches belong to the same 三合 group.
func IsTripleHe(a, b Zhi) bool {
	for _, tr := range TripleHeList {
		if inBranchList(tr.Branches, a) && inBranchList(tr.Branches, b) {
			return true
		}
	}
	return false
}

// IsTripleHui returns true if the two branches belong to the same 三会 group.
func IsTripleHui(a, b Zhi) bool {
	for _, tr := range TripleHuiList {
		if inBranchList(tr.Branches, a) && inBranchList(tr.Branches, b) {
			return true
		}
	}
	return false
}

// IsLiuChong returns true if the two branches form a 六冲 pair.
func IsLiuChong(a, b Zhi) bool {
	for _, p := range ChongPairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			return true
		}
	}
	return false
}

// IsXing returns true if the two branches are in a 相刑 relation.
func IsXing(a, b Zhi) bool {
	for _, x := range XingGroups {
		if x.Type == "zi" {
			if a == b && inBranchList(x.Branches, a) {
				return true
			}
		} else if a != b {
			if inBranchList(x.Branches, a) && inBranchList(x.Branches, b) {
				return true
			}
		}
	}
	return false
}

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

// IsHai returns true if the two branches form a 六害 pair.
func IsHai(a, b Zhi) bool {
	for _, p := range HaiPairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			return true
		}
	}
	return false
}

// IsAnHe returns true if the two branches form a 暗合 pair.
func IsAnHe(a, b Zhi) bool {
	for _, p := range anHePairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			return true
		}
	}
	return false
}

// IsPo returns true if the two branches form a 相破 pair.
func IsPo(a, b Zhi) bool {
	for _, p := range poPairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			return true
		}
	}
	return false
}

// -- hidden stems --

func loadHiddenStems() error {
	var cfg struct {
		Branches map[int]struct {
			Main  int  `json:"main"`
			Mid   *int `json:"mid"`
			Minor *int `json:"minor"`
		} `json:"branches"`
	}
	if err := json.Unmarshal(hiddenStemsJSON, &cfg); err != nil {
		return err
	}
	HiddenStemsTable = make(map[int]HiddenStems, len(cfg.Branches))
	for k, v := range cfg.Branches {
		HiddenStemsTable[k] = HiddenStems{
			Main:  &v.Main,
			Mid:   v.Mid,
			Minor: v.Minor,
		}
	}
	return nil
}

func loadLifeStages() error {
	var cfg struct {
		Stems map[int]struct {
			Stages []int `json:"stages"`
		} `json:"stems"`
	}
	if err := json.Unmarshal(lifeStagesJSON, &cfg); err != nil {
		return err
	}
	LifeStagesTable = make(map[int][]int, len(cfg.Stems))
	for k, v := range cfg.Stems {
		LifeStagesTable[k] = v.Stages
	}
	return nil
}

// HiddenStemsForBranch returns the hidden stems (藏干) for a branch.
func HiddenStemsForBranch(b Zhi) HiddenStems {
	if hs, ok := HiddenStemsTable[int(b)]; ok {
		return hs
	}
	return HiddenStems{}
}

func loadRenYuan() error {
	var cfg struct {
		Months map[int][]struct {
			Gan     int    `json:"gan"`
			GanName string `json:"gan_name"`
			Days    int    `json:"days"`
		} `json:"months"`
	}
	if err := json.Unmarshal(renYuanJSON, &cfg); err != nil {
		return err
	}
	renYuanTable = make(map[int][]RenYuanPhase, len(cfg.Months))
	for k, v := range cfg.Months {
		phases := make([]RenYuanPhase, len(v))
		for i, p := range v {
			phases[i] = RenYuanPhase{Gan: Gan(p.Gan), GanName: p.GanName, Days: p.Days}
		}
		renYuanTable[k] = phases
	}
	return nil
}

// RenYuanPhasesForBranch returns the 人元司令分野 phases for a month branch.
func RenYuanPhasesForBranch(branch Zhi) []RenYuanPhase {
	if phases, ok := renYuanTable[int(branch)]; ok {
		return phases
	}
	return nil
}
