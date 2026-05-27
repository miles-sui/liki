package bazi

import "github.com/25types/25types/internal/ganzhi"

// ShenShaEntry describes a single shensha hit on a pillar.
type ShenShaEntry struct {
	Name        string `json:"name"`
	Category    string `json:"category"` // "吉"|"凶"|"中性"
	Description string `json:"description"`
}

// Package-level lookup maps, shared by ComputeShenSha, ComputeDynamicShenSha, and ComputeLiuri.
var (
	taohuaBranchMap = map[int]int{
		3: 4, 7: 4, 11: 4, // 寅午戌→卯
		6: 7, 10: 7, 2: 7,  // 巳酉丑→午
		9: 10, 1: 10, 5: 10, // 申子辰→酉
		12: 1, 4: 1, 8: 1,   // 亥卯未→子
	}
	yimaBranchMap = map[int]int{
		3: 9, 7: 9, 11: 9,   // 寅午戌→申
		6: 12, 10: 12, 2: 12, // 巳酉丑→亥
		9: 3, 1: 3, 5: 3,     // 申子辰→寅
		12: 6, 4: 6, 8: 6,    // 亥卯未→巳
	}
	huagaiBranchMap = map[int]int{
		3: 11, 7: 11, 11: 11, // 寅午戌→戌
		6: 2, 10: 2, 2: 2,    // 巳酉丑→丑
		9: 5, 1: 5, 5: 5,     // 申子辰→辰
		12: 8, 4: 8, 8: 8,    // 亥卯未→未
	}
	yangRenLookup = map[int]int{1: 4, 2: 3, 3: 7, 4: 6, 5: 7, 6: 6, 7: 10, 8: 9, 9: 1, 10: 12}
	jieshaBranch  = map[int]int{
		3: 6, 7: 6, 11: 6,   // 寅午戌→巳
		6: 3, 10: 3, 2: 3,    // 巳酉丑→寅
		9: 12, 1: 12, 5: 12,  // 申子辰→亥
		12: 9, 4: 9, 8: 9,    // 亥卯未→申
	}
	zaishaBranch = map[int]int{
		3: 7, 7: 7, 11: 7,   // 寅午戌→午
		6: 10, 10: 10, 2: 10, // 巳酉丑→酉
		9: 1, 1: 1, 5: 1,    // 申子辰→子
		12: 4, 4: 4, 8: 4,   // 亥卯未→卯
	}
	hongluanLookup = map[int]int{
		1: 4, 2: 3, 3: 2, 4: 1, 5: 12, 6: 11,
		7: 10, 8: 9, 9: 8, 10: 7, 11: 6, 12: 5,
	}
	tianxiLookup = map[int]int{
		1: 10, 2: 9, 3: 8, 4: 7, 5: 6, 6: 5,
		7: 4, 8: 3, 9: 2, 10: 1, 11: 12, 12: 11,
	}
)

var tianYiLookup = map[int][]int{
	1: {2, 8}, 2: {1, 9}, 3: {12, 10}, 4: {12, 10}, 5: {2, 8},
	6: {1, 9}, 7: {7, 3}, 8: {7, 3}, 9: {4, 6}, 10: {4, 6},
}

var (
	tiandeStems = map[int][]int{
		1: {4}, 2: {7}, 3: {9}, 4: {8}, 5: {9}, 6: {1},
		7: {10}, 8: {1}, 9: {3}, 10: {2}, 11: {3}, 12: {7},
	}
	yuedeStem = map[int]int{
		3: 3, 7: 3, 11: 3, 12: 1, 4: 1, 8: 1,
		9: 9, 1: 9, 5: 9, 6: 7, 10: 7, 2: 7,
	}
	jiangxingLookup = map[int]int{
		3: 7, 7: 7, 11: 7, 6: 10, 10: 10, 2: 10,
		9: 1, 1: 1, 5: 1, 12: 4, 4: 4, 8: 4,
	}
	jinyuLookup = map[int][]int{
		1: {5}, 2: {6}, 3: {7}, 4: {8}, 5: {5},
		6: {6}, 7: {9}, 8: {10}, 9: {12}, 10: {12},
	}
	yueEnStems = map[int][]int{
		1: {3}, 2: {4}, 3: {5}, 4: {7}, 5: {8}, 6: {3},
		7: {5}, 8: {7}, 9: {9}, 10: {10}, 11: {9}, 12: {10},
	}
	xueRenLookup = map[int]int{1: 4, 2: 3, 3: 7, 4: 6, 5: 5, 6: 5, 7: 10, 8: 9, 9: 1, 10: 12}
	tianLuoDiWang = map[int]string{11: "天罗", 12: "天罗", 5: "地网", 6: "地网"}
	shiEDaBai    = map[int]struct{}{
		8: {}, 16: {}, 17: {}, 23: {}, 25: {}, 32: {}, 34: {}, 40: {}, 41: {}, 59: {},
	}
)

