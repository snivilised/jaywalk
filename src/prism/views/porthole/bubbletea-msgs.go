package porthole

import (
	"time"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
)

// OvertureMsg carries the initial setup data for a porthole session.
// It is sent once at the start of traversal and contains all the
// metadata needed to render the banner, header, and footer chrome.
type OvertureMsg struct {
	contract.OvertureMsg

	Banner banner.Info
}

// RenderParams stores the raw parameters passed to linear.RenderLine
// so the line can be re-rendered when the terminal is resized.
type RenderParams struct {
	contract.NodeParams
}

// ContentLineMsg carries a single line of content to render. The
// porthole view buffers these lines in memory and renders them via
// the viewport widget when requested. RenderParams stores the raw
// linear.RenderLine arguments so the line can be re-rendered on
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
