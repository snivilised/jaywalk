package porthole

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
	"github.com/snivilised/jaywalk/src/prism/views/linear"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
	"github.com/snivilised/jaywalk/src/prism/widgets/scrollbar"
	"github.com/snivilised/jaywalk/src/prism/widgets/status"
)

// WindowSizeCallback is called when the terminal is resized. The
// argument is the new usable content width inside borders and the
// scrollbar gutter. The presenter uses this to re-justify landing
// strips at the correct column after a resize.
type WindowSizeCallback func(bodyWidth uint)

func statusFieldSet() status.FieldSelectors {
	return status.FieldSelectors{
		ShowFiles:    true,
		ShowDirs:     true,
		ShowErrors:   true,
		ShowSkipped:  true,
		ShowProgress: true,
		ShowComplete: true,
		ShowElapsed:  true,
	}
}

// bufEntry pairs a rendered line with the raw parameters used to
// produce it. On terminal resize the model re-renders every buffered
// entry with the updated bodyWidth so landing strips stay correctly
// justified. branchStack is the branch state AFTER this line was
// rendered, used by the view to re-render the last line with the
// current activity frame.
type bufEntry struct {
	line        string
	params      RenderParams
	branchStack []bool
}

// Model is the bubbletea Model for the porthole view. It owns the
// chrome (banner, header, status, footer), the content buffer
// ([]bufEntry), and the viewport widget that renders the buffered
// lines. The flow renderer is used to convert Motif events into
// styled strings which are appended to the buffer.
type Model struct {
	contract.TraverseModel
	height int

	// contentBuf holds fully-rendered lines paired with the raw render
	// params. The buffer is truncated from the front when it exceeds
	// MaxContentBufferLines, dropping ContentBufferTruncateStep lines
	// at a time to keep navigation responsive.
	contentBuf []bufEntry

	status       status.Model
	bannerInfo   banner.Info
	bannerTicker *banner.Ticker

	// autoScroll tracks whether the viewport should automatically
	// follow new content. Set to true initially and when new content
	// arrives. Set to false when the user scrolls manually (arrow
	// keys, page-up/down). Re-enabled when the user scrolls back
	// to the very bottom.
	autoScroll bool

	// yOffset is the persisted viewport scroll offset. It is set
	// from the viewport after each render so that manual scroll
	// position survives across render cycles (the viewport is
	// recreated on every View call).
	yOffset int

	// onWindowSize is called when the terminal is resized so the
	// presenter can update its bodyWidth for landing strip
	// justification.
	onWindowSize WindowSizeCallback

	// Activity animation state. frameFn generates animation frames;
	// activityTick advances on each tickMsg; activityFrame holds the
	// current rendered frame string; gradientState drives the colour
	// sweep.
	frameFn          contract.FrameFunc
	activityTick     int
	activityFrame    string
	activityGradient *contract.ResolvedGradient
	gradientState    *effects.GradientState

	// countedFiles and countedDirs track how many unique nodes have
	// been visited via ContentLineMsg. Used to dispatch CountsMsg to
	// the status widget for the running file/dir tally and to drive
	// the IncDoneMsg that advances the progress bar.
	countedFiles int
	countedDirs  int
}

func NewModel(params contract.NewModelParams) Model {
	return Model{
		TraverseModel: contract.TraverseModel{
			Width:     80,
			RootPath:  params.RootPath,
			MaxDepth:  params.MaxDepth,
			NoRecurse: params.NoRecurse,
			TickRate:  TickRate,
			Theme:     params.Theme,
		},
		status: status.New(
			status.WithTheme(params.Theme),
			status.WithFields(statusFieldSet()),
			status.WithWidth(10),
		),
	}
}

// SetWindowSizeCallback registers a function that is called whenever
// the terminal is resized. The callback receives the new usable
// content width (window width minus borders and scrollbar gutter).
func (m *Model) SetWindowSizeCallback(fn WindowSizeCallback) {
	m.onWindowSize = fn
}

