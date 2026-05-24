package highway

import (
	"github.com/snivilised/jaywalk/src/prism"
)

// FrameFunc
//
// This type probably needs to move into traffic, but before doing so, we need to
// make sure that the direction of dependency is correct. Currently, I think this
// is wrong. It alway feels wrong to me to have a child package dependent upon it
// parent. It should alway be the other way around. traffic should be dependent on
// highway, but I fear this is not the case; this needs to be verified.
type FrameFunc func(tick int) string

// Lane
type Lane struct {
	Emoji           string
	Label           string
	FrameFn         FrameFunc
	SpinnerName     string
	Path            string
	Name            string
	IsDir           bool
	Depth           uint
	ActionName      string
	PipelineName    string
	CommandOutput   string
	ExecutionString string
	DryRun          bool
	Err             error

	// IntervalMs controls this lane's animation speed via the config
	// override mechanism. The highway model's global tick fires every
	// 50ms; a lane with IntervalMs=5000 advances its frame only once
	// every 100 ticks (every 5 seconds). Value 0 means advance every
	// tick (fastest, no skip). Set via HighwayConfig.Overrides in
	// the user's jay.ui.yml under spinners.override.<name>.interval.
	IntervalMs int

	tick        int
	skipCounter int

	// New gradient field for Phase 2 implementation.
	// nil means no gradient configured - use default styling.
	HighlightGradient *prism.ResolvedGradient

	// GradientState holds in-memory state for this lane's gradient animation.
	GradientState *GradientState
}

func (l *Lane) WindowSize() int {
	if len(l.CommandOutput) > 0 {
		return len([]rune(l.CommandOutput)) // use command output width
	}
	if l.ActionName != "" && !l.IsDir {
		return 6 // default window size for action animations
	}
	return 4 // default window size
}

func (l *Lane) ResetGradient() {
	if l.GradientState == nil {
		l.GradientState = NewGradientState()
	}
	l.GradientState.Reset()
}
