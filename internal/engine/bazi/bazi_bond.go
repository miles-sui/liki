package bazi

import "liki/internal/engine/ganzhi"

type zhuPairEntry struct {
	AZhu, BZhu, AStem, BStem, ABranch, BBranch string
	Stem    GanRelation
	Branch  ZhiRelation
}
type zhuCross struct{ Pairs []zhuPairEntry }
type tenGodCross struct {
	AToB map[string]string
	BToA map[string]string
}
type nayinPairEntry struct {
	AZhu, BZhu, ANaYin, BNaYin, Relation string
}
type yongShenEntry struct {
	Yong, Ji         string
	YongInOther, JiInOther int
}
type nayinCross struct {
	Pairs    []nayinPairEntry
	Elements struct{ A, B map[string]int }
	YongShen struct{ A, B yongShenEntry }
}
type shenshaMutual struct{ AInB, BInA bool }
type shenshaCross struct {
	TianYi, Lu, TaoHua, YiMa, KongWang, KuiGang, RiDe, RiGui shenshaMutual
}
type daYunCrossEntry struct {
	Gan    ganzhi.Gan
	Zhi    ganzhi.Zhi
	Name   string
	TenGod string
}
type daYunCross struct {
	ACurrent, BCurrent daYunCrossEntry
	StemRel  GanRelation
	BranchRel ZhiRelation
}
// XunGong describes whether two charts share the same xun (旬) or palace (宫).
type XunGong struct {
	SameXun  bool
	SameGong bool
}
type structureCross struct {
	DaYun   daYunCross
	XunGong XunGong
}
// Bond holds the compatibility analysis between two bazi charts.
type Bond struct {
	ZhuCross  zhuCross
	TenGodCross  tenGodCross
	NayinCross   nayinCross
	ShenshaCross shenshaCross
	Structure    structureCross
}

func ComputeBond(a, b ChartBase) Bond {
	return Bond{
		ZhuCross:  computeZhuCross(a, b),
		TenGodCross:  computeTenGodCross(a, b),
		NayinCross:   computeNayinCross(a, b),
		ShenshaCross: computeShenshaCross(a, b),
		Structure:    computeStructureCross(a, b),
	}
}

func computeZhuCross(a, b ChartBase) zhuCross {
	aG := [4]ganzhi.Gan{a.Year.Gan, a.Month.Gan, a.Day.Gan, a.Hour.Gan}
	bG := [4]ganzhi.Gan{b.Year.Gan, b.Month.Gan, b.Day.Gan, b.Hour.Gan}
	aZ := [4]ganzhi.Zhi{a.Year.Zhi, a.Month.Zhi, a.Day.Zhi, a.Hour.Zhi}
	bZ := [4]ganzhi.Zhi{b.Year.Zhi, b.Month.Zhi, b.Day.Zhi, b.Hour.Zhi}
	pairs := make([]zhuPairEntry, 0, 16)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			pairs = append(pairs, zhuPairEntry{
				AZhu: zhuNames[i], BZhu: zhuNames[j],
				AStem: ganzhi.GanName(aG[i]), BStem: ganzhi.GanName(bG[j]),
				ABranch: ganzhi.ZhiName(aZ[i]), BBranch: ganzhi.ZhiName(bZ[j]),
				Stem: analyzeGanRelation(aG[i], bG[j]),
				Branch: analyzeZhiRelation(aZ[i], bZ[j]),
			})
		}
	}
	return zhuCross{Pairs: pairs}
}

