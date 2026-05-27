package persona

import (
	"encoding/json"
	"math"
)

// Deviation is a five-element shape vector where Σ = 0.
// Positive = elevated element, negative = subdued element.
type Deviation [5]float64

// Proportion is a five-element display vector where Σ = 1, for presentation only.
type Proportion [5]float64

var elementKeys = [5]string{"wood", "fire", "earth", "metal", "water"}

// MarshalJSON serializes Deviation as a named-element object.
func (d Deviation) MarshalJSON() ([]byte, error) {
	m := map[string]float64{}
	for i, k := range elementKeys {
		m[k] = d[i]
	}
	return json.Marshal(m)
}

// UnmarshalJSON deserializes Deviation from either a named-element object or an array.
func (d *Deviation) UnmarshalJSON(b []byte) error {
	// Try named-element object first (matching MarshalJSON).
	var m map[string]float64
	if json.Unmarshal(b, &m) == nil && len(m) > 0 {
		for i, k := range elementKeys {
			d[i] = m[k]
		}
		return nil
	}
	// Fall back to plain array.
	var a [5]float64
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*d = a
	return nil
}

// MarshalJSON serializes Proportion as a named-element object.
func (p Proportion) MarshalJSON() ([]byte, error) {
	m := map[string]float64{}
	for i, k := range elementKeys {
		m[k] = math.Round(p[i]*100) / 100
	}
	return json.Marshal(m)
}

// UnmarshalJSON deserializes Proportion from either a named-element object or an array.
func (p *Proportion) UnmarshalJSON(b []byte) error {
	var m map[string]float64
	if json.Unmarshal(b, &m) == nil && len(m) > 0 {
		for i, k := range elementKeys {
			p[i] = m[k]
		}
		return nil
	}
	var a [5]float64
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*p = a
	return nil
}

func (d Deviation) Relu() Deviation {
	var o Deviation
	for i, x := range d {
		if x > 0 {
			o[i] = x
		}
	}
	return o
}

func (d Deviation) Add(o Deviation) Deviation {
	return Deviation{d[0] + o[0], d[1] + o[1], d[2] + o[2], d[3] + o[3], d[4] + o[4]}
}

func (d Deviation) Sub(o Deviation) Deviation {
	return Deviation{d[0] - o[0], d[1] - o[1], d[2] - o[2], d[3] - o[3], d[4] - o[4]}
}

func (d Deviation) Dot(o [5]float64) float64 {
	return d[0]*o[0] + d[1]*o[1] + d[2]*o[2] + d[3]*o[3] + d[4]*o[4]
}

func (d Deviation) Sum() float64 {
	return d[0] + d[1] + d[2] + d[3] + d[4]
}

// ToProportion converts d (Σ=0) to display proportions (Σ=1).
// p[i] = 0.2 + d[i]/5
func (d Deviation) ToProportion() Proportion {
	var p Proportion
	for i := range d {
		p[i] = 0.2 + d[i]/5
	}
	return p
}

// ToDeviation converts p (Σ=1) back to a deviation (Σ=0).
func (p Proportion) ToDeviation() Deviation {
	var d Deviation
	for i := range p {
		d[i] = 5 * (p[i] - 0.2)
	}
	return d
}

func (p Proportion) Sum() float64 {
	return p[0] + p[1] + p[2] + p[3] + p[4]
}

