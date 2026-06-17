package ganzhi

// Ten god type constants.
const (
	tgBiJian    = 0 // 比肩
	tgJieCai    = 1 // 劫财
	tgShiShen   = 2 // 食神
	tgShangGuan = 3 // 伤官
	tgPianCai   = 4 // 偏财
	tgZhengCai  = 5 // 正财
	tgQiSha     = 6 // 七杀
	tgZhengGuan = 7 // 正官
	tgPianYin   = 8 // 偏印
	tgZhengYin  = 9 // 正印
)

var tenGodNamesZH = [10]string{
	"比肩", "劫财", "食神", "伤官", "偏财",
	"正财", "七杀", "正官", "偏印", "正印",
}

// TenGodFromGan returns the ten god name for another stem relative to the day master.
func TenGodFromGan(dayMaster, other Gan) string {
	dmElem := GanWuxing(dayMaster)
	otherElem := GanWuxing(other)
	dmYY := GanYinYang(dayMaster)
	otherYY := GanYinYang(other)
	return TenGodName(TenGodType(dmElem, dmYY, otherElem, otherYY))
}

// TenGodName returns the Chinese name for a ten god type.
func TenGodName(tg int) string {
	if tg >= 0 && tg < 10 {
		return tenGodNamesZH[tg]
	}
	return ""
}

// TenGodType classifies the ten-god relationship between day master and another stem.
func TenGodType(dmElem Wuxing, dmYY YinYang, otherElem Wuxing, otherYY YinYang) int {
	switch {
	case dmElem == otherElem:
		if dmYY == otherYY {
			return tgBiJian
		}
		return tgJieCai
	case Sheng(dmElem, otherElem):
		if dmYY == otherYY {
			return tgShiShen
		}
		return tgShangGuan
	case Sheng(otherElem, dmElem):
		if dmYY == otherYY {
			return tgPianYin
		}
		return tgZhengYin
	case Ke(dmElem, otherElem):
		if dmYY == otherYY {
			return tgPianCai
		}
		return tgZhengCai
	default:
		if dmYY == otherYY {
			return tgQiSha
		}
		return tgZhengGuan
	}
}
