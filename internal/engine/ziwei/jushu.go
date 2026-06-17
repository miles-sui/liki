package ziwei

import "liki/internal/engine/ganzhi"

// determineJuShu returns the bureau number from ming palace stem and branch.
func determineJuShu(mingGan Gan, mingZhi Zhi) juShu {
	nayinName := ganzhi.NaYinLabel(mingGan, mingZhi)
	if nayinName == "未知" {
		return 0
	}
	runes := []rune(nayinName)
	lastChar := string(runes[len(runes)-1])
	wx := ganzhi.WuxingFromChinese(lastChar)
	return juShuFromWuxing(wx)
}
