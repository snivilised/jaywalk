package highway

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism"
	"github.com/snivilised/jaywalk/src/prism/widget"
)

func (m Model) renderLanes(b *strings.Builder) {
	laneBarWidth := LaneBarWidth
	barStyles := widget.SquareBarStyles{
		Filled: m.theme.BarFilledStyle,
		Empty:  m.theme.BarEmptyStyle,
	}

	borderStyle := m.theme.BorderStyle
	mutedStyle := m.theme.MutedStyle
	frameStyle := m.theme.FrameStyle

	dashes := strings.Repeat("─", maxInt(0, m.width-2))

	for i, lane := range m.lanes {
		emojiPart := lane.Emoji

		var fill int
		if m.maxDepth > 0 {
			fill = int(float64(lane.Depth) / float64(m.maxDepth) * float64(laneBarWidth))
			if fill > laneBarWidth {
				fill = laneBarWidth
			}
		} else {
			fill = (lane.tick * (7 + i*3)) % (laneBarWidth + 1)
		}
		laneBar := widget.RenderSquareBar(widget.SquareBarConfig{
			Width:  laneBarWidth,
			Fill:   fill,
			Styles: barStyles,
		})

		nodeIcon := ""
		if lane.Path != "" {
			nodeIcon = m.theme.TreeIcons[prism.TreeIconFile]
			if lane.IsDir {
				nodeIcon = m.theme.TreeIcons[prism.TreeIconDirectory]
			}
		}

		var actionInfo string
		if lane.Err != nil {
			actionInfo = m.theme.ErrorStyle.Render(" ! " + lane.Err.Error())
		} else if lane.ActionName != "" {
			actionInfo = m.theme.ActionStyle.Render(" • via " + lane.ActionName)
		} else if lane.PipelineName != "" {
			actionInfo = m.theme.PipelineStyle.Render(" • via " + lane.PipelineName)
		}

		var displayPath string
		if lane.Path != "" {
			displayPath = lane.Path
			if lane.IsDir {
				displayPath += "/"
			}
		} else {
			displayPath = lane.Label
		}

		frame := frameStyle.Render(lane.FrameFunc(lane.tick))

		landingStrip := m.renderExecutionInfo(lane)

		prefixWidth := lipgloss.Width(emojiPart) + 2 + laneBarWidth + 2
		if nodeIcon != "" {
			prefixWidth += lipgloss.Width(nodeIcon) + 1
		}
		if actionInfo != "" {
			prefixWidth += lipgloss.Width(actionInfo) + 2
		}
		suffixWidth := 2 + lipgloss.Width(frame)
		landingWidth := lipgloss.Width(landingStrip)
		if landingWidth > 0 {
			suffixWidth += 1 + landingWidth
		}

		availWidth := m.width - 4
		maxPathWidth := availWidth - prefixWidth - suffixWidth
		if maxPathWidth < 3 {
			maxPathWidth = 3
		}

		pathWidth := lipgloss.Width(displayPath)
		if pathWidth > maxPathWidth {
			keepWidth := maxPathWidth - 3
			runes := []rune(displayPath)
			width := 0
			start := len(runes)
			for i := len(runes) - 1; i >= 0; i-- {
				charWidth := lipgloss.Width(string(runes[i]))
				if width+charWidth > keepWidth {
					break
				}
				width += charWidth
				start = i
			}
			displayPath = "..." + string(runes[start:])
		}

		var styledPath string
		if lane.Path != "" {
			if lane.IsDir {
				styledPath = m.theme.DirStyle.Render(displayPath)
			} else {
				styledPath = m.theme.FileStyle.Render(displayPath)
			}
		} else {
			styledPath = mutedStyle.Render(lane.Label)
		}

		var leftBuf strings.Builder
		leftBuf.WriteString(emojiPart)
		leftBuf.WriteString("  ")
		leftBuf.WriteString(laneBar)
		leftBuf.WriteString("  ")
		if nodeIcon != "" {
			leftBuf.WriteString(nodeIcon)
			leftBuf.WriteString(" ")
		}
		leftBuf.WriteString(actionInfo)
		if actionInfo != "" {
			leftBuf.WriteString("  ")
		}
		leftBuf.WriteString(styledPath)
		leftBuf.WriteString("  ")
		leftBuf.WriteString(frame)
		leftContent := leftBuf.String()

		leftWidth := lipgloss.Width(leftContent)

		var content string
		if landingStrip != "" {
			padding := availWidth - leftWidth - landingWidth
			if padding < 1 {
				padding = 1
			}
			content = leftContent + strings.Repeat(" ", padding) + landingStrip
		} else {
			padding := availWidth - leftWidth
			if padding < 1 {
				padding = 1
			}
			content = leftContent + strings.Repeat(" ", padding)
		}

		b.WriteString(borderStyle.Render("│ "))
		b.WriteString(content)
		b.WriteString(borderStyle.Render(" │"))
		b.WriteString("\n")

		b.WriteString(borderStyle.Render("├" + dashes + "┤"))
		b.WriteString("\n")
	}
}

func (m Model) renderExecutionInfo(lane Lane) string {
	content := lane.CommandOutput
	if lane.DryRun {
		content = lane.ExecutionString
	}
	if content == "" {
		return ""
	}
	return m.theme.BranchStyle.Render(" [") +
		m.theme.LandingStripStyle.Render(content) +
		m.theme.BranchStyle.Render("]")
}
