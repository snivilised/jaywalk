package contract

import (
	"time"

	"github.com/snivilised/jaywalk/src/agenor/enums"
)

// NodeParams holds the data for a single tree node being rendered.
// It is used by the flow package's RenderLine function and shared
// between the highway and porthole renderers.
type NodeParams struct {
	Path            string
	Name            string
	ActionName      string
	PipelineName    string
	ExecutionString string
	CommandOutput   string
	Err             error
	Depth           uint
	VisualDepth     uint
	IsDir           bool
	IsLast          bool
	IsPipelineStep  bool
	IsLastStep      bool
	DryRun          bool
	ActivityFrame   string
}

// RenderParams holds the rendering configuration shared between
// flow, highway, and porthole renderers.
type RenderParams struct {
	BodyWidth uint
	Theme     Theme
}

// NewModelParams holds the common constructor arguments shared by
// the highway and porthole bubbletea models. Embed this struct or
// pass it to NewModel to avoid repeating the same four parameters
// across both view constructors.
type NewModelParams struct {
	RootPath  string
	MaxDepth  uint
	Theme     Theme
	NoRecurse bool
}

// OvertureMsg carries the initial setup data sent once at the start
// of traversal. It contains the metadata needed to render the banner,
// header, and footer chrome. Highway and porthole views embed this
// struct and add their own view-specific fields (e.g. ActionName,
// Banner).
type OvertureMsg struct {
	Root              string
	Caption           string
	SubscriptionLabel string
	StartedAt         time.Time
	DateFormat        string
	PipelineName      string
	Header            HeaderInfo
	FlagsRowPosition  string
}

// WorkerStateMsg carries a per-lane worker state update. Sent from the
// highway presenter to the model when a pool worker's activity state
// changes. The highway model dispatches this to the track widget which
// controls per-lane animation.
type WorkerStateMsg struct {
	// LaneID is the index of the lane (derived from WorkerID % NoW).
	LaneID int

	// State is the worker's current activity state.
	State enums.WorkerState
}

// MotifMsg carries a single per-node event. It is the shared payload
// embedded by both the highway and track MotifMsg types. The highway
// root constructs it; the track widget consumes it to update lanes,
// dedup on path, and increment files/dirs counters.
type MotifMsg struct {
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

	// WorkerID is the pool-assigned goroutine ID that processed this
	// job, formatted as "W#N". Used to route the motif to the correct
	// lane.
	WorkerID string

	// JobEmoji is the emoji associated with the incoming job,
	// rendered after the periscope bar.
	JobEmoji string

	// Gradient is the optional animation gradient to apply to
	// this lane's frame. Populated when
	// HighwayConfig.AnimationGradient is set in config; nil
	// otherwise.
	Gradient *ResolvedGradient

	// PeriscopeGradient is the optional gradient for this lane's
	// periscope bar. Looked up via
	// theme.GradientFor(GradientComponentPeriscope) in sendMotif.
	PeriscopeGradient *ResolvedGradient
}

// CensusMsg carries the total file/dir counts from a preview
// traversal. The porthole model uses this to seed the status
// widget's progress bar so it can display a meaningful done/total
// ratio during navigation. When both totals are zero (no preview),
// the progress bar is omitted until the next CensusMsg arrives.
type CensusMsg struct {
	TotalFiles uint
	TotalDirs  uint
	MaxDepth   uint
}

// CompleteMsg marks end-of-navigation. The porthole view uses this
// to stop receiving content lines and render the closing summary.
// It carries file/dir counts, error list, elapsed time, and whether
// the traversal completed successfully (no errors).
type CompleteMsg struct {
	Files   int
	Dirs    int
	Errs    []error
	Elapsed time.Duration
}
