package workshop

import (
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

// GradientVisualiser renders a realtime visual
// representation of a gradient for display in the
// gradient workshop seed screen. Implementations are
// interchangeable — the workshop calls Render on every
// state change and displays the result without knowing
// which visualiser is active.
type GradientVisualiser interface {
	// Render produces the terminal string for the
	// current gradient. steps is the pre-interpolated
	// colour slice from the working state. animFrame
	// is the current animation tick counter, used by
	// animated visualisers to advance their state.
	Render(
		steps []contract.Color,
		curve enums.CurveKind,
		easing enums.EasingKind,
		animFrame int,
	) string

	// Name returns the display name shown in the
	// visualiser picker.
	Name() string
}

var visualiserRegistry []GradientVisualiser

// Register adds a visualiser to the package-level
// registry. Intended for use only from
// RegisterVisualisers.
func Register(v GradientVisualiser) {
	visualiserRegistry = append(visualiserRegistry, v)
}

// Visualisers returns a copy of the registered
// visualisers. The slice is safe to iterate; mutations
// do not affect the registry.
func Visualisers() []GradientVisualiser {
	out := make([]GradientVisualiser, len(visualiserRegistry))
	copy(out, visualiserRegistry)
	return out
}

// Reset clears the registry.
// Intended for test isolation only.
func Reset() {
	visualiserRegistry = visualiserRegistry[:0]
}

// RegisterVisualisers registers all GradientVisualiser
// implementations for use in the gradient workshop.
// Called only from the tweak bootstrap path.
func RegisterVisualisers() {
	Register(&WaveformVisualiser{})
	Register(&SweepVisualiser{})
	Register(&BloomVisualiser{})
	Register(&BandsVisualiser{})
}
