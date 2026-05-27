package bazi

import (
	"github.com/25types/25types/internal/ganzhi"
)

// ---- Bond domain types ----

// PillarPairEntry is one cross-chart pillar pair.
type PillarPairEntry struct {
	APillar   string `json:"a_pillar"`
	BPillar   string `json:"b_pillar"`
	AStem     string `json:"a_stem"`
	BStem     string `json:"b_stem"`
	ABranch   string `json:"a_branch"`
	BBranch   string `json:"b_branch"`
	Stem      StemRelation  `json:"stem"`
	Branch    BranchRelation `json:"branch"`
}

// PillarCross holds all 16 pillar-pair interactions.
type PillarCross struct {
	Pairs []PillarPairEntry `json:"pairs"`
}

// TenGodCross holds mutual ten-god perspective.
type TenGodCross struct {
	AToB map[string]string `json:"a_to_b"`
	BToA map[string]string `json:"b_to_a"`
}

// NayinPairEntry is one cross-chart nayin pair.
type NayinPairEntry struct {
	APillar  string `json:"a_pillar"`
	BPillar  string `json:"b_pillar"`
	ANaYin   string `json:"a_nayin"`
	BNaYin   string `json:"b_nayin"`
	Relation string `json:"relation"`
}

// YongshenEntry holds one side of the yong/ji cross.
type YongshenEntry struct {
	Yong     string `json:"yong"`
	Ji       string `json:"ji"`
	YongInOther int `json:"yong_in_other"`
	JiInOther   int `json:"ji_in_other"`
}

// NayinCross holds nayin pairs, element counts, and yongshen cross.
type NayinCross struct {
	Pairs    []NayinPairEntry `json:"pairs"`
	Elements struct {
		A map[string]int `json:"a"`
		B map[string]int `json:"b"`
	} `json:"elements"`
	Yongshen struct {
		A YongshenEntry `json:"a"`
		B YongshenEntry `json:"b"`
	} `json:"yongshen"`
}

// ShenshaMutual is one mutual shensha check.
type ShenshaMutual struct {
	AInB bool `json:"a_in_b"`
	BInA bool `json:"b_in_a"`
}

// ShenshaCross holds mutual shensha occurrence.
type ShenshaCross struct {
	TianYi   ShenshaMutual `json:"tianyi"`
	Lu       ShenshaMutual `json:"lu"`
	TaoHua   ShenshaMutual `json:"taohua"`
	YiMa     ShenshaMutual `json:"yima"`
	KongWang ShenshaMutual `json:"kongwang"`
	KuiGang  ShenshaMutual `json:"kuigang"`
	RiDe     ShenshaMutual `json:"ride"`
	RiGui    ShenshaMutual `json:"rigui"`
}

// PillarStemBranch is a simple stem+branch pair.
type PillarStemBranch struct {
	Stem   int `json:"stem"`
	Branch int `json:"branch"`
}

// TYMGCrossEntry holds the cross for one of taiyuan/minggong/shengong.
type TYMGCrossEntry struct {
	Stem   StemRelation  `json:"stem"`
	Branch BranchRelation `json:"branch"`
}

// TYMGPerson holds the three extra pillars for one person.
type TYMGPerson struct {
	TaiYuan  PillarStemBranch `json:"tai_yuan"`
	MingGong PillarStemBranch `json:"ming_gong"`
	ShenGong PillarStemBranch `json:"shen_gong"`
}

// TaiYuanMingGongStruct holds per-person pillars plus cross analysis.
type TaiYuanMingGongStruct struct {
	A     TYMGPerson `json:"a"`
	B     TYMGPerson `json:"b"`
	Cross struct {
		TaiYuan  TYMGCrossEntry `json:"tai_yuan"`
		MingGong TYMGCrossEntry `json:"ming_gong"`
		ShenGong TYMGCrossEntry `json:"shen_gong"`
	} `json:"cross"`
}

// DayunCrossEntry describes the current dayun pillar for one person.
type DayunCrossEntry struct {
	Stem   int    `json:"stem"`
	Branch int    `json:"branch"`
	Name   string `json:"name"`
	TenGod string `json:"ten_god"`
}

// DayunCross holds current dayun for both people and their relation.
type DayunCross struct {
	ACurrent  DayunCrossEntry `json:"a_current"`
	BCurrent  DayunCrossEntry `json:"b_current"`
	StemRel   StemRelation    `json:"stem_rel"`
	BranchRel BranchRelation  `json:"branch_rel"`
}