// SetActivity configures the animation displayed next to the latest
// content line. The frameFn generates per-tick frame strings; the
// gradient (if non-nil) paints them with an interpolated colour sweep.
func (m *Model) SetActivity(frameFn contract.FrameFunc, gradient *contract.ResolvedGradient) {
	m.frameFn = frameFn
	m.activityGradient = gradient
	if gradient != nil && gradient.Steps > 0 {
		hiR, hiG, hiB, _ := gradient.Hi.RGBA()
		loR, loG, loB, _ := gradient.Lo.RGBA()
		steps := effects.InterpolateBetweenRGBA(
			uint8(hiR>>8), uint8(hiG>>8), uint8(hiB>>8), //nolint:gosec // channel values are 0-65535, >>8 yields 0-255
			uint8(loR>>8), uint8(loG>>8), uint8(loB>>8), //nolint:gosec // channel values are 0-65535, >>8 yields 0-255
			gradient.Steps,
		)
		gs := effects.NewGradientState()
		gs.SetSteps(steps)
		m.gradientState = gs
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Tick(m.TickRate, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Buffer() []string {
	lines := make([]string, len(m.contentBuf))
	for i, e := range m.contentBuf {
		lines[i] = e.line
	}
	return lines
}

func (m Model) IsDone() bool {
	return m.Done
}

func (m Model) BannerInfo() banner.Info {
	return m.bannerInfo
}

type tickMsg time.Time

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.height = msg.Height
		m.rerender()

		if m.onWindowSize != nil {
			bw := m.contentWidth()
			m.onWindowSize(uint(bw)) //nolint:gosec // contentWidth() is always non-negative and fits in uint
		}

		var cmd tea.Cmd
		m.status, cmd = m.dispatchStatus(status.WidthMsg{Width: msg.Width})
		return &m, cmd

	case tea.KeyMsg:
		switch {
		case m.Done && msg.String() == "space":
			return &m, tea.Quit
		case msg.String() == "ctrl+c":
			return &m, tea.Quit
		case msg.String() == "up":
			m.autoScroll = false
			m.yOffset--
			if m.yOffset < 0 {
				m.yOffset = 0
			}
			return &m, nil
		case msg.String() == "down":
			if !m.autoScroll {
				m.yOffset++
				m.autoScroll = m.atBottom()
			}
			return &m, nil
		case msg.String() == "pgup":
			m.autoScroll = false
			m.yOffset -= m.viewportHeight()
			if m.yOffset < 0 {
				m.yOffset = 0
			}
			return &m, nil
		case msg.String() == "pgdown":
			if !m.autoScroll {
				m.yOffset += m.viewportHeight()
				m.autoScroll = m.atBottom()
			}
			return &m, nil
		case msg.String() == "home":
			m.autoScroll = false
			m.yOffset = 0
			return &m, nil
		case msg.String() == "end":
			m.autoScroll = true
			m.yOffset = 0
			return &m, nil
		}
		return &m, nil

	case tickMsg:
		if m.Done {
			return &m, nil
		}
		if m.Start.IsZero() {
			m.Start = core.Now()
		}

		// Advance the activity animation.
		if m.frameFn != nil {
			m.activityTick++
			m.activityFrame = m.frameFn(m.activityTick)
			if m.gradientState != nil {
				m.gradientState.Update(1)
			}
		}

		if m.bannerTicker != nil {
			m.bannerTicker.Advance()
		}

		tickCmd := tea.Tick(m.TickRate, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})

		var elapsedCmd tea.Cmd
		m.status, elapsedCmd = m.dispatchStatus(status.ElapsedMsg{
			Elapsed: time.Since(m.Start),
		})
		if elapsedCmd != nil {
			return &m, tea.Batch(tickCmd, elapsedCmd)
		}
		return &m, tickCmd

	case OvertureMsg:
		m.ApplyOverture(&msg.OvertureMsg)
		m.bannerInfo = msg.Banner
		if !msg.Banner.Disable && msg.Banner.State != nil && msg.Banner.Gradient != nil {
			m.bannerTicker = banner.NewTicker(msg.Banner.State, msg.Banner.Tick, m.TickRate)
		} else {
			m.bannerTicker = nil
		}

		return &m, nil

	case contract.CensusMsg:
		if msg.MaxDepth > m.MaxDepth {
			m.MaxDepth = msg.MaxDepth
		}

		// Seed the status widget's total with files + dirs. The
		// progress bar ratio is done / (files + dirs) because
		// every ContentLineMsg (file OR dir) increments done.
		total := int(msg.TotalFiles + msg.TotalDirs) //nolint:gosec // ok
		var cmd tea.Cmd
		m.status, cmd = m.dispatchStatus(status.TotalMsg{Total: total})
		if cmd != nil {
			return &m, tea.Batch(cmd)
		}
		return &m, nil

	case contract.CompleteMsg:
		m.ApplyCompletion(msg.Errs, msg.Elapsed)

		m.status, _ = m.dispatchStatus(status.CountsMsg{
			Files: msg.Files, Dirs: msg.Dirs, Errors: len(msg.Errs),
		})
		m.status, _ = m.dispatchStatus(status.ElapsedMsg{
			Elapsed: msg.Elapsed,
		})

		var cmd tea.Cmd
		m.status, cmd = m.dispatchStatus(status.DoneMsg{
			Done: msg.Files + msg.Dirs, IsDone: true, Err: m.ErrMsg,
		})
		if cmd != nil {
			return &m, tea.Batch(cmd)
		}

	case ContentLineMsg:
		line := strings.TrimRight(msg.Line, "\n")
		if line == "" {
			break
		}

		// Split multi-line entries (action output with embedded newlines)
		// into individual lines so each buffer entry is a single display
		// line. This matches the linear view which writes each line
		// directly to the terminal.
		maxW := m.contentWidth()
		for sub := range strings.SplitSeq(line, "\n") {
			sub = truncateStyled(sub, maxW)
			if sub != "" {
				m.contentBuf = append(m.contentBuf, bufEntry{
					line:        sub,
					params:      msg.Params,
					branchStack: msg.BranchStack,
				})
			}
		}

		if len(m.contentBuf) > MaxContentBufferLines {
			m.contentBuf = m.contentBuf[len(m.contentBuf)-ContentBufferTruncateStep:]
		}

		// Count this node for the progress bar. Each ContentLineMsg
		// corresponds to one node event (file or dir). We track
		// running tallies for the status CountsMsg and increment
		// the done counter by 1 for every node.
		if msg.Params.IsDir {
			m.countedDirs++
		} else {
			m.countedFiles++
		}

		cmds := make([]tea.Cmd, 0, 2)
		var incCmd tea.Cmd
		m.status, incCmd = m.dispatchStatus(status.IncDoneMsg{N: 1})
		if incCmd != nil {
			cmds = append(cmds, incCmd)
		}
		m.status, _ = m.dispatchStatus(status.CountsMsg{
			Files: m.countedFiles, Dirs: m.countedDirs, Errors: m.Errors,
		})
		if len(cmds) > 0 {
			return &m, tea.Batch(cmds...)
		}

	case progress.FrameMsg:
		// Forward the bubbles progress spring's animation
		// frames to the status widget. Without this, the spring
		// cmd returned from status.Update loops back through the
		// bubbletea program but is dropped at the default arm,
		// so the bar never advances past its first frame.
		var cmd tea.Cmd
		m.status, cmd = m.dispatchStatus(msg)
		return &m, cmd

	default:
		return &m, nil
	}
	return &m, nil
}

