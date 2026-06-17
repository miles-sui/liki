package liuyao

import (
	"math/rand"
	"time"

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

func shakeCoins(rng *rand.Rand) [6]YaoType {
	var yaos [6]YaoType
	for i := 0; i < 6; i++ {
		sum := 0
		for j := 0; j < 3; j++ {
			if rng.Intn(2) == 0 { sum += 2 } else { sum += 3 }
		}
		yaos[i] = YaoType(sum)
	}
	return yaos
}

func shakeCoinsFixed(results [6]int) [6]YaoType {
	var y [6]YaoType
	for i := 0; i < 6; i++ { y[i] = YaoType(results[i]) }
	return y
}

func yaoTypeToYang(y YaoType) int {
	if y.IsYang() { return 1 }
	return 0
}

func yaosToGua(yaos [6]YaoType) guaIndex {
	upper, lower := 0, 0
	for i := 0; i < 3; i++ { upper = upper<<1 | yaoTypeToYang(yaos[5-i]) }
	for i := 0; i < 3; i++ { lower = lower<<1 | yaoTypeToYang(yaos[2-i]) }
	return guaIndex(upper*8 + lower)
}

func dongYao(yaos [6]YaoType) []int {
	var dy []int
	for i := 0; i < 6; i++ {
		if yaos[i].IsChanging() { dy = append(dy, i+1) }
	}
	return dy
}

func invertDongYao(benGua guaIndex, dy []int) (guaIndex, bool) {
	if len(dy) == 0 { return 0, false }
	val := int(benGua)
	for _, pos := range dy { val ^= 1 << (5 - (pos - 1)) }
	return guaIndex(val), true
}

func computeGuaPan(yaos [6]YaoType, year, month, day int) Chart {
	benGua := yaosToGua(yaos)
	meta := guaTable[benGua]
	dy := dongYao(yaos)
	bianGua, hasBian := invertDongYao(benGua, dy)
	dayPillar := tianwen.DayPillar(year, month, day)

	lines := zhuangGua(benGua, dayPillar.Gan, false)
	for i := 0; i < 6; i++ { lines[i].Position, lines[i].Type = i+1, yaos[i] }

	var bianLines [6]Line
	if hasBian {
		bianLines = zhuangGua(bianGua, dayPillar.Gan, true)
		for i := 0; i < 6; i++ {
			bianLines[i].Position = i + 1
			if yaos[i].IsChanging() {
				if yaos[i] == LaoYang {
					bianLines[i].Type = ShaoYin
				} else {
					bianLines[i].Type = ShaoYang
				}
			} else {
				bianLines[i].Type = yaos[i]
			}
		}
	}

	return Chart{
		Name:         meta.Name,
		BenGua:       benGua,
		BianGua:      bianGua,
		Palace:       palaceNames[meta.PalaceIdx],
		PalaceWuxing: palaceWuxing[meta.PalaceIdx],
		Lines:        lines,
		BianLines:    bianLines,
		DayGan:       dayPillar.Gan,
		DayZhi:       dayPillar.Zhi,
		DongYao:      dy,
	}
}

func zhuangGua(gua guaIndex, dayGan ganzhi.Gan, isBian bool) [6]Line {
	meta := guaTable[gua]
	elem := palaceWuxing[meta.PalaceIdx]
	naZhi := naZhiTable[meta.PalaceIdx]
	naGan := naGanTable[meta.PalaceIdx]
	shouOrder := dayGanShouOrder(dayGan)
	var lines [6]Line
	for i := 0; i < 6; i++ {
		z := ganzhi.Zhi(naZhi[i])
		zwx := ganzhi.ZhiWuxing(z)
		qin := computeLiuQin(zwx, elem)
		shiYing := ""
		if !isBian {
			shi := meta.ShiPos - 1
			if i == shi { shiYing = "世" } else if i == (shi+3)%6 { shiYing = "应" }
		}
		lines[i] = Line{Gan: naGan, Zhi: z, Wuxing: zwx, LiuQin: qin, ShiYing: shiYing, LiuShou: shouOrder[i]}
	}
	return lines
}

func computeLiuQin(lineElem, palaceElem ganzhi.Wuxing) LiuQin {
	le, pe := int(lineElem), int(palaceElem)
	if le == pe { return QinXiongDi }
	if sheng(pe, le) { return QinZiSun }
	if sheng(le, pe) { return QinFumu }
	if ke(pe, le) { return QinQiCai }
	return QinGuanGui
}

func sheng(a, b int) bool { return b == a%5+1 }
func ke(a, b int) bool    { return b == (a+1)%5+1 }

// YongShenResult holds the 用神 analysis result.
type YongShenResult struct {
	Type     YongShen `json:"type"`
	Position int      `json:"position"` // line position 1-6, 0 if not found
	FuShen   *FuShen  `json:"fu_shen,omitempty"`
}

// ComputeChart computes a complete 六爻 chart from solar time, question type, and optional fixed yaos.
func ComputeChart(st tianwen.SolarTime, yongShen YongShen, fixed [6]int) Chart {
	t := st.Time()
	y, m, d := t.Date()

	var yaos [6]YaoType
	if fixed != [6]int{} {
		yaos = shakeCoinsFixed(fixed)
	} else {
		yaos = shakeCoins(rand.New(rand.NewSource(time.Now().UnixNano())))
	}

	chart := computeGuaPan(yaos, y, int(m), d)

	// Month building from bazi.
	chart.MonthZhi = tianwen.ComputeBazi(st).Yue.Zhi

	// 用神.
	pos := chart.findYongShen(yongShen)
	chart.YongShen = YongShenResult{Type: yongShen, Position: pos}
	if pos == 0 {
		chart.YongShen.FuShen = chart.findFuShen(yongShen)
	}

	// 旺衰 + 日建关系.
	for i := 0; i < 6; i++ {
		chart.WangShuai[i] = monthWangShuai(chart.Lines[i].Zhi, chart.MonthZhi)
		chart.DayRelations[i] = dayInteraction(chart.Lines[i].Zhi, chart.DayZhi)
	}

	// 应期.
	chart.YingQi = computeYingQi(&chart, yongShen)

	return chart
}
