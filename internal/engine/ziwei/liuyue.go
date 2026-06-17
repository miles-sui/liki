package ziwei

// LiuYue holds monthly fate analysis.
type LiuYue struct {
	MingGong     palaceIndex `json:"ming_gong"`
	MingGongName string      `json:"ming_gong_name"`
	SiHua        siHuaResult `json:"si_hua"`
}

func ComputeLiuYue(liuYear, lunarMonth int, chart Chart) LiuYue {
	liuNianMing := liuNianMingGong(liuYear, chart.BirthYear)
	ming := liuYueMingGong(lunarMonth, liuNianMing)
	liuYearGan := yearGan(liuYear)
	return LiuYue{
		MingGong:     ming,
		MingGongName: PalaceNames[ming],
		SiHua:        liuYueSiHua(lunarMonth, liuYearGan),
	}
}

func liuYueMingGong(lunarMonth int, liuNianMing palaceIndex) palaceIndex {
	return (liuNianMing + palaceIndex(lunarMonth-1)) % 12
}

func liuYueSiHua(lunarMonth int, liuYearGan Gan) siHuaResult {
	yg := yinGan(liuYearGan)
	monthGan := Gan(((int(yg)-1+lunarMonth-1)%10+10)%10 + 1)
	return computeSiHua(monthGan)
}

// LiuRi holds daily fate analysis.
type LiuRi struct {
	MingGong     palaceIndex `json:"ming_gong"`
	MingGongName string      `json:"ming_gong_name"`
	SiHua        siHuaResult `json:"si_hua"`
}

func ComputeLiuRi(liuYear, lunarMonth, lunarDay int, chart Chart) LiuRi {
	liuNianMing := liuNianMingGong(liuYear, chart.BirthYear)
	liuYueMing := liuYueMingGong(lunarMonth, liuNianMing)
	ming := liuRiMingGong(lunarDay, liuYueMing)
	dayGan := riGan(lunarDay, liuYear, lunarMonth)
	return LiuRi{
		MingGong:     ming,
		MingGongName: PalaceNames[ming],
		SiHua:        liuRiSiHua(dayGan),
	}
}

func liuRiMingGong(lunarDay int, liuYueMing palaceIndex) palaceIndex {
	return (liuYueMing + palaceIndex(lunarDay-1)) % 12
}

func yearGan(year int) Gan { ys := (year - 4) % 60; return Gan((ys+9)%10 + 1) }
func riGan(day, year, month int) Gan { return Gan(day) } // TODO: placeholder
func liuRiSiHua(dayGan Gan) siHuaResult { return computeSiHua(dayGan) }
