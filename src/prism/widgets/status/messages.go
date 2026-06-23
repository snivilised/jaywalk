package status

import (
	"time"
)

// Messages understood only by Model.Update. The status widget knows
// nothing about highway's MotifMsg/CompleteMsg/CensusMsg vocabulary;
// the highway root translates them into the messages defined here.

// WidthMsg updates the rendered row width. The root forwards
// tea.WindowSizeMsg as WidthMsg so the row caps re-linear.
type WidthMsg struct {
	Width int
}

// CountsMsg sets the four count fields atomically. The root uses this
// when a CompleteMsg arrives and on each MotifMsg that flips the
// counted map. Incremental updates use IncDoneMsg instead.
type CountsMsg struct {
	Files   int
	Dirs    int
	Errors  int
	Skipped int
}

// ElapsedMsg updates the elapsed duration. The root pushes this on
// every tick (live elapsed) and on CompleteMsg (final elapsed).
type ElapsedMsg struct {
	Elapsed time.Duration
}

// PercentMsg sets the displayed percentage directly. Used by the
// highway root in demo mode where the percent is a function of
// elapsed time, not a real file-count. Clamped to [0, 100].
type PercentMsg struct {
	Percent int
}

// TotalMsg declares the expected total file and directory counts
// from a preview traversal. Total is the sum (TotalFiles + TotalDirs)
// used for progress bar computation. TotalFiles and TotalDirs are the
// individual preview counts displayed in the status row so the user
// can compare live progress against the preview estimate.
type TotalMsg struct {
	Total      int
	TotalFiles int
	TotalDirs  int
}

// DoneMsg records the final counts and marks the widget as
// completed. IsDone drives the "✔ complete" / "❌ Failed" segment
// in the rendered row. Err populates the error message when
// non-empty.
type DoneMsg struct {
	Done   int
	IsDone bool
	Err    string
}

// IncDoneMsg increments the done counter atomically. When N is
// zero it defaults to 1. After incrementing, the widget
// recomputes percent from done/total (if a TotalMsg was previously
// seen) and re-targets the embedded spring.
type IncDoneMsg struct {
	N int
}

// ResetMsg returns the widget to its initial state. Used between
// sessions.
type ResetMsg struct{}
