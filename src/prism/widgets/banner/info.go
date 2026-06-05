package banner

import (
	"time"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
)

// Info carries the per-render and per-session banner configuration.
// The presenter builds one of these once in OnBegin; the highway model
// stores it and reads from it on every render.
//
// Disable: hides the banner entirely (View returns "").
// Position: "top" or "bottom" - rendered outside the bordered region.
// Justify: "left" | "center" | "right" - horizontal alignment.
// Width: terminal width used for justification.
// Aspects: the random visual aspects chosen once at startup.
// Gradient: resolved colour endpoints from the "banner-control" theme.
// State: the per-widget gradient state (advanced on every banner tick).
// Tick: the per-tick interval (used to compute the skip factor).
type Info struct {
	Disable  bool
	Position string
	Justify  string
	Width    int
	Aspects  Aspects
	Gradient *contract.ResolvedGradient
	State    *effects.GradientState
	Tick     time.Duration
}
