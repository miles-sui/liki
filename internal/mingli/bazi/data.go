package bazi

import (
	_ "embed"
	"log"

	"gopkg.in/yaml.v3"
)

// -- embedded data --

//go:embed data/nayin.yaml
var nayinYAML []byte

//go:embed data/hidden_stems.yaml
var hiddenStemsYAML []byte

//go:embed data/life_stages.yaml
var lifeStagesYAML []byte

//go:embed data/he_hua.yaml
var heHuaYAML []byte

//go:embed data/chong_xing_hai.yaml
var chongXingHaiYAML []byte

// -- parsed data --

var NayinTable map[int]string
var HiddenStemsTable map[int]HiddenStemsQi
var LifeStagesTable map[int][]int
var StemHePairs []StemHePair
var BranchHePairs []BranchHePair
var TripleHeList []TripleRecord
var TripleHuiList []TripleRecord
var ChongPairs []BranchPair
var XingGroups []XingRecord
var HaiPairs []BranchPair

// StemHePair describes a 天干五合 pair and its resulting element.
type StemHePair struct {
	A, B   Stem
	Result Element
}

// BranchHePair describes a 地支六合 pair and its resulting element.
type BranchHePair struct {
	A, B   Branch
	Result Element
}

// BranchPair describes a pair of branches (used for 六冲, 六害, 暗合, 破).
type BranchPair struct {
	A, B Branch
}

// HiddenStemsQi holds the hidden (藏干) stems for a branch.
type HiddenStemsQi struct {
	Main  *int
	Mid   *int
	Minor *int
}

// Slice returns the three hidden stems as a [3]*int for indexed access.
func (h HiddenStemsQi) Slice() [3]*int {
	return [3]*int{h.Main, h.Mid, h.Minor}
}

type TripleRecord struct {
	Branches []int
	Element  int
}

type XingRecord struct {
	Type     string
	Branches []int
}

// defaultData provides a local struct for backward-compatible data access.
var defaultData = struct {
	NayinTable       map[int]string
	HiddenStemsTable map[int]HiddenStemsQi
	LifeStagesTable  map[int][]int
	StemHePairs      []StemHePair
	BranchHePairs    []BranchHePair
	TripleHeList     []TripleRecord
	TripleHuiList    []TripleRecord
	ChongPairs       []BranchPair
	XingGroups       []XingRecord
	HaiPairs         []BranchPair
}{}

func eng() *struct {
	NayinTable       map[int]string
	HiddenStemsTable map[int]HiddenStemsQi
	LifeStagesTable  map[int][]int
	StemHePairs      []StemHePair
	BranchHePairs    []BranchHePair
	TripleHeList     []TripleRecord
	TripleHuiList    []TripleRecord
	ChongPairs       []BranchPair
	XingGroups       []XingRecord
	HaiPairs         []BranchPair
} {
	return &defaultData
}

// defaultEngine is never nil; provided for backward compatibility.
var defaultEngine = &defaultData

func init() {
	if err := loadNayin(); err != nil {
		log.Printf("bazi: load nayin: %v", err)
	}
	if err := loadHiddenStems(); err != nil {
		log.Printf("bazi: load hidden_stems: %v", err)
	}
	if err := loadLifeStages(); err != nil {
		log.Printf("bazi: load life_stages: %v", err)
	}
	if err := loadHeHua(); err != nil {
		log.Printf("bazi: load he_hua: %v", err)
	}
	if err := loadChongXingHai(); err != nil {
		log.Printf("bazi: load chong_xing_hai: %v", err)
	}

	defaultData.NayinTable = NayinTable
	defaultData.HiddenStemsTable = HiddenStemsTable
	defaultData.LifeStagesTable = LifeStagesTable
	defaultData.StemHePairs = StemHePairs
	defaultData.BranchHePairs = BranchHePairs
	defaultData.TripleHeList = TripleHeList
	defaultData.TripleHuiList = TripleHuiList
	defaultData.ChongPairs = ChongPairs
	defaultData.XingGroups = XingGroups
	defaultData.HaiPairs = HaiPairs
}

