package questionnaire

import (
	"testing"
)

func TestLoad_EN(t *testing.T) {
	q, err := Load("en")
	if err != nil {
		t.Fatalf("Load(en): %v", err)
	}
	if q.Meta.Version != "4.0" {
		t.Errorf("version = %s, want 4.0", q.Meta.Version)
	}
	if q.Meta.Language != "en" {
		t.Errorf("language = %s, want en", q.Meta.Language)
	}
	if len(q.Rounds) < 6 {
		t.Errorf("rounds = %d, want at least 6", len(q.Rounds))
	}
	// Each round must have exactly 5 triads and each triad has 3 options.
	for i, r := range q.Rounds {
		if len(r.Questions) != 5 {
			t.Errorf("round %d has %d questions, want 5", i, len(r.Questions))
		}
		for _, qi := range r.Questions {
			if len(qi.Options) != 3 {
				t.Errorf("q %s has %d options, want 3", qi.ID, len(qi.Options))
			}
		}
	}
}

func TestLoad_ZH(t *testing.T) {
	q, err := Load("zh-CN")
	if err != nil {
		t.Fatalf("Load(zh-CN): %v", err)
	}
	if q.Meta.Language != "zh-CN" {
		t.Errorf("language = %s, want zh-CN", q.Meta.Language)
	}
	if len(q.Rounds) < 6 {
		t.Errorf("rounds = %d, want at least 6", len(q.Rounds))
	}
}

func TestAllQIDs(t *testing.T) {
	_, err := Load("en")
	if err != nil {
		t.Fatalf("Load(en): %v", err)
	}
	qids := AllQIDs()
	if len(qids) < 30 {
		t.Errorf("AllQIDs = %d, want at least 30", len(qids))
	}
}

func TestRoundBalance(t *testing.T) {
	q, err := Load("en")
	if err != nil {
		t.Fatalf("Load(en): %v", err)
	}
	// Each round has 5 triads with 3 options each = 15 element references.
	// With 5 elements, each should appear exactly 3 times per round.
	for i, round := range q.Rounds {
		counts := map[string]int{}
		for _, qi := range round.Questions {
			for _, opt := range qi.Options {
				counts[opt.Element]++
			}
		}
		expected := 3 // 15 total / 5 elements
		for elem, c := range counts {
			if c != expected {
				t.Errorf("round %d element %q appears %d times, want %d", i, elem, c, expected)
			}
		}
	}
}

func TestGlobalPairBalance(t *testing.T) {
	q, err := Load("en")
	if err != nil {
		t.Fatalf("Load(en): %v", err)
	}
	pairs := map[string]int{}
	for _, round := range q.Rounds {
		for _, qi := range round.Questions {
			var els []string
			for _, opt := range qi.Options {
				els = append(els, opt.Element)
			}
			for i := 0; i < len(els); i++ {
				for j := i + 1; j < len(els); j++ {
					pairs[els[i]+"+"+els[j]]++
				}
			}
		}
	}
	if len(pairs) != 10 {
		t.Errorf("unique pairs = %d, want 10 (5 choose 2)", len(pairs))
	}
	for pair, c := range pairs {
		if c == 0 {
			t.Errorf("pair %s appears 0 times", pair)
		}
	}
}

func TestEachQuestionHasThreeOptions(t *testing.T) {
	q, err := Load("en")
	if err != nil {
		t.Fatalf("Load(en): %v", err)
	}
	for _, round := range q.Rounds {
		for _, qi := range round.Questions {
			if len(qi.Options) != 3 {
				t.Errorf("q %s has %d options, want 3", qi.ID, len(qi.Options))
			}
		}
	}
}
