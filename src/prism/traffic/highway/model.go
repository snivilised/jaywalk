package highway

import (
	"time"

	"charm.land/bubbletea/v2"
	"github.com/charmbracelet/bubbles/progress"

	"github.com/snivilised/jaywalk/src/prism"
)

type tickMsg time.Time

type Model struct {
	lanes              []Lane
	width              int
	start              time.Time
	tickRate           time.Duration
	totalTicks         int64
	rootPath           string
	progress           progress.Model
	percent            int
	realMode           bool
	done               bool
	files              int
	dirs               int
	errors             int
	elapsed            time.Duration
	currentLaneIdx     int
	totalFiles         uint
	totalDirs          uint
	maxDepth           uint
	pipelineName       string
	subscriptionLabel  string
	startedAt          time.Time
	caption            string
	dateFormat         string
	theme              prism.Theme
	counted            map[string]bool
	errMsg             string
}

func NewModel(lanes []Lane, tickRate time.Duration, rootPath string,
	maxDepth uint, theme prism.Theme) Model {
	return Model{
		lanes:    lanes,
		tickRate: tickRate,
		width:    80,
		rootPath: rootPath,
		maxDepth: maxDepth,
		theme:    theme,
		counted:  make(map[string]bool),
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
		if m.done && msg.String() == "space" {
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
		for i := range m.lanes {
			m.lanes[i].tick++
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
