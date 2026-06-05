package banner

import (
	"time"

	"github.com/snivilised/jaywalk/src/prism/effects"
)

// Ticker encapsulates the per-tick gradient advance with skip-factor
// logic. The factor is computed from (bannerTick, globalTick) so the
// banner advances less often than the global tick rate, producing a
// warmer glow than the lane animations.
//
// nil-safe: (*Ticker).Advance is a no-op when t is nil or t.state is nil.
type Ticker struct {
	state   *effects.GradientState
	factor  int
	counter int
}

// NewTicker constructs a Ticker from the supplied gradient state and
// per-tick interval. skipFactor is computed from the global tickRate
// (passed in) and the banner's own tick. A skipFactor of 0 means the
// banner advances on every global tick (full speed); larger values
// mean it advances less often (slower glow). Returns nil when state
// is nil.
func NewTicker(state *effects.GradientState, bannerTick, globalTick time.Duration) *Ticker {
	if state == nil {
		return nil
	}
	if bannerTick <= 0 {
		bannerTick = DefaultBannerTick
	}
	factor := 0
	if globalTick > 0 {
		ns := bannerTick.Nanoseconds() / globalTick.Nanoseconds()
		if ns > 0 {
			factor = int(ns)
		}
	}
	return &Ticker{state: state, factor: factor}
}

// Advance should be called on every global tick. It increments the
// skipCounter and, when the threshold is reached, advances the
// gradient state by one step. The advance uses a windowSize of 1
// (one rune column at a time) so the sweep is gentle. nil-safe.
func (t *Ticker) Advance() {
	if t == nil || t.state == nil {
		return
	}
	if t.factor <= 0 {
		t.state.Update(1)
		return
	}
	t.counter++
	if t.counter >= t.factor {
		t.counter = 0
		t.state.Update(1)
	}
}

// State returns the wrapped gradient state. Useful for tests and
// for callers that need to seed the state with steps before the
// ticker is constructed.
func (t *Ticker) State() *effects.GradientState { return t.state }

// Factor returns the computed skip factor (banner tick / global tick).
// Returns 0 when factor was not computed (e.g. bannerTick < globalTick).
func (t *Ticker) Factor() int { return t.factor }

// Counter returns the current skip counter. Useful for tests.
func (t *Ticker) Counter() int { return t.counter }
