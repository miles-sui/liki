package ziwei


import "liki/internal/engine/ganzhi"
// chartParams holds the input needed for ziwei chart computation.
type chartParams struct {
	LunarMonth int
	LunarDay   int
	HourZhi    Zhi
	YearGan    Gan
	YearZhi    Zhi
	Gender     ganzhi.Gender
}

// computeChart builds the core ziwei chart (palaces + stars, no brightness/patterns).
func computeChart(p chartParams) Chart {
	mingZhi, shenZhi := computeMingShen(p.LunarMonth, p.HourZhi)
	palaceZhis := arrangePalaceZhis(mingZhi)
	shenGong := findShenGongIndex(palaceZhis, shenZhi)

	mingGan, palaceGans := arrangePalaceGans(p.YearGan, mingZhi)
	ju := determineJuShu(mingGan, mingZhi)
	ziweiPos := findZiwei(ju, p.LunarDay)

	mainByPalace := placeMainStars(ziweiPos)
	minorByPalace := placeMinorStars(p.YearGan, p.YearZhi, p.LunarMonth, p.HourZhi)

	var palaces [12]palace
	for i := 0; i < 12; i++ {
		var starInfos []starInfo
		for _, s := range mainByPalace[palaceIndex(i)] {
			starInfos = append(starInfos, starInfo{Star: s, Name: starName(s), IsMajor: true})
		}
		for _, s := range minorByPalace[palaceIndex(i)] {
			starInfos = append(starInfos, starInfo{Star: s, Name: starName(s), IsMajor: false})
		}
		palaces[i] = palace{
			Index:        palaceIndex(i),
			Name:         PalaceNames[i],
			Gan:          palaceGans[i],
			Zhi:          palaceZhis[i],
			IsBodyPalace: palaceIndex(i) == shenGong,
			Stars:        starInfos,
		}
	}

	return Chart{
		Palaces:   palaces,
		MingGong:  0,
		ShenGong:  shenGong,
		JuShu:     ju,
		JuShuName: juShuName(ju),
		ZiweiPos:  ziweiPos,
		YearGan:   p.YearGan,
		HourZhi:   p.HourZhi,
	}
}

// buildChartDetail enriches a core chart with siHua, brightness, and patterns.
func buildChartDetail(chart Chart) Chart {
	siHua := computeSiHua(chart.YearGan)
	for i := range chart.Palaces {
		for j, s := range chart.Palaces[i].Stars {
			if s.IsMajor {
				chart.Palaces[i].Stars[j].Brightness = miaoWang(s.Star, chart.Palaces[i].Zhi).String()
				if h, ok := siHua[s.Star]; ok {
					chart.Palaces[i].Stars[j].SiHua = string(h)
				}
			}
		}
	}
	chart.SiHua = siHua
	chart.Patterns = findPatterns(chart.Palaces)
	return chart
}


// ComputeChart computes a complete紫微命盘.
func ComputeChart(birthYear, lunarMonth, lunarDay int, hourZhi Zhi, yearGan Gan, yearZhi Zhi, gender ganzhi.Gender) Chart {
	p := chartParams{
		LunarMonth: lunarMonth, LunarDay: lunarDay,
		HourZhi: hourZhi, YearGan: yearGan, YearZhi: yearZhi, Gender: gender,
	}
	chart := computeChart(p)
	chart.BirthYear = birthYear
	chart.Gender = gender
	chart = buildChartDetail(chart)
	return chart
}
