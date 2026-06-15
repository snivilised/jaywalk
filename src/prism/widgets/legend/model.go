package legend

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/layout"
	"github.com/snivilised/jaywalk/src/prism/widgets/cascade"
	"github.com/snivilised/jaywalk/src/prism/widgets/filter"
	"github.com/snivilised/jaywalk/src/prism/widgets/sampler"
)

// separator is the string that joins consecutive flag entries on the
// same wrapped line. The surrounding single spaces distinguish each
// entry from its neighbours, as required by the spec.
const separator = " | "

// Model is a thin rendering wrapper for the flags row (cascade /
// filter / sampler composition). It is value-typed and intended to
// be constructed on the fly inside a parent view's render method.
// There is no Init/Update interface - the data lives in the
// contract.HeaderInfo held by the caller and read on every render.
//
// Model is not a bubbletea tea.Model. The user has explicitly asked
// for the "create on the fly" pattern, not the long-lived child
// pattern used by track and status.
type Model struct {
	width  int
	info   Info
	styles Styles
}

// Option configures a Model at construction time.
type Option func(*Model)

// WithInfo attaches the frozen-per-session Info to the model.
func WithInfo(info Info) Option {
	return func(m *Model) { m.info = info }
}

// WithWidth overrides the render width. The highway view uses this
// to inject the current terminal width on every render (the value
// stored in the model is per-session and may go stale across
// WindowSizeMsg events).
func WithWidth(width int) Option {
	return func(m *Model) { m.width = width }
}

// WithStyles attaches the styles used to render the row. The caller
// is responsible for stripping the column width from the label style
// (see legend.Styles for details).
func WithStyles(styles Styles) Option {
	return func(m *Model) { m.styles = styles }
}

// NewModel constructs a Model. The returned value is discarded by
// the caller immediately after the View() call.
func NewModel(opts ...Option) Model {
	m := Model{}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// Height returns the number of terminal rows the rendered flag entries
// will occupy. The result is 0 when the position is empty/unrecognised
// or when no flag is active. When at least one flag is active, the
// height is the number of entry lines (no surrounding borders are
// included - those are the caller's responsibility). Callers (e.g.
// porthole's viewport body-height calculation) use this to reserve
// the right number of rows for the flag content.
func (m Model) Height() int {
	if m.info.Position != contract.PositionTop && m.info.Position != contract.PositionBottom {
		return 0
	}
	entries := m.compose()
	if len(entries) == 0 {
		return 0
	}
	available := m.width - 4
	if available < 1 {
		available = 1
	}
	lines := wrap(entries, available)
	return len(lines)
}

// View returns the rendered flags-row string, or "" when the
// position is empty / unrecognised or when no flag is active.
// The widget renders the flag entries only - surrounding borders
// are the caller's responsibility, since the layout around the
// flags section depends on the parent view's overall structure.
func (m Model) View() string {
	if m.info.Position != contract.PositionTop && m.info.Position != contract.PositionBottom {
		return ""
	}

	// width-4 accounts for the "│ " left cap and " │" right cap.
	available := m.width - 4
	if available < 1 {
		available = 1
	}

	entries := m.compose()
	if len(entries) == 0 {
		return ""
	}

	lines := wrap(entries, available)

	var b strings.Builder
	borderStyle := m.styles.BorderStyle
	for _, line := range lines {
		row := layout.NewRow(m.width-4).
			Caps(borderStyle.Render("│ "), borderStyle.Render(" │")).
			Content(line)
		row.RenderTo(&b)
		b.WriteString("\n")
	}

	return b.String()
}

// compose returns the individual flag entries in spec-mandated
// order. Each entry is already rendered with the appropriate style.
// Returns an empty slice when no flag is set.
//
// Order is fixed and matches the spec:
//  1. cascade (lock or depth)  - one entry, no label
//  2. filter labels             - one entry bundling all active patterns
//  3. sampler labels            - one entry with the 🐌 indicator and counts
func (m Model) compose() []string {
	hdr := m.info.Header
	labelStyle := m.styles.LabelStyle
	valueStyle := m.styles.ValueStyle

	var entries []string

	if entry := cascade.Render(hdr.CascadeDisplay, cascade.Styles{
		ValueStyle: valueStyle,
	}); entry != "" {
		entries = append(entries, entry)
	}

	if entry := filter.Render(filter.RenderParams{
		FilesGlob:    hdr.FilesGlob,
		FilesRegex:   hdr.FilesRegex,
		DirsGlob:     hdr.DirsGlob,
		DirsRegex:    hdr.DirsRegex,
		FileTypeMode: hdr.FileTypeMode,
		DirTypeMode:  hdr.DirTypeMode,
		Styles: filter.Styles{
			LabelStyle: labelStyle,
			ValueStyle: valueStyle,
		},
	}); entry != "" {
		entries = append(entries, entry)
	}

	if entry := sampler.Render(hdr.NumFiles, hdr.NumFolders, hdr.SampleLast, sampler.Styles{
		LabelStyle: labelStyle,
		ValueStyle: valueStyle,
	}); entry != "" {
		entries = append(entries, entry)
	}

	return entries
}

// wrap partitions a sequence of pre-rendered entries into lines
// whose visual width is no greater than maxWidth. Entries are
// joined by the supplied separator. When a single entry is itself
// wider than maxWidth, it is placed alone on its own line (we do
// not split within an entry).
func wrap(entries []string, maxWidth int) []string {
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
