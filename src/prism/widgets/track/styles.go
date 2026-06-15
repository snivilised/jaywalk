package track

import (
	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// Styles holds the lipgloss styles used by the track widget at
// render time. Populated either directly via WithStyles or from a
// resolved theme via WithTheme. Mirrors the status.Styles pattern.
type Styles struct {
	BarFilledStyle    lipgloss.Style
	BarEmptyStyle     lipgloss.Style
	ErrorStyle        lipgloss.Style
	ActionStyle       lipgloss.Style
	PipelineStyle     lipgloss.Style
	DirStyle          lipgloss.Style
	FileStyle         lipgloss.Style
	MutedStyle        lipgloss.Style
	TreeIcons         contract.TreeIcons
	FrameStyle        lipgloss.Style
	BorderStyle       lipgloss.Style
	BranchStyle       lipgloss.Style
	LandingStripStyle lipgloss.Style

	// IdleStyle is applied to lane content when the worker is idle.
	// Typically a muted/dimmed style. Wired from theme.WorkerIdleStyle.
	IdleStyle lipgloss.Style

	// WorkingStyle is applied to lane content when the worker is
	// working. Wired from theme.WorkerStyle.
	WorkingStyle lipgloss.Style
}
