// Package scrollbar provides a lightweight vertical scrollbar widget
// for bubbletea views. It is intentionally a transient, value-typed
// renderer (no Init/Update) that hosts compose on the fly inside
// their render method, mirroring the pattern used by
// prism/widgets/banner. State (scroll position, content height) is
// injected on each call rather than held inside the widget.
//
// The widget draws its rail using TreeIconBranchVertical from the
// host's theme, painted with the host's BranchStyle. The thumb is
// drawn with the host's MutedStyle so it reads as a subtle track
// marker rather than a heavy grab handle, matching the minimalist
// visual language of the rest of prism.
//
// When content fits within the viewport (or the host passes a
// non-positive height), the widget renders nothing and returns
// false from Visible. Hosts typically replace the entire gutter
// with a single blank column in that case.
package scrollbar

import (
	"strings"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// ScrollbarGutterWidth is the width (in columns) allocated for the
// scrollbar gutter column in the porthole view's body row. The
// scrollbar widget renders into this space; when content fits, the
// host replaces it with a blank column.
const ScrollbarGutterWidth = 1

// State carries the inputs the widget needs to compute thumb
// position and visibility. All values are read-only snapshots of
// the host's viewport state at the moment View is called; the
// widget does not retain references between calls.
type State struct {
	// Height is the visible viewport height in rows. A non-positive
	// value means the scrollbar is suppressed entirely.
	Height int

	// ContentLines is the total number of lines in the underlying
	// content. Zero or negative means there is no content.
	ContentLines int

	// Offset is the current top-row offset of the viewport into the
	// content. Clamped to [0, ContentLines-1] by the host before
	// passing in.
	Offset int
}

// Config carries the static rendering configuration. The widget
// resolves lipgloss styles lazily from the theme's BranchStyle and
// the optional thumb style; nothing else is needed.
type Config struct {
	// Theme provides BranchStyle, MutedStyle and the branch glyph.
	Theme contract.Theme
}

// View returns the rendered scrollbar as a single string with one
// row per viewport row. The string ends with a trailing newline so
// callers can write it into a strings.Builder without an extra
// separator. Returns "" when Height is non-positive or content fits
// in the viewport (the host should hide the gutter in that case).
func View(state State, cfg Config) string {
	if !Visible(state) {
		return ""
	}

	branch := cfg.Theme.TreeIcons[contract.TreeIconBranchVertical]
	if branch == "" {
		branch = "│"
	}

	thumb := cfg.Theme.MutedStyle

	thumbRow := computeThumbRow(state)
	rail := cfg.Theme.BranchStyle.Render(branch)

	var b strings.Builder
	for row := 0; row < state.Height; row++ {
		if row == thumbRow {
			b.WriteString(thumb.Render("█"))
		} else {
			b.WriteString(rail)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// Visible reports whether the widget would emit any output for the
// given state. Hosts consult this to decide whether to allocate
// the gutter column at all.
func Visible(state State) bool {
	return state.Height > 0 && state.ContentLines > state.Height
}

// computeThumbRow maps the viewport offset into the rail's
// [0, Height) range. The thumb is the rail cell that visually
// represents the current top edge of the viewport within the
// scrollable region.
//
// The mapping is "proportional on the scrollable range": the
// position of the viewport's top edge in the scrollable range
// (ContentLines - Height) is projected onto the rail (Height rows).
// The thumb is clamped to the last row when the viewport is at
// the bottom of the content.
func computeThumbRow(state State) int {
	scrollable := state.ContentLines - state.Height
	if scrollable <= 0 {
		return state.Height - 1
	}
	if state.Offset <= 0 {
		return 0
	}
	if state.Offset >= scrollable {
		return state.Height - 1
	}
	row := (state.Offset * (state.Height - 1)) / scrollable
	if row < 0 {
		return 0
	}
	if row >= state.Height {
		return state.Height - 1
	}
	return row
}
