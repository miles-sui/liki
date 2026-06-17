package qimen

import "liki/internal/engine/ganzhi"

// YingQi holds timing prediction information.
type YingQi struct {
	MaXing   string `json:"ma_xing_dir"`   // 马星应期方向
	KongWang string `json:"kong_wang_fill"` // 空亡填实时机
	DutyMove string `json:"duty_move"`      // 值符值使推动
	Summary  string `json:"summary"`        // 综合应期判断
}

// computeYingQi computes 应期 timing from the pan.
func computeYingQi(pan pan) YingQi {
	var yq YingQi

	// 马星应期: 冲马星的地支年份/月份/时辰.
	mataZhi := palaceZhi(pan.MaXing)
	chongZhi := chongBranch(mataZhi)
	yq.MaXing = "马星在" + ganzhi.ZhiName(mataZhi) + "，冲则动，应期在" + ganzhi.ZhiName(chongZhi) + "（年月日时）"

	// 空亡填实: 空亡宫被填实时应事.
	for _, kw := range pan.KongWang {
		z := palaceZhi(kw)
		yq.KongWang += ganzhi.ZhiName(z) + " "
	}
	if yq.KongWang != "" {
		yq.KongWang = "空亡在" + yq.KongWang + "，填实或冲空之时应事"
	}

	// 值符值使推动.
	yq.DutyMove = "值符" + pan.DutyStar.String() + "加" + pan.DutyDoor.String() + "，以时干" + ganzhi.GanName(ganzhi.Gan(int(pan.DriveZhi)%10+1)) + "为应"

	yq.Summary = yq.MaXing + "；" + yq.KongWang

	return yq
}

// chongBranch returns the opposing branch (冲).
func chongBranch(z ganzhi.Zhi) ganzhi.Zhi {
	return ganzhi.Zhi((int(z) + 6) % 12)
}
