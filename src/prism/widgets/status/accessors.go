package status

import (
	"time"

	"charm.land/bubbles/v2/progress"
)

// Public accessors for Model state. These exist so callers
// (e.g. the highway root) can query the widget without going
// through View() parsing. They are read-only.

// Files returns the current files count.
func (m Model) Files() int { return m.files }

// Dirs returns the current directories count.
func (m Model) Dirs() int { return m.dirs }

// Errors returns the current errors count.
func (m Model) Errors() int { return m.errors }

// Skipped returns the current skipped count.
func (m Model) Skipped() int { return m.skipped }

// Elapsed returns the current elapsed duration.
func (m Model) Elapsed() time.Duration { return m.elapsed }

// Percent returns the current percent (0-100).
func (m Model) Percent() int { return m.percent }

// Total returns the current total count (files + dirs sum).
func (m Model) Total() int { return m.total }

// TotalFiles returns the preview file count, if set.
func (m Model) TotalFiles() int { return m.totalFiles }

// TotalDirs returns the preview directory count, if set.
func (m Model) TotalDirs() int { return m.totalDirs }

// Done returns the current done count.
func (m Model) Done() int { return m.done }

// HasTotal reports whether a TotalMsg has been received.
func (m Model) HasTotal() bool { return m.hasTotal }

// ErrMsg returns the captured error message, if any.
func (m Model) ErrMsg() string { return m.errMsg }

// Inner returns a pointer to the embedded bubbles progress model
// for tests and callers that need to inspect it directly. Not
// intended for production use.
func (m Model) Inner() *progress.Model { return &m.inner }
