package highway

import (
	"github.com/snivilised/jaywalk/src/prism/contract"
)

// Lane
type Lane struct {
	Emoji           string
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
	GradientState *GradientState

	// PeriscopeGradient is the animation gradient for this lane's periscope bar.
	// nil means no gradient configured - use default styling.
	PeriscopeGradient *contract.ResolvedGradient

	// PeriscopeGradientState holds in-memory state for this lane's periscope gradient animation.
	PeriscopeGradientState *GradientState
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
