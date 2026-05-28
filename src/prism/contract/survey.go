package contract

// SurveyResult carries the output of a two-phase navigation survey
// pass. Populated by controller/dispatch after the survey phase and
// passed to the renderer via Overture. Nil means single-phase
// navigation - no survey was performed.
type SurveyResult struct {
	// NodeCount is the total nodes to be visited in the execute phase.
	// Enables accurate progress reporting.
	NodeCount uint

	// MaxDepth is the deepest level seen during the survey phase.
	// Used by views for layout calculations.
	MaxDepth uint
}
