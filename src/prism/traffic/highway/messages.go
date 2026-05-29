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

	// NEW: header information for filter widgets and depth indicator
	CascadeDisplay   string  // "🔒", "depth:<n>", or ""
	FilesGlob        string  // pattern value when --files-glob or -f used
	FilesRegex       string  // pattern value when --files-regex used
	DirsGlob         string  // pattern value when --dirs-glob used
	DirsRegex        string  // pattern value when --dirs-regex used
	FileTypeMode     string  // "glob" or "regex" for files
	DirTypeMode      string  // "glob" or "regex" for dirs
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
