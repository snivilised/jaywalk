package status

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/layout"
)

// segment represents a single segment in the status row. Kept
// unexported; only this file uses it.
type segment struct {
	content   string
	separator bool // true to render a separator after this segment
}

// View produces the status row as a tea.View (bubbletea v2). All
// inputs come from the Model's fields. The composition is the same
// as the previous stateless Render, except the progress bar segment
// is now driven by the embedded bubbles progress model and the
// percent/total state lives on the widget.
func (m Model) View() tea.View {
	borderStyle := m.styles.BorderStyle

	segments := make([]segment, 0, 9)
	var elapsedContent string

	// Files segment
	if m.fields.ShowFiles {
		icon := m.styles.TreeIcons[contract.TreeIconFile]
		label := m.styles.SummaryLabelStyle.Render(icon + " files:")
		value := m.styles.SummaryValueStyle.Render(fmt.Sprintf("%4d", m.files))
		segments = append(segments, segment{
			content:   " " + label + " " + value + " ",
			separator: true,
		})
	}

	// Dirs segment
	if m.fields.ShowDirs {
		icon := m.styles.TreeIcons[contract.TreeIconDirectory]
		label := m.styles.SummaryLabelStyle.Render(icon + " dirs:")
		value := m.styles.SummaryValueStyle.Render(fmt.Sprintf("%3d", m.dirs))
		segments = append(segments, segment{
			content:   " " + label + " " + value + " ",
			separator: true,
		})
	}

	// Errors segment
	if m.fields.ShowErrors {
		icon := m.styles.TreeIcons[contract.TreeIconError]
		label := m.styles.ErrorStyle.Render(icon + " errors:")
		value := m.styles.SummaryValueStyle.Render(fmt.Sprintf("%3d", m.errors))
		segments = append(segments, segment{
			content:   " " + label + " " + value + " ",
			separator: true,
		})
	}

	// Skipped segment
	if m.fields.ShowSkipped {
		icon := m.styles.TreeIcons[contract.TreeIconSkipped]
		label := m.styles.SummaryLabelStyle.Render(icon + " skipped:")
		value := m.styles.SummaryValueStyle.Render(fmt.Sprintf("%3d", m.skipped))
		segments = append(segments, segment{
			content:   " " + label + " " + value + " ",
			separator: true,
		})
	}

	// Progress segment - driven by the embedded bubbles progress.
	// The bar is shown whenever we have something to display:
	//   - real mode: a TotalMsg has arrived (m.hasTotal), so the
	//     bar fills with done/total as MotifMsg deliveries come in.
	//   - demo mode: PercentMsg has set m.percent > 0 directly.
	// On a fresh model with neither, the segment is omitted so
	// the row doesn't render an empty 10-cell track. View() is
	// used (rather than ViewAs) so the rendered bar reflects the
	// spring's animated percentShown rather than snapping to the
	// target.
	if m.fields.ShowProgress && (m.percent > 0 || (m.hasTotal && m.total > 0)) {
		progressView := m.inner.View()
		if progressView != "" {
			pctLabel := m.styles.ProgressStyle.Render(fmt.Sprintf("%3d%%", m.percent))
			segments = append(segments, segment{
				content:   " " + progressView + "  " + pctLabel + " ",
				separator: true,
			})
		}
	}

	// Complete/Failed message segment
	if m.fields.ShowComplete && m.isDone {
		var msg string
		if m.errMsg != "" {
			label := "❌ Failed: " + m.errMsg
			if m.errors > 1 {
				label += fmt.Sprintf(" (+%d more)", m.errors-1)
			}
			msg = " " + m.styles.ErrorStyle.Render(label) + " "
		} else {
			msg = " " + m.styles.ProgressStyle.Render("✔ complete") + " "
		}
		segments = append(segments, segment{
			content:   msg,
			separator: true,
		})
	}

	// Elapsed segment (always last, right-aligned)
	if m.fields.ShowElapsed {
		icon := m.styles.TreeIcons[contract.TreeIconElapsed]
		label := m.styles.SummaryLabelStyle.Render(icon + " elapsed:")
		elapsedStr := formatDuration(m.elapsed)
		value := m.styles.SummaryValueStyle.Render(elapsedStr)
		elapsedContent = " " + label + " " + value + " "
	}

	// Build the row using layout.NewRow
	row := layout.NewRow(m.width-4).
		Caps(borderStyle.Render("│ "), borderStyle.Render(" │"))

	for _, seg := range segments {
		row.Content(seg.content)
		if seg.separator {
			row.Content(borderStyle.Render("│"))
		}
	}

	if elapsedContent != "" {
		row.RightContent(elapsedContent)
	}

	return tea.NewView(row.Render())
}

// formatDuration formats a duration as a human-readable string.
// Mirrors the helper previously inlined in the stateless Render.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Milliseconds()))
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	mins := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%02ds", mins, secs)
}
