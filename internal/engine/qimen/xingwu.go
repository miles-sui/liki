package qimen

// StarInteraction holds star-palace interaction data.
type StarInteraction struct {
	Star       string
	Palace     string
	Name       string
	Meaning    string
	Auspicious bool
}

// starPalaceTable maps (star, palace) → XingInteraction 克应.
var starPalaceTable = map[[2]int]StarInteraction{
	// 天蓬星
	{int(StarTianPeng), 0}: {"天蓬", "坎", "水星入水宫", "安居之象，宜修造安葬", true},
	{int(StarTianPeng), 8}: {"天蓬", "离", "水星入火宫", "水火相激，多事之秋", false},
	{int(StarTianPeng), 1}: {"天蓬", "坤", "水星入土宫", "土能克水，事宜缓行", false},
	{int(StarTianPeng), 5}: {"天蓬", "乾", "水星入金宫", "金生水旺，谋事可成", true},
	{int(StarTianPeng), 2}: {"天蓬", "震", "水星入木宫", "木得水生，文书之喜", true},

	// 天芮星
	{int(StarTianRui), 1}: {"天芮", "坤", "土星入土宫", "比和之象，宜修造种植", true},
	{int(StarTianRui), 0}: {"天芮", "坎", "土星入水宫", "土能克水，事宜稳重", false},
	{int(StarTianRui), 8}: {"天芮", "离", "土星入火宫", "火能生土，求财得利", true},
	{int(StarTianRui), 5}: {"天芮", "乾", "土星入金宫", "土能生金，借力可成", true},

	// 天冲星
	{int(StarTianChong), 2}: {"天冲", "震", "木星入木宫", "比和之象，宜出师征战", true},
	{int(StarTianChong), 1}: {"天冲", "坤", "木星入土宫", "木能克土，宜田宅之事", false},
	{int(StarTianChong), 8}: {"天冲", "离", "木星入火宫", "木能生火，文书有成", true},

	// 天辅星
	{int(StarTianFu), 3}: {"天辅", "巽", "木星入木宫", "比和之象，宜修造入宅", true},
	{int(StarTianFu), 2}: {"天辅", "震", "木星入木宫", "兄弟同心，合作得利", true},
	{int(StarTianFu), 8}: {"天辅", "离", "木星入火宫", "木火通明，学业有成", true},

	// 天禽星 (大吉)
	{int(StarTianQin), 4}: {"天禽", "中", "土星入中宫", "居中得位，万事亨通", true},
	{int(StarTianQin), 1}: {"天禽", "坤", "土星入土宫", "比和之象，百事皆宜", true},
	{int(StarTianQin), 5}: {"天禽", "乾", "土星入金宫", "土生金旺，利谒贵求财", true},

	// 天心星 (大吉)
	{int(StarTianXin), 5}: {"天心", "乾", "金星入金宫", "比和之象，宜疗病行军", true},
	{int(StarTianXin), 2}: {"天心", "震", "金星入木宫", "金能克木，宜征战讨伐", false},
	{int(StarTianXin), 3}: {"天心", "巽", "金星入木宫", "金克木象，事宜谨慎", false},

	// 天柱星
	{int(StarTianZhu), 6}: {"天柱", "兑", "金星入金宫", "比和之象，宜守不宜攻", false},
	{int(StarTianZhu), 2}: {"天柱", "震", "金星入木宫", "金能克木，征战得利", false},
	{int(StarTianZhu), 3}: {"天柱", "巽", "金星入木宫", "金克木象，宜隐蔽行事", false},

	// 天任星 (吉)
	{int(StarTianRen), 7}: {"天任", "艮", "土星入土宫", "比和之象，宜安葬修造", true},
	{int(StarTianRen), 5}: {"天任", "乾", "土星入金宫", "土生金旺，求财得利", true},
	{int(StarTianRen), 0}: {"天任", "坎", "土星入水宫", "土能克水，事宜稳重", false},

	// 天英星
	{int(StarTianYing), 8}: {"天英", "离", "火星入火宫", "比和之象，宜文书宴乐", true},
	{int(StarTianYing), 5}: {"天英", "乾", "火星入金宫", "火能克金，出行有碍", false},
	{int(StarTianYing), 1}: {"天英", "坤", "火星入土宫", "火能生土，宜求财", true},
}

// computeStarInteractions returns star-palace 克应 for each palace.
func computeStarInteractions(pan pan) [9]StarInteraction {
	var result [9]StarInteraction
	for i := 0; i < 9; i++ {
		p := pan.Palaces[i]
		if p.Star == 0 {
			continue
		}
		key := [2]int{int(p.Star), i}
		if entry, ok := starPalaceTable[key]; ok {
			result[i] = entry
		} else {
			// Generic five-element-based description.
			result[i] = genericStarInteraction(p.Star, PalaceIndex(i+1))
		}
	}
	return result
}

func genericStarInteraction(star StarIndex, pal PalaceIndex) StarInteraction {
	return StarInteraction{
		Star:       star.String(),
		Palace:     pal.String(),
		Name:       star.String() + "加" + pal.String(),
		Meaning:    starNature(star) + "临" + pal.String() + "宫",
		Auspicious: isAuspiciousStar(star),
	}
}

func starNature(s StarIndex) string {
	switch s {
	case StarTianPeng:
		return "水性之精"
	case StarTianRui:
		return "土性之精"
	case StarTianChong:
		return "木性之精"
	case StarTianFu:
		return "木性文明"
	case StarTianQin:
		return "土性中和"
	case StarTianXin:
		return "金性肃杀"
	case StarTianZhu:
		return "金性锐利"
	case StarTianRen:
		return "土性厚重"
	case StarTianYing:
		return "火性光明"
	}
	return ""
}

func isAuspiciousStar(s StarIndex) bool {
	switch s {
	case StarTianFu, StarTianQin, StarTianXin, StarTianRen:
		return true
	default:
		return false
	}
}

// WangShuai represents 旺衰 state of a star in a palace.
type WangShuai struct {
	Star   StarIndex   `json:"star"`
	Palace PalaceIndex `json:"palace"`
	State  string      `json:"state"` // 旺/相/休/囚/废
}

// computeWangShuai computes the 旺衰 state for each star in the pan.
func computeWangShuai(pan pan) [9]WangShuai {
	var result [9]WangShuai
	for i, p := range pan.Palaces {
		if p.Star == 0 {
			continue
		}
		sw := starWuxing(p.Star)
		pw := palaceWuxing(PalaceIndex(i + 1))
		result[i] = WangShuai{
			Star:   p.Star,
			Palace: PalaceIndex(i + 1),
			State:  wuxingState(sw, pw),
		}
	}
	return result
}

// starWuxing returns the element of a star.
func starWuxing(s StarIndex) int {
	switch s {
	case StarTianPeng:
		return wxShui
	case StarTianRui, StarTianQin, StarTianRen:
		return wxTu
	case StarTianChong, StarTianFu:
		return wxMu
	case StarTianXin, StarTianZhu:
		return wxJin
	case StarTianYing:
		return wxHuo
	}
	return 0
}

// wuxingState returns 旺/相/休/囚/废 for star(at starElem) in palace(at palElem).
func wuxingState(starElem, palElem int) string {
	if starElem == palElem {
		return "旺"
	}
	if starElem == (palElem%5)+1 { // palace generates star → 相
		return "相"
	}
	if palElem == (starElem%5)+1 { // star generates palace → 休
		return "休"
	}
	if palElem == (starElem+1)%5+1 { // star overcomes palace → 囚
		return "囚"
	}
	return "废"
}
