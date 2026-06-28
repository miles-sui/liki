package qiming

import (
	"testing"
)

func BenchmarkPrepareWuGe(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := PrepareWuGe("王", "金", []string{"土", "金"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDetailNames(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := DetailNames("王", []string{"浩然", "明哲"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEvaluateName(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := EvaluateName("王", "浩然", "金"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkComposeNames(b *testing.B) {
	// Pre-compute wuge data for realistic compose benchmark.
	wuge, err := PrepareWuGe("王", "金", []string{"土", "金"})
	if err != nil || len(wuge.Combos) == 0 {
		b.Skip("wuge returned no combos")
	}
	combo := wuge.Combos[:min(3, len(wuge.Combos))]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComposeNames("王", combo, wuge.YongChars, wuge.XiChars)
	}
}
