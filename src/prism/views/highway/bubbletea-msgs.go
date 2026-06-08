package highway

import (
	"time"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
)

type OvertureMsg struct {
	contract.OvertureMsg

	ActionName string

	// Banner carries the optional ANSI shadow banner info. The
	// highway model uses banner.Info directly (no highway-specific
	// wrapper type). When the Disable flag is true or the gradient
	// is nil, the view skips it.
	Banner banner.Info
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
