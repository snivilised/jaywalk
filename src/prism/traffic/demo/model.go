package demo

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/bubbles/progress"

	"github.com/snivilised/jaywalk/src/prism/widget"
)

type tickMsg time.Time

const joinTicks = 5

type OnboardMsg struct {
	Lane Lane
}

// TODO(PortholeView): use bubbles/viewport for the scrollable content area
// with a static header and footer:
//
//	header (static content)
//	viewport.Model (scrollable content)
//	footer (static — keybinding help via bubbles/help)
type Model struct {
	lanes      []Lane
	width      int
	start      time.Time
	tickRate   time.Duration
	totalTicks int64
	rootPath   string
	progress   progress.Model
	percent    int
}

func NewModel(lanes []Lane, tickRate time.Duration, rootPath string) Model {
	for i := range lanes {
		lanes[i].tick = joinTicks
	}
	return Model{
		lanes:    lanes,
		tickRate: tickRate,
		width:    80,
		rootPath: rootPath,
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
		return m, tea.Quit

	case tickMsg:
		if m.start.IsZero() {
			m.start = time.Now()
		}
		m.totalTicks++
		for i := range m.lanes {
			m.lanes[i].tick++
		}
		tickCmd := tea.Tick(m.tickRate, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
		m.percent = int(time.Since(m.start).Seconds()) * 2 % 100
		return m, tickCmd

	case OnboardMsg:
		m.lanes = append(m.lanes, msg.Lane)
		return m, nil

	default:
		return m, nil
	}
}

func (m Model) View() tea.View {
	var b strings.Builder

	hStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
	lStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	fStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("bright-black"))
	landingStripStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FBF8CC"))
	slStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FDE4CF"))
	svStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4D6D"))
	progStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#B9FBC0"))

	dashes := strings.Repeat("─", maxInt(0, m.width-2))
	pathDisplay := m.rootPath
	if pathDisplay == "" {
		pathDisplay = "."
	}
	pathWidth := lipgloss.Width(pathDisplay)

	// ╭───[ path ]───.★..─╮
	// Max path width: subtract fixed chars (╭=1, [sp=2, sp]=2, .★..─=5, ╮=1 = 11)
	// plus minimum 1 dash on each side (2 total)
	maxPathWidth := m.width - 13
	if pathWidth > maxPathWidth {
		keep := maxInt(0, maxPathWidth-3)
		pathDisplay = "..." + pathDisplay[lipgloss.Width(pathDisplay)-keep:]
		pathWidth = maxPathWidth
	}

	avail := maxInt(2, m.width-pathWidth-11)
	L := avail / 2
	R := avail - L

	b.WriteString(borderStyle.Render("╭" + strings.Repeat("─", L) + "[ "))
	b.WriteString(landingStripStyle.Render(pathDisplay))
	b.WriteString(borderStyle.Render(" ]" + strings.Repeat("─", R) + ".★..─╮"))
	b.WriteString("\n")

	// │  Header Example  │
	header := hStyle.Render("Header Example")

	b.WriteString(borderStyle.Render("│ "))
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left,
		header,
		strings.Repeat(" ", maxInt(1, m.width-4-lipgloss.Width(header))),
	))
	b.WriteString(borderStyle.Render(" │"))
	b.WriteString("\n")

	// ├─ separator ──┤
	b.WriteString(borderStyle.Render("├" + dashes + "┤"))
	b.WriteString("\n")

	const laneBarWidth = 10
	barStyles := widget.SquareBarStyles{
		Filled: lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")),
		Empty:  lipgloss.NewStyle().Foreground(lipgloss.Color("#585B70")),
	}

	// lanes
	for i, lane := range m.lanes {
		emoji := lane.Emoji
		label := lStyle.Render(lane.Label)
		frame := fStyle.Render(lane.FrameFunc(lane.tick))

		if lane.tick < joinTicks {
			joiningFrames := []string{"○", "◔", "◐", "◕", "●"}
			idx := min(lane.tick, joinTicks-1)
			emoji = joiningFrames[idx]
			label = lStyle.Render("Joining...")
			frame = ""
		}

		fill := (lane.tick * (7 + i*3)) % (laneBarWidth + 1)
		laneBar := widget.RenderSquareBar(widget.SquareBarConfig{
			Width:  laneBarWidth,
			Fill:   fill,
			Styles: barStyles,
		})

		content := fmt.Sprintf(
			"%s  %s  %s  %s",
			emoji,
			laneBar,
			label,
			frame,
		)
		padding := maxInt(1, m.width-4-lipgloss.Width(content))

		b.WriteString(borderStyle.Render("│ "))
		b.WriteString(content)
		b.WriteString(borderStyle.Render(strings.Repeat(" ", padding) + " │"))
		b.WriteString("\n")

		// horizontal lane separator
		b.WriteString(borderStyle.Render("├" + dashes + "┤"))
		b.WriteString("\n")
	}

	// │  files:  97 │ dirs:  30 │ errors:   0 │ ██████░░░░  42%  │      elapsed: 4s  │
	elapsedSecs := 0
	if !m.start.IsZero() {
		elapsedSecs = int(time.Since(m.start).Seconds())
	}
	files := elapsedSecs*23 + 5
	dirs := elapsedSecs*7 + 2
	errors := elapsedSecs / 15

	pct := m.percent
	barView := m.progress.ViewAs(float64(m.percent) / 100.0)

	seg1 := " " + slStyle.Render("files:") + " " + svStyle.Render(fmt.Sprintf("%4d", files)) + " "
	seg2 := " " + slStyle.Render("dirs:") + " " + svStyle.Render(fmt.Sprintf("%3d", dirs)) + " "
	seg3 := " " + errStyle.Render("errors:") + " " + svStyle.Render(fmt.Sprintf("%3d", errors)) + " "
	seg4 := " " + barView + "  " + progStyle.Render(fmt.Sprintf("%3d%%", pct)) + " "
	elapsedText := " " + slStyle.Render("elapsed:") + " " + svStyle.Render(fmt.Sprintf("%ds", elapsedSecs)) + " "

	inner := m.width - 4
	leftContent := lipgloss.JoinHorizontal(lipgloss.Left,
		seg1, borderStyle.Render("│"),
		seg2, borderStyle.Render("│"),
		seg3, borderStyle.Render("│"),
		seg4, borderStyle.Render("│"),
	)
	body := lipgloss.JoinHorizontal(lipgloss.Left,
		leftContent,
		strings.Repeat(" ", maxInt(1, inner-lipgloss.Width(leftContent)-lipgloss.Width(elapsedText))),
		elapsedText,
	)

	b.WriteString(borderStyle.Render("│ "))
	b.WriteString(body)
	b.WriteString(borderStyle.Render(" │"))
	b.WriteString("\n")

	// ╰─..★.──────────╯
	N := maxInt(0, m.width-7)
	b.WriteString(borderStyle.Render("╰─..★." + strings.Repeat("─", N) + "╯"))

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}
