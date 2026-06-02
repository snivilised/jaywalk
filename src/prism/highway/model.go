package highway

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/bubbles/progress"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
)

type tickMsg time.Time

type Model struct {
	lanes             []Lane
	skip              []int
	width             int
	start             time.Time
	tickRate          time.Duration
	totalTicks        int64
	rootPath          string
	progress          progress.Model
	percent           int
	realMode          bool
	done              bool
	noRecurse         bool
	files             int
	dirs              int
	errors            int
	elapsed           time.Duration
	currentLaneIdx    int
	totalFiles        uint
	totalDirs         uint
	maxDepth          uint
	pipelineName      string
	subscriptionLabel string
	startedAt         time.Time
	caption           string
	dateFormat        string
	theme             contract.Theme
	counted           map[string]bool
	errMsg            string

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
		progress: progress.New(
			progress.WithSolidFill("#B9FBC0"), //TODO: theme.progress
			progress.WithoutPercentage(),
			progress.WithWidth(10),
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
		return m, nil

	case tea.KeyMsg:
		switch {
		case m.done && msg.String() == "space":
			return m, tea.Quit
		case msg.String() == "ctrl+c":
			return m, tea.Quit
		}
		return m, nil

	case tickMsg:
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
		if !m.start.IsZero() {
			if !m.realMode {
				m.percent = int(time.Since(m.start).Seconds()) * 2 % 100
			}
		}
		return m, tickCmd

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
		return m, nil

	case MotifMsg:
		if !m.counted[msg.Data.Path] {
			m.counted[msg.Data.Path] = true
			if msg.Data.IsDir {
				m.dirs++
			} else {
				m.files++
			}
		}
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
		if m.totalFiles > 0 {
			m.percent = int(float64(m.files) / float64(m.totalFiles) * 100)
		}
		return m, nil

	case CompleteMsg:
		m.files = msg.Files
		m.dirs = msg.Dirs
		m.errors = len(msg.Errs)
		m.elapsed = msg.Elapsed
		m.done = true
		for _, e := range msg.Errs {
			if e != nil {
				m.errMsg = e.Error()
				break
			}
		}
		if m.totalFiles > 0 {
			m.percent = int(float64(m.files) / float64(m.totalFiles) * 100)
		}
		return m, nil

	default:
		return m, nil
	}
}
