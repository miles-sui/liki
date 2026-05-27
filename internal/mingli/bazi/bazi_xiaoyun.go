package bazi

// XiaoYunPillar is a single year of minor fortune (小运).
type XiaoYunPillar struct {
	Age       int    `json:"age"`
	Stem      Stem   `json:"stem"`
	Branch    Branch `json:"branch"`
	Name      string `json:"name"`
	TenGod    string `json:"ten_god"`
}

// ComputeXiaoYun computes the minor fortune (小运) pillars for each age starting from 1.
// Male: start from 丙寅 (stem=3, branch=3) and go forward.
// Female: start from 壬申 (stem=9, branch=9) and go backward.
// Returns up to maxAge pillars (typically up to 12 for childhood).
func ComputeXiaoYun(gender Gender, dayMaster Stem, maxAge int) []XiaoYunPillar {
	if maxAge <= 0 {
		maxAge = 12
	}

	var startIdx int
	if gender == Male {
		startIdx = sixtyCycleName(3, 3) // 丙寅
	} else {
		startIdx = sixtyCycleName(9, 9) // 壬申
	}

	dmElem := StemElement(dayMaster)
	dmYY := StemYinYang(dayMaster)

	pillars := make([]XiaoYunPillar, 0, maxAge)
	for age := 1; age <= maxAge; age++ {
		var idx int
		if gender == Male {
			idx = (startIdx + (age - 1)) % 60
		} else {
			idx = (startIdx - (age - 1) + 60) % 60
		}
		stem := Stem(idx%10 + 1)
		branch := Branch(idx%12 + 1)
		name := stemNameStr(stem) + branchNameStr(branch)

		sElem := StemElement(stem)
		sYY := StemYinYang(stem)
		tg := TenGodName(TenGodType(dmElem, dmYY, sElem, sYY))

		pillars = append(pillars, XiaoYunPillar{
			Age:    age,
			Stem:   stem,
			Branch: branch,
			Name:   name,
			TenGod: tg,
		})
	}
	return pillars
}
