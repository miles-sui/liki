package qiming

import "liki/internal/engine/ganzhi"

type Wuxing = ganzhi.Wuxing

func wuxingFromChinese(name string) Wuxing { return ganzhi.WuxingFromChinese(name) }
