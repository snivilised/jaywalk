package highway

import (
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
	"github.com/snivilised/jaywalk/src/prism/widgets/status"
	"github.com/snivilised/jaywalk/src/prism/widgets/track"
)

type tickMsg time.Time

// statusFieldSet is the canonical FieldSelectors used by the
// status widget in the highway view. Hoisted to a package-level
// value so the constructor and any tests share a single source of
// truth.
func statusFieldSet() status.FieldSelectors {
	return status.FieldSelectors{
		ShowFiles:    true,
		ShowDirs:     true,
		ShowErrors:   true,
		ShowSkipped:  false,
		ShowProgress: true,
		ShowComplete: true,
		ShowElapsed:  true,
	}
}

// Model is the bubbletea Model for the highway view. It owns the
// chrome (header, flags row, summary, banner), the status child
// widget, and the track child widget. The track child is
// responsible for lane data, per-tick advance, motif application
// and per-lane rendering.
type Model struct {
	contract.TraverseModel

	// track is the child widget owning lane data, per-tick
	// advance, motif application and per-lane rendering. The
	// root forwards the relevant tea.Msg values to the child
	// (see Message flow in
	// make-lanes-its-own-child-widget.implementation-plan.issue-604.md).
	track track.Model

	// bannerInfo is the (immutable for the session) configuration
	// received via OvertureMsg. The view reads this each render to
	// construct a transient banner.Model on the fly. The Ticker is
	// the long-lived state driver; nil when the banner is disabled,
	// when the gradient binding is absent, or when the state pointer
	// is nil. Advance() is nil-safe.
	bannerInfo   banner.Info
	bannerTicker *banner.Ticker

	// status is the child status widget. It owns the
	// files/dirs/errors/elapsed/isDone/errMsg/percent/total state
	// and the embedded bubbles progress bar. The root
	// translates highway messages into status.* messages
	// (see update.go's translation helpers).
	status status.Model

	totalTicks int64
	realMode   bool
	totalFiles uint
	totalDirs  uint
}

