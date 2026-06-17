package bazi

import "liki/internal/engine/ganzhi"

// XiaoYunPillar is a single year of minor fortune (小运).
type XiaoYunPillar struct {
	Age    int    `json:"age"`
	Gan ganzhi.Gan    `json:"gan"`
	Zhi ganzhi.Zhi    `json:"zhi"`
	Name   string `json:"name"`
	TenGod string `json:"shishen"`
}

// ComputeXiaoYun computes the minor fortune (小运) pillars for each age starting from 1.
// ganzhi.Male: start from 丙寅 (stem=3, branch=3) and go forward.
// ganzhi.Female: start from 壬申 (stem=9, branch=9) and go backward.
// Returns up to maxAge pillars (typically up to 12 for childhood).
func ComputeXiaoYun(gender ganzhi.Gender, dayMaster ganzhi.Gan, maxAge int) []XiaoYunPillar {
	if maxAge <= 0 {
		maxAge = 12
	}

	var startIdx int
	if gender == ganzhi.Male {
		startIdx = ganzhi.SixtyCycleName(3, 3) // 丙寅
	} else {
		startIdx = ganzhi.SixtyCycleName(9, 9) // 壬申
	}

	pillars := make([]XiaoYunPillar, 0, maxAge)
	for age := 1; age <= maxAge; age++ {
		var idx int
		if gender == ganzhi.Male {
			idx = (startIdx + (age - 1)) % 60
		} else {
			idx = (startIdx - (age - 1) + 60) % 60
		}
		pillar := ganzhi.SixtyToZhu(idx)
		name := ganzhi.GanName(pillar.Gan) + ganzhi.ZhiName(pillar.Zhi)

		tg := ganzhi.TenGodFromGan(dayMaster, pillar.Gan)

		pillars = append(pillars, XiaoYunPillar{
			Age:    age,
			Gan:    pillar.Gan,
			Zhi:    pillar.Zhi,
			Name:   name,
			TenGod: tg,
		})
	}
	return pillars
}