// XunGong holds same-xun and same-gong checks.
type XunGong struct {
	SameXun bool `json:"same_xun"`
	SameGong bool `json:"same_gong"`
}

// StructureCross holds all structural cross comparisons.
type StructureCross struct {
	TaiYuanMingGong TaiYuanMingGongStruct `json:"tai_yuan_ming_gong"`
	Dayun           DayunCross            `json:"dayun"`
	XunGong         XunGong               `json:"xun_gong"`
}

// BondResult holds the complete cross-chart bond analysis.
type BondResult struct {
	PillarCross  PillarCross    `json:"pillar_cross"`
	TenGodCross  TenGodCross    `json:"ten_god_cross"`
	NayinCross   NayinCross     `json:"nayin_cross"`
	ShenshaCross ShenshaCross   `json:"shensha_cross"`
	Structure    StructureCross `json:"structure"`
}

// ---- Main entry point ----

var pillarNames = [4]string{"year", "month", "day", "hour"}

// ComputeBond computes the full cross-chart bond analysis.
func ComputeBond(a, b ChartResult, aBirthYear, aBirthMonth, aBirthHour, bBirthYear, bBirthMonth, bBirthHour int) BondResult {
	return BondResult{
		PillarCross:  computePillarCross(a, b),
		TenGodCross:  computeTenGodCross(a, b),
		NayinCross:   computeNayinCross(a, b),
		ShenshaCross: computeShenshaCross(a, b),
		Structure:    computeStructureCross(a, b, aBirthYear, aBirthMonth, aBirthHour, bBirthYear, bBirthMonth, bBirthHour),
	}
}

// ---- 一、pillar_cross ----

func computePillarCross(a, b ChartResult) PillarCross {
	aStems := [4]Stem{a.Year.Stem, a.Month.Stem, a.Day.Stem, a.Hour.Stem}
	bStems := [4]Stem{b.Year.Stem, b.Month.Stem, b.Day.Stem, b.Hour.Stem}
	aBranches := [4]Branch{a.Year.Branch, a.Month.Branch, a.Day.Branch, a.Hour.Branch}
	bBranches := [4]Branch{b.Year.Branch, b.Month.Branch, b.Day.Branch, b.Hour.Branch}

	pairs := make([]PillarPairEntry, 0, 16)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			pairs = append(pairs, PillarPairEntry{
				APillar: pillarNames[i],
				BPillar: pillarNames[j],
				AStem:   stemNameStr(aStems[i]),
				BStem:   stemNameStr(bStems[j]),
				ABranch: branchNameStr(aBranches[i]),
				BBranch: branchNameStr(bBranches[j]),
				Stem:    AnalyzeStemRelation(aStems[i], bStems[j]),
				Branch:  AnalyzeBranchRelation(aBranches[i], bBranches[j]),
			})
		}
	}
	return PillarCross{Pairs: pairs}
}

// ---- 二、ten_god_cross ----

func computeTenGodCross(a, b ChartResult) TenGodCross {
	aStems := [4]Stem{a.Year.Stem, a.Month.Stem, a.Day.Stem, a.Hour.Stem}
	bStems := [4]Stem{b.Year.Stem, b.Month.Stem, b.Day.Stem, b.Hour.Stem}

	aElem, aYY := StemElement(a.DayMaster), StemYinYang(a.DayMaster)
	bElem, bYY := StemElement(b.DayMaster), StemYinYang(b.DayMaster)

	aToB := make(map[string]string, 4)
	bToA := make(map[string]string, 4)
	for i := 0; i < 4; i++ {
		aToB[pillarNames[i]+"_stem"] = TenGodName(TenGodType(aElem, aYY, StemElement(bStems[i]), StemYinYang(bStems[i])))
		bToA[pillarNames[i]+"_stem"] = TenGodName(TenGodType(bElem, bYY, StemElement(aStems[i]), StemYinYang(aStems[i])))
	}

	return TenGodCross{AToB: aToB, BToA: bToA}
}

// ---- 三、nayin_cross ----