func (m Model) dispatchStatus(msg tea.Msg) (status.Model, tea.Cmd) {
	r, cmd := m.status.Update(msg)
	return r.(status.Model), cmd //nolint:errcheck // known concrete type
}

// lastLineBranchStack returns the branch state BEFORE the last line
// in the buffer, which is what linear.RenderLine needs to re-render
// that line. Returns nil when the buffer has fewer than 2 lines.
func (m Model) lastLineBranchStack() []bool {
	if len(m.contentBuf) < 2 {
		return nil
	}
	return m.contentBuf[len(m.contentBuf)-2].branchStack
}

// contentWidth returns the usable content width inside the viewport,
// accounting for left border (1), right border (1), and scrollbar gutter.
func (m Model) contentWidth() int {
	w := m.Width - scrollbar.ScrollbarGutterWidth - 2
	if w < 0 {
		return 0
	}
	return w
}

// viewportHeight returns the number of visible rows in the content
// body, used as the scroll step for Page-Up / Page-Down.
func (m Model) viewportHeight() int {
	h := m.height - 6 - m.legendSectionHeight()
	if h < 1 {
		return 1
	}
	return h
}

// atBottom reports whether the viewport's bottom edge is at or past
// the last line of content. Used to re-enable auto-scroll when the
// user scrolls down to the latest content.
func (m Model) atBottom() bool {
	vh := m.viewportHeight()
	if vh < 1 {
		return true
	}
	return m.yOffset+vh >= len(m.contentBuf)
}

