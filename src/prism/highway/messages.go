package highway

import (
	"time"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
)

type OvertureMsg struct {
	Root              string
	Caption           string
	SubscriptionLabel string
	StartedAt         time.Time
	DateFormat        string
	ActionName        string
	PipelineName      string

	// Header groups the supplementary flag values for the flags row
	// renderer. See contract.HeaderInfo for the field semantics.
	Header contract.HeaderInfo

	// FlagsRowPosition selects where the flags row is rendered. Allowed
	// values are "top" and "bottom"; any other value is treated as
	// "bottom" (the default).
	FlagsRowPosition string

	// Banner carries the optional ANSI shadow banner info. The
	// highway model uses banner.Info directly (no highway-specific
	// wrapper type). When the Disable flag is true or the gradient
	// is nil, the view skips it.
	Banner banner.Info
}

// MotifData is the per-node payload sent to the track child widget.
// The root re-wraps incoming highway.MotifMsg values into
// track.MotifMsg before forwarding.
type MotifData struct {
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

	// JobEmoji is the emoji associated with the incoming job, rendered
	// after the periscope bar.
	JobEmoji string

	// Gradient is the optional animation gradient to apply to this lane's frame.
	// Populated when HighwayConfig.AnimationGradient is set in config; nil otherwise.
	Gradient *contract.ResolvedGradient

	// PeriscopeGradient is the optional gradient for this lane's periscope bar.
	// Looked up via theme.GradientFor(GradientComponentPeriscope) in sendMotif.
	PeriscopeGradient *contract.ResolvedGradient
}

// MotifMsg is the highway-level per-node event message. The root
// translates it into a track.MotifMsg before forwarding to the
// track child. Kept in the highway package because external
// callers (e.g. src/app/ui/highway.go) construct it directly.
type MotifMsg struct {
	Data MotifData
}

// CompleteMsg marks end-of-navigation. Carries the final
// file/dir counts, error list and elapsed time. The root stores
// errors/elapsed/errMsg/done in its own state, forwards a
// track.CompleteMsg{} to the track child (flush signal) and
// translates the relevant fields into status.CountsMsg,
// status.ElapsedMsg and status.DoneMsg for the status child.
type CompleteMsg struct {
	Files   int
	Dirs    int
	Errs    []error
	Elapsed time.Duration
}

// CensusMsg carries the total file/dir counts from a preview
// traversal. The root uses the counts to seed the status widget's
// progress total; the MaxDepth is forwarded to the track child
// for the periscope bar fill formula.
type CensusMsg struct {
	TotalFiles uint
	TotalDirs  uint
	MaxDepth   uint
}
