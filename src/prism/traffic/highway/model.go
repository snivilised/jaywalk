package highway

import (
	"time"

	"charm.land/bubbletea/v2"
	"github.com/charmbracelet/bubbles/progress"

	"github.com/snivilised/jaywalk/src/prism"
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
	theme             prism.Theme
	counted           map[string]bool
	errMsg            string
}

// initLaneSkip computes the per-lane skip factor from each lane's
// IntervalMs. The skip factor = IntervalMs / tickRate (in ms). A lane
// with no override (IntervalMs=0) gets factor 0 — it advances every
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
	maxDepth uint, theme prism.Theme, noRecurse bool) Model {
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
			progress.WithSolidFill("#B9FBC0"),
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
			m.start = time.Now()
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
	}
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
		m.start = time.Now()
		m.realMode = true
		m.pipelineName = msg.PipelineName
		m.subscriptionLabel = msg.SubscriptionLabel
		m.startedAt = msg.StartedAt
		m.caption = msg.Caption
		m.dateFormat = msg.DateFormat
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
					m.lanes[m.currentLaneIdx].GradientState = NewGradientState()
				}
				m.lanes[m.currentLaneIdx].GradientState.TotalSteps = msg.Data.Gradient.Steps
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
