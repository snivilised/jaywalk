package controller

import (
	"sync"
	"time"

	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/app/report"
)

// WorkerStateTracker tracks per-worker activity state and notifies the
// UI presenter when each worker completes a job. The tracker is created
// when the shell pool executor is wired up and is nil otherwise.
//
// The only notification is [WorkerState.Idle] — we learn the WorkerID
// only at output time (the synchronous executor limitation). Setting
// the lane to idle on output, combined with the MotifMsg (which sets
// the lane to Working in applyMotifData), gives the UI a clean
// idle→working transition per job completion.
type WorkerStateTracker struct {
	mu         sync.Mutex
	lastWorker map[string]time.Time
	ui         report.Presenter
	now        uint
}

// NewWorkerStateTracker creates a tracker bound to the given presenter
// and pool size. NoW is the number of workers in the pool.
func NewWorkerStateTracker(ui report.Presenter, now uint) *WorkerStateTracker {
	return &WorkerStateTracker{
		lastWorker: make(map[string]time.Time),
		ui:         ui,
		now:        now,
	}
}

// OnOutput is called when a job output arrives via the executor. It
// records the worker's last activity time and notifies the UI that this
// specific worker is now idle.
// workerID must be formatted as "W#N".
func (t *WorkerStateTracker) OnOutput(workerID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.lastWorker[workerID] = time.Now()
	t.ui.OnWorkerState(enums.WorkerStateIdle, workerID)
}
