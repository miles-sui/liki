package bazi

import (
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

type DaYunPillar struct {
	Gan      ganzhi.Gan  `json:"gan"`
	Zhi      ganzhi.Zhi  `json:"zhi"`
	AgeStart int         `json:"age_start"`
	AgeEnd   int         `json:"age_end"`
	Name     string      `json:"name"`
	Element  string      `json:"element"`
	TenGod   string      `json:"ten_god"`
}

type DaYun struct {
	StartAge           int           `json:"start_age"`
	Direction          string        `json:"direction"`
	Pillars            []DaYunPillar `json:"pillars"`
	CurrentPillarIndex int           `json:"current_pillar_index"` // set by caller if needed
}

// computeDaYun computes the labeled big fortune (大运) pillars.
func computeDaYun(st tianwen.SolarTime, month ganzhi.Zhu, nianGan, riGan ganzhi.Gan, gender ganzhi.Gender) *DaYun {
	bf := computeDaYunPillars(st, month, nianGan, gender)
	r := &DaYun{
		StartAge:  bf.startAge,
		Direction: bf.direction,
	}
	for i, pillar := range bf.pillars {
		ageStart := bf.startAge + i*10
		r.Pillars = append(r.Pillars, DaYunPillar{
			Gan:      pillar.Gan,
			Zhi:      pillar.Zhi,
			AgeStart: ageStart,
			AgeEnd:   ageStart + 9,
			Name:     ganzhi.GanName(pillar.Gan) + ganzhi.ZhiName(pillar.Zhi),
			Element:  ganzhi.GanWuxing(pillar.Gan).String(),
			TenGod:   daYunTenGodLabel(riGan, pillar.Gan),
		})
	}
	return r
}

func daYunTenGodLabel(dayMaster, other ganzhi.Gan) string {
	if tg := ganzhi.TenGodFromGan(dayMaster, other); tg != "" {
		return tg + "运"
	}
	return "未知运"
}
