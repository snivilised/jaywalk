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
}
