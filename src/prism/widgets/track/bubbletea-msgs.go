package track

import (
	"time"

	"github.com/snivilised/jaywalk/src/agenor/enums"
)

// Messages understood only by Model.Update. The track widget knows
// nothing about highway's OvertureMsg / BannerInfo vocabulary; the
// highway root translates them into the messages defined here.

// TickMsg is the per-tick forward from the highway root. The
// root drives tea.Tick (it owns the ticker) and wraps the
// time.Time payload in TickMsg before forwarding to the track
// child. The child uses the payload only as a signal; the
// payload value is not interpreted.
type TickMsg time.Time

// WidthMsg updates the rendered row width. The root forwards
// tea.WindowSizeMsg as WidthMsg so the per-lane layout re-flows.
type WidthMsg struct {
	Width int
}

// CensusMsg declares the maximum tree depth observed during the
// preview pass. Used by the periscope bar fill formula. TotalFiles
// and TotalDirs stay in the highway-level CensusMsg because the
// status widget (not track) needs them for the progress bar.
type CensusMsg struct {
	MaxDepth uint
}

// CompleteMsg is a flush signal: the root has signalled end of
// navigation. The track widget clears its counted map so any
// further MotifMsg re-becomes a no-op dedup-wise.
type CompleteMsg struct{}

// WorkerStateMsg carries a per-lane worker state update from the
// highway root. The track widget updates the specified lane's state.
// The gradient is NOT reset on transition to working — it continues
// from its current position to avoid harsh visual contrast.
type WorkerStateMsg struct {
	LaneID int
	State  enums.WorkerState
}

// enums imported from github.com/snivilised/jaywalk/src/agenor/enums
