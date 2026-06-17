package bazi

import (
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// ComputeChart produces a full Chart from solar birth time and gender.
func ComputeChart(st tianwen.SolarTime, gender ganzhi.Gender) Chart {
	bz := tianwen.ComputeBazi(st)
	dm := bz.Ri.Gan
	hs := computeHiddenStems(bz)
	ny := computeNaYin(bz)
	ls := computeLifeStages(dm)
	ec := computeElementCount(bz, hs)
	tgTable := computeTenGodsTable(bz, hs)
	lsTable := computeLifeStageTable(bz, hs)
	shensha := computeShenSha(bz)
	voidHits := computeKongWang(bz)
	ps := bz.Slice()

	makePI := func(i int) pillarInfo {
		isVoid := false
		for _, vh := range voidHits {
			if vh == i {
				isVoid = true
				break
			}
		}
		pz := ganzhi.Zhu{Gan: ps[i].Gan, Zhi: ps[i].Zhi}
		pi := pillarInfo{Gan: ps[i].Gan, Zhi: ps[i].Zhi, NaYin: ny[i], HiddenStems: hs[i], TenGods: tgTable[i], LifeStages: lsTable[i], ShenSha: shensha[i], IsVoid: isVoid}
		pi.IsSelfHe = isSelfHe(pz)
		if pi.IsSelfHe {
			pi.SelfHeName = selfHeName(pz)
		}
		pi.IsKuiGang = isKuiGang(pz)
		return pi
	}

	cr := Chart{
		ChartBase: ChartBase{
			Year: makePI(0), Month: makePI(1), Day: makePI(2), Hour: makePI(3),
			DayMaster:   dm,
			LifeStages:  ls,
			WuxingCount: ec,
		},
		HeHui:     computeFullTripleHeHui(bz),
		GongJia:   computeGongJia(bz),
		NayinRel:  computeNaYinRelations(ny),
		SanQiName: sanQiName(sanQiType(bz)),
		WangShuai: map[string]string{
			ganzhi.WxMu.String():   monthWangShuai(ganzhi.WxMu, bz.Yue.Zhi),
			ganzhi.WxHuo.String():  monthWangShuai(ganzhi.WxHuo, bz.Yue.Zhi),
			ganzhi.WxTu.String():   monthWangShuai(ganzhi.WxTu, bz.Yue.Zhi),
			ganzhi.WxJin.String():  monthWangShuai(ganzhi.WxJin, bz.Yue.Zhi),
			ganzhi.WxShui.String(): monthWangShuai(ganzhi.WxShui, bz.Yue.Zhi),
		},
	}
	cr.FuYi = computeFuYi(cr)
	cr.TiaoHou, _ = computeTiaohou(cr.DayMaster, cr.Month.Zhi)
	cr.DaYun = computeDaYun(st, bz.Yue, bz.Nian.Gan, bz.Ri.Gan, gender)
	birthMonth := (int(bz.Yue.Zhi) + 9) % 12 + 1
	cr.TaiYuanMingGong = computeTaiYuanMingGong(bz.Yue, bz.Nian.Gan, birthMonth, int(bz.Shi.Zhi))
	return cr
}

func computeHiddenStems(bz ganzhi.Bazi) [4]hiddenStemsOut {
	var hs [4]hiddenStemsOut
	for i, z := range bz.Slice() {
		qi := ganzhi.HiddenStemsForBranch(z.Zhi)
		hs[i] = hiddenStemsOut{Main: ganzhi.Gan(*qi.Main)}
		if qi.Mid != nil {
			mid := ganzhi.Gan(*qi.Mid)
			hs[i].Mid = &mid
		}
		if qi.Minor != nil {
			minor := ganzhi.Gan(*qi.Minor)
			hs[i].Minor = &minor
		}
	}
	return hs
}

func computeNaYin(bz ganzhi.Bazi) [4]string {
	var ny [4]string
	for i, z := range bz.Slice() {
		ny[i] = ganzhi.NaYinLabel(z.Gan, z.Zhi)
	}
	return ny
}

func computeLifeStages(dm ganzhi.Gan) [12]stageOut {
	var out [12]stageOut
	branches := ganzhi.LifeStagesTable[int(dm)]
	for i := 0; i < 12; i++ {
		out[i] = stageOut{
			Name:  ganzhi.StageNamesZH[i],
			Index: branches[i],
		}
	}
	return out
}

func computeElementCount(bz ganzhi.Bazi, hs [4]hiddenStemsOut) map[ganzhi.Wuxing]int {
	wc := make(map[ganzhi.Wuxing]int)
	for _, z := range bz.Slice() {
		wc[ganzhi.GanWuxing(z.Gan)]++
	}
	for _, h := range hs {
		wc[ganzhi.GanWuxing(h.Main)]++
		if h.Mid != nil {
			wc[ganzhi.GanWuxing(*h.Mid)]++
		}
		if h.Minor != nil {
			wc[ganzhi.GanWuxing(*h.Minor)]++
		}
	}
	return wc
}

func computeTenGodsTable(bz ganzhi.Bazi, hs [4]hiddenStemsOut) [4][]tenGodEntry {
	dm := bz.Ri.Gan
	var table [4][]tenGodEntry
	ps := bz.Slice()
	for i := range ps {
		var entries []tenGodEntry
		entries = append(entries, tenGodEntry{
			TenGod: ganzhi.TenGodFromGan(dm, ps[i].Gan),
			Name:   ganzhi.GanName(ps[i].Gan),
			Source: sourceGan,
			Gan:    ps[i].Gan,
		})
		entries = append(entries, tenGodEntry{
			TenGod: ganzhi.TenGodFromGan(dm, hs[i].Main),
			Name:   ganzhi.GanName(hs[i].Main),
			Source: sourceMainQi,
			Gan:    hs[i].Main,
		})
		if hs[i].Mid != nil {
			entries = append(entries, tenGodEntry{
				TenGod: ganzhi.TenGodFromGan(dm, *hs[i].Mid),
				Name:   ganzhi.GanName(*hs[i].Mid),
				Source: sourceMidQi,
				Gan:    *hs[i].Mid,
			})
		}
		if hs[i].Minor != nil {
			entries = append(entries, tenGodEntry{
				TenGod: ganzhi.TenGodFromGan(dm, *hs[i].Minor),
				Name:   ganzhi.GanName(*hs[i].Minor),
				Source: sourceMinQi,
				Gan:    *hs[i].Minor,
			})
		}
		table[i] = entries
	}
	return table
}

func computeLifeStageTable(bz ganzhi.Bazi, hs [4]hiddenStemsOut) [4][]lifeStageEntry {
	var table [4][]lifeStageEntry
	for i, z := range bz.Slice() {
		stages, ok := ganzhi.LifeStagesTable[int(z.Gan)]
		if !ok {
			continue
		}
		for stageIdx, b := range stages {
			if b == int(z.Zhi) {
				table[i] = []lifeStageEntry{{
					Stage: ganzhi.StageNamesZH[stageIdx],
					Gan:   z.Gan,
				}}
				break
			}
		}
	}
	return table
}

func computeNaYinRelations(nayins [4]string) []naYinRelation {
	var rels []naYinRelation
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			ae := nayinElement(nayins[i])
			be := nayinElement(nayins[j])
			rel := "相同"
			if ae != 0 && be != 0 && ae != be {
				if ganzhi.Sheng(ae, be) || ganzhi.Sheng(be, ae) {
					rel = "相生"
				} else {
					rel = "相克"
				}
			}
			rels = append(rels, naYinRelation{A: pillarNames[i], B: pillarNames[j], Relation: rel})
		}
	}
	return rels
}
