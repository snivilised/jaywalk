package legend

import "charm.land/lipgloss/v2"

// Styles defines the styles used to render the legend widget.
//
// LabelStyle is applied to label text (e.g. "files glob" in
// "files glob: *.go") and to any indicator that has no associated
// value (e.g. the 🐌 emoji from the sampler widget).
//
// ValueStyle is applied to value text (e.g. the pattern strings,
// the numeric counts). The cascade widget has no associated label,
// so its single value is rendered with ValueStyle.
//
// BorderStyle is applied to the "│ " left cap, the " │" right cap
// and the closing "├─────┤" separator below the row.
//
// IMPORTANT: LabelStyle.Width(0) should be set by the caller.
// The closing summary uses Width(16) to align labels in a column;
// the flags row renders inline, so padding the label to 16 would
// push the colon away from the label text (e.g. "files glob      :
// *.go" instead of the desired "files glob: *.go"). The highway
// view is responsible for stripping the column width before
// handing the style to the legend widget.
type Styles struct {
	LabelStyle  lipgloss.Style
	ValueStyle  lipgloss.Style
	BorderStyle lipgloss.Style
}
