package ziwei

// --- 紫微定位 (0.4) ---

var ziweiStartPos = map[juShu]int{
	JuWater: 2, JuWood: 4, JuMetal: 11, JuEarth: 6, JuFire: 9,
}

func findZiwei(ju juShu, lunarDay int) palaceIndex {
	start, ok := ziweiStartPos[ju]
	if !ok {
		return 0
	}
	n := (lunarDay + int(ju) - 1) / int(ju)
	pos := (start - (n - 1)) % 12
	if pos < 0 {
		pos += 12
	}
	return palaceIndex(pos)
}

// --- 主星安星 (0.5) ---

var ziweiOffsets = []struct {
	star   starIndex
	offset int
}{
	{ZiWei, 0}, {TianJi, 1}, {TaiYang, 3}, {WuQu, 4}, {TianTong, 5}, {LianZhen, 8},
}

var tianfuOffsets = []struct {
	star   starIndex
	offset int
}{
	{TianFu, 0}, {TaiYin, 1}, {TanLang, 2}, {JuMen, 3},
	{TianXiang, 4}, {TianLiang, 5}, {QiSha, 6}, {PoJun, 10},
}

func placeMainStars(ziweiPos palaceIndex) map[palaceIndex][]starIndex {
	tianfuPos := palaceIndex((int(ziweiPos) + 2) % 12)
	m := make(map[palaceIndex][]starIndex)
	for _, e := range ziweiOffsets {
		pos := palaceIndex(((int(ziweiPos) - e.offset) + 12) % 12)
		m[pos] = append(m[pos], e.star)
	}
	for _, e := range tianfuOffsets {
		pos := palaceIndex((int(tianfuPos) + e.offset) % 12)
		m[pos] = append(m[pos], e.star)
	}
	return m
}

// --- 辅星安星 (0.6) — 每颗一个函数 + 装配 ---

func luCunPos(yearGan Gan) palaceIndex {
	table := [10]int{2, 3, 5, 6, 5, 6, 8, 9, 11, 0}
	if yg := int(yearGan); yg >= 1 && yg <= 10 {
		return palaceIndex(table[yg-1])
	}
	return 0
}

func tianKuiPos(yearGan Gan) palaceIndex {
	table := [10]int{1, 0, 11, 9, 1, 0, 1, 6, 3, 5}
	if yg := int(yearGan); yg >= 1 && yg <= 10 {
		return palaceIndex(table[yg-1])
	}
	return 0
}

func tianYuePos(tk palaceIndex) palaceIndex { return (tk + 6) % 12 }

func qingYangPos(yearZhi Zhi) palaceIndex { return palaceIndex(int(yearZhi) % 12) }
func tuoLuoPos(yearZhi Zhi) palaceIndex   { return palaceIndex(((int(yearZhi)-2+12)%12+12)%12) }

func tianMaPos(yearZhi Zhi) palaceIndex {
	table := [12]int{2, 11, 8, 5, 2, 11, 8, 5, 2, 11, 8, 5}
	if yz := int(yearZhi); yz >= 1 && yz <= 12 {
		return palaceIndex(table[yz-1])
	}
	return 0
}

func zuoFuPos(lunarMonth int) palaceIndex  { return palaceIndex((lunarMonth + 2) % 12) }
func youBiPos(lunarMonth int) palaceIndex   { return palaceIndex((11 - lunarMonth + 12) % 12) }
func wenChangPos(hourZhi Zhi) palaceIndex   { return palaceIndex((11 - int(hourZhi) + 12) % 12) }
func wenQuPos(hourZhi Zhi) palaceIndex      { return palaceIndex((int(hourZhi) + 3) % 12) }
func diKongPos(hourZhi Zhi) palaceIndex     { return palaceIndex((12 - int(hourZhi) + 12) % 12) }
func diJiePos(hourZhi Zhi) palaceIndex      { return palaceIndex((int(hourZhi) + 10) % 12) }

func huoXingPos(yearZhi Zhi, hourZhi Zhi) palaceIndex {
	return palaceIndex(marsIndex(int(yearZhi), int(hourZhi)))
}

func lingXingPos(yearZhi Zhi, hourZhi Zhi) palaceIndex {
	return palaceIndex(lingxingIndex(int(yearZhi), int(hourZhi)))
}

func marsIndex(yearZhi, hourZhi int) int {
	switch {
	case inGroup(yearZhi, 3, 7, 11):
		return (hourZhi + 1) % 12
	case inGroup(yearZhi, 9, 1, 5):
		return (3 - hourZhi + 12) % 12
	case inGroup(yearZhi, 6, 10, 2):
		return (hourZhi + 2) % 12
	case inGroup(yearZhi, 12, 4, 8):
		return (hourZhi + 8) % 12
	}
	return 0
}

func lingxingIndex(yearZhi, hourZhi int) int {
	switch {
	case inGroup(yearZhi, 3, 7, 11):
		return (hourZhi + 2) % 12
	case inGroup(yearZhi, 9, 1, 5):
		return (hourZhi + 9) % 12
	case inGroup(yearZhi, 6, 10, 2):
		return (hourZhi + 9) % 12
	case inGroup(yearZhi, 12, 4, 8):
		return (10 - hourZhi + 12) % 12
	}
	return 0
}

func inGroup(zhi, a, b, c int) bool { return zhi == a || zhi == b || zhi == c }

// placeMinorStars collects all 14 minor star placements.
func placeMinorStars(yearGan Gan, yearZhi Zhi, lunarMonth int, hourZhi Zhi) map[palaceIndex][]starIndex {
	m := make(map[palaceIndex][]starIndex)
	add := func(pos palaceIndex, s starIndex) {
		m[pos] = append(m[pos], s)
	}
	tk := tianKuiPos(yearGan)
	add(luCunPos(yearGan), LuCun)
	add(tk, TianKui)
	add(tianYuePos(tk), TianYue)
	add(qingYangPos(yearZhi), QingYang)
	add(tuoLuoPos(yearZhi), TuoLuo)
	add(tianMaPos(yearZhi), TianMa)
	add(zuoFuPos(lunarMonth), ZuoFu)
	add(youBiPos(lunarMonth), YouBi)
	add(wenChangPos(hourZhi), WenChang)
	add(wenQuPos(hourZhi), WenQu)
	add(diKongPos(hourZhi), DiKong)
	add(diJiePos(hourZhi), DiJie)
	add(huoXingPos(yearZhi, hourZhi), HuoXing)
	add(lingXingPos(yearZhi, hourZhi), LingXing)
	return m
}
