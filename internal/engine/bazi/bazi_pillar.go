package bazi

import (
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

func computeChart(bz ganzhi.Bazi, st tianwen.SolarTime, gender ganzhi.Gender) Chart {
	dm := bz.Ri.Gan
	hs := computeCangGan(bz)
	ny := computeNaYin(bz)
	ls := computeChangSheng(dm)
	ec := computeElementCount(bz, hs)
	tgTable := computeShiShensTable(bz, hs)
	lsTable := computeChangShengTable(bz, hs)
	shensha := computeShenSha(bz)
	voidHits := computeKongWang(bz)
	ps := bz.Slice()

	makePI := func(i int) zhuInfo {
		isVoid := false
		for _, vh := range voidHits {
			if vh == i {
				isVoid = true
				break
			}
		}
		pz := ganzhi.Zhu{Gan: ps[i].Gan, Zhi: ps[i].Zhi}
		pi := zhuInfo{Gan: ps[i].Gan, Zhi: ps[i].Zhi, NaYin: ny[i], CangGan: hs[i], ShiShens: tgTable[i], ChangSheng: lsTable[i], ShenSha: shensha[i], IsVoid: isVoid}
		pi.IsSelfHe = isSelfHe(pz)
		if pi.IsSelfHe {
			pi.SelfHeName = selfHeName(pz)
		}
		pi.IsKuiGang = isKuiGang(pz)
		return pi
	}

	cr := Chart{
		ChartBase: ChartBase{
			Nian: makePI(0), Yue: makePI(1), Ri: makePI(2), Shi: makePI(3),
		},
		SolarTime:  st,
		ChangSheng: ls,
		WuxingCount: ec,
		HeHui:     computeFullTripleHeHui(bz),
		GongJia:   computeGongJia(bz),
		NayinRel:  computeNaYinRelations(ny),
		SanQiName: sanQiName(sanQiType(bz)),
		WangShuai: map[string]string{
			ganzhi.WxMu.String():   ganzhi.WangShuaiOf(ganzhi.WxMu, bz.Yue.Zhi).String(),
			ganzhi.WxHuo.String():  ganzhi.WangShuaiOf(ganzhi.WxHuo, bz.Yue.Zhi).String(),
			ganzhi.WxTu.String():   ganzhi.WangShuaiOf(ganzhi.WxTu, bz.Yue.Zhi).String(),
			ganzhi.WxJin.String():  ganzhi.WangShuaiOf(ganzhi.WxJin, bz.Yue.Zhi).String(),
			ganzhi.WxShui.String(): ganzhi.WangShuaiOf(ganzhi.WxShui, bz.Yue.Zhi).String(),
		},
	}
	cr.FuYi = computeFuYi(cr)
	cr.TiaoHou, _ = computeTiaohou(cr.Ri.Gan, cr.Yue.Zhi)
	cr.DaYun = computeDaYun(st, bz.Yue, bz.Nian.Gan, bz.Ri.Gan, gender)
	birthMonth := (int(bz.Yue.Zhi) + 9) % 12 + 1
	cr.TaiYuanMingGong = computeTaiYuanMingGong(bz.Yue, bz.Nian.Gan, birthMonth, int(bz.Shi.Zhi))
	return cr
}

func computeCangGan(bz ganzhi.Bazi) [4]cangGanOut {
	var hs [4]cangGanOut
	for i, z := range bz.Slice() {
		qi := ganzhi.CangGanForZhi(z.Zhi)
		hs[i] = cangGanOut{Main: *qi.Main}
		if qi.Mid != nil {
			mid := *qi.Mid
			hs[i].Mid = &mid
		}
		if qi.Minor != nil {
			minor := *qi.Minor
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

func computeChangSheng(dm ganzhi.Gan) [12]stageOut {
	var out [12]stageOut
	branches := ganzhi.ChangShengTable[dm]
	for i := 0; i < 12; i++ {
		out[i] = stageOut{
			Name:  ganzhi.StageNamesZH[i],
			Index: branches[i],
		}
	}
	return out
}

func computeElementCount(bz ganzhi.Bazi, hs [4]cangGanOut) map[ganzhi.Wuxing]int {
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

func computeShiShensTable(bz ganzhi.Bazi, hs [4]cangGanOut) [4][]shiShenEntry {
	dm := bz.Ri.Gan
	var table [4][]shiShenEntry
	ps := bz.Slice()
	for i := range ps {
		var entries []shiShenEntry
		entries = append(entries, shiShenEntry{
			ShiShen: ganzhi.ShiShenFromGan(dm, ps[i].Gan),
			Name:   ganzhi.GanName(ps[i].Gan),
			Source: sourceGan,
			Gan:    ps[i].Gan,
		})
		entries = append(entries, shiShenEntry{
			ShiShen: ganzhi.ShiShenFromGan(dm, hs[i].Main),
			Name:   ganzhi.GanName(hs[i].Main),
			Source: sourceMainQi,
			Gan:    hs[i].Main,
		})
		if hs[i].Mid != nil {
			entries = append(entries, shiShenEntry{
				ShiShen: ganzhi.ShiShenFromGan(dm, *hs[i].Mid),
				Name:   ganzhi.GanName(*hs[i].Mid),
				Source: sourceMidQi,
				Gan:    *hs[i].Mid,
			})
		}
		if hs[i].Minor != nil {
			entries = append(entries, shiShenEntry{
				ShiShen: ganzhi.ShiShenFromGan(dm, *hs[i].Minor),
				Name:   ganzhi.GanName(*hs[i].Minor),
				Source: sourceMinQi,
				Gan:    *hs[i].Minor,
			})
		}
		table[i] = entries
	}
	return table
}

func computeChangShengTable(bz ganzhi.Bazi, hs [4]cangGanOut) [4][]changShengEntry {
	var table [4][]changShengEntry
	for i, z := range bz.Slice() {
		stages, ok := ganzhi.ChangShengTable[z.Gan]
		if !ok {
			continue
		}
		for stageIdx, b := range stages {
			if b == z.Zhi {
				table[i] = []changShengEntry{{
					Stage: ganzhi.StageNamesZH[stageIdx],
					Gan:   z.Gan,
				}}
				break
			}
		}
	}
	return table
}

func computeNaYinRelations(nayins [4]string) []naYinGuanXi {
	var rels []naYinGuanXi
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			ae := ganzhi.NaYinWuxing(nayins[i])
			be := ganzhi.NaYinWuxing(nayins[j])
			rel := "相同"
			if ae != 0 && be != 0 && ae != be {
				if ganzhi.Sheng(ae, be) || ganzhi.Sheng(be, ae) {
					rel = "相生"
				} else {
					rel = "相克"
				}
			}
			rels = append(rels, naYinGuanXi{A: zhuNames[i], B: zhuNames[j], Relation: rel})
		}
	}
	return rels
}