// ComputeShenSha computes all shensha for the bazi chart, grouped by pillar.
func ComputeShenSha(bz ganzhi.Bazi, dayMaster Stem, monthBranch Branch) [4][]ShenShaEntry {
	pillars := bz.Slice()
	var out [4][]ShenShaEntry
	branches := [4]int{int(pillars[0].Branch), int(pillars[1].Branch), int(pillars[2].Branch), int(pillars[3].Branch)}
	seasonIdx := (int(monthBranch) - 1) / 3
	yearBranch := int(pillars[0].Branch)

	addTianYi(&out, bz, dayMaster, pillars[0].Stem)
	addWenChang(&out, bz, dayMaster)
	addXueTang(&out, bz, dayMaster)
	addLuShen(&out, bz, dayMaster)
	addYangRen(&out, bz, dayMaster)
	addTianDe(&out, bz, monthBranch)
	addYueDe(&out, bz, monthBranch)
	addTaoHua(&out, bz, branches)
	addYiMa(&out, bz, branches)
	addHuaGai(&out, bz, branches)
	addJiangXing(&out, bz, branches)
	addJieSha(&out, bz, branches)
	addZaiSha(&out, bz, branches)
	addGuChenGuaSu(&out, bz, seasonIdx)
	addHongLuanTianXi(&out, bz, yearBranch)
	addJinYu(&out, bz, dayMaster)
	addCiGuan(&out, bz, dayMaster)
	addYueEn(&out, bz, monthBranch)
	addTianShe(&out, bz, monthBranch)
	addTianLuoDiWang(&out, bz)
	addGouJiao(&out, bz, yearBranch)
	addYuanChen(&out, bz, yearBranch)
	addXueRen(&out, bz, dayMaster)
	addSiFei(&out, bz, seasonIdx)
	addShiEDaBai(&out, bz)

	return out
}

func addTianYi(out *[4][]ShenShaEntry, bz ganzhi.Bazi, dayMaster, yearStem Stem) {
	appendShenShaByStemLookup(out, bz, dayMaster, tianYiLookup, "天乙贵人", "吉", "主贵人相助，逢凶化吉")
	appendShenShaByStemLookup(out, bz, yearStem, tianYiLookup, "天乙贵人", "吉", "主贵人相助，逢凶化吉")
}

var wenChangLookup = map[int][]int{
	1: {6}, 2: {7}, 3: {6}, 4: {7}, 5: {9},
	6: {10}, 7: {12}, 8: {1}, 9: {3}, 10: {4},
}

func addWenChang(out *[4][]ShenShaEntry, bz ganzhi.Bazi, dayMaster Stem) {
	appendShenShaByStemLookup(out, bz, dayMaster, wenChangLookup, "文昌", "吉", "主学业、文书、才华")
}

func addXueTang(out *[4][]ShenShaEntry, bz ganzhi.Bazi, dayMaster Stem) {
	addLifeStageShenSha(out, bz, dayMaster, 0, "学堂", "吉", "日主长生之位，主学业聪颖")
}

func addLuShen(out *[4][]ShenShaEntry, bz ganzhi.Bazi, dayMaster Stem) {
	addLifeStageShenSha(out, bz, dayMaster, 3, "禄神", "吉", "日主临官之位，主福禄安康")
}

func addCiGuan(out *[4][]ShenShaEntry, bz ganzhi.Bazi, dayMaster Stem) {
	addLifeStageShenSha(out, bz, dayMaster, 3, "词馆", "吉", "主文章、口才、文职")
}

func addLifeStageShenSha(out *[4][]ShenShaEntry, bz ganzhi.Bazi, dayMaster Stem, stageIdx int, name, cat, desc string) {
	pillars := bz.Slice()
	stageRow := defaultEngine.LifeStagesTable[int(dayMaster)]
	if len(stageRow) != 12 {
		return
	}
	for pi, p := range pillars {
		if bn := int(p.Branch); bn >= 1 && bn <= 12 && stageRow[stageIdx] == bn {
			(*out)[pi] = append((*out)[pi], ShenShaEntry{Name: name, Category: cat, Description: desc})
		}
	}
}

