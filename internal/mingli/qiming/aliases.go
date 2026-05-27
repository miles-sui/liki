package qiming

import "github.com/25types/25types/internal/ganzhi"

// -- type aliases --

type Element = ganzhi.Element
type Branch = ganzhi.Branch

// -- function wrappers --

func ElementFromChinese(name string) Element { return ganzhi.ElementFromChinese(name) }
