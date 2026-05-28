package highway

import (
	"fmt"
	"strings"

	"github.com/snivilised/jaywalk/src/prism/layout"
	"github.com/snivilised/jaywalk/src/prism/widget"
)

// renderLanes renders each highway lane with animation frames and event data.
// Rendering order per lane (left to right inside one frame):
// | 🤖  ◼◻◻◻◻◻◻◻◻◻  🌀  • via boo  braillewave       <Path>  ⠁⠂⠄⡀                                    [👾 sleep 1.0s] |
//
//  1. Left border segment │
//  2. Emoji indicator (e.g., 🔄, ⏳)
//  3. Periscope bar (indentation depth or animated fill pattern)
//  4. Node icon (file/folder tree icon, if path is present)
//  5. Action info (error message / action name / pipeline name)
//  6. Spinner column (fixed width label for spinner type: "default", "bounce", etc.)
//  7. Styled path or label (files get DirStyle/FileStyle, empty lanes get MutedStyle)
//  8. Animation frame (frame content from FrameFunc in quotes)
//  9. Landing strip
//
// 10. Right border segment │
func (m Model) renderLanes(b *strings.Builder) {
	laneBarWidth := LaneBarWidth

	periscopeStyles := widget.PeriscopeStyles{
		Filled: m.theme.BarFilledStyle,
		Empty:  m.theme.BarEmptyStyle,
	}
	actionStyles := widget.ActionStyles{
		ErrorStyle:    m.theme.ErrorStyle,
		ActionStyle:   m.theme.ActionStyle,
		PipelineStyle: m.theme.PipelineStyle,
	}
	nodePathStyles := widget.NodePathStyles{
		DirStyle:   m.theme.DirStyle,
		FileStyle:  m.theme.FileStyle,
		MutedStyle: m.theme.MutedStyle,
		TreeIcons:  m.theme.TreeIcons,
	}
	activityStyles := widget.ActivityStyles{
		FrameStyle: m.theme.FrameStyle,
	}
	landingStripStyles := widget.LandingStripStyles{
		BranchStyle:       m.theme.BranchStyle,
		LandingStripStyle: m.theme.LandingStripStyle,
	}

	borderStyle := m.theme.BorderStyle
	mutedStyle := m.theme.MutedStyle

	for i, lane := range m.lanes {
		emojiPart := lane.Emoji

		// Periscope bar (depth indicator)
		laneBar := m.renderPeriscopeBar(lane, i, laneBarWidth, periscopeStyles)

		// Node icon (determined before layout calc for prefix width)
		nodeIcon := renderNodeIcon(lane, m.theme.TreeIcons)

		// Action info (error / action name / pipeline name)
		actionInfo := widget.RenderAction(widget.ActionConfig{
			Error:        lane.Err,
			ActionName:   lane.ActionName,
			PipelineName: lane.PipelineName,
		}, actionStyles)

		// Spinner name column (fixed width)
		spinnerNameCol := fmt.Sprintf("%-*s", SpinnerNameWidth, lane.SpinnerName)
		spinnerNameCol = mutedStyle.Render(spinnerNameCol)

		// Animation frame with optional gradient
		frame := m.renderActivityFrame(lane, activityStyles)

		// Landing strip (execution info)
		landingStrip := widget.RenderLandingStrip(widget.LandingStripConfig{
			CommandOutput:   lane.CommandOutput,
			ExecutionString: lane.ExecutionString,
			DryRun:          lane.DryRun,
		}, landingStripStyles)

		// Build the row layout declaratively
		row := layout.NewRow(m.width - 4).
			Caps(borderStyle.Render("│ "), borderStyle.Render(" │"))
		row.
			Content(emojiPart).Gap(2).
			Content(laneBar).Gap(2)
		if nodeIcon != "" {
			row.Content(nodeIcon).Gap(1)
		}
		if actionInfo != "" {
			row.Content(actionInfo).Gap(2)
		}
		row.
			Fixed(SpinnerNameWidth, spinnerNameCol).
			Flex(true).Gap(2).     // path: flex with gap(2) after
			Content(frame).Gap(1).
			RightContent(landingStrip)

		// Render the path content with the allocated flex width
		_, styledPath := widget.RenderNodePath(widget.NodePathConfig{
			Path:     lane.Path,
			IsDir:    lane.IsDir,
			Label:    lane.Label,
			MaxWidth: row.FlexWidth(),
		}, nodePathStyles)
		row.SetFlexContent(styledPath)

		row.RenderTo(b)
		b.WriteString("\n")

		dashes := strings.Repeat("─", max(0, m.width-2))
		b.WriteString(borderStyle.Render("├" + dashes + "┤"))
		b.WriteString("\n")
	}
}

func (m Model) renderPeriscopeBar(lane Lane, idx int, laneBarWidth int, styles widget.PeriscopeStyles) string {
	if m.noRecurse {
		return styles.Filled.Render("◼")
	}
	var fill int
	if m.maxDepth > 0 {
		fill = int(float64(lane.Depth) / float64(m.maxDepth) * float64(laneBarWidth))
		if fill > laneBarWidth {
			fill = laneBarWidth
		}
	} else {
		fill = (lane.tick * (7 + idx*3)) % (laneBarWidth + 1)
	}
	return widget.RenderPeriscope(widget.PeriscopeConfig{
		Width:  laneBarWidth,
		Fill:   fill,
		Styles: styles,
	})
}

func renderNodeIcon(lane Lane, treeIcons map[string]string) string {
	if lane.Path == "" {
		return ""
	}
	if lane.IsDir {
		return treeIcons["directory"]
	}
	return treeIcons["file"]
}

func (m Model) renderActivityFrame(lane Lane, styles widget.ActivityStyles) string {
	var frameContent string
	if lane.FrameFn != nil {
		frameContent = lane.FrameFn(lane.tick)
	}
	if frameContent == "" {
		return ""
	}

	if lane.HighlightGradient != nil && lane.GradientState != nil {
		gradientRuns := ApplyGradient(
			lane.HighlightGradient.Hi,
			lane.HighlightGradient.Lo,
			frameContent,
			lane.GradientState,
		)
		if gradientRuns != nil {
			styledFrame := ApplyGradientStyled(gradientRuns)
			return m.theme.FrameStyle.Render(styledFrame)
		}
	}
	return widget.RenderActivity(widget.ActivityConfig{Content: frameContent}, styles)
}
