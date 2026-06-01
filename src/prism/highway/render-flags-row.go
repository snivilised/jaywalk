package highway

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/layout"
	"github.com/snivilised/jaywalk/src/prism/widgets/cascade"
	"github.com/snivilised/jaywalk/src/prism/widgets/filter"
	"github.com/snivilised/jaywalk/src/prism/widgets/sampler"
)

// flagsRowSeparator is the separator that joins consecutive flag
// entries on the same wrapped line. The surrounding single spaces
// distinguish each entry from its neighbours, as required by the spec.
const flagsRowSeparator = " | "

// renderFlagsRow emits zero or more wrapped lines representing the
// supplementary flags (cascade / filters / sampler) and finishes with
// a border separator that closes the section. When no flag is active,
// the function is a no-op so callers do not need to guard.
func (m Model) renderFlagsRow(b *strings.Builder) {
	borderStyle := m.theme.BorderStyle
	// Strip the column width from SummaryLabelStyle: the closing summary
	// uses Width(16) to align labels in a column, but the flags row is
	// inline, so padding the label to 16 would push the colon away from
	// the label text (e.g. "files glob      : *.go" instead of the
	// desired "files glob: *.go"). Setting Width(0) on a lipgloss style
	// removes the width constraint.
	summaryLabelStyle := m.theme.SummaryLabelStyle.Width(0)
	summaryValueStyle := m.theme.SummaryValueStyle

	// Collect the active flag entries in the order specified by the spec:
	// cascade (lock or depth) → filter labels → sampler labels.
	entries := m.composeFlagsRowEntries(summaryLabelStyle, summaryValueStyle)
	if len(entries) == 0 {
		return
	}

	// Wrap the entries into lines that fit within the available width.
	// width-4 accounts for the "│ " left cap and " │" right cap.
	available := m.width - 4
	if available < 1 {
		available = 1
	}
	lines := wrapFlagsRow(entries, flagsRowSeparator, available)

	for _, line := range lines {
		row := layout.NewRow(m.width - 4).
			Caps(borderStyle.Render("│ "), borderStyle.Render(" │")).
			Content(line)
		row.RenderTo(b)
		b.WriteString("\n")
	}

	// Close the flags row section with a separator border, matching the
	// look of the existing top/lane separators.
	dashes := strings.Repeat("─", max(0, m.width-2))
	b.WriteString(borderStyle.Render("├" + dashes + "┤"))
	b.WriteString("\n")
}

// composeFlagsRowEntries returns the individual flag entries in
// spec-mandated order. Each entry is already rendered with the
// appropriate style. Returns an empty slice when no flag is set.
//
// summaryLabelStyle is applied to label text (e.g. "files glob",
// "#files", "🐌") and summaryValueStyle is applied to the value text
// (e.g. the pattern strings, the numeric counts). The cascade widget
// has no associated label, so its single value is rendered with
// summaryValueStyle.
func (m Model) composeFlagsRowEntries(
	summaryLabelStyle lipgloss.Style,
	summaryValueStyle lipgloss.Style,
) []string {
	var entries []string

	if entry := cascade.Render(m.header.CascadeDisplay, cascade.Styles{
		ValueStyle: summaryValueStyle,
	}); entry != "" {
		entries = append(entries, entry)
	}

	if entry := filter.Render(m.header.FilesGlob, m.header.FilesRegex, m.header.DirsGlob, m.header.DirsRegex,
		m.header.FileTypeMode, m.header.DirTypeMode, filter.Styles{
			LabelStyle: summaryLabelStyle,
			ValueStyle: summaryValueStyle,
		}); entry != "" {
		entries = append(entries, entry)
	}

	if entry := sampler.Render(m.header.NumFiles, m.header.NumFolders, m.header.SampleLast, sampler.Styles{
		LabelStyle: summaryLabelStyle,
		ValueStyle: summaryValueStyle,
	}); entry != "" {
		entries = append(entries, entry)
	}

	return entries
}

// wrapFlagsRow partitions a sequence of pre-rendered entries into
// lines whose visual width is no greater than maxWidth. Entries are
// joined by the supplied separator. When a single entry is itself
// wider than maxWidth, it is placed alone on its own line (we do not
// split within an entry).
func wrapFlagsRow(entries []string, separator string, maxWidth int) []string {
	if len(entries) == 0 {
		return nil
	}

	var lines []string
	var current []string
	currentWidth := 0

	for _, entry := range entries {
		w := lipgloss.Width(entry)
		sepWidth := lipgloss.Width(separator)

		switch {
		case len(current) == 0:
			current = append(current, entry)
			currentWidth = w

		case currentWidth+sepWidth+w <= maxWidth:
			current = append(current, entry)
			currentWidth += sepWidth + w

		default:
			lines = append(lines, strings.Join(current, separator))
			current = []string{entry}
			currentWidth = w
		}
	}

	if len(current) > 0 {
		lines = append(lines, strings.Join(current, separator))
	}

	return lines
}
