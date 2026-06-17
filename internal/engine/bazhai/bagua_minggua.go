package bazhai

import (
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// gua is a bagua trigram.
type gua struct {
	Index   int    `json:"index"`
	Name    string `json:"name"`
	Wuxing  string `json:"wuxing"`
	YinYang string `json:"yin_yang"`
}

var guaTable = [10]gua{
	{},
	{1, "乾", "金", "阳"},
	{2, "兑", "金", "阴"},
	{3, "离", "火", "阴"},
	{4, "震", "木", "阳"},
	{5, "巽", "木", "阴"},
	{6, "坎", "水", "阳"},
	{7, "艮", "土", "阳"},
	{8, "坤", "土", "阴"},
	{9, "离", "火", "阴"},
}

func ganNaJia(stem ganzhi.Gan) gua {
	switch stem {
	case 1, 9: return guaTable[1]  // 甲壬→乾
	case 2, 10: return guaTable[8] // 乙癸→坤
	case 3: return guaTable[7]     // 丙→艮
	case 4: return guaTable[2]     // 丁→兑
	case 5: return guaTable[6]     // 戊→坎
	case 6: return guaTable[9]     // 己→离
	case 7: return guaTable[4]     // 庚→震
	case 8: return guaTable[5]     // 辛→巽
	default: return guaTable[0]
	}
}

func zhuNaJia(p ganzhi.Zhu) gua { return ganNaJia(p.Gan) }

// MingGua is the 命卦 result.
type MingGua struct {
	Gua       gua    `json:"gua"`
	GuaNumber int    `json:"gua_number"`
	Group     string `json:"group"`
}

var westGroup = map[int]bool{2: true, 6: true, 7: true, 8: true}

// ComputeMingGua computes the 命卦 from gender and birth year.
func ComputeMingGua(gender ganzhi.Gender, birthYear int) MingGua {
	n := (birthYear%100 - 4) % 9
	if n <= 0 { n += 9 }
	if gender == ganzhi.Female {
		n = 11 - n
		if n > 9 { n = 1 + (n-1)%9 }
	}
	g := guaTable[n]
	if n == 5 {
		if gender == ganzhi.Male { g = guaTable[8] } else { g = guaTable[7] }
	}
	group := "东四命"
	if westGroup[n] || (n == 5 && gender == ganzhi.Male) || (n == 5 && gender == ganzhi.Female) {
		group = "西四命"
	}
	return MingGua{Gua: g, GuaNumber: n, Group: group}
}

// Chart is the complete八宅合参 result.
type Chart struct {
	MingGua     MingGua          `json:"ming_gua"`
	BaZhaiDirs  baZhaiDirections `json:"ba_zhai_dirs"`
	YearStars   yearStarResult   `json:"year_stars"`
	PillarBagua [4]gua           `json:"pillar_bagua"`
}

// ComputeChart computes a complete八宅合参 from solar time and gender.
func ComputeChart(st tianwen.SolarTime, gender ganzhi.Gender) Chart {
	t := st.Time()
	birthYear := t.Year()
	year := t.Year()
	bz := tianwen.ComputeBazi(st)

	mg := ComputeMingGua(gender, birthYear)
	return Chart{
		MingGua:    mg,
		BaZhaiDirs: baZhaiDirectionsForGua(mg.GuaNumber),
		YearStars:  computeYearStars(year),
		PillarBagua: [4]gua{
			zhuNaJia(bz.Nian),
			zhuNaJia(bz.Yue),
			zhuNaJia(bz.Ri),
			zhuNaJia(bz.Shi),
		},
	}
}