// NewModel constructs a highway view Model. The lanes slice is
// passed to the track child widget which owns it from then on.
// The theme is split: a copy of the theme fields the track widget
// needs is captured into track.WithTheme; the rest stay on the
// root for chrome rendering.
func NewModel(params contract.NewModelParams, lanes []track.Lane, tickRate time.Duration) Model {
	return Model{
		TraverseModel: contract.TraverseModel{
			Width:     80,
			RootPath:  params.RootPath,
			MaxDepth:  params.MaxDepth,
			NoRecurse: params.NoRecurse,
			TickRate:  tickRate,
			Theme:     params.Theme,
		},
		status: status.New(
			status.WithTheme(params.Theme),
			status.WithFields(statusFieldSet()),
			status.WithWidth(10),
		),
		track: track.New(
			track.WithLanes(lanes),
			track.WithTheme(params.Theme),
			track.WithTickRate(tickRate),
			track.WithMaxDepth(params.MaxDepth),
			track.WithNoRecurse(params.NoRecurse),
			track.WithWidth(80),
		),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Tick(m.TickRate, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		var cmd tea.Cmd
		m.status, cmd = m.dispatchStatus(status.WidthMsg{Width: msg.Width})
		m.track, _ = m.dispatchTrack(track.WidthMsg{Width: msg.Width})
		return m, cmd

	case tea.KeyMsg:
		switch {
		case m.Done && msg.String() == "space":
			return m, tea.Quit
		case msg.String() == "ctrl+c":
			return m, tea.Quit
		}
		return m, nil

	case tickMsg:
		cmds := make([]tea.Cmd, 0, 2)
		if m.Done {
			return m, nil
		}
		if m.Start.IsZero() && !m.realMode {
			m.Start = core.Now()
		}
		m.totalTicks++
		// Forward the tick to the track child. Track advances
		// each lane's tick counter, applies the per-lane skip
		// factor and advances gradient states.
		m.track, _ = m.dispatchTrack(track.TickMsg(time.Time(msg)))
		// Advance the banner's gradient state on its own slower
		// tick so its warm glow is visibly different from the
		// lane animations. The Ticker encapsulates the skip
		// factor and the skip counter; Advance is nil-safe.
		m.bannerTicker.Advance()
		tickCmd := tea.Tick(m.TickRate, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
		cmds = append(cmds, tickCmd)
		if !m.Start.IsZero() {
			// Elapsed time is real in both demo and real mode,
			// so push it on every tick. Without this the status
			// row's elapsed segment would stay at 0 in real
			// mode until the final CompleteMsg arrived.
			elapsed := time.Since(m.Start)
			var elapsedCmd tea.Cmd
			m.status, elapsedCmd = m.dispatchStatus(status.ElapsedMsg{Elapsed: elapsed})
			if elapsedCmd != nil {
				cmds = append(cmds, elapsedCmd)
			}
			if !m.realMode {
				// Demo mode only: push a time-derived percent
				// so the bar animates without real traversal
				// data. In real mode the percent is driven by
				// TotalMsg + IncDoneMsg (the done/total ratio).
				// TODO(realMode-cleanup): once the demo-mode
				// fake percent generation moves into
				// status.Update (or track.Update), the root
				// no longer needs to know about it.
				m.status, _ = m.dispatchStatus(status.PercentMsg{
					Percent: int(elapsed.Seconds()) * 2 % 100,
				})
			}
		}
		return m, tea.Batch(cmds...)

	case OvertureMsg:
		m.Start = core.Now()
		m.realMode = true
		m.ApplyOverture(&msg.OvertureMsg)

		// Initialise the banner from the OvertureMsg. The bannerInfo
		// is stored verbatim; bannerTicker is nil when the banner is
		// disabled, when the gradient binding is absent, or when
		// the state pointer is nil. The view constructs a transient
		// banner.Model on the fly from bannerInfo per render.
		m.bannerInfo = msg.Banner
		if !msg.Banner.Disable && msg.Banner.State != nil && msg.Banner.Gradient != nil {
			m.bannerTicker = banner.NewTicker(msg.Banner.State, msg.Banner.Tick, m.TickRate)
		} else {
			m.bannerTicker = nil
		}

		return m, nil

	case contract.CensusMsg:
		m.totalFiles = msg.TotalFiles
		m.totalDirs = msg.TotalDirs
		if msg.MaxDepth > m.MaxDepth {
			m.MaxDepth = msg.MaxDepth
		}
		// Forward the max depth to the track child for the
		// periscope bar fill formula. The track package only
		// needs MaxDepth, not the file/dir counts (those stay
		// at the root and go to the status widget).
		m.track, _ = m.dispatchTrack(track.CensusMsg{MaxDepth: msg.MaxDepth})
		// Forward the total to the status widget so it can
		// compute the percent on its own. The total must include
		// both files AND dirs because every MotifMsg (regardless
		// of IsDir) translates to a single IncDoneMsg{N:1}
		// downstream. Seeding the total with only TotalFiles
		// makes done exceed total during navigation, clamping
		// the bar to 100% well before completion.
		totalForProgress := msg.TotalFiles + msg.TotalDirs
		if totalForProgress > 0 {
			var cmd tea.Cmd
			m.status, cmd = m.dispatchStatus(status.TotalMsg{Total: int(totalForProgress)}) //nolint:gosec // cast ok
			return m, cmd
		}
		return m, nil

	case contract.WorkerStateMsg:
		m.track, _ = m.dispatchTrack(track.WorkerStateMsg{
			LaneID: msg.LaneID,
			State:  msg.State,
		})
		return m, nil

	case contract.MotifMsg:
		// Track the pre-dispatch file/dir counts so we can
		// detect whether the motif was new (i.e. the track
		// child's dedup map saw the path for the first time).
		// Only motifs that are new should drive the status
		// progress bar; duplicates still apply their data to
		// the current lane but must NOT re-target the spring.
		prevFiles := m.track.Files()
		prevDirs := m.track.Dirs()

		// Forward to the track child. Track dedupes on path,
		// increments files/dirs (only on first sighting),
		// applies the motif data to the current lane and
		// rotates the lane index.
		m.track, _ = m.dispatchTrack(msg)

		// If neither files nor dirs changed, the path was a
		// duplicate - the track child applied data to the
		// current lane and rotated the index, but did not
		// increment any counter, so the status widget must
		// not see an IncDoneMsg.
		isNew := m.track.Files() > prevFiles || m.track.Dirs() > prevDirs
		if !isNew {
			return m, nil
		}

		// New motif: forward the increment to the status
		// widget so the percent / progress bar reflects the
		// new total. We read the post-update counts from the
		// track child (track.Files() / track.Dirs()).
		cmds := make([]tea.Cmd, 0, 1)
		var cmd tea.Cmd
		m.status, cmd = m.dispatchStatus(status.IncDoneMsg{N: 1})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.status, _ = m.dispatchStatus(status.CountsMsg{
			Files: m.track.Files(), Dirs: m.track.Dirs(), Errors: m.Errors,
		})
		return m, tea.Batch(cmds...)

	case contract.CompleteMsg:
		// Capture the first non-nil error message for the
		// "press space to exit" footer (still rendered by the
		// highway chrome; see renderSummary).
		m.ApplyCompletion(msg.Errs, msg.Elapsed)

		// Forward the flush signal to the track child so it
		// clears its counted map. The track child does not need
		// the file/dir counts (they are not displayed in any
		// lane); those go to status via CountsMsg below.
		m.track, _ = m.dispatchTrack(track.CompleteMsg{})

		// Forward the final counts and elapsed to the status
		// widget. The widget owns the percent calculation and
		// the isDone flag from here on. The status DoneMsg
		// handler only overwrites its errMsg when msg.Err is
		// non-empty, so passing the root's potentially-empty
		// errMsg is safe.
		cmds := make([]tea.Cmd, 0, 2)
		m.status, _ = m.dispatchStatus(status.CountsMsg{
			Files: msg.Files, Dirs: msg.Dirs, Errors: len(msg.Errs),
		})
		m.status, _ = m.dispatchStatus(status.ElapsedMsg{Elapsed: msg.Elapsed})
		var cmd tea.Cmd
		m.status, cmd = m.dispatchStatus(status.DoneMsg{
			Done: msg.Files, IsDone: true, Err: m.ErrMsg,
		})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case progress.FrameMsg:
		// Forward the bubbles progress spring's animation
		// frames to the status widget. Without this, the spring
		// cmd returned from status.Update loops back through the
		// bubbletea program but is dropped at the default arm,
		// so the bar never advances past its first frame.
		var cmd tea.Cmd
		m.status, cmd = m.dispatchStatus(msg)
		return m, cmd

	default:
		return m, nil
	}
}

// dispatchStatus forwards a status-owned message to the child
// status widget and returns the resulting (status.Model, tea.Cmd)
// tuple, type-asserting the bubbletea Model back to a concrete
// status.Model. The caller appends cmd to its running cmds slice
// when non-nil and assigns the returned model back to m.status.
func (m Model) dispatchStatus(msg tea.Msg) (status.Model, tea.Cmd) {
	r, cmd := m.status.Update(msg)
	return r.(status.Model), cmd //nolint:errcheck // known concrete type
}

// dispatchTrack forwards a track-owned message to the child track
// widget and returns the resulting (track.Model, tea.Cmd) tuple.
// The caller assigns the returned model back to m.track. Mirrors
// dispatchStatus.
func (m Model) dispatchTrack(msg tea.Msg) (track.Model, tea.Cmd) { //nolint:unparam // ok
	r, cmd := m.track.Update(msg)
	return r.(track.Model), cmd //nolint:errcheck // known concrete type
}
