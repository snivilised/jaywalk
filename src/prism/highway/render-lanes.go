package highway

import (
	"fmt"
	"strings"

	"github.com/snivilised/jaywalk/src/prism/layout"
	"github.com/snivilised/jaywalk/src/prism/widgets/action"
	"github.com/snivilised/jaywalk/src/prism/widgets/activity"
	"github.com/snivilised/jaywalk/src/prism/widgets/landing"
	"github.com/snivilised/jaywalk/src/prism/widgets/node"
	"github.com/snivilised/jaywalk/src/prism/widgets/periscope"
)

// renderLanes renders each highway lane with animation frames and event data.
// Rendering order per lane (left to right inside one frame):
// | 🤖  ◼◻◻◻◻◻◻◻◻◻  🍎  📁  • via boo  braillewave       <Path>  ⠁⠂⠄⡀        [👾 sleep 1.0s] |
//
//  1. Left border segment │
//  2. Worker emoji indicator (changes per job arrival)
//  3. Periscope bar (indentation depth or animated fill pattern)
//  4. Job emoji indicator (changes per job arrival)
//  5. Node icon (file/folder tree icon, if path is present)
//  6. Action info (error message / action name / pipeline name)
//  7. Spinner column (fixed width label for spinner type: "default", "bounce", etc.)
//  8. Styled path or label (files get DirStyle/FileStyle, empty lanes get MutedStyle)
//  9. Animation frame (frame content from FrameFunc in quotes)
//
// 10. Landing strip
//
// 11. Right border segment │
func (m Model) renderLanes(b *strings.Builder) {
	laneBarWidth := LaneBarWidth

	periscopeStyles := periscope.Styles{
		Filled: m.theme.BarFilledStyle,
		Empty:  m.theme.BarEmptyStyle,
	}
	actionStyles := action.Styles{
		ErrorStyle:    m.theme.ErrorStyle,
		ActionStyle:   m.theme.ActionStyle,
		PipelineStyle: m.theme.PipelineStyle,
	}
	nodePathStyles := node.Styles{
		DirStyle:   m.theme.DirStyle,
		FileStyle:  m.theme.FileStyle,
		MutedStyle: m.theme.MutedStyle,
		TreeIcons:  m.theme.TreeIcons,
	}
	activityStyles := activity.Styles{
		FrameStyle: m.theme.FrameStyle,
	}
	landingStripStyles := landing.Styles{
		BranchStyle:       m.theme.BranchStyle,
		LandingStripStyle: m.theme.LandingStripStyle,
	}

	borderStyle := m.theme.BorderStyle
	mutedStyle := m.theme.MutedStyle

	for i, lane := range m.lanes {
		emojiPart := lane.Emoji

		// Periscope bar (depth indicator)
		periscopeContent := m.renderPeriscopeBar(lane, i, laneBarWidth, periscopeStyles)

		// Action info (error / action name / pipeline name)
		actionContent := action.Render(action.Config{
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
		landingStripContent := landing.Render(landing.Config{
			CommandOutput:   lane.CommandOutput,
			ExecutionString: lane.ExecutionString,
			DryRun:          lane.DryRun,
		}, landingStripStyles)

		// Build the row layout declaratively
		row := layout.NewRow(m.width-4).
			Caps(borderStyle.Render("│ "), borderStyle.Render(" │"))
		row.
			Content(emojiPart).Gap(2).
			Content(periscopeContent).Gap(2)
		if lane.JobEmoji != "" {
			row.Content(lane.JobEmoji).Gap(2)
		}
		if actionContent != "" {
			row.Content(actionContent).Gap(2)
		}
		row.
			Fixed(SpinnerNameWidth, spinnerNameCol).
			Flex(true).Gap(2). // path: flex with gap(2) before
			Content(frame).Gap(1).
			RightContent(landingStripContent)

		// Render the path content - returns complete formatted string with icon
		pathContent := node.Render(node.Config{
			Path:     lane.Path,
			IsDir:    lane.IsDir,
			Label:    lane.Label,
			MaxWidth: row.FlexWidth(),
		}, nodePathStyles)

		row.SetFlexContent(pathContent)

		row.RenderTo(b)
		b.WriteString("\n")

		dashes := strings.Repeat("─", max(0, m.width-2))
		b.WriteString(borderStyle.Render("├" + dashes + "┤"))
		b.WriteString("\n")
	}
}

func (m Model) renderPeriscopeBar(lane Lane, idx int, width int, styles periscope.Styles) string {
	if m.noRecurse {
		return styles.Filled.Render("◼")
	}
	var fill int
	if m.maxDepth > 0 {
		fill = int(float64(lane.Depth) / float64(m.maxDepth) * float64(width))
		if fill > width {
			fill = width
		}
	} else {
		fill = (lane.tick * (7 + idx*3)) % (width + 1)
	}

	return periscope.Render(periscope.Config{
		Width:  width,
		Fill:   fill,
		Styles: styles,
	}, styles, periscope.Effect{
		Gradient: lane.PeriscopeGradient,
		State:    lane.PeriscopeGradientState,
	})
}

func (m Model) renderActivityFrame(lane Lane, styles activity.Styles) string {
	var frameContent string
	if lane.FrameFn != nil {
		frameContent = lane.FrameFn(lane.tick)
	}
	if frameContent == "" {
		return ""
	}

	return activity.Render(activity.Config{
		Content: frameContent,
	},
		styles,
		activity.Effect{
			Gradient: lane.HighlightGradient,
			State:    lane.GradientState,
		},
	)
}
