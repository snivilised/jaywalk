package banner

import "math/rand/v2"

// RandomiseAspectsForTest is a thin alias for RandomiseAspects that
// exists for symmetry with other test exports in the codebase. Both
// names are interchangeable; the production code path uses
// RandomiseAspects directly.
func RandomiseAspectsForTest(rng *rand.Rand) Aspects {
	return RandomiseAspects(rng)
}
