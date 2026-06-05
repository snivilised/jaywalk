package scroll

import (
	"strings"

	bp "charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/flow"
	"github.com/snivilised/jaywalk/src/prism/layout"
	"github.com/snivilised/jaywalk/src/prism/widgets/activity"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
	"github.com/snivilised/jaywalk/src/prism/widgets/border"
	"github.com/snivilised/jaywalk/src/prism/widgets/intro"
	"github.com/snivilised/jaywalk/src/prism/widgets/pipeline"
	"github.com/snivilised/jaywalk/src/prism/widgets/scrollbar"
)

func (m *Model) View() tea.View {
	var b strings.Builder

	if m.bannerInfo.Position == contract.PositionTop {
		m.writeBanner(&b)
	}

	m.renderHeader(&b)
	if m.flagsRowPosition == contract.PositionTop && m.legendHeight() > 0 {
		m.writeSeparator(&b)
		m.writeLegend(&b)
		m.writeSeparator(&b)
	}

	b.WriteString(m.renderBody())

	if (m.flagsRowPosition == contract.PositionBottom || m.flagsRowPosition == "") &&
		m.legendHeight() > 0 {
		m.writeSeparator(&b)
		m.writeLegend(&b)
		m.writeSeparator(&b)
	}

	statusContent := m.status.View().Content
	b.WriteString(statusContent)
	b.WriteString("\n")

	b.WriteString(border.RenderBottom(m.width, border.Styles{
		BorderStyle: m.theme.BorderStyle,
	}))

	if m.status.IsDone() {
		b.WriteString("\n")
		b.WriteString(m.theme.MutedStyle.Render(" • press space to exit"))
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

// writeSeparator emits a horizontal "├─────┤" border line, used to
// frame the legend section on both sides. The legend widget itself
// is layout-agnostic; the surrounding borders are the view's concern.
func (m Model) writeSeparator(b *strings.Builder) {
	dashes := strings.Repeat("─", max(0, m.width-2))
	b.WriteString(m.theme.BorderStyle.Render("├" + dashes + "┤"))
	b.WriteString("\n")
}

func (m Model) writeBanner(b *strings.Builder) {
	bm := banner.NewModel(
		banner.WithInfo(m.bannerInfo),
		banner.WithWidth(m.width),
	)
	if bm.Disabled() {
		return
	}
	if out := bm.View(); out != "" {
		b.WriteString(out)
	}
}

func (m Model) renderHeader(b *strings.Builder) {
	borderStyle := m.theme.BorderStyle
	headerStyle := m.theme.HeaderStyle
	summaryValueStyle := m.theme.SummaryValueStyle
	pipelineStyle := m.theme.PipelineStyle

	dashes := strings.Repeat("─", max(0, m.width-2))

	topBorderContent := border.RenderTop(m.rootPath, m.width, border.Styles{
		BorderStyle: borderStyle,
		PathStyle:   m.theme.RootStyle,
		CornerStyle: borderStyle,
	})
	b.WriteString(topBorderContent)

	introContent := intro.Render(m.subscriptionLabel, m.startedAt, m.dateFormat, intro.Styles{
		InfoStyle: summaryValueStyle,
	})

	header := headerStyle.Render("Processing")
	middle := header + introContent

	row := layout.NewRow(m.width-4).
		Caps(borderStyle.Render("│ "), borderStyle.Render(" │")).
		Content(middle)

	pipelineContent := pipeline.Render(m.pipeline, pipeline.Styles{
		PipelineStyle: pipelineStyle,
	})
	if pipelineContent != "" {
		row.RightContent(pipelineContent)
	}

	row.RenderTo(b)
	b.WriteString("\n")

	b.WriteString(borderStyle.Render("├" + dashes + "┤"))
	b.WriteString("\n")
}

func (m *Model) renderBody() string {
	// bodyWidth is the usable content width inside the left/right borders
	// and the scrollbar gutter column.
	bodyWidth := m.width - scrollbar.ScrollbarGutterWidth - 2
	if bodyWidth < 0 {
		bodyWidth = 0
	}

	// Account for all fixed chrome: top border (1), header row (1),
	// separator (1), status (1), bottom border (1), "press space
	// to exit" (1) = 6 lines, plus the legend section (when active):
	// entry lines + 2 separator borders framing it (one above, one
	// below). legendSectionHeight returns 0 when no flags are present
	// so the body claims the full available space in that case.
	bodyHeight := m.height - 6 - m.legendSectionHeight()
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	truncated := make([]string, len(m.contentBuf))
	for i, entry := range m.contentBuf {
		truncated[i] = truncateStyled(entry.line, bodyWidth)
	}
	content := strings.Join(truncated, "\n")
	viewport := bp.New(
		bp.WithWidth(bodyWidth),
		bp.WithHeight(bodyHeight),
	)
	viewport.SetContent(content)
	if m.autoScroll {
		viewport.GotoBottom()
	} else {
		viewport.SetYOffset(m.yOffset)
	}

	scrollbarState := scrollbar.State{
		Height:       viewport.Height(),
		ContentLines: viewport.TotalLineCount(),
		Offset:       viewport.YOffset(),
	}

	// Collect gutter lines (one per row).
	var gutterLines []string
	if scrollbar.Visible(scrollbarState) {
		gutterStr := scrollbar.View(scrollbarState, scrollbar.Config{Theme: m.theme})
		gutterLines = strings.Split(strings.TrimRight(gutterStr, "\n"), "\n")
	}

	// Collect viewport lines.
	vpLines := strings.Split(viewport.View(), "\n")
	if len(vpLines) > 0 && vpLines[len(vpLines)-1] == "" {
		vpLines = vpLines[:len(vpLines)-1]
	}

	// Persist the viewport's scroll offset so the next render
	// starts from the same position.
	m.yOffset = viewport.YOffset()

	// Determine whether the activity spinner should be shown.
	// It only appears next to the last content line when the
	// viewport is scrolled to the bottom.
	atBottom := m.autoScroll || m.yOffset+bodyHeight >= len(m.contentBuf)
	showActivity := atBottom && m.frameFn != nil && m.activityFrame != ""

	// Find the index of the last content-bearing line in the
	// viewport. When the viewport has more rows than content,
	// trailing rows are empty padding; the spinner must attach
	// to the actual last node line, not an empty row.
	lastContentIdx := -1
	if showActivity {
		for i := len(vpLines) - 1; i >= 0; i-- {
			if strings.TrimSpace(vpLines[i]) != "" {
				lastContentIdx = i
				break
			}
		}
	}

	// When showing activity, re-render the last content line via
	// flow.RenderLine with the current activity frame. This places
	// the spinner after the action/pipeline text (matching the
	// highway's layout) rather than at the far right edge.
	var activityLine string
	if showActivity && lastContentIdx >= 0 && len(m.contentBuf) > 0 {
		last := m.contentBuf[len(m.contentBuf)-1]
		branchStack := m.lastLineBranchStack()
		spinner := " " + m.renderActivity()
		result := flow.RenderLine(
			last.params.Path,
			last.params.Name,
			last.params.IsDir,
			last.params.Depth,
			last.params.ActionName,
			last.params.PipelineName,
			last.params.CommandOutput,
			last.params.ExecutionString,
			last.params.DryRun,
			last.params.Err,
			last.params.IsLast,
			last.params.IsPipelineStep,
			last.params.IsLastStep,
			last.params.VisualDepth,
			branchStack,
			uint(bodyWidth),
			m.theme,
			spinner,
		)
		activityLine = strings.TrimRight(result.Line, "\n")
		activityLine = truncateStyled(activityLine, bodyWidth)
	}

	borderStyle := m.theme.BorderStyle
	leftBar := borderStyle.Render("│")
	rightBar := borderStyle.Render("│")

	var b strings.Builder
	for i := 0; i < bodyHeight; i++ {
		b.WriteString(leftBar)

		// Content line, padded to bodyWidth.
		if i < len(vpLines) {
			vpLine := vpLines[i]
			if i == lastContentIdx && activityLine != "" {
				vpLine = activityLine
			}
			if pad := bodyWidth - lipgloss.Width(vpLine); pad > 0 {
				vpLine += strings.Repeat(" ", pad)
			}
			b.WriteString(vpLine)
		} else {
			b.WriteString(strings.Repeat(" ", bodyWidth))
		}

		// Scrollbar gutter column.
		if i < len(gutterLines) {
			b.WriteString(gutterLines[i])
		} else {
			b.WriteString(strings.Repeat(" ", scrollbar.ScrollbarGutterWidth))
		}

		b.WriteString(rightBar)
		b.WriteByte('\n')
	}

	return b.String()
}

// renderActivity renders the current animation frame with optional
// gradient colouring. Returns "" when no animation is configured.
func (m Model) renderActivity() string {
	if m.frameFn == nil || m.activityFrame == "" {
		return ""
	}
	return activity.Render(activity.Config{
		Content: m.activityFrame,
	}, activity.Styles{}, activity.Effect{
		Gradient: m.activityGradient,
		State:    m.gradientState,
	})
}
