package ziwei


// yearStemBranch returns the stem and branch for a Gregorian year.
func yearStemBranch(year int) (Gan, Zhi) {
	g := Gan(((year-4)%10+10)%10 + 1)
	z := Zhi(((year-4)%12+12)%12 + 1)
	return g, z
}

// liuNianMingGong returns the annual ming gong index.
func liuNianMingGong(liuYear, birthYear int) palaceIndex {
	xuSui := liuYear - birthYear + 1
	return palaceIndex((xuSui - 1) % 12)
}

// liuNianSiHua computes the annual four transformations.
func liuNianSiHua(liuYear int) siHuaResult {
	liuGan, _ := yearStemBranch(liuYear)
	return computeSiHua(liuGan)
}

// liuNianMinors computes the annual minor stars.
func liuNianMinors(liuYearZhi, hourZhi Zhi) map[starIndex]palaceIndex {
	yz := int(liuYearZhi)
	h := int(hourZhi)
	return map[starIndex]palaceIndex{
		QingYang: qingYangPos(liuYearZhi),
		TuoLuo:   tuoLuoPos(liuYearZhi),
		HuoXing:  palaceIndex(marsIndex(yz, h)),
		LingXing: palaceIndex(lingxingIndex(yz, h)),
	}
}

// ComputeLiuNian assembles the full annual analysis.
func ComputeLiuNian(liuYear int, chart Chart) LiuNian {
	mingGong := liuNianMingGong(liuYear, chart.BirthYear)
	siHua := liuNianSiHua(liuYear)
	siHuaPalace := make(map[starIndex]palaceIndex)
	for _, p := range chart.Palaces {
		for _, s := range p.Stars {
			if _, ok := siHua[s.Star]; ok {
				siHuaPalace[s.Star] = p.Index
			}
		}
	}
	_, liuYearZhi := yearStemBranch(liuYear)
	minorStars := liuNianMinors(liuYearZhi, chart.HourZhi)
	return LiuNian{
		MingGong:     mingGong,
		MingGongName: PalaceNames[mingGong],
		SiHua:        siHua,
		SiHuaPalace:  siHuaPalace,
		MinorStars:   minorStars,
	}
}
