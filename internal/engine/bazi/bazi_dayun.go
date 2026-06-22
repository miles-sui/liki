package bazi

import (
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// DaYunZhu holds one 10-year fortune pillar in the big fortune cycle.
type DaYunZhu struct {
	Gan      ganzhi.Gan  `json:"gan"`
	Zhi      ganzhi.Zhi  `json:"zhi"`
	AgeStart int         `json:"age_start"`
	AgeEnd   int         `json:"age_end"`
	Name     string      `json:"name"`
	Element  string      `json:"element"`
	ShiShen   string      `json:"shi_shen"`
}

// DaYun holds the big fortune (大运) cycle for a bazi chart.
type DaYun struct {
	StartAge           int           `json:"start_age"`
	Direction          string        `json:"direction"`
	Zhu            []DaYunZhu `json:"zhu"`
	CurrentZhuIndex int           `json:"current_zhu_index"` // set by caller if needed
}

// computeDaYun computes the labeled big fortune (大运) zhus.
func computeDaYun(st tianwen.SolarTime, month ganzhi.Zhu, nianGan, riGan ganzhi.Gan, gender ganzhi.Gender) *DaYun {
	bf := computeDaYunZhus(st, month, nianGan, gender)
	r := &DaYun{
		StartAge:  bf.startAge,
		Direction: bf.direction,
	}
	for i, zhu := range bf.zhus {
		ageStart := bf.startAge + i*10
		r.Zhu = append(r.Zhu, DaYunZhu{
			Gan:      zhu.Gan,
			Zhi:      zhu.Zhi,
			AgeStart: ageStart,
			AgeEnd:   ageStart + 9,
			Name:     ganzhi.GanName(zhu.Gan) + ganzhi.ZhiName(zhu.Zhi),
			Element:  ganzhi.GanWuxing(zhu.Gan).String(),
			ShiShen:   daYunShiShenLabel(riGan, zhu.Gan),
		})
	}
	return r
}

func daYunShiShenLabel(riYuan, other ganzhi.Gan) string {
	if tg := ganzhi.ShiShenFromGan(riYuan, other).String(); tg != "" {
		return tg + "运"
	}
	return "未知运"
}
