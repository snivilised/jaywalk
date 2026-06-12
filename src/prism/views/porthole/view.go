package porthole

import (
	"strings"

	bp "charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/layout"
	"github.com/snivilised/jaywalk/src/prism/views/linear"
	"github.com/snivilised/jaywalk/src/prism/views/shared"
	"github.com/snivilised/jaywalk/src/prism/widgets/activity"
	"github.com/snivilised/jaywalk/src/prism/widgets/border"
	"github.com/snivilised/jaywalk/src/prism/widgets/intro"
	"github.com/snivilised/jaywalk/src/prism/widgets/legend"
	"github.com/snivilised/jaywalk/src/prism/widgets/pipeline"
	"github.com/snivilised/jaywalk/src/prism/widgets/scrollbar"
)

func (m *Model) View() tea.View {
	var b strings.Builder

	// Banner at top: rendered OUTSIDE the bordered region, above the
	// top border. This is independent of the flags row position
	// because the flags row lives INSIDE the border.
	if m.bannerInfo.Position == contract.PositionTop {
		shared.WriteBanner(&b, m.bannerInfo, m.Width)
	}

	m.renderHeader(&b)
	if m.FlagsRowPosition == contract.PositionTop && m.legendHeight() > 0 {
		m.WriteSeparator(&b)
		m.writeLegend(&b)
		m.WriteSeparator(&b)
	}

	b.WriteString(m.renderBody())

	if (m.FlagsRowPosition == contract.PositionBottom || m.FlagsRowPosition == "") &&
		m.legendHeight() > 0 {
		m.WriteSeparator(&b)
		m.writeLegend(&b)
		m.WriteSeparator(&b)
	}

	statusContent := m.status.View().Content
	b.WriteString(statusContent)
	b.WriteString("\n")

	b.WriteString(border.RenderBottom(m.Width, border.Styles{
		BorderStyle: m.Theme.BorderStyle,
	}))

	if m.status.IsDone() {
		b.WriteString("\n")
		b.WriteString(m.Theme.MutedStyle.Render(" • press space to exit"))
	}

	// Banner at bottom: rendered OUTSIDE the bordered region, below
	// the summary. The summary's last line is the bottom border
	// (no trailing newline), so we emit a separator newline to
	// keep the banner from overwriting the border.
	if m.bannerInfo.Position == contract.PositionBottom {
		b.WriteByte('\n')
		shared.WriteBanner(&b, m.bannerInfo, m.Width)
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func (m Model) renderHeader(b *strings.Builder) {
	borderStyle := m.Theme.BorderStyle
	headerStyle := m.Theme.HeaderStyle
	summaryValueStyle := m.Theme.SummaryValueStyle
	pipelineStyle := m.Theme.PipelineStyle

	dashes := strings.Repeat("─", max(0, m.Width-2))

	topBorderContent := border.RenderTop(m.RootPath, m.Width, border.Styles{
		BorderStyle: borderStyle,
		PathStyle:   m.Theme.RootStyle,
		CornerStyle: borderStyle,
	})
	b.WriteString(topBorderContent)

	introContent := intro.Render(m.SubscriptionLabel, m.StartedAt, m.DateFormat, intro.Styles{
		InfoStyle: summaryValueStyle,
	})

	header := headerStyle.Render("Processing")
	middle := header + introContent

	row := layout.NewRow(m.Width-4).
		Caps(borderStyle.Render("│ "), borderStyle.Render(" │")).
		Content(middle)

	pipelineContent := pipeline.Render(m.PipelineName, pipeline.Styles{
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
	bodyWidth := max(m.Width-scrollbar.ScrollbarGutterWidth-2, 0)

	// Account for all fixed chrome: top border (1), header row (1),
	// separator (1), status (1), bottom border (1), "press space
	// to exit" (1) = 6 lines, plus the legend section (when active):
	// entry lines + 2 separator borders framing it (one above, one
	// below). legendSectionHeight returns 0 when no flags are present
	// so the body claims the full available space in that case.
	bodyHeight := max(m.height-6-m.legendSectionHeight(), 1)

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
		gutterStr := scrollbar.View(scrollbarState, scrollbar.Config{Theme: m.Theme})
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
	// linear.RenderLine with the current activity frame. This places
	// the spinner after the action/pipeline text (matching the
	// highway's layout) rather than at the far right edge.
	var activityLine string
	if showActivity && lastContentIdx >= 0 && len(m.contentBuf) > 0 {
		last := m.contentBuf[len(m.contentBuf)-1]
		branchStack := m.lastLineBranchStack()
		spinner := " " + m.renderActivity()
		nodeParams := last.params.NodeParams
		nodeParams.ActivityFrame = spinner
		result := linear.RenderLine(linear.LineParams{
			NodeParams: nodeParams,
			RenderParams: contract.RenderParams{
				BodyWidth: uint(bodyWidth),
				Theme:     m.Theme,
			},
			BranchStack: branchStack,
		})
		activityLine = strings.TrimRight(result.Line, "\n")
		activityLine = truncateStyled(activityLine, bodyWidth)
	}

	borderStyle := m.Theme.BorderStyle
	leftBar := borderStyle.Render("│")
	rightBar := borderStyle.Render("│")

	var b strings.Builder
	for i := range bodyHeight {
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

// writeLegend renders the flags/legend row into b using the shared
// chrome utility. The row is a no-op when no flag is active.
func (m Model) writeLegend(b *strings.Builder) {
	shared.WriteLegend(b, m.legendParams())
}

// legendHeight returns the number of terminal rows the legend will
// occupy (entry lines only). Returns 0 when no flags are active so
// the viewport can claim the full available height.
func (m Model) legendHeight() int {
	return shared.LegendHeight(m.legendParams())
}

// legendSectionHeight returns the total number of rows the legend
// section will occupy in the view: entry lines plus the two separator
// borders that frame it (one above, one below). Returns 0 when no
// flags are active so callers can skip the section entirely.
func (m Model) legendSectionHeight() int {
	return shared.LegendSectionHeight(m.legendParams())
}

// legendParams returns the LegendParams for the current model state.
// Centralised here so all legend-related methods stay consistent.
func (m Model) legendParams() shared.LegendParams {
	return shared.LegendParams{
		Info: legend.Info{
			Position: m.FlagsRowPosition,
			Header:   m.Header,
		},
		Width: m.Width,
		Styles: legend.Styles{
			LabelStyle:  m.Theme.SummaryLabelStyle.Width(0),
			ValueStyle:  m.Theme.SummaryValueStyle,
			BorderStyle: m.Theme.BorderStyle,
		},
	}
}
