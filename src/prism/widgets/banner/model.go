package banner

import (
	"strings"
)

// Model is a thin rendering wrapper around the existing Render
// function. It is value-typed and intended to be constructed on the
// fly inside a parent view's render method. There is no Init/Update
// interface - the mutable state lives in *effects.GradientState, held
// by the caller and advanced via the Ticker helper.
//
// Model is not a bubbletea tea.Model. The user has explicitly asked
// for the "create on the fly" pattern, not the long-lived child
// pattern used by track and status.
type Model struct {
	width int
	info  Info
}

// Option configures a Model at construction time.
type Option func(*Model)

// WithInfo attaches the frozen-per-session Info to the model.
func WithInfo(info Info) Option {
	return func(m *Model) { m.info = info }
}

// WithWidth overrides the Width field on the captured Info. The
// highway view uses this to inject the current terminal width on
// every render (the Width stored in Info is the configured width and
// may go stale across WindowSizeMsg events).
func WithWidth(width int) Option {
	return func(m *Model) { m.width = width }
}

// NewModel constructs a Model. The returned value is discarded by
// the caller immediately after the View() call.
func NewModel(opts ...Option) Model {
	m := Model{}
	for _, opt := range opts {
		opt(&m)
	}
	if m.width == 0 {
		m.width = m.info.Width
	}
	return m
}

// Disabled reports whether the model should emit no output. View()
// returns "" in this case so callers do not need to guard.
func (m Model) Disabled() bool {
	return m.info.Disable || m.info.Gradient == nil || m.info.State == nil
}

// View returns the rendered banner string, or "" when disabled. The
// existing Render function is the single rendering core; NewModel is
// a thin per-render wrapper.
func (m Model) View() string {
	if m.Disabled() {
		return ""
	}
	return Render(
		Config{Width: m.width, Justify: m.info.Justify},
		Styles{},
		Effect{
			Gradient: m.info.Gradient,
			State:    m.info.State,
			Aspects:  m.info.Aspects,
		},
	)
}

// Height returns the number of terminal rows the rendered banner
// occupies. Returns 0 when disabled. Used by host views (highway,
// porthole) to budget vertical space for the banner without
// re-rendering it.
func (m Model) Height() int {
	if m.Disabled() {
		return 0
	}
	return strings.Count(m.View(), "\n")
}