func addYangRen(out *[4][]ShenShaEntry, bz ganzhi.Bazi, dayMaster Stem) {
	pillars := bz.Slice()
	for pi, p := range pillars {
		if yangRenLookup[int(dayMaster)] == int(p.Branch) {
			(*out)[pi] = append((*out)[pi], ShenShaEntry{
				Name: "羊刃", Category: "凶", Description: "日干帝旺/刃位，主刚强果断，但易冲动",
			})
		}
	}
}

func addTianDe(out *[4][]ShenShaEntry, bz ganzhi.Bazi, monthBranch Branch) {
	pillars := bz.Slice()
	targets, ok := tiandeStems[int(monthBranch)]
	if !ok {
		return
	}
	for _, ts := range targets {
		for pi, p := range pillars {
			if int(p.Stem) == ts {
				(*out)[pi] = append((*out)[pi], ShenShaEntry{
					Name: "天德", Category: "吉", Description: "天德贵人，主福泽深厚，化险为夷",
				})
			}
		}
	}
}

func addYueDe(out *[4][]ShenShaEntry, bz ganzhi.Bazi, monthBranch Branch) {
	pillars := bz.Slice()
	targetStem, ok := yuedeStem[int(monthBranch)]
	if !ok {
		return
	}
	for pi, p := range pillars {
		if int(p.Stem) == targetStem {
			(*out)[pi] = append((*out)[pi], ShenShaEntry{
				Name: "月德", Category: "吉", Description: "月德贵人，主月令之德，人缘佳",
			})
		}
	}
}

func addTriadShenSha(out *[4][]ShenShaEntry, bz ganzhi.Bazi, branches [4]int, lookup map[int]int, name, cat, desc string) {
	pillars := bz.Slice()
	for _, refIdx := range []int{0, 2} { // year & day branch
		if tb, ok := lookup[branches[refIdx]]; ok {
			for pi, p := range pillars {
				if int(p.Branch) == tb {
					(*out)[pi] = append((*out)[pi], ShenShaEntry{Name: name, Category: cat, Description: desc})
				}
			}
		}
	}
}

func addTaoHua(out *[4][]ShenShaEntry, bz ganzhi.Bazi, branches [4]int) {
	addTriadShenSha(out, bz, branches, taohuaBranchMap, "桃花", "中性", "主异性缘佳，浪漫多情")
}

func addYiMa(out *[4][]ShenShaEntry, bz ganzhi.Bazi, branches [4]int) {
	addTriadShenSha(out, bz, branches, yimaBranchMap, "驿马", "中性", "主动荡、奔波、迁移")
}

func addHuaGai(out *[4][]ShenShaEntry, bz ganzhi.Bazi, branches [4]int) {
	addTriadShenSha(out, bz, branches, huagaiBranchMap, "华盖", "中性", "主孤独清高，聪明好学，有艺术天赋")
}

func addJiangXing(out *[4][]ShenShaEntry, bz ganzhi.Bazi, branches [4]int) {
	addTriadShenSha(out, bz, branches, jiangxingLookup, "将星", "吉", "主领导才能，有权威")
}

func addJieSha(out *[4][]ShenShaEntry, bz ganzhi.Bazi, branches [4]int) {
	addTriadShenSha(out, bz, branches, jieshaBranch, "劫煞", "凶", "主破财、意外、是非")
}

func addZaiSha(out *[4][]ShenShaEntry, bz ganzhi.Bazi, branches [4]int) {
	addTriadShenSha(out, bz, branches, zaishaBranch, "灾煞", "凶", "主灾祸、疾病、横事")
}

func addGuChenGuaSu(out *[4][]ShenShaEntry, bz ganzhi.Bazi, seasonIdx int) {
	pillars := bz.Slice()
	guchenBranches := [4]int{6, 9, 12, 3}
	guasuBranches := [4]int{2, 5, 8, 11}
	for pi, p := range pillars {
		if int(p.Branch) == guchenBranches[seasonIdx] {
			(*out)[pi] = append((*out)[pi], ShenShaEntry{
				Name: "孤辰", Category: "凶", Description: "主性格孤僻，晚婚或婚姻不顺",
			})
		}
		if int(p.Branch) == guasuBranches[seasonIdx] {
			(*out)[pi] = append((*out)[pi], ShenShaEntry{
				Name: "寡宿", Category: "凶", Description: "主孤独寂寞，夫妻缘薄",
			})
		}
	}
}

func addHongLuanTianXi(out *[4][]ShenShaEntry, bz ganzhi.Bazi, yearBranch int) {
	pillars := bz.Slice()
	if target, ok := hongluanLookup[yearBranch]; ok {
		for pi, p := range pillars {
			if int(p.Branch) == target {
				(*out)[pi] = append((*out)[pi], ShenShaEntry{
					Name: "红鸾", Category: "吉", Description: "主婚喜、恋爱、添丁",
				})
			}
		}
	}
	if target, ok := tianxiLookup[yearBranch]; ok {
		for pi, p := range pillars {
			if int(p.Branch) == target {
				(*out)[pi] = append((*out)[pi], ShenShaEntry{
					Name: "天喜", Category: "吉", Description: "主喜庆之事，婚恋吉兆",
				})
			}
		}
	}
}

