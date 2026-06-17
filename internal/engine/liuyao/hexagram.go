package liuyao
import "liki/internal/engine/ganzhi"

// guaTable holds all 64 hexagrams, ordered by palace (八宫).
// Each palace: 本宫, 一世, 二世, 三世, 四世, 五世, 游魂, 归魂.
var guaTable = [64]guaMeta{
	// ---- 乾宫 (金) ----
	{Name: "乾为天", PalaceIdx: 0, ShiPos: 6}, // 本宫
	{Name: "天风姤", PalaceIdx: 0, ShiPos: 1},
	{Name: "天山遁", PalaceIdx: 0, ShiPos: 2},
	{Name: "天地否", PalaceIdx: 0, ShiPos: 3},
	{Name: "风地观", PalaceIdx: 0, ShiPos: 4},
	{Name: "山地剥", PalaceIdx: 0, ShiPos: 5},
	{Name: "火地晋", PalaceIdx: 0, ShiPos: 4}, // 游魂
	{Name: "火天大有", PalaceIdx: 0, ShiPos: 3}, // 归魂

	// ---- 兑宫 (金) ----
	{Name: "兑为泽", PalaceIdx: 1, ShiPos: 6},
	{Name: "泽水困", PalaceIdx: 1, ShiPos: 1},
	{Name: "泽地萃", PalaceIdx: 1, ShiPos: 2},
	{Name: "泽山咸", PalaceIdx: 1, ShiPos: 3},
	{Name: "水山蹇", PalaceIdx: 1, ShiPos: 4},
	{Name: "地山谦", PalaceIdx: 1, ShiPos: 5},
	{Name: "雷山小过", PalaceIdx: 1, ShiPos: 4},
	{Name: "雷泽归妹", PalaceIdx: 1, ShiPos: 3},

	// ---- 离宫 (火) ----
	{Name: "离为火", PalaceIdx: 2, ShiPos: 6},
	{Name: "火山旅", PalaceIdx: 2, ShiPos: 1},
	{Name: "火风鼎", PalaceIdx: 2, ShiPos: 2},
	{Name: "火水未济", PalaceIdx: 2, ShiPos: 3},
	{Name: "山水蒙", PalaceIdx: 2, ShiPos: 4},
	{Name: "风水涣", PalaceIdx: 2, ShiPos: 5},
	{Name: "天水讼", PalaceIdx: 2, ShiPos: 4},
	{Name: "天火同人", PalaceIdx: 2, ShiPos: 3},

	// ---- 震宫 (木) ----
	{Name: "震为雷", PalaceIdx: 3, ShiPos: 6},
	{Name: "雷地豫", PalaceIdx: 3, ShiPos: 1},
	{Name: "雷水解", PalaceIdx: 3, ShiPos: 2},
	{Name: "雷风恒", PalaceIdx: 3, ShiPos: 3},
	{Name: "地风升", PalaceIdx: 3, ShiPos: 4},
	{Name: "水风井", PalaceIdx: 3, ShiPos: 5},
	{Name: "泽风大过", PalaceIdx: 3, ShiPos: 4},
	{Name: "泽雷随", PalaceIdx: 3, ShiPos: 3},

	// ---- 巽宫 (木) ----
	{Name: "巽为风", PalaceIdx: 4, ShiPos: 6},
	{Name: "风天小畜", PalaceIdx: 4, ShiPos: 1},
	{Name: "风火家人", PalaceIdx: 4, ShiPos: 2},
	{Name: "风雷益", PalaceIdx: 4, ShiPos: 3},
	{Name: "天雷无妄", PalaceIdx: 4, ShiPos: 4},
	{Name: "火雷噬嗑", PalaceIdx: 4, ShiPos: 5},
	{Name: "山雷颐", PalaceIdx: 4, ShiPos: 4},
	{Name: "山风蛊", PalaceIdx: 4, ShiPos: 3},

	// ---- 坎宫 (水) ----
	{Name: "坎为水", PalaceIdx: 5, ShiPos: 6},
	{Name: "水泽节", PalaceIdx: 5, ShiPos: 1},
	{Name: "水雷屯", PalaceIdx: 5, ShiPos: 2},
	{Name: "水火既济", PalaceIdx: 5, ShiPos: 3},
	{Name: "泽火革", PalaceIdx: 5, ShiPos: 4},
	{Name: "雷火丰", PalaceIdx: 5, ShiPos: 5},
	{Name: "地火明夷", PalaceIdx: 5, ShiPos: 4},
	{Name: "地水师", PalaceIdx: 5, ShiPos: 3},

	// ---- 艮宫 (土) ----
	{Name: "艮为山", PalaceIdx: 6, ShiPos: 6},
	{Name: "山火贲", PalaceIdx: 6, ShiPos: 1},
	{Name: "山天大畜", PalaceIdx: 6, ShiPos: 2},
	{Name: "山泽损", PalaceIdx: 6, ShiPos: 3},
	{Name: "火泽睽", PalaceIdx: 6, ShiPos: 4},
	{Name: "天泽履", PalaceIdx: 6, ShiPos: 5},
	{Name: "风泽中孚", PalaceIdx: 6, ShiPos: 4},
	{Name: "风山渐", PalaceIdx: 6, ShiPos: 3},

	// ---- 坤宫 (土) ----
	{Name: "坤为地", PalaceIdx: 7, ShiPos: 6},
	{Name: "地雷复", PalaceIdx: 7, ShiPos: 1},
	{Name: "地泽临", PalaceIdx: 7, ShiPos: 2},
	{Name: "地天泰", PalaceIdx: 7, ShiPos: 3},
	{Name: "雷天大壮", PalaceIdx: 7, ShiPos: 4},
	{Name: "泽天夬", PalaceIdx: 7, ShiPos: 5},
	{Name: "水天需", PalaceIdx: 7, ShiPos: 4},
	{Name: "水地比", PalaceIdx: 7, ShiPos: 3},
}