func computeTenGodCross(a, b ChartBase) tenGodCross {
	aG := [4]ganzhi.Gan{a.Year.Gan, a.Month.Gan, a.Day.Gan, a.Hour.Gan}
	bG := [4]ganzhi.Gan{b.Year.Gan, b.Month.Gan, b.Day.Gan, b.Hour.Gan}
	aElem, aYY := ganzhi.GanWuxing(a.DayMaster), ganzhi.GanYinYang(a.DayMaster)
	bElem, bYY := ganzhi.GanWuxing(b.DayMaster), ganzhi.GanYinYang(b.DayMaster)
	aToB, bToA := make(map[string]string, 4), make(map[string]string, 4)
	for i := 0; i < 4; i++ {
		aToB[zhuNames[i]+"_stem"] = ganzhi.TenGodName(ganzhi.TenGodType(aElem, aYY, ganzhi.GanWuxing(bG[i]), ganzhi.GanYinYang(bG[i])))
		bToA[zhuNames[i]+"_stem"] = ganzhi.TenGodName(ganzhi.TenGodType(bElem, bYY, ganzhi.GanWuxing(aG[i]), ganzhi.GanYinYang(aG[i])))
	}
	return tenGodCross{AToB: aToB, BToA: bToA}
}

func computeNayinCross(a, b ChartBase) nayinCross {
	aNy, bNy := a.NaYinArray(), b.NaYinArray()
	pairs := make([]nayinPairEntry, 0, 16)
	for i := 0; i < 4; i++ {
		ae := ganzhi.NaYinWuxing(aNy[i])
		for j := 0; j < 4; j++ {
			be := ganzhi.NaYinWuxing(bNy[j])
			rel := "相同"
			if ae != be {
				if ganzhi.Sheng(ae, be) || ganzhi.Sheng(be, ae) { rel = "相生" } else { rel = "相克" }
			}
			pairs = append(pairs, nayinPairEntry{AZhu: zhuNames[i], BZhu: zhuNames[j], ANaYin: aNy[i], BNaYin: bNy[j], Relation: rel})
		}
	}
	nc := nayinCross{Pairs: pairs}
	nc.Elements.A, nc.Elements.B = countNayinElems(a.WuxingCount), countNayinElems(b.WuxingCount)
	nc.YongShen.A = makeyongShenEntry(a.FuYi.Yong, a.FuYi.Ji, b.WuxingCount)
	nc.YongShen.B = makeyongShenEntry(b.FuYi.Yong, b.FuYi.Ji, a.WuxingCount)
	return nc
}

func countNayinElems(wc map[ganzhi.Wuxing]int) map[string]int {
	m := make(map[string]int, 5)
	for e, c := range wc { m[e.String()] = c }
	return m
}

func makeyongShenEntry(yong, ji string, wc map[ganzhi.Wuxing]int) yongShenEntry {
	ys, err := ganzhi.ParseWuxing(yong)
	if err != nil {
		ys = 0
	}
	js, err := ganzhi.ParseWuxing(ji)
	if err != nil {
		js = 0
	}
	return yongShenEntry{Yong: yong, Ji: ji, YongInOther: wc[ys], JiInOther: wc[js]}
}

func computeShenshaCross(a, b ChartBase) shenshaCross {
	aBz, bBz := a.ToBazi(), b.ToBazi()
	aPs, bPs := aBz.Slice(), bBz.Slice()
	mut := func(aZ, bZ []ganzhi.Zhi) shenshaMutual { return shenshaMutual{branchesInZhus(aZ, bBz), branchesInZhus(bZ, aBz)} }
	return shenshaCross{
		TianYi: mut(tianYiBranches(a), tianYiBranches(b)),
		Lu: mut(luBranches(a), luBranches(b)),
		TaoHua: mut(zhiLookup(a, taohuaBranchMap), zhiLookup(b, taohuaBranchMap)),
		YiMa: mut(zhiLookup(a, yimaBranchMap), zhiLookup(b, yimaBranchMap)),
		KongWang: mut(kongwangBranches(aBz), kongwangBranches(bBz)),
		KuiGang: mut(collectBranches(aPs, isKuiGang), collectBranches(bPs, isKuiGang)),
		RiDe: mut(collectBranches(aPs, isRiDe), collectBranches(bPs, isRiDe)),
		RiGui: mut(collectBranches(aPs, isRiGui), collectBranches(bPs, isRiGui)),
	}
}

