package ziwei

import (
	"testing"
	"time"

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

func BenchmarkComputeChart(b *testing.B) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeChart(st, ganzhi.Male)
	}
}

func BenchmarkComputeDaXian(b *testing.B) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeDaXian(chart)
	}
}

func BenchmarkComputeLiuNian(b *testing.B) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeLiuNian(2026, chart)
	}
}
