package contract

import (
	"time"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
)

// Summary carries the result of a completed traversal. Passed to
// Renderer.End to render the closing display.
type Summary struct {
	// FilesVisited is the count of files encountered.
	FilesVisited core.MetricValue

	// DirsVisited is the count of directories encountered.
	DirsVisited core.MetricValue

	// Skipped is the count of nodes for which actions were skipped.
	Skipped core.MetricValue

	// Elapsed is the total duration of the traversal.
	Elapsed time.Duration

	// Errors contains any errors encountered during traversal.
	Errors []error

	// Kind mirrors Overture.Kind so the renderer can label counts
	// appropriately in the closing summary.
	Kind enums.NavigationKind
}
