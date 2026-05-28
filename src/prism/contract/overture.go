package contract

import "time"

// Overture carries the metadata known at the start of a traversal.
// Passed to Renderer.Begin to render the opening display.
type Overture struct {
	// Root is the top-level path being traversed.
	Root string

	// Caption is a human-readable description of the traversal options,
	// e.g. "files and folders".
	Caption string

	// StartedAt is the time the traversal began.
	StartedAt time.Time

	// Kind indicates whether this is a prime or resume traversal.
	Kind NavigationKind

	// ResumeFrom is the path from which a resume traversal continues.
	// Populated only when Kind == ResumeNavigation.
	ResumeFrom string

	// Survey holds the results of a prior survey phase. Nil for
	// single-phase navigations such as the linear view.
	Survey *SurveyResult

	// DateFormat is the Go time format string for rendering StartedAt.
	// Empty means use the default (time.RFC1123).
	DateFormat string
}
