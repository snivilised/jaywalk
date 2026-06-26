package status

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

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
//
// Width management: the widget uses a compact label style (Width(0))
// to avoid the 16-cell padding from SummaryLabelStyle which would
// overflow on a standard 80-column terminal when all segments are
// active. The progress segment is also conditionally removed when
// the available row width is insufficient.
func (m Model) View() tea.View {
	borderStyle := m.styles.BorderStyle

	// Use per-segment fixed-width label styles so the files and dirs
	// labels always occupy the same number of cells regardless of
	// whether the total is shown in parentheses ("files:") or not
	// ("files(20):"). This prevents the layout from jumping ~5 cells
	// when CensusMsg arrives, which shifts the dirs value left/right
	// on every frame. Skipped and elapsed don't have a hasTotal
	// transition, so they use natural width (Width(0)) to avoid the
	// theme's 16-cell default padding.
	compactLabel := m.styles.SummaryLabelStyle.Width(0)
	filesLabel := m.styles.SummaryLabelStyle.Width(16)
	dirsLabel := m.styles.SummaryLabelStyle.Width(14)

	segments := make([]segment, 0, 9)
	progressIdx := -1
	var elapsedContent string

	// Files segment
	if m.fields.ShowFiles {
		icon := m.styles.TreeIcons[contract.TreeIconFile]
		label := filesLabel.Render(m.labelText(icon, " files", m.totalFiles, m.hasTotal))
		value := m.styles.SummaryValueStyle.Render(fmt.Sprintf("%4d", m.files))
		segments = append(segments, segment{
			content:   " " + label + " " + value + " ",
			separator: true,
		})
	}

	// Dirs segment
	if m.fields.ShowDirs {
		icon := m.styles.TreeIcons[contract.TreeIconDirectory]
		label := dirsLabel.Render(m.labelText(icon, " dirs", m.totalDirs, m.hasTotal))
		value := m.styles.SummaryValueStyle.Render(fmt.Sprintf("%4d", m.dirs))
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

	// Skipped segment (no hasTotal transition, uses natural width)
	if m.fields.ShowSkipped {
		icon := m.styles.TreeIcons[contract.TreeIconSkipped]
		label := compactLabel.Render(icon + " skipped:")
		value := m.styles.SummaryValueStyle.Render(fmt.Sprintf("%3d", m.skipped))
		segments = append(segments, segment{
			content:   " " + label + " " + value + " ",
			separator: true,
		})
	}

	// Progress segment
	if m.fields.ShowProgress && (m.percent > 0 || (m.hasTotal && m.total > 0)) {
		progressView := m.inner.View()
		if progressView != "" {
			pctLabel := m.styles.ProgressStyle.Render(fmt.Sprintf("%3d%%", m.percent))
			progressIdx = len(segments)
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
		} else if m.hasTotal && m.percent >= 100 {
			msg = " " + m.styles.ProgressStyle.Render("✔ complete") + " "
		}
		if msg != "" {
			segments = append(segments, segment{
				content:   msg,
				separator: true,
			})
		}
	}

	// Elapsed segment (always last, right-aligned, no hasTotal transition)
	if m.fields.ShowElapsed {
		icon := m.styles.TreeIcons[contract.TreeIconElapsed]
		label := compactLabel.Render(icon + " elapsed:")
		elapsedStr := formatDuration(m.elapsed)
		value := m.styles.SummaryValueStyle.Render(elapsedStr)
		elapsedContent = " " + label + " " + value + " "
	}

	// Estimate whether the row fits within the available width.
	// Drop the progress segment only when dropping it makes the
	// row fit — that is, when without progress the total width
	// is ≤ rowWidth but with progress it exceeds rowWidth.
	rowWidth := m.width - 4
	needWidth := 0
	for _, seg := range segments {
		needWidth += lipgloss.Width(seg.content)
		if seg.separator {
			needWidth++
		}
	}
	if elapsedContent != "" {
		needWidth += lipgloss.Width(elapsedContent)
	}

	if needWidth > rowWidth && progressIdx >= 0 {
		widthWithoutProgress := needWidth - lipgloss.Width(segments[progressIdx].content)
		if segments[progressIdx].separator {
			widthWithoutProgress--
		}
		if widthWithoutProgress <= rowWidth {
			segments = append(segments[:progressIdx], segments[progressIdx+1:]...)
		}
	}

	// Build the row using layout.NewRow
	row := layout.NewRow(rowWidth).
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
