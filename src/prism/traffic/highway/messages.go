package highway

import (
	"time"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

type OvertureMsg struct {
	Root              string
	Caption           string
	SubscriptionLabel string
	StartedAt         time.Time
	DateFormat        string
	ActionName        string
	PipelineName      string
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