func branchesInZhus(ts []ganzhi.Zhi, bz ganzhi.Bazi) bool {
	for _, t := range ts { for _, p := range bz.Slice() { if p.Zhi == t { return true } } }
	return false
}
func tianYiBranches(c ChartBase) []ganzhi.Zhi {
	var bs []ganzhi.Zhi
	if b, ok := tianYiLookup[c.Year.Gan]; ok { bs = append(bs, b[0], b[1]) }
	if b, ok := tianYiLookup[c.DayMaster]; ok { bs = append(bs, b[0], b[1]) }
	return bs
}
func luBranches(c ChartBase) []ganzhi.Zhi {
	if c.DayMaster < 1 || c.DayMaster > 10 { return nil }
	return []ganzhi.Zhi{ganzhi.ChangShengTable[c.DayMaster][3]}
}
func zhiLookup(c ChartBase, m map[ganzhi.Zhi]ganzhi.Zhi) []ganzhi.Zhi {
	var bs []ganzhi.Zhi
	if b, ok := m[c.Year.Zhi]; ok { bs = append(bs, b) }
	if b, ok := m[c.Day.Zhi]; ok { bs = append(bs, b) }
	return bs
}
func kongwangBranches(bz ganzhi.Bazi) []ganzhi.Zhi {
	ps := bz.Slice(); vh := computeKongWang(bz)
	bs := make([]ganzhi.Zhi, 0, len(vh))
	for _, idx := range vh { bs = append(bs, ps[idx].Zhi) }
	return bs
}
func collectBranches(ps [4]ganzhi.Zhu, f func(ganzhi.Zhu) bool) []ganzhi.Zhi {
	var bs []ganzhi.Zhi; for _, p := range ps { if f(p) { bs = append(bs, p.Zhi) } }; return bs
}
func isRiDe(p ganzhi.Zhu) bool { _, ok := riDeSet[[2]int{int(p.Gan), int(p.Zhi)}]; return ok }
func isRiGui(p ganzhi.Zhu) bool { _, ok := riGuiSet[[2]int{int(p.Gan), int(p.Zhi)}]; return ok }
var riDeSet map[[2]int]bool
var riGuiSet map[[2]int]bool

func computeStructureCross(a, b ChartBase) structureCross {
	return structureCross{DaYun: computeDaYunCross(a, b), XunGong: computeXunGong(a, b)}
}
func computeDaYunCross(a, b ChartBase) daYunCross {
	if a.DaYun == nil || b.DaYun == nil { return daYunCross{} }
	dc := daYunCross{ACurrent: currentDaYunEntry(a.DaYun), BCurrent: currentDaYunEntry(b.DaYun)}
	dc.StemRel = analyzeGanRelation(dc.ACurrent.Gan, dc.BCurrent.Gan)
	dc.BranchRel = analyzeZhiRelation(dc.ACurrent.Zhi, dc.BCurrent.Zhi)
	return dc
}
func currentDaYunEntry(dr *DaYun) daYunCrossEntry {
	if dr == nil || dr.CurrentZhuIndex < 0 || dr.CurrentZhuIndex >= len(dr.Zhus) { return daYunCrossEntry{} }
	p := dr.Zhus[dr.CurrentZhuIndex]
	return daYunCrossEntry{Gan: p.Gan, Zhi: p.Zhi, Name: p.Name, TenGod: p.TenGod}
}
func computeXunGong(a, b ChartBase) XunGong {
	return XunGong{SameXun: xunIndex(ganzhi.Zhu{Gan:a.Day.Gan,Zhi:a.Day.Zhi}) == xunIndex(ganzhi.Zhu{Gan:b.Day.Gan,Zhi:b.Day.Zhi}), SameGong: a.Day.Zhi == b.Day.Zhi}
}