func addJinYu(out *[4][]ShenShaEntry, bz ganzhi.Bazi, dayMaster Stem) {
	appendShenShaByStemLookup(out, bz, dayMaster, jinyuLookup, "金舆", "吉", "主财运、车辆、出行顺利")
}

func addYueEn(out *[4][]ShenShaEntry, bz ganzhi.Bazi, monthBranch Branch) {
	pillars := bz.Slice()
	targets, ok := yueEnStems[int(monthBranch)]
	if !ok {
		return
	}
	for _, ts := range targets {
		for pi, p := range pillars {
			if int(p.Stem) == ts {
				(*out)[pi] = append((*out)[pi], ShenShaEntry{
					Name: "月恩", Category: "吉", Description: "月令之恩，主福佑加持",
				})
			}
		}
	}
}

func addTianShe(out *[4][]ShenShaEntry, bz ganzhi.Bazi, monthBranch Branch) {
	pillars := bz.Slice()
	season := (int(monthBranch) - 1) / 3
	tianSheChecks := [4][2]int{{5, 3}, {1, 7}, {5, 9}, {1, 1}} // 戊寅, 甲午, 戊申, 甲子
	if season >= 0 && season < 4 {
		ds, db := tianSheChecks[season][0], tianSheChecks[season][1]
		if int(pillars[2].Stem) == ds && int(pillars[2].Branch) == db {
			(*out)[2] = append((*out)[2], ShenShaEntry{
				Name: "天赦", Category: "吉", Description: "天赦日出生，主逢凶化吉，宽恕赦免",
			})
		}
	}
}

func addTianLuoDiWang(out *[4][]ShenShaEntry, bz ganzhi.Bazi) {
	pillars := bz.Slice()
	for pi, p := range pillars {
		if label, ok := tianLuoDiWang[int(p.Branch)]; ok {
			(*out)[pi] = append((*out)[pi], ShenShaEntry{
				Name: label, Category: "凶", Description: "主运势阻滞，有志难伸",
			})
		}
	}
}

func addGouJiao(out *[4][]ShenShaEntry, bz ganzhi.Bazi, yearBranch int) {
	pillars := bz.Slice()
	gouShen := (yearBranch+2)%12 + 1
	jiaoShen := (yearBranch+4)%12 + 1
	for pi, p := range pillars {
		if int(p.Branch) == gouShen {
			(*out)[pi] = append((*out)[pi], ShenShaEntry{
				Name: "勾神", Category: "凶", Description: "主纠缠牵连，是非官讼",
			})
		}
		if int(p.Branch) == jiaoShen {
			(*out)[pi] = append((*out)[pi], ShenShaEntry{
				Name: "绞神", Category: "凶", Description: "主受困被缚，身不由己",
			})
		}
	}
}

func addYuanChen(out *[4][]ShenShaEntry, bz ganzhi.Bazi, yearBranch int) {
	pillars := bz.Slice()
	ycBranch := yuanChenBranch(yearBranch)
	for pi, p := range pillars {
		if int(p.Branch) == ycBranch {
			(*out)[pi] = append((*out)[pi], ShenShaEntry{
				Name: "元辰", Category: "凶", Description: "主波折反复，好事多磨",
			})
		}
	}
}

func addXueRen(out *[4][]ShenShaEntry, bz ganzhi.Bazi, dayMaster Stem) {
	pillars := bz.Slice()
	for pi, p := range pillars {
		if xueRenLookup[int(dayMaster)] == int(p.Branch) {
			(*out)[pi] = append((*out)[pi], ShenShaEntry{
				Name: "血刃", Category: "凶", Description: "主意外血光，手术外伤",
			})
		}
	}
}

func addSiFei(out *[4][]ShenShaEntry, bz ganzhi.Bazi, seasonIdx int) {
	pillars := bz.Slice()
	siFeiPillars := [4][][2]int{
		{{7, 9}, {8, 10}}, {{9, 1}, {10, 12}}, {{1, 3}, {2, 4}}, {{3, 7}, {4, 6}},
	}
	if seasonIdx < 0 || seasonIdx >= 4 {
		return
	}
	for pi, p := range pillars {
		for _, pair := range siFeiPillars[seasonIdx] {
			if int(p.Stem) == pair[0] && int(p.Branch) == pair[1] {
				(*out)[pi] = append((*out)[pi], ShenShaEntry{
					Name: "四废", Category: "凶", Description: "四季废日，主事业阻滞，有志难伸",
				})
			}
		}
	}
}

