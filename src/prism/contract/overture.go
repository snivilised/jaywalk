package contract

import (
	"time"

	"github.com/snivilised/jaywalk/src/agenor/enums"
)

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
	Kind enums.NavigationKind

	// ResumeFrom is the path from which a resume traversal continues.
	// Populated only when Kind == ResumeNavigation.
	ResumeFrom string

	// Survey holds the results of a prior survey phase. Nil for
	// single-phase navigations such as the linear view.
	Survey *SurveyResult

	// DateFormat is the Go time format string for rendering StartedAt.
	// Empty means use the default (time.RFC1123).
	DateFormat string

	// Banner carries the optional ANSI shadow banner configuration.
	// Nil when the banner is disabled or not configured.
	Banner *BannerInfo
}

// BannerInfo carries the ANSI shadow banner configuration for a renderer.
// The linear renderer uses this to render a static (non-animated) banner.
type BannerInfo struct {
	// Disable hides the banner when true.
	Disable bool

	// Position selects where the banner is rendered: "top" or "bottom".
	Position string

	// Justify is the horizontal alignment: "right", "left", or "center".
	Justify string

	// Width is the terminal width for justification padding.
	Width int

	// Aspects are the random visual aspects chosen once at startup.
	Aspects BannerAspects

	// Gradient is the resolved colour endpoints (Hi/Lo) bound to
	// the "banner-control" theme component.
	Gradient *ResolvedGradient
}

// BannerAspects captures the visual aspects for the banner.
type BannerAspects struct {
	Orientation int
	Banding     int
	Unity       int
	FixedEnd    int
}
