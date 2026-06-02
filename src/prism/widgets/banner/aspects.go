package banner

import "math/rand/v2"

// Orientation selects whether the gradient sweeps across rows or down
// columns of the banner art.
type Orientation int

const (
	// OrientationHorizontal sweeps left-to-right (or right-to-left,
	// depending on the gradient state's initial direction) across the
	// banner. The position used to look up the gradient step is the
	// rune's column index.
	OrientationHorizontal Orientation = iota

	// OrientationVertical sweeps top-to-bottom (or bottom-to-top)
	// across the banner. The position used to look up the gradient
	// step is the rune's row index.
	OrientationVertical
)

// Banding controls how the gradient repeats once it reaches the end of
// the banner.
type Banding int

const (
	// BandingWithout renders a single sweep from one end of the
	// banner to the other. The gradient state's natural reverse-on-end
	// behaviour still applies (so the sweep bounces back rather than
	// jumping to the start).
	BandingWithout Banding = iota

	// BandingWith repeats the gradient using the number of steps in
	// the gradient definition. The reverse-on-end behaviour avoids
	// hard contrast jumps at the turnaround.
	BandingWith
)

// Unity controls which class of characters receives the gradient sweep.
type Unity int

const (
	// UnityUnified applies the gradient to both face and shadow
	// characters. FixedEnd is ignored.
	UnityUnified Unity = iota

	// UnityGradientFace applies the gradient to face characters only;
	// shadow characters are fixed to either the gradient's Hi or Lo
	// colour depending on FixedEnd.
	UnityGradientFace

	// UnityShadowFace applies the gradient to shadow characters only;
	// face characters are fixed to either the gradient's Hi or Lo
	// colour depending on FixedEnd.
	UnityShadowFace
)

// FixedEnd selects which end of the gradient the non-swept character
// class is locked to. It is only meaningful when Unity is not Unified.
type FixedEnd int

const (
	// FixedEndUnfixed indicates that the FixedEnd is irrelevant
	// (Unity is Unified). It is the value assigned by randomiseAspects
	// when Unity is Unified.
	FixedEndUnfixed FixedEnd = iota

	// FixedEndHi locks the non-swept character class to the gradient's
	// Hi colour.
	FixedEndHi

	// FixedEndLo locks the non-swept character class to the gradient's
	// Lo colour.
	FixedEndLo
)

// Aspects captures the three orthogonal visual aspects selected at
// startup. They remain in effect for the entire application lifetime -
// the random selection happens once, not per-render.
type Aspects struct {
	Orientation Orientation
	Banding     Banding
	Unity       Unity
	// FixedEnd is only meaningful when Unity is not Unified; it is
	// set to FixedEndUnfixed by randomiseAspects when Unity is Unified.
	FixedEnd FixedEnd
}

// randomiseAspects picks the four aspect values that govern the banner's
// animation for the duration of the process. The call is expected to
// happen exactly once during startup (in the highway presenter) so the
// chosen aspects remain in effect for the entire application lifetime.
//
// Selection logic:
//   - Orientation - 50/50 horizontal/vertical
//   - Banding     - 50/50 without/with
//   - Unity       - uniform pick over the three values
//   - FixedEnd    - 50/50 hi/lo, but only when Unity is not Unified.
//     When Unity is Unified, FixedEnd is set to FixedEndUnfixed so
//     downstream code can detect "irrelevant" by a single equality.
func randomiseAspects(rng *rand.Rand) Aspects {
	orientation := OrientationHorizontal
	if rng.IntN(2) == 1 {
		orientation = OrientationVertical
	}

	banding := BandingWithout
	if rng.IntN(2) == 1 {
		banding = BandingWith
	}

	unity := Unity(rng.IntN(3))

	fixedEnd := FixedEndUnfixed
	if unity != UnityUnified {
		if rng.IntN(2) == 0 {
			fixedEnd = FixedEndHi
		} else {
			fixedEnd = FixedEndLo
		}
	}

	return Aspects{
		Orientation: orientation,
		Banding:     banding,
		Unity:       unity,
		FixedEnd:    fixedEnd,
	}
}

// RandomiseAspects is the public entry point for aspect randomisation.
// Callers (typically the highway presenter) are expected to call this
// exactly once at startup with a freshly-seeded *rand.Rand so the
// chosen aspects remain in effect for the entire application
// lifetime. The test-only export_test.go wrapper delegates here.
func RandomiseAspects(rng *rand.Rand) Aspects {
	return randomiseAspects(rng)
}
