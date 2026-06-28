package bazi

import (
	"time"

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

type daYunZhus struct {
	startAge  int
	direction string
	zhus      []ganzhi.Zhu
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
	monthIdx := ganzhi.SixtyCycleIndex(month.Gan, month.Zhi)
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
