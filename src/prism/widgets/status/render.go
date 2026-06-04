package status

import "time"

// Config is the one-shot stateless input set for Render. The
// highway view does NOT use Render — it constructs a Model, drives
// it via messages, and calls View(). The linear view (and any
// other non-bubbletea caller) uses Render so the row-composition
// logic is not duplicated.
type Config struct {
	// Files is the number of files visited.
	Files int

	// Dirs is the number of directories visited.
	Dirs int

	// Errors is the number of errors encountered.
	Errors int

	// Skipped is the number of items skipped.
	Skipped int

	// Elapsed is the time elapsed since traversal started.
	Elapsed time.Duration

	// Percent is the completion percentage (0-100). Only shown
	// when fields.ShowProgress is true.
	Percent int

	// IsDone indicates whether the traversal is complete.
	IsDone bool

	// ErrMsg is the error message shown when done with errors.
	ErrMsg string
}

// Render is a thin wrapper that builds a Model, applies the
// stateless inputs from cfg, and returns the rendered string.
// All row composition logic lives in Model.View() — Render is
// provided so the linear view (and any future non-bubbletea
// caller) can use the widget without managing bubbletea state.
// The returned string is the tea.View's Content field; callers
// that need a tea.View directly should construct the Model and
// call View() themselves.
func Render(cfg Config, styles Styles, fields FieldSelectors, width int) string {
	m := New(
		WithStyles(styles),
		WithFields(fields),
		WithWidth(width),
	)
	m.files = cfg.Files
	m.dirs = cfg.Dirs
	m.errors = cfg.Errors
	m.skipped = cfg.Skipped
	m.elapsed = cfg.Elapsed
	m.percent = clamp(cfg.Percent, 0, 100)
	m.isDone = cfg.IsDone
	m.errMsg = cfg.ErrMsg
	return m.View().Content
}