func loadNayin() error {
	var cfg struct {
		Nayin map[int]string `yaml:"nayin"`
	}
	if err := yaml.Unmarshal(nayinYAML, &cfg); err != nil {
		return err
	}
	NayinTable = cfg.Nayin
	return nil
}

func loadHiddenStems() error {
	var cfg struct {
		Branches map[int]struct {
			Main  int  `yaml:"main"`
			Mid   *int `yaml:"mid"`
			Minor *int `yaml:"minor"`
		} `yaml:"branches"`
	}
	if err := yaml.Unmarshal(hiddenStemsYAML, &cfg); err != nil {
		return err
	}
	HiddenStemsTable = make(map[int]HiddenStemsQi, len(cfg.Branches))
	for k, v := range cfg.Branches {
		HiddenStemsTable[k] = HiddenStemsQi{
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
			Stages []int `yaml:"stages"`
		} `yaml:"stems"`
	}
	if err := yaml.Unmarshal(lifeStagesYAML, &cfg); err != nil {
		return err
	}
	LifeStagesTable = make(map[int][]int, len(cfg.Stems))
	for k, v := range cfg.Stems {
		LifeStagesTable[k] = v.Stages
	}
	return nil
}

func loadHeHua() error {
	var cfg struct {
		StemHe struct {
			Pairs [][3]int `yaml:"pairs"`
		} `yaml:"stem_he"`
		BranchHe struct {
			Pairs [][3]int `yaml:"pairs"`
		} `yaml:"branch_he"`
		TripleHe []struct {
			Branches []int `yaml:"branches"`
			Element  int   `yaml:"element"`
		} `yaml:"triple_he"`
		TripleHui []struct {
			Branches []int `yaml:"branches"`
			Element  int   `yaml:"element"`
		} `yaml:"triple_hui"`
	}
	if err := yaml.Unmarshal(heHuaYAML, &cfg); err != nil {
		return err
	}
	StemHePairs = make([]StemHePair, len(cfg.StemHe.Pairs))
	for i, p := range cfg.StemHe.Pairs {
		StemHePairs[i] = StemHePair{A: Stem(p[0]), B: Stem(p[1]), Result: Element(p[2])}
	}
	BranchHePairs = make([]BranchHePair, len(cfg.BranchHe.Pairs))
	for i, p := range cfg.BranchHe.Pairs {
		BranchHePairs[i] = BranchHePair{A: Branch(p[0]), B: Branch(p[1]), Result: Element(p[2])}
	}
	for _, th := range cfg.TripleHe {
		TripleHeList = append(TripleHeList, TripleRecord{th.Branches, th.Element})
	}
	for _, th := range cfg.TripleHui {
		TripleHuiList = append(TripleHuiList, TripleRecord{th.Branches, th.Element})
	}
	return nil
}

func loadChongXingHai() error {
	var cfg struct {
		Chong struct {
			Pairs [][2]int `yaml:"pairs"`
		} `yaml:"chong"`
		Xing []struct {
			Type     string `yaml:"type"`
			Branches []int  `yaml:"branches"`
		} `yaml:"xing"`
		Hai struct {
			Pairs [][2]int `yaml:"pairs"`
		} `yaml:"hai"`
	}
	if err := yaml.Unmarshal(chongXingHaiYAML, &cfg); err != nil {
		return err
	}
	ChongPairs = make([]BranchPair, len(cfg.Chong.Pairs))
	for i, p := range cfg.Chong.Pairs {
		ChongPairs[i] = BranchPair{A: Branch(p[0]), B: Branch(p[1])}
	}
	for _, x := range cfg.Xing {
		XingGroups = append(XingGroups, XingRecord{x.Type, x.Branches})
	}
	HaiPairs = make([]BranchPair, len(cfg.Hai.Pairs))
	for i, p := range cfg.Hai.Pairs {
		HaiPairs[i] = BranchPair{A: Branch(p[0]), B: Branch(p[1])}
	}
	return nil
}
