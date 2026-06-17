package qimen

import "liki/internal/engine/ganzhi"

// zhiPalace returns the palace index (1-9) where a branch sits.
func zhiPalace(z ganzhi.Zhi) PalaceIndex {
	switch int(z) {
	case 1: // 子 → 坎
		return PalaceKan
	case 2, 3: // 丑, 寅 → 艮
		return PalaceGen
	case 4: // 卯 → 震
		return PalaceZhen
	case 5, 6: // 辰, 巳 → 巽
		return PalaceXun
	case 7: // 午 → 离
		return PalaceLi
	case 8, 9: // 未, 申 → 坤
		return PalaceKun
	case 10: // 酉 → 兑
		return PalaceDui
	case 11, 12: // 戌, 亥 → 乾
		return PalaceQian
	}
	return PalaceKan
}

// palaceZhi returns the principal branch of a palace.
func palaceZhi(p PalaceIndex) ganzhi.Zhi {
	switch p {
	case PalaceKan:
		return ganzhi.ZhiZi // 子
	case PalaceKun:
		return ganzhi.ZhiWei // 未
	case PalaceZhen:
		return ganzhi.ZhiMao // 卯
	case PalaceXun:
		return ganzhi.ZhiSi // 巳
	case PalaceQian:
		return ganzhi.ZhiXu // 戌
	case PalaceDui:
		return ganzhi.ZhiYou // 酉
	case PalaceGen:
		return ganzhi.ZhiYin // 寅
	case PalaceLi:
		return ganzhi.ZhiWu // 午
	}
	return ganzhi.ZhiZi
}

// palaceWuxing returns the five-element index of a palace.
func palaceWuxing(p PalaceIndex) int {
	switch p {
	case PalaceKan:
		return wxShui
	case PalaceKun, PalaceZhong, PalaceGen:
		return wxTu
	case PalaceZhen, PalaceXun:
		return wxMu
	case PalaceQian, PalaceDui:
		return wxJin
	case PalaceLi:
		return wxHuo
	}
	return wxTu
}

// starHomePalace returns the home palace index (0-based internal) for a star.
func starHomePalace(s StarIndex) int {
	switch s {
	case StarTianPeng:
		return 0 // 坎
	case StarTianRui:
		return 1 // 坤
	case StarTianChong:
		return 2 // 震
	case StarTianFu:
		return 3 // 巽
	case StarTianQin:
		return 4 // 中
	case StarTianXin:
		return 5 // 乾
	case StarTianZhu:
		return 6 // 兑
	case StarTianRen:
		return 7 // 艮
	case StarTianYing:
		return 8 // 离
	}
	return 4
}
