package highway

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism"
	"github.com/snivilised/jaywalk/src/prism/widget"
)

// renderLanes renders each highway lane with animation frames and event data.
// Rendering order per lane (left to right inside one frame):
// | 🤖  ◼◻◻◻◻◻◻◻◻◻  🌀  • via boo  braillewave       <Path>  ⠁⠂⠄⡀                                    [👾 sleep 1.0s] |
//
//  1. Left border segment │
//  2. Emoji indicator (e.g., 🔄, ⏳)
//  3. Lane bar (indentation depth or animated fill pattern)
//  4. Node icon (file/folder tree icon, if path is present)
//  5. Activity info (error message / action name / pipeline name)
//  6. Spinner column (fixed width label for spinner type: "default", "bounce", etc.)
//  7. Styled path or label (files get DirStyle/FileStyle, empty lanes get MutedStyle)
//  8. Animation frame (frame content from FrameFunc in quotes)
//  9. Landing strip
//
// 10. Right border segment │
//
// Gradient rendering: The HighlightGradient field holds the gradient config; however,
// the current implementation does NOT apply the gradient to the animation frames.
// This is a known limitation - the gradient state advances each tick but never applies
// colour interpolation to the rendered frame content. To fix: call ApplyGradient() on
// frameContent when HighlightGradient != nil and insert styled result into leftBuf.
func (m Model) renderLanes(b *strings.Builder) {
	// + Left border
	laneBarWidth := LaneBarWidth
	barStyles := widget.SquareBarStyles{
		Filled: m.theme.BarFilledStyle,
		Empty:  m.theme.BarEmptyStyle,
	}

	borderStyle := m.theme.BorderStyle
	mutedStyle := m.theme.MutedStyle
	frameStyle := m.theme.FrameStyle

	dashes := strings.Repeat("─", max(0, m.width-2))

	// + Lanes
	for i, lane := range m.lanes {
		emojiPart := lane.Emoji

		// Lane bar (indentation indicator)
		var laneBar string
		if m.noRecurse {
			laneBar = barStyles.Filled.Render("◼") // ◼ (■)
		} else {
			var fill int
			if m.maxDepth > 0 {
				fill = int(float64(lane.Depth) / float64(m.maxDepth) * float64(laneBarWidth))
				if fill > laneBarWidth {
					fill = laneBarWidth
				}
			} else {
				fill = (lane.tick * (7 + i*3)) % (laneBarWidth + 1)
			}
			laneBar = widget.RenderSquareBar(widget.SquareBarConfig{
				Width:  laneBarWidth,
				Fill:   fill,
				Styles: barStyles,
			})
		}

		// + Node icon
		nodeIcon := ""
		if lane.Path != "" {
			nodeIcon = m.theme.TreeIcons[prism.TreeIconFile]
			if lane.IsDir {
				nodeIcon = m.theme.TreeIcons[prism.TreeIconDirectory]
			}
		}

		// + Action
		var actionInfo string
		if lane.Err != nil {
			actionInfo = m.theme.ErrorStyle.Render(" ! " + lane.Err.Error())
		} else if lane.ActionName != "" {
			actionInfo = m.theme.ActionStyle.Render(" • via " + lane.ActionName)
		} else if lane.PipelineName != "" {
			actionInfo = m.theme.PipelineStyle.Render(" • via " + lane.PipelineName)
		}

		// + Node path
		var displayPath string
		if lane.Path != "" {
			displayPath = lane.Path
			if lane.IsDir {
				displayPath += "/"
			}
		} else {
			displayPath = lane.Label
		}

		// Get frame content from FrameFunc (this is what gets gradient applied)
		var frameContent string
		if lane.FrameFn != nil {
			frameContent = lane.FrameFn(lane.tick)
		}

		// Apply gradient if configured.
		// The HighlightGradient field is set when MotifMsg arrives with Gradient != nil.
		// We interpolate the frame content using gradient Hi/Lo endpoints and
		// advance the state per tick so characters cycle through the gradient colours.
		var frame string
		if len(frameContent) > 0 {
			// Apply gradient if configured
			if lane.HighlightGradient != nil && lane.GradientState != nil {
				gradientRuns := ApplyGradient(
					lane.HighlightGradient.Hi,
					lane.HighlightGradient.Lo,
					frameContent,
					lane.GradientState,
				)
				if gradientRuns != nil {
					styledFrame := ApplyGradientStyled(gradientRuns)
					frame = m.theme.FrameStyle.Render(styledFrame)
				} else {
					// Fallback to plain rendering if gradient application failed
					frame = frameStyle.Render(frameContent)
				}
			} else {
				// No gradient configured, use plain frame style
				frame = frameStyle.Render(frameContent)
			}
		}

		landingStrip := m.renderExecutionInfo(lane)

		prefixWidth := lipgloss.Width(emojiPart) + 2 + laneBarWidth + 2
		if nodeIcon != "" {
			prefixWidth += lipgloss.Width(nodeIcon) + 1
		}
		if actionInfo != "" {
			prefixWidth += lipgloss.Width(actionInfo) + 2
		}
		prefixWidth += SpinnerNameWidth
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

		spinnerNameCol := fmt.Sprintf("%-*s", SpinnerNameWidth, lane.SpinnerName)
		spinnerNameCol = mutedStyle.Render(spinnerNameCol)

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
		leftBuf.WriteString(spinnerNameCol)
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

		// Right border
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