func computeNayinCross(a, b ChartResult) NayinCross {
	aNayin := a.NaYinArray()
	bNayin := b.NaYinArray()

	pairs := make([]NayinPairEntry, 0, 16)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			aElem := nayinElement(aNayin[i])
			bElem := nayinElement(bNayin[j])
			rel := "相同"
			if aElem != bElem {
				if ganzhi.Sheng(aElem, bElem) || ganzhi.Sheng(bElem, aElem) {
					rel = "相生"
				} else if ganzhi.Ke(aElem, bElem) || ganzhi.Ke(bElem, aElem) {
					rel = "相克"
				}
			}
			pairs = append(pairs, NayinPairEntry{
				APillar:  pillarNames[i],
				BPillar:  pillarNames[j],
				ANaYin:   aNayin[i],
				BNaYin:   bNayin[j],
				Relation: rel,
			})
		}
	}

	// Element counts
	aElemCount := make(map[string]int, 5)
	bElemCount := make(map[string]int, 5)
	for _, name := range []string{"木", "火", "土", "金", "水"} {
		aElemCount[name] = 0
		bElemCount[name] = 0
	}
	for e, c := range a.ElementCount {
		aElemCount[e.String()] = c
	}
	for e, c := range b.ElementCount {
		bElemCount[e.String()] = c
	}

	// Yongshen cross
	aYS := ComputeYongShen(a)
	bYS := ComputeYongShen(b)

	nc := NayinCross{Pairs: pairs}
	nc.Elements.A = aElemCount
	nc.Elements.B = bElemCount
	nc.Yongshen.A = makeYongshenEntry(aYS.FuYi.Yong, aYS.FuYi.Ji, b.ElementCount)
	nc.Yongshen.B = makeYongshenEntry(bYS.FuYi.Yong, bYS.FuYi.Ji, a.ElementCount)

	return nc
}

func makeYongshenEntry(yong, ji string, otherCount map[Element]int) YongshenEntry {
	elem := elementFromChinese(yong)
	yongIn := otherCount[elem]
	jiElem := elementFromChinese(ji)
	jiIn := otherCount[jiElem]
	return YongshenEntry{Yong: yong, Ji: ji, YongInOther: yongIn, JiInOther: jiIn}
}

// ---- 四、shensha_cross ----

func computeShenshaCross(a, b ChartResult) ShenshaCross {
	aBz, bBz := a.ToBazi(), b.ToBazi()
	return ShenshaCross{
		TianYi:   mutualShensha(tianYiBranches(a), tianYiBranches(b), aBz, bBz),
		Lu:       mutualShensha(luBranches(a), luBranches(b), aBz, bBz),
		TaoHua:   mutualShensha(taohuaBranches(a), taohuaBranches(b), aBz, bBz),
		YiMa:     mutualShensha(yimaBranches(a), yimaBranches(b), aBz, bBz),
		KongWang: mutualShensha(kongwangBranches(a), kongwangBranches(b), aBz, bBz),
		KuiGang:  mutualShensha(kuigangBranches(a), kuigangBranches(b), aBz, bBz),
		RiDe:     mutualShensha(rideBranches(a), rideBranches(b), aBz, bBz),
		RiGui:    mutualShensha(riguiBranches(a), riguiBranches(b), aBz, bBz),
	}
}

func mutualShensha(aBranches, bBranches []Branch, aBz, bBz Bazi) ShenshaMutual {
	return ShenshaMutual{
		AInB: branchesInPillars(aBranches, bBz),
		BInA: branchesInPillars(bBranches, aBz),
	}
}

func branchesInPillars(targets []Branch, bz Bazi) bool {
	ps := bz.Slice()
	for _, t := range targets {
		for _, p := range ps {
			if p.Branch == t {
				return true
			}
		}
	}
	return false
}

func tianYiBranches(cr ChartResult) []Branch {
	branches := []Branch{}
	if b, ok := tianYiLookup[int(cr.Year.Stem)]; ok {
		branches = append(branches, Branch(b[0]), Branch(b[1]))
	}
	if b, ok := tianYiLookup[int(cr.DayMaster)]; ok {
		branches = append(branches, Branch(b[0]), Branch(b[1]))
	}
	return branches
}

func luBranches(cr ChartResult) []Branch {
	if int(cr.DayMaster) < 1 || int(cr.DayMaster) > 10 {
		return nil
	}
	return []Branch{Branch(LifeStagesTable[int(cr.DayMaster)][3])}
}

