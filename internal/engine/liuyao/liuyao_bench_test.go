package liuyao

import (
	"testing"
	"time"

	"liki/internal/engine/tianwen"
)

func BenchmarkComputeChart(b *testing.B) {
	st := tianwen.GregorianToSolar(
		time.Date(2026, 6, 28, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeChart(st, YongGuanGui, [6]int{})
	}
}
