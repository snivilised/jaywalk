package highway

import (
	"time"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
)

// BannerInfo is the per-render state for the ANSI shadow banner. The
// presenter populates one of these in OnBegin (picking the random
// aspects once) and sends it via OvertureMsg. The model then drives
// the gradient state on every tick and renders the banner above or
// below the bordered region of the highway view.
type BannerInfo struct {
	// Disable hides the banner when true. The widget is still
	// constructed but the view skips it.
	Disable bool

	// Position selects whether the banner is rendered above the top
	// border ("top") or below the bottom border ("bottom"). The view
	// treats any other value as "top".
	Position string

	// Justify is the horizontal alignment of the banner: "right",
	// "left" or "center". The widget handles all three.
	Justify string

	// Width is the terminal width. The widget uses it for
	// justification padding.
	Width int

	// Aspects are the random visual aspects chosen once at startup.
	// See banner.Aspects for the orthogonal values.
	Aspects BannerAspects

	// Gradient is the resolved colour endpoints (Hi/Lo) bound to
	// the "banner-control" theme component.
	Gradient *contract.ResolvedGradient

	// State is the per-widget gradient state advanced on each
	// banner tick. The highway model owns the lifecycle of this
	// pointer.
	State *effects.GradientState

	// Tick is the per-tick interval in milliseconds for the banner's
	// gradient animation. The model uses this to compute the skip
	// factor relative to the global tick rate.
	Tick time.Duration
}

// BannerAspects is the subset of banner.Aspects the highway model
// needs. The full type lives in the banner package; this is the
// pre-computed view produced by the presenter.
type BannerAspects struct {
	Orientation int
	Banding     int
	Unity       int
	FixedEnd    int
}

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

	// Banner carries the optional ANSI shadow banner info. When the
	// Disable flag is true or the gradient is nil, the view skips it.
	Banner BannerInfo
}

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

type MotifMsg struct {
	Data MotifData
}

type CompleteMsg struct {
	Files   int
	Dirs    int
	Errs    []error
	Elapsed time.Duration
}

// CensusMsg carries the total file/dir counts from a preview traversal.
// The model uses these to calculate progress percentage during the live pass.
type CensusMsg struct {
	TotalFiles uint
	TotalDirs  uint
	MaxDepth   uint
}
