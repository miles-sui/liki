package domain

import (
	"context"

	"github.com/25types/25types/internal/25types"
	"github.com/25types/25types/internal/ganzhi"
)

// ProfileLoader loads a user's latest personality profile.
// Used by both Bond computation (match context) and Flow computation (flow context).
type ProfileLoader interface {
	LoadProfile(ctx context.Context, userID int64) (*PersonalityProfile, error)
}

// PersonalityProfile is a value object representing a user's current personality state.
type PersonalityProfile struct {
	D        persona.Deviation  `json:"d"`
	P        persona.Proportion `json:"p"`
	Identity persona.Identity   `json:"identity"`
}

// NewProfile creates a PersonalityProfile from engine results.
func NewProfile(d persona.Deviation, p persona.Proportion, id persona.Identity) PersonalityProfile {
	return PersonalityProfile{D: d, P: p, Identity: id}
}

// ComputeProfileFromAnswers computes a full profile from raw answers.
func ComputeProfileFromAnswers(answers []persona.Answer) PersonalityProfile {
	d := persona.ComputeD(answers)
	p := persona.ComputeP(answers)
	id := persona.ClassifyIdentity(d, ganzhi.BuiltinPrototypes)
	return PersonalityProfile{D: d, P: p, Identity: id}
}

// PeerProfile is a peer-aggregated view of a user's personality.
type PeerProfile struct {
	Self      PersonalityProfile  `json:"self"`
	Peers     *PersonalityProfile `json:"peers_aggregated,omitempty"`
	Combined  *PersonalityProfile `json:"combined,omitempty"`
	PeerCount int                 `json:"peer_count"`
}

// Bond represents the interpersonal dynamics between two profiles.
type Bond struct {
	Self     persona.Deviation `json:"self"`
	Other    persona.Deviation `json:"other"`
	DeltaA   persona.Deviation `json:"delta_a"`
	DeltaB   persona.Deviation `json:"delta_b"`
	Concord persona.Concord  `json:"concord"`
}

// NewBond computes a Bond from two profiles.
func NewBond(a, b PersonalityProfile) Bond {
	result := persona.ComputeBond(a.D, b.D)
	return Bond{
		Self:     result.DEffA,
		Other:    result.DEffB,
		DeltaA:   result.DeltaA,
		DeltaB:   result.DeltaB,
		Concord: persona.ComputeConcord(a.D, b.D),
	}
}