func taohuaBranches(cr ChartResult) []Branch {
	branches := []Branch{}
	if b, ok := taohuaBranchMap[int(cr.Year.Branch)]; ok {
		branches = append(branches, Branch(b))
	}
	if b, ok := taohuaBranchMap[int(cr.Day.Branch)]; ok {
		branches = append(branches, Branch(b))
	}
	return branches
}

func yimaBranches(cr ChartResult) []Branch {
	branches := []Branch{}
	if b, ok := yimaBranchMap[int(cr.Year.Branch)]; ok {
		branches = append(branches, Branch(b))
	}
	if b, ok := yimaBranchMap[int(cr.Day.Branch)]; ok {
		branches = append(branches, Branch(b))
	}
	return branches
}

func kongwangBranches(cr ChartResult) []Branch {
	bz := cr.ToBazi()
	ps := bz.Slice()
	voidHits := ComputeKongWang(Pillar{Stem: cr.Day.Stem, Branch: cr.Day.Branch}, bz)
	branches := make([]Branch, 0, len(voidHits))
	for _, idx := range voidHits {
		branches = append(branches, ps[idx].Branch)
	}
	return branches
}

func kuigangBranches(cr ChartResult) []Branch {
	ps := cr.ToBazi().Slice()
	branches := []Branch{}
	for _, p := range ps {
		if IsKuiGang(p) {
			branches = append(branches, p.Branch)
		}
	}
	return branches
}

func rideBranches(cr ChartResult) []Branch {
	branches := []Branch{}
	for _, p := range cr.ToBazi().Slice() {
		if isRiDe(p) {
			branches = append(branches, p.Branch)
		}
	}
	return branches
}

func riguiBranches(cr ChartResult) []Branch {
	branches := []Branch{}
	for _, p := range cr.ToBazi().Slice() {
		if isRiGui(p) {
			branches = append(branches, p.Branch)
		}
	}
	return branches
}

func isRiDe(p Pillar) bool {
	// 日德: 甲寅(1,3), 丙辰(3,5), 戊辰(5,5), 庚辰(7,5), 壬戌(9,11)
	switch {
	case p.Stem == 1 && p.Branch == 3:
		return true
	case p.Stem == 3 && p.Branch == 5:
		return true
	case p.Stem == 5 && p.Branch == 5:
		return true
	case p.Stem == 7 && p.Branch == 5:
		return true
	case p.Stem == 9 && p.Branch == 11:
		return true
	}
	return false
}

func isRiGui(p Pillar) bool {
	// 日贵: 丁酉(4,10), 丁亥(4,12), 癸巳(10,6), 癸卯(10,4)
	switch {
	case p.Stem == 4 && p.Branch == 10:
		return true
	case p.Stem == 4 && p.Branch == 12:
		return true
	case p.Stem == 10 && p.Branch == 6:
		return true
	case p.Stem == 10 && p.Branch == 4:
		return true
	}
	return false
}

// ---- 五、structure ----

func computeStructureCross(a, b ChartResult, aBirthYear, aBirthMonth, aBirthHour, bBirthYear, bBirthMonth, bBirthHour int) StructureCross {
	return StructureCross{
		TaiYuanMingGong: computeTYMG(a, b, aBirthMonth, aBirthHour, bBirthMonth, bBirthHour),
		Dayun:           computeDayunCross(a, b, aBirthYear, bBirthYear),
		XunGong:         computeXunGong(a, b),
	}
}

