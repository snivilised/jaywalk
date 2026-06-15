package contract

import (
	"github.com/snivilised/jaywalk/src/agenor/core"
)

// Motif is the unit of render-able content passed to Renderer.Show for
// each item encountered during traversal. Fields are generic filesystem
// and execution concepts - no jaywalk or agenor types appear here.
// Depth is sourced from node.Extension.Level in agenor.
type Motif struct {
	// Path is the full path of the item.
	Path string

	// Name is the base name of the item.
	Name string

	// IsDir indicates whether the item is a directory.
	IsDir bool

	// Depth is the number of levels below the traversal root, sourced
	// from node.Extension.Level in agenor.
	Depth core.TraversalDepth

	// VisualDepth is the visual indent level for this item. For directories
	// this is the same as Depth, but for files it is Depth+1 since they are
	// visually one level deeper than their parent directory.
	VisualDepth core.TraversalDepth

	// ActionName is the name of the action executed against this node.
	// Empty when no action was configured.
	ActionName string

	// PipelineName is the name of the pipeline executed against this node.
	// Empty when no pipeline was configured.
	PipelineName string

	// ExecutionString is the expanded command string for dry-run display.
	ExecutionString string

	// CommandOutput is the captured output of the command execution.
	CommandOutput string

	// DryRun indicates if this is a dry-run execution.
	DryRun bool

	// Skipped is true when an action or pipeline was skipped because a
	// placeholder resolved to a path at or above the traversal root.
	Skipped bool

	// Placeholder is the placeholder string that caused the skip.
	// Populated only when Skipped is true.
	Placeholder string

	// ResolvedPath is the path the placeholder resolved to.
	// Populated only when Skipped is true.
	ResolvedPath string

	// Err is any error produced by the action or pipeline for this node.
	// Nil when the node was visited without error.
	Err error

	// IsLast is true when this is the last motif to be rendered in the traversal.
	IsLast bool

	// IsPipelineHeader is true when this motif represents the start of a pipeline.
	IsPipelineHeader bool

	// IsPipelineStep is true when this motif represents an individual action
	// within a pipeline.
	IsPipelineStep bool

	// IsLastStep is true when this is the last action in a pipeline.
	IsLastStep bool
}
