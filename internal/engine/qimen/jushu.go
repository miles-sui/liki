package qimen

import "liki/internal/engine/tianwen"

// solarTermBureau maps 24 solar terms (0=冬至) → [上元, 中元, 下元, yangDun(1/0)].
var solarTermBureau = [24][4]int{
	{1, 7, 4, 1}, // 冬至 yang
	{2, 8, 5, 1}, // 小寒 yang
	{3, 9, 6, 1}, // 大寒 yang
	{8, 5, 2, 1}, // 立春 yang
	{9, 6, 3, 1}, // 雨水 yang
	{1, 7, 4, 1}, // 惊蛰 yang
	{3, 9, 6, 1}, // 春分 yang
	{4, 1, 7, 1}, // 清明 yang
	{5, 2, 8, 1}, // 谷雨 yang
	{4, 1, 7, 1}, // 立夏 yang
	{5, 2, 8, 1}, // 小满 yang
	{6, 3, 9, 1}, // 芒种 yang
	{9, 3, 6, 0}, // 夏至 yin
	{8, 2, 5, 0}, // 小暑 yin
	{7, 1, 4, 0}, // 大暑 yin
	{2, 5, 8, 0}, // 立秋 yin
	{1, 4, 7, 0}, // 处暑 yin
	{9, 3, 6, 0}, // 白露 yin
	{7, 1, 4, 0}, // 秋分 yin
	{6, 9, 3, 0}, // 寒露 yin
	{5, 8, 2, 0}, // 霜降 yin
	{6, 9, 3, 0}, // 立冬 yin
	{5, 8, 2, 0}, // 小雪 yin
	{4, 7, 1, 0}, // 大雪 yin
}

// determineJuShu computes the bureau number and yin/yang dun for a given date.
// dayGan/dayZhi are the day pillar indices from bazi, avoiding redundant computation.
func determineJuShu(year, month, day, dayGan, dayZhi int) juShu {
	idx := tianwen.SolarTermIndex(year, month, day)
	entry := solarTermBureau[idx]

	yuan := determineYuan(dayGan, dayZhi)

	var ju int
	var yuanName string
	switch yuan {
	case 0:
		ju, yuanName = entry[0], "上元"
	case 1:
		ju, yuanName = entry[1], "中元"
	default:
		ju, yuanName = entry[2], "下元"
	}

	return juShu{
		Number: ju,
		YinDun: entry[3] == 0,
		Yuan:   yuanName,
	}
}

// determineYuan returns 0=上元, 1=中元, 2=下元 based on the day pillar's position in the 60-cycle.
func determineYuan(dayGan, dayZhi int) int {
	dayIdx := dayGan*6 + dayZhi
	if dayIdx >= 60 {
		dayIdx -= 60
	}
	if dayIdx < 0 {
		dayIdx += 60
	}

	for _, start := range []int{0, 15, 30, 45} {
		if inCycleRange(dayIdx, start, 5) {
			return 0
		}
	}
	for _, start := range []int{10, 25, 40, 55} {
		if inCycleRange(dayIdx, start, 5) {
			return 1
		}
	}
	return 2
}

func inCycleRange(idx, start, length int) bool {
	for i := 0; i < length; i++ {
		if (idx+60)%60 == (start+i)%60 {
			return true
		}
	}
	return false
}