func addShiEDaBai(out *[4][]ShenShaEntry, bz ganzhi.Bazi) {
	pillars := bz.Slice()
	if _, ok := shiEDaBai[sixtyCycleName(pillars[2].Stem, pillars[2].Branch)]; ok {
		(*out)[2] = append((*out)[2], ShenShaEntry{
			Name: "十恶大败", Category: "凶", Description: "日柱十恶大败日，主财库不聚，须谨慎理财",
		})
	}
}

func appendShenShaByStemLookup(out *[4][]ShenShaEntry, bz ganzhi.Bazi, s Stem, lookup map[int][]int, name, cat, desc string) {
	pillars := bz.Slice()
	targets, ok := lookup[int(s)]
	if !ok {
		return
	}
	for pi, p := range pillars {
		for _, t := range targets {
			if int(p.Branch) == t {
				(*out)[pi] = append((*out)[pi], ShenShaEntry{Name: name, Category: cat, Description: desc})
			}
		}
	}
}

// ComputeKongWang returns pillar indices whose branches fall in the void (空亡)
// of the day pillar's 旬.
func ComputeKongWang(dayPillar Pillar, bz Bazi) []int {
	sbIdx := sixtyCycleName(dayPillar.Stem, dayPillar.Branch)
	xunIdx := sbIdx / 10

	voidPairs := [6][2]int{
		{11, 12}, {9, 10}, {7, 8}, {5, 6}, {3, 4}, {1, 2},
	}
	v1, v2 := voidPairs[xunIdx][0], voidPairs[xunIdx][1]

	var hits []int
	for pi, p := range bz.Slice() {
		b := int(p.Branch)
		if b == v1 || b == v2 {
			hits = append(hits, pi)
		}
	}
	return hits
}

// ComputeDynamicShenSha computes shensha triggered by an external branch against the bazi chart.
func ComputeDynamicShenSha(b Branch, yearBranch Branch, dayMaster Stem) []ShenShaEntry {
	var result []ShenShaEntry
	bi := int(b)
	yb := int(yearBranch)

	if tb, ok := taohuaBranchMap[yb]; ok && tb == bi {
		result = append(result, ShenShaEntry{Name: "桃花", Category: "中性", Description: "流运桃花，异性缘佳"})
	}
	if tb, ok := yimaBranchMap[yb]; ok && tb == bi {
		result = append(result, ShenShaEntry{Name: "驿马", Category: "中性", Description: "流运驿马，动象奔波"})
	}
	if tb, ok := huagaiBranchMap[yb]; ok && tb == bi {
		result = append(result, ShenShaEntry{Name: "华盖", Category: "中性", Description: "流运华盖，宜静思"})
	}
	if targets, ok := tianYiLookup[int(dayMaster)]; ok {
		for _, t := range targets {
			if t == bi {
				result = append(result, ShenShaEntry{Name: "天乙贵人", Category: "吉", Description: "流运天乙贵人，有贵人相助"})
				break
			}
		}
	}
	if yr, ok := yangRenLookup[int(dayMaster)]; ok && yr == bi {
		result = append(result, ShenShaEntry{Name: "羊刃", Category: "凶", Description: "流运羊刃，防冲动冲突"})
	}
	if js, ok := jieshaBranch[yb]; ok && js == bi {
		result = append(result, ShenShaEntry{Name: "劫煞", Category: "凶", Description: "流运劫煞，防破财是非"})
	}
	if zs, ok := zaishaBranch[yb]; ok && zs == bi {
		result = append(result, ShenShaEntry{Name: "灾煞", Category: "凶", Description: "流运灾煞，防意外灾祸"})
	}
	if hl, ok := hongluanLookup[yb]; ok && hl == bi {
		result = append(result, ShenShaEntry{Name: "红鸾", Category: "吉", Description: "流运红鸾，主婚喜添丁"})
	}
	if tx, ok := tianxiLookup[yb]; ok && tx == bi {
		result = append(result, ShenShaEntry{Name: "天喜", Category: "吉", Description: "流运天喜，喜庆之事"})
	}

	return result
}

func yuanChenBranch(yearBranch int) int {
	for _, p := range defaultData.ChongPairs {
		if int(p.A) == yearBranch {
			return (int(p.B) % 12) + 1
		}
		if int(p.B) == yearBranch {
			return (int(p.A) % 12) + 1
		}
	}
	return 0
}