// naGanTable maps palace → stem for 纳甲.
var naGanTable = [8]ganzhi.Gan{
	ganzhi.GanJia,  // 乾纳甲
	ganzhi.GanDing, // 兑纳丁
	ganzhi.GanJi,   // 离纳己
	ganzhi.GanGeng, // 震纳庚
	ganzhi.GanXin,  // 巽纳辛
	ganzhi.GanWu,   // 坎纳戊
	ganzhi.GanBing, // 艮纳丙
	ganzhi.GanYi,   // 坤纳乙
}

// naZhiTable maps palace → 6 branch indices (1=子..12=亥) for lines 1-6.
// Order follows the八纯卦 pattern.
var naZhiTable = [8][6]int{
	{1, 3, 5, 7, 9, 11},  // 乾宫: 子寅辰 午申戌
	{6, 4, 2, 12, 10, 8},  // 兑宫: 巳卯丑 亥酉未
	{4, 2, 12, 10, 8, 6},  // 离宫: 卯丑亥 酉未巳
	{1, 3, 5, 7, 9, 11},  // 震宫: 子寅辰 午申戌
	{2, 12, 10, 8, 6, 4},  // 巽宫: 丑亥酉 未巳卯
	{3, 5, 7, 9, 11, 1},   // 坎宫: 寅辰午 申戌子
	{5, 7, 9, 11, 1, 3},   // 艮宫: 辰午申 戌子寅
	{8, 6, 4, 2, 12, 10},  // 坤宫: 未巳卯 丑亥酉
}

// dayGanShouOrder returns the六兽 assignment order starting from a given day stem.
func dayGanShouOrder(d ganzhi.Gan) [6]LiuShou {
	// 甲乙起青龙, 丙丁起朱雀, 戊起勾陈, 己起螣蛇, 庚辛起白虎, 壬癸起玄武
	// 甲乙起青龙, 丙丁起朱雀, 戊起勾陈, 己起螣蛇, 庚辛起白虎, 壬癸起玄武
	base := int(d)          // 甲(1)乙(2)...
	start := (base - 1) / 2 // 甲乙→0, 丙丁→1, 戊→2, 己→3(shift), 庚辛→4, 壬癸→5
	if base >= 6 {           // 己起螣蛇，跳过钩陈的 pair
		start++
	}
	var order [6]LiuShou
	for i := 0; i < 6; i++ {
		order[i] = LiuShou((start + i) % 6)
	}
	return order
}
