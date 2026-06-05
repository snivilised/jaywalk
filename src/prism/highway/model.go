package highway

import (
	"time"

	bp "charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
	"github.com/snivilised/jaywalk/src/prism/widgets/status"
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

type Model struct {
	lanes          []Lane
	skip           []int
	width          int
	start          time.Time
	tickRate       time.Duration
	totalTicks     int64
	rootPath       string
	realMode       bool
	done           bool
	noRecurse      bool
	files          int
	dirs           int
	errors         int
	elapsed        time.Duration
	currentLaneIdx int
	totalDirs      uint
	maxDepth       uint
	pipelineName   string
	// TODO(status-widget-state-migration): totalFiles is still
	// owned by the root in this PR. The CensusMsg path
	// (highway/model.go Update) forwards it into the status
	// widget via status.TotalMsg so the widget can compute
	// percent on its own. The root's local copy remains for
	// now because the lanes update path still references it
	// for fake-mode rendering. Move the root copy out once
	// lanes becomes a proper child model.
	totalFiles        uint
	subscriptionLabel string
	startedAt         time.Time
	caption           string
	dateFormat        string
	theme             contract.Theme
	counted           map[string]bool
	errMsg            string

	// status is the child status widget. It owns the
	// files/dirs/errors/elapsed/isDone/errMsg/percent/total state
	// and the embedded bubbles progress bar. The root
	// translates highway messages into status.* messages
	// (see update.go's translation helpers).
	status status.Model

	// header is the supplementary flag info carried on the OvertureMsg.
	// Stored on the model so renderFlagsRow and other renderers can
	// access it without further plumbing. See contract.HeaderInfo for
	// the field semantics.
	header contract.HeaderInfo

	// FlagsRowPosition controls where the flags row is rendered. See
	// contract/... or the theme config; "top" places it after the top
	// border, "bottom" places it above the status line. The default
	// applied by the loader is "bottom".
	FlagsRowPosition string

	// banner is the per-render state for the optional ANSI shadow
	// banner. Nil when the banner is disabled. The gradient state
	// advances on a slower tick than the lane animations.
	banner *bannerState

	// bannerInfo is the (immutable for the session) configuration
	// received via OvertureMsg. The view reads this each render.
	bannerInfo BannerInfo
}

// initLaneSkip computes the per-lane skip factor from each lane's
// IntervalMs. The skip factor = IntervalMs / tickRate (in ms). A lane
// with no override (IntervalMs=0) gets factor 0 - it advances every
// tick. A lane with IntervalMs=5000 at 50ms tick rate gets factor 100,
// advancing one frame every 100 ticks (every 5 seconds).
// See Lane.IntervalMs for the config override path.
func initLaneSkip(lanes []Lane, tickRate time.Duration) []int {
	factors := make([]int, len(lanes))
	tickMs := int(tickRate.Milliseconds())
	if tickMs == 0 {
		tickMs = 50
	}
	for i, lane := range lanes {
		if lane.IntervalMs > 0 {
			factors[i] = lane.IntervalMs / tickMs
			if factors[i] < 1 {
				factors[i] = 1
			}
		}
	}
	return factors
}

func NewModel(lanes []Lane, tickRate time.Duration, rootPath string,
	maxDepth uint, theme contract.Theme, noRecurse bool) Model {
	return Model{
		lanes:     lanes,
		skip:      initLaneSkip(lanes, tickRate),
		tickRate:  tickRate,
		noRecurse: noRecurse,
		width:     80,
		rootPath:  rootPath,
		maxDepth:  maxDepth,
		theme:     theme,
		counted:   make(map[string]bool),
		status: status.New(
			status.WithTheme(theme),
			status.WithFields(statusFieldSet()),
			status.WithWidth(10),
		),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Tick(m.tickRate, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		var cmd tea.Cmd
		m.status, cmd = m.dispatchStatus(status.WidthMsg{Width: msg.Width})
		return m, cmd

	case tea.KeyMsg:
		switch {
		case m.done && msg.String() == "space":
			return m, tea.Quit
		case msg.String() == "ctrl+c":
			return m, tea.Quit
		}
		return m, nil

	case tickMsg:
		cmds := make([]tea.Cmd, 0, 2)
		if m.done {
			return m, nil
		}
		if m.start.IsZero() && !m.realMode {
			m.start = core.Now()
		}
		m.totalTicks++
		// Advance each lane's frame counter independently.
		// Lanes with a skip factor > 0 (set via IntervalMs override)
		// only advance their tick every N global ticks, producing a
		// visibly slower animation. Lanes with skip factor 0 advance
		// every tick (full speed).
		for i := range m.lanes {
			if m.skip != nil && i < len(m.skip) && m.skip[i] > 0 {
				m.lanes[i].skipCounter++
				if m.lanes[i].skipCounter >= m.skip[i] {
					m.lanes[i].skipCounter = 0
					m.lanes[i].tick++
				}
			} else {
				m.lanes[i].tick++
			}

			// Advance gradient state for lanes with configured gradients.
			// IMPORTANT: This updates the state (offset/index) but does NOT apply
			// the gradient colours to frameContent - that happens in renderLanes().
			// The GradientState.Offset tracks current position in gradient array;
			// ApplyGradient() uses this offset to interpolate characters from Hi->Lo.
			if m.lanes[i].HighlightGradient != nil {
				windowSize := m.lanes[i].WindowSize()
				if windowSize <= 0 {
					windowSize = 4
				}
				m.lanes[i].GradientState.Update(windowSize)
			}

			if m.lanes[i].PeriscopeGradient != nil {
				windowSize := m.lanes[i].WindowSize()
				if windowSize <= 0 {
					windowSize = 4
				}
				m.lanes[i].PeriscopeGradientState.Update(windowSize)
			}
		}

		// Advance the banner's gradient state on its own slower tick
		// so its warm glow is visibly different from the lane
		// animations. skipFactor handles the speed difference.
		m.banner.advance()
		tickCmd := tea.Tick(m.tickRate, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
		cmds = append(cmds, tickCmd)
		if !m.start.IsZero() {
			// Elapsed time is real in both demo and real mode,
			// so push it on every tick. Without this the status
			// row's elapsed segment would stay at 0 in real
			// mode until the final CompleteMsg arrived.
			elapsed := time.Since(m.start)
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
				// TODO(realMode-cleanup): once lanes becomes a
				// proper child model and owns its own ticking,
				// move the demo-mode fake percent generation
				// into lanes.Update so the root doesn't need
				// to know about it.
				m.status, _ = m.dispatchStatus(status.PercentMsg{
					Percent: int(elapsed.Seconds()) * 2 % 100,
				})
			}
		}
		return m, tea.Batch(cmds...)

	case OvertureMsg:
		m.rootPath = msg.Root
		m.start = core.Now()
		m.realMode = true
		m.pipelineName = msg.PipelineName
		m.subscriptionLabel = msg.SubscriptionLabel
		m.startedAt = msg.StartedAt
		m.caption = msg.Caption
		m.dateFormat = msg.DateFormat

		// Cache the header info for renderers.
		m.header = msg.Header

		// store flags row position; default to "bottom" for empty/invalid
		m.FlagsRowPosition = msg.FlagsRowPosition
		if m.FlagsRowPosition != FlagsRowPositionTop && m.FlagsRowPosition != FlagsRowPositionBottom {
			m.FlagsRowPosition = FlagsRowPositionBottom
		}

		// Initialise the banner state from the OvertureMsg. The
		// bannerState is nil when the banner is disabled, when the
		// gradient binding is absent, or when the state pointer is
		// nil. The view checks for nil before rendering.
		m.bannerInfo = msg.Banner
		if !msg.Banner.Disable && msg.Banner.State != nil && msg.Banner.Gradient != nil {
			m.banner = newBannerState(msg.Banner.State, msg.Banner.Tick, m.tickRate)
		} else {
			m.banner = nil
		}

		return m, nil

	case CensusMsg:
		m.totalFiles = msg.TotalFiles
		m.totalDirs = msg.TotalDirs
		if msg.MaxDepth > m.maxDepth {
			m.maxDepth = msg.MaxDepth
		}
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

	case MotifMsg:
		isNew := !m.counted[msg.Data.Path]
		if isNew {
			m.counted[msg.Data.Path] = true
			cmds := make([]tea.Cmd, 0, 1)
			if msg.Data.IsDir {
				m.dirs++
			} else {
				m.files++
			}
			// Forward the new count to the status widget so the
			// percent / progress bar reflects the new total.
			// TODO(lanes-widget-state-migration): the counted
			// map, m.files and m.dirs will all move to the
			// lanes child widget in a future PR; this
			// translation will then happen inside lanes.Update
			// and the root's local copies can be removed.
			var cmd tea.Cmd
			m.status, cmd = m.dispatchStatus(status.IncDoneMsg{N: 1})
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			m.status, _ = m.dispatchStatus(status.CountsMsg{
				Files: m.files, Dirs: m.dirs, Errors: m.errors,
			})
			return m.applyMotifData(msg, cmds)
		}
		return m.applyMotifData(msg, nil)

	case CompleteMsg:
		// Capture the first non-nil error message for the
		// "press space to exit" footer (still rendered by the
		// highway chrome; see renderSummary).
		for _, e := range msg.Errs {
			if e != nil {
				m.errMsg = e.Error()
				break
			}
		}
		m.errors = len(msg.Errs)
		m.elapsed = msg.Elapsed
		m.done = true

		// Forward the final counts and elapsed to the status
		// widget. The widget owns the percent calculation and
		// the isDone flag from here on.
		cmds := make([]tea.Cmd, 0, 2)
		m.status, _ = m.dispatchStatus(status.CountsMsg{
			Files: msg.Files, Dirs: msg.Dirs, Errors: len(msg.Errs),
		})
		m.status, _ = m.dispatchStatus(status.ElapsedMsg{Elapsed: msg.Elapsed})
		var cmd tea.Cmd
		m.status, cmd = m.dispatchStatus(status.DoneMsg{
			Done: msg.Files, IsDone: true, Err: m.errMsg,
		})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case bp.FrameMsg:
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

// applyMotifData updates the per-lane fields from a MotifMsg. The
// caller is responsible for any status translation (IncDoneMsg,
// CountsMsg) before calling this. Extracted from the MotifMsg
// switch arm to keep the update flow readable.
func (m Model) applyMotifData(msg MotifMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	if len(m.lanes) > 0 {
		m.lanes[m.currentLaneIdx].JobEmoji = msg.Data.JobEmoji
		m.lanes[m.currentLaneIdx].Path = msg.Data.Path
		m.lanes[m.currentLaneIdx].Name = msg.Data.Name
		m.lanes[m.currentLaneIdx].IsDir = msg.Data.IsDir
		m.lanes[m.currentLaneIdx].Depth = msg.Data.Depth
		m.lanes[m.currentLaneIdx].ActionName = msg.Data.ActionName
		m.lanes[m.currentLaneIdx].PipelineName = msg.Data.PipelineName
		m.lanes[m.currentLaneIdx].CommandOutput = msg.Data.CommandOutput
		m.lanes[m.currentLaneIdx].ExecutionString = msg.Data.ExecutionString
		m.lanes[m.currentLaneIdx].DryRun = msg.Data.DryRun
		m.lanes[m.currentLaneIdx].Err = msg.Data.Err
		// Copy gradient from message to lane if provided.
		// The gradient is a ResolvedGradient {Steps, Hi, Lo} from the theme palette.
		// It holds colour endpoint info; we apply it in renderLanes() using ApplyGradient().
		if msg.Data.Gradient != nil {
			m.lanes[m.currentLaneIdx].HighlightGradient = msg.Data.Gradient
			// Also ensure GradientState exists and is configured with steps.
			if m.lanes[m.currentLaneIdx].GradientState == nil {
				m.lanes[m.currentLaneIdx].GradientState = effects.NewGradientState()
			}
			m.lanes[m.currentLaneIdx].GradientState.TotalSteps = msg.Data.Gradient.Steps
		}

		if msg.Data.PeriscopeGradient != nil {
			m.lanes[m.currentLaneIdx].PeriscopeGradient = msg.Data.PeriscopeGradient
			if m.lanes[m.currentLaneIdx].PeriscopeGradientState == nil {
				m.lanes[m.currentLaneIdx].PeriscopeGradientState = effects.NewGradientState()
			}
			m.lanes[m.currentLaneIdx].PeriscopeGradientState.TotalSteps = msg.Data.PeriscopeGradient.Steps
		}
		m.currentLaneIdx = (m.currentLaneIdx + 1) % len(m.lanes)
	}
	if len(cmds) == 0 {
		return m, nil
	}
	return m, tea.Batch(cmds...)
}
