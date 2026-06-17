package track

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/prism/layout"
	"github.com/snivilised/jaywalk/src/prism/widgets/action"
	"github.com/snivilised/jaywalk/src/prism/widgets/activity"
	"github.com/snivilised/jaywalk/src/prism/widgets/landing"
	"github.com/snivilised/jaywalk/src/prism/widgets/node"
	"github.com/snivilised/jaywalk/src/prism/widgets/periscope"
)

// View renders every lane in the track, one row per lane with a
// `├──┤` separator between rows. The full multi-lane string is
// returned as a single tea.View so the highway root can embed it
// in its bordered layout.
//
// Rendering order per lane (left to right inside one frame):
//
//		🤖  ◼◻◻◻◻◻◻◻◻◻  🍎  📁  • via boo  braille-wave       <Path>  ⠁⠂⠄⡀        [👾 sleep 1.0s]
//
//	 1. Left border segment │
//	 2. Worker emoji indicator (changes per job arrival)
//	 3. Periscope bar (indentation depth or animated fill pattern)
//	 4. Job emoji indicator (changes per job arrival)
//	 5. Node icon (file/folder tree icon, if path is present)
//	 6. Action info (error message / action name / pipeline name)
//	 7. Worker-ID column (fixed width label: "W#1", "W#2", etc. Shows pool-assigned
//	    goroutine ID; replaced by a proper user-facing ID once pants exposes one.)
//	 8. Styled path or label (files get DirStyle/FileStyle, empty lanes get MutedStyle)
//	 9. Animation frame (frame content from FrameFunc in quotes)
//
// 10. Landing strip
// 11. Right border segment │
func (m Model) View() tea.View {
	var b strings.Builder
	m.renderLanes(&b)
	return tea.NewView(b.String())
}

// renderLanes writes the lane rows to b. The body is the former
// highway.Model.renderLanes moved verbatim, with the only edit
// being the theme read replaced by m.styles.
func (m Model) renderLanes(b *strings.Builder) {
	laneBarWidth := LaneBarWidth

	periscopeStyles := periscope.Styles{
		Filled: m.styles.BarFilledStyle,
		Empty:  m.styles.BarEmptyStyle,
	}
	actionStyles := action.Styles{
		ErrorStyle:    m.styles.ErrorStyle,
		ActionStyle:   m.styles.ActionStyle,
		PipelineStyle: m.styles.PipelineStyle,
	}
	nodePathStyles := node.Styles{
		DirStyle:   m.styles.DirStyle,
		FileStyle:  m.styles.FileStyle,
		MutedStyle: m.styles.MutedStyle,
		TreeIcons:  m.styles.TreeIcons,
	}
	activityStyles := activity.Styles{
		FrameStyle: m.styles.FrameStyle,
	}
	landingStripStyles := landing.Styles{
		BranchStyle:       m.styles.BranchStyle,
		LandingStripStyle: m.styles.LandingStripStyle,
	}

	borderStyle := m.styles.BorderStyle
	mutedStyle := m.styles.MutedStyle

	for i, lane := range m.lanes[:m.visibleCount] {
		emojiPart := lane.Emoji

		periscopeContent := m.renderPeriscopeBar(lane, i, laneBarWidth, periscopeStyles)

		actionContent := action.Render(action.Config{
			Error:        lane.Err,
			ActionName:   lane.ActionName,
			PipelineName: lane.PipelineName,
		}, actionStyles)

		// Apply idle/working style based on worker state.
		if actionContent != "" {
			stateStyle := m.styles.WorkingStyle
			if lane.State == enums.WorkerStateIdle || lane.State == enums.WorkerStateUndefined {
				stateStyle = m.styles.IdleStyle
			}
			actionContent = stateStyle.Render(actionContent)
		}

		workerID := lane.WorkerID
		if len([]rune(workerID)) > SpinnerNameWidth {
			workerID = string([]rune(workerID)[:SpinnerNameWidth])
		}
		workerIDCol := fmt.Sprintf("%-*s", SpinnerNameWidth, workerID)
		workerIDCol = mutedStyle.Render(workerIDCol)

		frame := m.renderActivityFrame(lane, activityStyles)

		landingStripContent := landing.Render(landing.Config{
			CommandOutput:   lane.CommandOutput,
			ExecutionString: lane.ExecutionString,
			DryRun:          lane.DryRun,
		}, landingStripStyles)

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
			Fixed(SpinnerNameWidth, workerIDCol).
			Flex(true).Gap(2).
			Content(frame).Gap(1).
			RightContent(landingStripContent)

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

// renderPeriscopeBar returns the periscope bar content for the
// given lane. Moved verbatim from highway.Model.renderPeriscopeBar.
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

// renderActivityFrame returns the styled animation frame content
// for the given lane. Moved verbatim from
// highway.Model.renderActivityFrame.
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
	}, styles, activity.Effect{
		Gradient: lane.HighlightGradient,
		State:    lane.GradientState,
	})
}
