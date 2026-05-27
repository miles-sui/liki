package bazi

import "github.com/25types/25types/internal/ganzhi"

// IsSelfHe returns true if the pillar's stem and its branch's main hidden stem
// form a 天干五合 (stem-he harmony within a single pillar — 干支自合).
//
// Traditional pairs:
//
//	甲午 (甲+午中己→甲己合), 乙巳 (乙+巳中庚→乙庚合), 丙戌 (丙+戌中辛→丙辛合),
//	丁亥 (丁+亥中壬→丁壬合), 戊子 (戊+子中癸→戊癸合),
//	庚辰 (庚+辰中乙→乙庚合), 辛巳 (辛+巳中丙→丙辛合),
//	壬戌 (壬+戌中丁→丁壬合), 癸巳 (癸+巳中戊→戊癸合).
func IsSelfHe(p Pillar) bool {
	hs := HiddenStemsForBranch(p.Branch)
	for _, h := range hs.Slice() {
		if h != nil && isStemHePair(int(p.Stem), *h) {
			return true
		}
	}
	return false
}

// SelfHeName returns the 干支自合 description string (e.g. "甲己合").
func SelfHeName(p Pillar) string {
	hs := HiddenStemsForBranch(p.Branch)
	for _, h := range hs.Slice() {
		if h != nil && isStemHePair(int(p.Stem), *h) {
			return stemNameStr(p.Stem) + stemNameStr(Stem(*h)) + "合"
		}
	}
	return ""
}

func isStemHePair(a, b int) bool { return ganzhi.IsStemHe(Stem(a), Stem(b)) }

// IsKuiGang checks if the pillar is a 魁罡 day pillar.
// 魁罡: 庚辰, 庚戌, 壬辰, 戊戌.
func IsKuiGang(p Pillar) bool {
	s, b := int(p.Stem), int(p.Branch)
	return (s == 7 && b == 5) || // 庚辰
		(s == 7 && b == 11) || // 庚戌
		(s == 9 && b == 5) || // 壬辰
		(s == 5 && b == 11) // 戊戌
}

// SanQiType checks if the four-pillar stem set contains a 三奇贵人 pattern.
// Returns "天上" (甲戊庚), "地下" (乙丙丁), or "人中" (壬癸辛), or empty.
func SanQiType(bz ganzhi.Bazi) string {
	pillars := bz.Slice()
	stemSet := [11]bool{}
	for _, p := range pillars {
		if s := int(p.Stem); s >= 1 && s <= 10 {
			stemSet[s] = true
		}
	}

	// 天上三奇: 甲(1)+戊(5)+庚(7)
	if stemSet[1] && stemSet[5] && stemSet[7] {
		return "天上"
	}
	// 地下三奇: 乙(2)+丙(3)+丁(4)
	if stemSet[2] && stemSet[3] && stemSet[4] {
		return "地下"
	}
	// 人中三奇: 壬(9)+癸(10)+辛(8)
	if stemSet[9] && stemSet[10] && stemSet[8] {
		return "人中"
	}
	return ""
}

// SanQiName returns the full name for a sanqi type code.
func SanQiName(typ string) string {
	switch typ {
	case "天上":
		return "天上三奇（甲戊庚）"
	case "地下":
		return "地下三奇（乙丙丁）"
	case "人中":
		return "人中三奇（壬癸辛）"
	}
	return ""
}
