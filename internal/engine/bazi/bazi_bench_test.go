package bazi

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

func BenchmarkComputeBond(b *testing.B) {
	stA := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	stB := tianwen.GregorianToSolar(
		time.Date(1986, 8, 20, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		121.5, 8,
	)
	chartA := ComputeChart(stA, ganzhi.Male)
	chartB := ComputeChart(stB, ganzhi.Female)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeBond(chartA.ChartBase, chartB.ChartBase)
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
		if _, err := ComputeLiuNian(chart.ChartBase, 2026); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkComputeXiaoYun(b *testing.B) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeXiaoYun(st, ganzhi.Male, 120)
	}
}
