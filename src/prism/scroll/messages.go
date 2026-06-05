package scroll

import (
	"time"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
)

// OvertureMsg carries the initial setup data for a porthole session.
// It is sent once at the start of traversal and contains all the
// metadata needed to render the banner, header, and footer chrome.
type OvertureMsg struct {
	Root              string
	Caption           string
	SubscriptionLabel string
	StartedAt         time.Time
	DateFormat        string
	PipelineName      string

	Header contract.HeaderInfo

	Banner banner.Info

	// FlagsRowPosition selects where the flags row is rendered.
	// contract.PositionTop places it after the top border;
	// contract.PositionBottom places it above the status line.
	// Empty or invalid values default to PositionBottom.
	FlagsRowPosition string
}

// RenderParams stores the raw parameters passed to flow.RenderLine
// so the line can be re-rendered when the terminal is resized.
type RenderParams struct {
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
	IsLast          bool
	IsPipelineStep  bool
	IsLastStep      bool
	VisualDepth     uint
}

// ContentLineMsg carries a single line of content to render. The
// porthole view buffers these lines in memory and renders them via
// the viewport widget when requested. RenderParams stores the raw
// flow.RenderLine arguments so the line can be re-rendered on
// terminal resize with the updated bodyWidth. BranchStack is the
// branch state AFTER this line was rendered, stored so the view can
// re-render the last line with the current activity frame.
type ContentLineMsg struct {
	Line        string
	Params      RenderParams
	BranchStack []bool
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