func computeTYMG(a, b ChartResult, aBirthMonth, aBirthHour, bBirthMonth, bBirthHour int) TaiYuanMingGongStruct {
	aTYMG := ComputeTaiYuanMingGong(Pillar{Stem: a.Month.Stem, Branch: a.Month.Branch}, a.Year.Stem, aBirthMonth, aBirthHour)
	bTYMG := ComputeTaiYuanMingGong(Pillar{Stem: b.Month.Stem, Branch: b.Month.Branch}, b.Year.Stem, bBirthMonth, bBirthHour)

	result := TaiYuanMingGongStruct{
		A: TYMGPerson{
			TaiYuan:  PillarStemBranch{Stem: int(aTYMG.TaiYuan.Stem), Branch: int(aTYMG.TaiYuan.Branch)},
			MingGong: PillarStemBranch{Stem: int(aTYMG.MingGong.Stem), Branch: int(aTYMG.MingGong.Branch)},
			ShenGong: PillarStemBranch{Stem: int(aTYMG.ShenGong.Stem), Branch: int(aTYMG.ShenGong.Branch)},
		},
		B: TYMGPerson{
			TaiYuan:  PillarStemBranch{Stem: int(bTYMG.TaiYuan.Stem), Branch: int(bTYMG.TaiYuan.Branch)},
			MingGong: PillarStemBranch{Stem: int(bTYMG.MingGong.Stem), Branch: int(bTYMG.MingGong.Branch)},
			ShenGong: PillarStemBranch{Stem: int(bTYMG.ShenGong.Stem), Branch: int(bTYMG.ShenGong.Branch)},
		},
	}

	result.Cross.TaiYuan = TYMGCrossEntry{
		Stem:   AnalyzeStemRelation(aTYMG.TaiYuan.Stem, bTYMG.TaiYuan.Stem),
		Branch: AnalyzeBranchRelation(aTYMG.TaiYuan.Branch, bTYMG.TaiYuan.Branch),
	}
	result.Cross.MingGong = TYMGCrossEntry{
		Stem:   AnalyzeStemRelation(aTYMG.MingGong.Stem, bTYMG.MingGong.Stem),
		Branch: AnalyzeBranchRelation(aTYMG.MingGong.Branch, bTYMG.MingGong.Branch),
	}
	result.Cross.ShenGong = TYMGCrossEntry{
		Stem:   AnalyzeStemRelation(aTYMG.ShenGong.Stem, bTYMG.ShenGong.Stem),
		Branch: AnalyzeBranchRelation(aTYMG.ShenGong.Branch, bTYMG.ShenGong.Branch),
	}

	return result
}

func computeDayunCross(a, b ChartResult, aBirthYear, bBirthYear int) DayunCross {
	currentYear := aBirthYear
	if bBirthYear > currentYear {
		currentYear = bBirthYear
	}
	if currentYear < 2024 {
		currentYear = 2024
	}

	aDayunResult := ComputeDayunResult(a.Dayun, a.DayMaster, aBirthYear, currentYear, a.ToBazi())
	bDayunResult := ComputeDayunResult(b.Dayun, b.DayMaster, bBirthYear, currentYear, b.ToBazi())

	dc := DayunCross{}

	if aDayunResult != nil && aDayunResult.CurrentPillarIndex >= 0 && aDayunResult.CurrentPillarIndex < len(aDayunResult.Pillars) {
		dp := aDayunResult.Pillars[aDayunResult.CurrentPillarIndex]
		dc.ACurrent = DayunCrossEntry{Stem: int(dp.Stem), Branch: int(dp.Branch), Name: dp.Name, TenGod: dp.TenGod}
	}
	if bDayunResult != nil && bDayunResult.CurrentPillarIndex >= 0 && bDayunResult.CurrentPillarIndex < len(bDayunResult.Pillars) {
		dp := bDayunResult.Pillars[bDayunResult.CurrentPillarIndex]
		dc.BCurrent = DayunCrossEntry{Stem: int(dp.Stem), Branch: int(dp.Branch), Name: dp.Name, TenGod: dp.TenGod}
	}

	dc.StemRel = AnalyzeStemRelation(Stem(dc.ACurrent.Stem), Stem(dc.BCurrent.Stem))
	dc.BranchRel = AnalyzeBranchRelation(Branch(dc.ACurrent.Branch), Branch(dc.BCurrent.Branch))

	return dc
}

func computeXunGong(a, b ChartResult) XunGong {
	return XunGong{
		SameXun:  XunIndex(Pillar{Stem: a.Day.Stem, Branch: a.Day.Branch}) == XunIndex(Pillar{Stem: b.Day.Stem, Branch: b.Day.Branch}),
		SameGong: a.Day.Branch == b.Day.Branch,
	}
}

// ---- helpers ----

func elementFromChinese(s string) Element {
	switch s {
	case "木":
		return ElemWood
	case "火":
		return ElemFire
	case "土":
		return ElemEarth
	case "金":
		return ElemMetal
	case "水":
		return ElemWater
	}
	return 0
}

func nayinElement(nayin string) Element {
	if len(nayin) < 3 {
		return 0
	}
	rs := []rune(nayin)
	last := string(rs[len(rs)-1:])
	return elementFromChinese(last)
}
