package track

import (
	"time"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// Messages understood only by Model.Update. The track widget knows
// nothing about highway's OvertureMsg / BannerInfo vocabulary; the
// highway root translates them into the messages defined here.

// TickMsg is the per-tick forward from the highway root. The
// root drives tea.Tick (it owns the ticker) and wraps the
// time.Time payload in TickMsg before forwarding to the track
// child. The child uses the payload only as a signal; the
// payload value is not interpreted.
type TickMsg time.Time

// WidthMsg updates the rendered row width. The root forwards
// tea.WindowSizeMsg as WidthMsg so the per-lane layout re-flows.
type WidthMsg struct {
	Width int
}

// MotifMsg carries a single per-node event into the widget. The
// widget applies the data to the current lane, advances
// currentLaneIdx round-robin, dedups on Path, and increments
// files/dirs.
type MotifMsg struct {
	Data MotifData
}

// MotifData is the per-node payload. Moved verbatim from the
// highway package.
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

	// JobEmoji is the emoji associated with the incoming job,
	// rendered after the periscope bar.
	JobEmoji string

	// Gradient is the optional animation gradient to apply to
	// this lane's frame. Populated when
	// HighwayConfig.AnimationGradient is set in config; nil
	// otherwise.
	Gradient *contract.ResolvedGradient

	// PeriscopeGradient is the optional gradient for this lane's
	// periscope bar. Looked up via
	// theme.GradientFor(GradientComponentPeriscope) in sendMotif.
	PeriscopeGradient *contract.ResolvedGradient
}

// CensusMsg declares the maximum tree depth observed during the
// preview pass. Used by the periscope bar fill formula. TotalFiles
// and TotalDirs stay in the highway-level CensusMsg because the
// status widget (not track) needs them for the progress bar.
type CensusMsg struct {
	MaxDepth uint
}

// CompleteMsg is a flush signal: the root has signalled end of
// navigation. The track widget clears its counted map so any
// further MotifMsg re-becomes a no-op dedup-wise.
type CompleteMsg struct{}
