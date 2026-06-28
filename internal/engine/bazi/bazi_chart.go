package bazi

import (
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// --- per-pillar data ---

type cangGanOut struct {
	Main  ganzhi.Gan  `json:"main"`
	Mid   *ganzhi.Gan `json:"mid"`
	Minor *ganzhi.Gan `json:"minor"`
}

type zhuInfo struct {
	ganzhi.Zhu
	NaYin      string             `json:"na_yin"`
	CangGan    cangGanOut         `json:"cang_gan"`
	ShiShens   []shiShenEntry     `json:"shi_shens"`
	ChangSheng []changShengEntry  `json:"chang_sheng"`
	ShenSha    []shenShaEntry     `json:"shen_sha"`
	IsVoid     bool               `json:"is_void"`
	IsSelfHe   bool               `json:"is_self_he"`
	IsKuiGang  bool               `json:"is_kui_gang"`
	SelfHeName string             `json:"self_he_name"`
}
type shiShenEntry struct {
	ShiShen ganzhi.ShiShen `json:"shi_shen"`
	Name    string         `json:"name"`
	Source  string         `json:"source"`
	Gan     ganzhi.Gan     `json:"gan"`
}
type changShengEntry struct {
	Stage string     `json:"stage"`
	Name  string     `json:"name"`
	Gan   ganzhi.Gan `json:"gan"`
}
type stageOut struct {
	Name  string     `json:"name"`
	Index ganzhi.Zhi `json:"index"`
}
type naYinGuanXi struct {
	A, B, Relation string
}
// Ten god source constants.
const (
	sourceGan     = "stem"
	sourceMainQi  = "main_qi"
	sourceMidQi   = "mid_qi"
	sourceMinQi   = "minor_qi"
)

// --- chart ---

// ChartBase holds the core bazi chart data without display-only decorations.
type ChartBase struct {
	Nian        zhuInfo                `json:"nian"`
	Yue         zhuInfo                `json:"yue"`
	Ri          zhuInfo                `json:"ri"`
	Shi         zhuInfo                `json:"shi"`
	DaYun       *DaYun                 `json:"da_yun"`
}
// Chart holds a complete bazi chart including display fields and analysis.
type Chart struct {
	ChartBase
	SolarTime       tianwen.SolarTime      `json:"solar_time"`
	FuYi            FuYi                   `json:"fu_yi"`
	TiaoHou         TiaoHou                `json:"tiao_hou"`
	ChangSheng      [12]stageOut           `json:"chang_sheng"`
	WuxingCount     map[ganzhi.Wuxing]int  `json:"wuxing_count"`
	SanYuan SanYuan        `json:"san_yuan"`
	HeHui           []HeHuiCombination         `json:"he_hui"`
	GongJia         []GongJia              `json:"gong_jia"`
	SanQiName       string                 `json:"san_qi_name"`
	WangShuai       map[string]string      `json:"wang_shuai"`
	NayinRel        []naYinGuanXi          `json:"nayin_rel"`
}
var zhuNames = [4]string{"nian", "yue", "ri", "shi"}
func (cb ChartBase) ToBazi() ganzhi.Bazi {
	return ganzhi.Bazi{
		Nian: cb.Nian.Zhu,
		Yue:  cb.Yue.Zhu,
		Ri:   cb.Ri.Zhu,
		Shi:  cb.Shi.Zhu,
	}
}
func (cb ChartBase) NaYinArray() [4]string {
	return [4]string{cb.Nian.NaYin, cb.Yue.NaYin, cb.Ri.NaYin, cb.Shi.NaYin}
}

func (cb ChartBase) CangGanArray() [4]cangGanOut {
	return [4]cangGanOut{cb.Nian.CangGan, cb.Yue.CangGan, cb.Ri.CangGan, cb.Shi.CangGan}
}

