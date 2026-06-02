package highway

import (
	"time"

	"github.com/snivilised/jaywalk/src/prism/effects"
)

// bannerState holds the per-render state of the ANSI shadow banner.
// The GradientState is the same primitive that the lane animations
// use (see effects/gradient-state.go). The skipCounter advances one
// step every tickRate/bannerTick global ticks, producing a slower,
// warmer glow than the lane animations.
type bannerState struct {
	gradient    *effects.GradientState
	skipCounter int
	skipFactor  int
	tick        time.Duration
}

// newBannerState constructs a bannerState from the supplied gradient
// state and per-tick interval. skipFactor is computed from the
// global tickRate (passed in) and the banner's own tick. A skipFactor
// of 0 means the banner advances on every global tick (full speed);
// larger values mean it advances less often (slower glow).
func newBannerState(gs *effects.GradientState, bannerTick, globalTick time.Duration) *bannerState {
	if gs == nil {
		return nil
	}
	if bannerTick <= 0 {
		bannerTick = 500 * time.Millisecond
	}
	factor := 0
	if globalTick > 0 {
		ns := bannerTick.Nanoseconds() / globalTick.Nanoseconds()
		if ns > 0 {
			factor = int(ns)
		}
	}
	return &bannerState{
		gradient:    gs,
		skipFactor:  factor,
		tick:        bannerTick,
		skipCounter: 0,
	}
}

// advance should be called on every global tick. It increments the
// skipCounter and, when the threshold is reached, advances the
// gradient state by one step. The advance uses a windowSize of 1
// (one rune column at a time) so the sweep is gentle.
func (b *bannerState) advance() {
	if b == nil || b.gradient == nil {
		return
	}
	if b.skipFactor <= 0 {
		b.gradient.Update(1)
		return
	}
	b.skipCounter++
	if b.skipCounter >= b.skipFactor {
		b.skipCounter = 0
		b.gradient.Update(1)
	}
}
