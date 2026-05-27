package persona

import "github.com/25types/25types/internal/ganzhi"

// Answer is a triad forced-choice answer.
type Answer struct {
	QID        string   `json:"qid"`
	Selections []string `json:"selections"`
}

func countElements(answers []Answer) (raw [5]float64, total float64) {
	for _, a := range answers {
		for _, sel := range a.Selections {
			if idx := ganzhi.ElementCodeIndex(sel); idx >= 0 {
				raw[idx] += 1
			}
		}
	}
	for _, v := range raw {
		total += v
	}
	return
}

// ComputeD aggregates triad forced-choice answers into the shape vector d (Σ=0).
func ComputeD(answers []Answer) Deviation {
	raw, total := countElements(answers)
	if total == 0 {
		return Deviation{}
	}
	var d Deviation
	for i := 0; i < 5; i++ {
		d[i] = 5 * (raw[i]/total - 0.2)
	}
	return d
}

// ComputeP computes display proportions (Σ=1) from raw answer counts.
func ComputeP(answers []Answer) Proportion {
	raw, total := countElements(answers)
	if total == 0 {
		return Proportion{0.2, 0.2, 0.2, 0.2, 0.2}
	}
	var p Proportion
	for i := 0; i < 5; i++ {
		p[i] = raw[i] / total
	}
	return p
}

// ApplyTranspose computes Mᵀv.
func ApplyTranspose(m *[5][5]float64, v Deviation) Deviation {
	var r Deviation
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			r[i] += m[j][i] * v[j]
		}
	}
	return r
}
