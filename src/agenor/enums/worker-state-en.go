package enums

//go:generate stringer -type=WorkerState -linecomment -trimprefix=WorkerState -output worker-state-en-auto.go

// WorkerState represents the lifecycle state of a pool worker.
type WorkerState uint

const (
	// WorkerStateUndefined represents the undefined worker state
	WorkerStateUndefined WorkerState = iota // undefined

	// WorkerStateIdle indicates the worker is not currently processing a job
	//
	WorkerStateIdle // idle

	// WorkerStateWorking indicates the worker is currently processing a job
	//
	WorkerStateWorking // working
)

// TransitionWorkerState returns desired if the transition from current to
// desired is valid, otherwise returns current. Valid transitions:
//
//	Idle    → Working   (job dispatched to pool)
//	Working → Idle      (job output received)
//
// Any other transition (including Undefined) returns current unchanged.
func TransitionWorkerState(current, desired WorkerState) WorkerState {
	switch current { //nolint:exhaustive // don't need to account for WorkerStateUndefined
	case WorkerStateIdle:
		if desired == WorkerStateWorking {
			return desired
		}
	case WorkerStateWorking:
		if desired == WorkerStateIdle {
			return desired
		}
	}
	return current
}
