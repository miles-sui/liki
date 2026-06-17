package ziwei

var siHuaTable = map[Gan][4]starIndex{
	1:  {LianZhen, PoJun, WuQu, TaiYang},
	2:  {TianJi, TianLiang, ZiWei, TaiYin},
	3:  {TianTong, TianJi, WenChang, LianZhen},
	4:  {TaiYin, TianTong, TianJi, JuMen},
	5:  {TanLang, TaiYin, YouBi, TianJi},
	6:  {WuQu, TanLang, TianLiang, WenQu},
	7:  {TaiYang, WuQu, TaiYin, TianTong},
	8:  {JuMen, TaiYang, WenQu, WenChang},
	9:  {TianLiang, ZiWei, ZuoFu, WuQu},
	10: {PoJun, JuMen, TaiYin, TanLang},
}

func computeSiHua(yearGan Gan) siHuaResult {
	stars, ok := siHuaTable[yearGan]
	if !ok {
		return nil
	}
	return siHuaResult{
		stars[0]: HuaLu, stars[1]: HuaQuan, stars[2]: HuaKe, stars[3]: HuaJi,
	}
}
