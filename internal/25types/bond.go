package persona

import "github.com/25types/25types/internal/ganzhi"

type BondResult struct {
	DeltaA Deviation `json:"delta_a"`
	DeltaB Deviation `json:"delta_b"`
	DEffA  Deviation `json:"d_eff_a"`
	DEffB  Deviation `json:"d_eff_b"`
}

// ComputeBond computes the bidirectional Bond between two users.
// Only elevated elements (d_i > 0) output force via ReLU.
// ΣΔ = 0 for both DeltaA and DeltaB (conservation).
func ComputeBond(dA, dB Deviation) BondResult {
	dApos := dA.Relu()
	dBpos := dB.Relu()

	deltaA := ApplyTranspose(&ganzhi.S, dBpos).Sub(ApplyTranspose(&ganzhi.C, dBpos))
	deltaB := ApplyTranspose(&ganzhi.S, dApos).Sub(ApplyTranspose(&ganzhi.C, dApos))

	return BondResult{
		DeltaA: deltaA,
		DeltaB: deltaB,
		DEffA:  dA.Add(deltaA),
		DEffB:  dB.Add(deltaB),
	}
}
