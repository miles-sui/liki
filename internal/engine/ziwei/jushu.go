package ziwei

import "liki/internal/engine/ganzhi"

// determineJuShu returns the bureau number from ming palace stem and branch.
func determineJuShu(mingGan Gan, mingZhi Zhi) juShu {
	nayinName := ganzhi.NaYinLabel(mingGan, mingZhi)
	wx := ganzhi.NaYinWuxing(nayinName)
	if wx == 0 {
		return 0
	}
	return juShuFromWuxing(wx)
}
