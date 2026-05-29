package report

import (
	"context"
	"time"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
)

// DisplayEvent is the base event embedded into all UI events. It carries
// the node that triggered the event and an optional name identifying the
// action or pipeline that was invoked. Name is empty for NodeEvent.
type DisplayEvent struct {
	Node             *core.Node
	IsLast           bool
	Name             string
	IsPipelineHeader bool
	IsPipelineStep   bool
	IsLastStep       bool
}

// BeginEvent carries the metadata known at the start of a traversal.
// It is passed to Presenter.OnBegin before any node events are fired.
type BeginEvent struct {
	// Root is the top-level path being traversed.
	Root string

	// Caption is a human-readable description of the traversal options.
	Caption string

	// StartedAt is the time the traversal began.
	StartedAt time.Time

	// IsPrime indicates whether this is a fresh traversal. False means
	// this is a resume from a checkpoint.
	IsPrime bool

	// ResumeFrom is the path from which a resume traversal continues.
	// Populated only when IsPrime is false.
	ResumeFrom string

	// Subscription is the type of nodes being visited.
	Subscription enums.Subscription

	// ActionName is the name of the configured action, if any.
	ActionName string

	// PipelineName is the name of the configured pipeline, if any.
	PipelineName string

	// DateFormat is the Go time format string for rendering StartedAt
	// in the view header. Empty means use the default (time.RFC1123).
	DateFormat string

	// Cancel is the context cancellation function for the traversal.
	// When set, the presenter may call Cancel to abort the traversal
	// (e.g., in response to a user interrupt like Ctrl-C in the TUI).
	Cancel context.CancelFunc

	// NEW: Header display fields for filter flags
	CascadeDisplay string  // "🔒", "depth:<n>", or empty
	FilesGlob      string  // Pattern when --files-glob or -f used  
	FilesRegex     string  // Pattern when --files-regex used
	DirsGlob       string  // Pattern when --dirs-glob used
	DirsRegex      string  // Pattern when --dirs-regex used
	FileTypeMode   string  // "glob" or "regex" for files
	DirTypeMode    string  // "glob" or "regex" for dirs
}

// NeutralEvent is emitted per node visit when no action or pipeline is
// configured. The UI decides how to render the node information.
type NeutralEvent struct {
	DisplayEvent
}

// ActionEvent is emitted when a configured action has been executed
// against a node. ExecutionString is the composed CLI string that was
// (or would be) run - population of this field is a future concern.
type ActionEvent struct {
	DisplayEvent
	ExecutionString string
	CommandOutput   string
	DryRun          bool
	Err             error
}

// PipelineEvent is emitted when a configured pipeline has been executed
// against a node. A pipeline is a sequence of actions against the same
// node. ExecutionString is the composed CLI string - population of this
// field is a future concern.
type PipelineEvent struct {
	DisplayEvent
	ExecutionString string
	CommandOutput   string
	DryRun          bool
	Err             error
}

// SkipEvent is emitted when an action is skipped for a node because a
// placeholder in the action's cmd string resolved to a path at or above
// the traversal root. The navigation continues; this node is not counted
// as a successful action invocation.
type SkipEvent struct {
	DisplayEvent

	// Placeholder is the token that caused the breach, e.g. "{{.grand}}".
	Placeholder string

	// ResolvedPath is the path the offending placeholder resolved to.
	ResolvedPath string
}

// Traversal captures the outcome of a completed directory traversal.
// It is populated by the controller and handed to the UI via OnComplete.
// The UI decides how to present each field - colour, layout, and
// formatting are entirely its own concern.
type Traversal struct {
	// FilesVisited is the number of file nodes invoked during traversal.
	FilesVisited core.MetricValue

	// DirsVisited is the number of directory nodes invoked during traversal.
	DirsVisited core.MetricValue

	// ActionsSkipped is the number of nodes for which an action was skipped
	// because a placeholder breached the traversal root.
	ActionsSkipped core.NavigationMetric

	// Elapsed is the total wall-clock time taken for the traversal.
	Elapsed time.Duration

	// Err holds the traversal error, if any. Nil on success.
	Err error
}
