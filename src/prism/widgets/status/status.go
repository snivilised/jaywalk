package status

import (
	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// Styles holds the lipgloss styles for the status row. Shared by
// both the bubbletea Model (via WithStyles/WithTheme) and the
// stateless Render wrapper used by the linear view.
type Styles struct {
	// TreeIcons are the glyphs used for field icons.
	TreeIcons contract.TreeIcons

	// SummaryLabelStyle is applied to field labels.
	SummaryLabelStyle lipgloss.Style

	// SummaryValueStyle is applied to field values.
	SummaryValueStyle lipgloss.Style

	// ErrorStyle is applied to error-related content.
	ErrorStyle lipgloss.Style

	// ProgressStyle is applied to progress-related content.
	ProgressStyle lipgloss.Style

	// BorderStyle is applied to border characters.
	BorderStyle lipgloss.Style

	// MutedStyle is applied to secondary text.
	MutedStyle lipgloss.Style
}

// FieldSelectors controls which segments are rendered in the
// status row. Shared by both the bubbletea Model (via WithFields)
// and the stateless Render wrapper.
type FieldSelectors struct {
	// ShowFiles enables the files count segment.
	ShowFiles bool

	// ShowDirs enables the directories count segment.
	ShowDirs bool

	// ShowErrors enables the errors count segment.
	ShowErrors bool

	// ShowSkipped enables the skipped count segment.
	ShowSkipped bool

	// ShowProgress enables the progress bar and percentage segment.
	ShowProgress bool

	// ShowComplete enables the complete/failed message segment.
	ShowComplete bool

	// ShowElapsed enables the elapsed time segment.
	ShowElapsed bool
}
