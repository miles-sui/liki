package persona

import (
	"math"
	"sort"

	"github.com/25types/25types/internal/ganzhi"
)

// Identity is the prototype classification result.
// ID is the only persisted field. Label and Category are derived at query time.
type Identity struct {
	Label    string `json:"label"`
	ID       string `json:"id"`
	Category string `json:"category"`
}

// ClassifyIdentity classifies d by cosine similarity with simplex d-prototypes.
// Both profile and prototypes live in d-space (Σ=0). Cosine removes magnitude bias.
// Tie-breaking: pick the candidate whose first-letter element has the largest d value.
func ClassifyIdentity(d Deviation, prototypes map[string][5]float64) Identity {
	dNorm := math.Sqrt(d.Dot([5]float64(d)))
	if dNorm < 1e-12 {
		// Degenerate: d ≈ 0. Fall back to max element.
		best := 0
		for i := 1; i < 5; i++ {
			if d[i] > d[best] {
				best = i
			}
		}
		fallback := string([]byte{'W', 'F', 'E', 'M', 'R'}[best])
		return Identity{Label: fallback, ID: fallback, Category: fallback}
	}

	const epsilon = 1e-12
	var bestIDs []string
	bestScore := -2.0 // cos ∈ [-1, 1]
	for id, proto := range prototypes {
		pDev := Deviation(proto)
		pNorm := math.Sqrt(pDev.Dot(proto))
		score := d.Dot(proto) / (dNorm * pNorm)
		if score > bestScore+epsilon {
			bestScore = score
			bestIDs = []string{id}
		} else if math.Abs(score-bestScore) <= epsilon {
			bestIDs = append(bestIDs, id)
		}
	}

	if len(bestIDs) > 1 {
		sort.Slice(bestIDs, func(i, j int) bool {
			di := d[ganzhi.ElementCodeIndex(string(bestIDs[i][0]))]
			dj := d[ganzhi.ElementCodeIndex(string(bestIDs[j][0]))]
			if di != dj {
				return di > dj
			}
			// Same first-letter d value: use sheng/ke priority.
			return tieKey(bestIDs[i]) < tieKey(bestIDs[j])
		})
	}
	bestID := bestIDs[0]

	return Identity{Label: bestID, ID: bestID, Category: DeriveCategory(bestID)}
}

// DeriveCategory returns the primary element code from a prototype ID.
func DeriveCategory(id string) string {
	if len(id) == 0 {
		return ""
	}
	return string(id[0])
}

// tieKey returns a sort key for tie-breaking when d values are equal.
// Preference: pure > sheng_wo > wo_sheng > wo_ke > ke_wo.
// Within category: sheng-cycle order W→F→E→M→R, primary element first.
func tieKey(id string) int {
	if len(id) != 2 {
		primary := ganzhi.ElementCodeIndex(id)
		if primary < 0 {
			primary = 0
		}
		return primary
	}

	primary := ganzhi.ElementCodeIndex(string(id[0]))
	secondary := ganzhi.ElementCodeIndex(string(id[1]))
	if primary < 0 {
		primary = 0
	}
	if secondary < 0 {
		secondary = 0
	}

	cat := relationshipCategory(primary, secondary)
	return 5 + cat*100 + primary*10 + secondary
}

// relationshipCategory determines the wuxing relationship.
// Returns: 1=sheng_wo, 2=wo_sheng, 3=wo_ke, 4=ke_wo
func relationshipCategory(a, b int) int {
	if ganzhi.S[a][b] == 1 {
		return 2 // wo_sheng: a nourishes b
	}
	if ganzhi.S[b][a] == 1 {
		return 1 // sheng_wo: b nourishes a
	}
	if ganzhi.C[a][b] == 1 {
		return 3 // wo_ke: a controls b
	}
	if ganzhi.C[b][a] == 1 {
		return 4 // ke_wo: b controls a
	}
	return 0
}
