package persona

import "github.com/25types/25types/internal/ganzhi"

// Concord represents the bidirectional sheng-ke relationship between two people.
type Concord string

const (
	ConcordShun Concord = "顺"
	ConcordNi   Concord = "逆"
	ConcordPing Concord = "平"
)

const concordEpsilon = 0.02

// ComputeConcord computes the bidirectional sheng-ke relationship between two profiles.
func ComputeConcord(dA, dB Deviation) Concord {
	dAp := dA.Relu()
	dBp := dB.Relu()

	sMid := applyMatrix(&ganzhi.SSym, dBp)
	sh := dAp.Dot(sMid)

	cMid := applyMatrix(&ganzhi.CSym, dBp)
	ke := dAp.Dot(cMid)

	if sh > ke+concordEpsilon {
		return ConcordShun
	}
	if ke > sh+concordEpsilon {
		return ConcordNi
	}
	return ConcordPing
}

func applyMatrix(m *[5][5]float64, v Deviation) [5]float64 {
	var r [5]float64
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			r[i] += m[i][j] * v[j]
		}
	}
	return r
}
