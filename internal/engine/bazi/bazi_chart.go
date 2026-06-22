package bazi

import (
	"time"

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
	Gan        ganzhi.Gan         `json:"gan"`
	Zhi        ganzhi.Zhi         `json:"zhi"`
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
type daYunZhus struct {
	startAge  int
	direction string
	zhus      []ganzhi.Zhu
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
	TaiYuanMingGong TaiYuanMingGong        `json:"tai_yuan_ming_gong"`
	HeHui           []TripleHeFull         `json:"he_hui"`
	GongJia         []GongJia              `json:"gong_jia"`
	SanQiName       string                 `json:"san_qi_name"`
	WangShuai       map[string]string      `json:"wang_shuai"`
	NayinRel        []naYinGuanXi          `json:"nayin_rel"`
}
var zhuNames = [4]string{"nian", "yue", "ri", "shi"}
func (cb ChartBase) ToBazi() ganzhi.Bazi {
	return ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: cb.Nian.Gan, Zhi: cb.Nian.Zhi},
		Yue:  ganzhi.Zhu{Gan: cb.Yue.Gan, Zhi: cb.Yue.Zhi},
		Ri:   ganzhi.Zhu{Gan: cb.Ri.Gan, Zhi: cb.Ri.Zhi},
		Shi:  ganzhi.Zhu{Gan: cb.Shi.Gan, Zhi: cb.Shi.Zhi},
	}
}
func (cb ChartBase) NaYinArray() [4]string {
	return [4]string{cb.Nian.NaYin, cb.Yue.NaYin, cb.Ri.NaYin, cb.Shi.NaYin}
}

func (cb ChartBase) CangGanArray() [4]cangGanOut {
	return [4]cangGanOut{cb.Nian.CangGan, cb.Yue.CangGan, cb.Ri.CangGan, cb.Shi.CangGan}
}

func computeDaYunZhus(st tianwen.SolarTime, month ganzhi.Zhu, nianGan ganzhi.Gan, gender ganzhi.Gender) daYunZhus {
	isYang := int(nianGan)%2 == 1
	isMale := gender == ganzhi.Male
	forward := (isMale && isYang) || (!isMale && !isYang)

	birthTime := st.Time()
	birthYear := birthTime.Year()
	jz := tianwen.JianYue(tianwen.GregorianTime(birthTime))
	mi := (int(jz) + 9) % 12 // 0=寅月..11=丑月

	// The jie index in JieQiLongitudes for month mi is mi*2.
	jieIdx := mi * 2
	var targetJie time.Time
	var dir string
	if forward {
		nextIdx := ((mi + 1) % 12) * 2
		targetJie = tianwen.SolarTermTime(birthYear, tianwen.JieQiLongitudes[nextIdx])
		if !targetJie.After(birthTime) {
			targetJie = tianwen.SolarTermTime(birthYear+1, tianwen.JieQiLongitudes[nextIdx])
		}
		dir = "顺排"
	} else {
		targetJie = tianwen.SolarTermTime(birthYear, tianwen.JieQiLongitudes[jieIdx])
		if targetJie.After(birthTime) {
			targetJie = tianwen.SolarTermTime(birthYear-1, tianwen.JieQiLongitudes[jieIdx])
		}
		dir = "逆排"
	}

	days := targetJie.Sub(birthTime).Hours() / 24
	if days < 0 {
		days = -days
	}
	startAge := int(days/3 + 0.5) // 3 days = 1 year, round to nearest

	// Generate 8 zhus from month pillar.
	monthIdx := ganzhi.SixtyCycleName(month.Gan, month.Zhi)
	zhus := make([]ganzhi.Zhu, 0, 8)
	for i := 1; i <= 8; i++ {
		var idx int
		if forward {
			idx = (monthIdx + i) % 60
		} else {
			idx = (monthIdx - i + 60) % 60
		}
		g := ganzhi.Gan((idx % 10) + 1)
		z := ganzhi.Zhi((idx % 12) + 1)
		zhus = append(zhus, ganzhi.Zhu{Gan: g, Zhi: z})
	}

	return daYunZhus{
		startAge:  startAge,
		direction: dir,
		zhus:      zhus,
	}
}
