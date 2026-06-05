package track

import (
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
)

// Lane holds the per-worker state for a single track lane. The
// highway root constructs the initial slice and the track widget
// owns the per-tick and per-motif updates from there on.
type Lane struct {
	Emoji           string
	JobEmoji        string
	Label           string
	FrameFn         contract.FrameFunc
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

	// HighlightGradient is the animation gradient for this lane's activity frame.
	// nil means no gradient configured - use default styling.
	HighlightGradient *contract.ResolvedGradient

	// GradientState holds in-memory state for this lane's activity gradient animation.
	GradientState *effects.GradientState

	// PeriscopeGradient is the animation gradient for this lane's periscope bar.
	// nil means no gradient configured - use default styling.
	PeriscopeGradient *contract.ResolvedGradient

	// PeriscopeGradientState holds in-memory state for this lane's periscope gradient animation.
	PeriscopeGradientState *effects.GradientState
}

// WindowSize returns the gradient window size appropriate for the
// lane's current content. The same heuristic that previously lived
// on the highway Lane type: command output length when present,
// otherwise 6 for action animations on files, otherwise 4.
func (l *Lane) WindowSize() int {
	if len(l.CommandOutput) > 0 {
		return len([]rune(l.CommandOutput))
	}
	if l.ActionName != "" && !l.IsDir {
		return 6
	}
	return 4
}

// ResetGradient re-initialises the highlight gradient state pointer
// (creating one if absent) and resets it. Used when the same lane
// starts a new motif cycle.
func (l *Lane) ResetGradient() {
	if l.GradientState == nil {
		l.GradientState = effects.NewGradientState()
	}
	l.GradientState.Reset()
}