// rerender re-renders every buffered line with the current bodyWidth
// so landing strips stay correctly justified after a terminal resize.
// It replays the branch stack progression sequentially because each
// line's rendering depends on the branch state left by the previous line.
func (m *Model) rerender() {
	if len(m.contentBuf) == 0 {
		return
	}

	bodyWidth := uint(m.contentWidth()) //nolint:gosec // contentWidth() is always non-negative and fits in uint
	branchStack := make([]bool, 0)
	newBuf := make([]bufEntry, 0, len(m.contentBuf))

	for _, entry := range m.contentBuf {
		p := entry.params
		result := linear.RenderLine(linear.LineParams{
			NodeParams: p.NodeParams,
			RenderParams: contract.RenderParams{
				BodyWidth: bodyWidth,
				Theme:     m.Theme,
			},
			BranchStack: branchStack,
		})
		branchStack = result.BranchStack

		line := strings.TrimRight(result.Line, "\n")
		if line == "" {
			continue
		}
		for sub := range strings.SplitSeq(line, "\n") {
			sub = truncateStyled(sub, int(bodyWidth)) //nolint:gosec // bodyWidth is always a reasonable terminal width
			if sub != "" {
				newBuf = append(newBuf, bufEntry{
					line:        sub,
					params:      entry.params,
					branchStack: result.BranchStack,
				})
			}
		}
	}
	m.contentBuf = newBuf
}

// truncateStyled truncates a lipgloss-styled string to maxVisible visible
// characters, correctly skipping ANSI escape sequences when measuring width.
func truncateStyled(s string, maxVisible int) string {
	if maxVisible <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= maxVisible {
		return s
	}

	visible := 0
	runes := []rune(s)
	var b strings.Builder

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Skip ANSI CSI sequences: ESC [ <params 0x30-0x3F> <final 0x40-0x7E>
		if r == 0x1b && i+1 < len(runes) && runes[i+1] == '[' {
			b.WriteRune(r)
			i++
			b.WriteRune(runes[i])
			i++
			for i < len(runes) && runes[i] >= 0x30 && runes[i] <= 0x3f {
				b.WriteRune(runes[i])
				i++
			}
			if i < len(runes) {
				b.WriteRune(runes[i])
			}
			continue
		}

		rw := lipgloss.Width(string(r))
		if visible+rw > maxVisible {
			break
		}
		b.WriteRune(r)
		visible += rw
	}

	return b.String()
}
