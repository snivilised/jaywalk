package highway

import (
	"fmt"
	"strings"
	"time"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/layout"
	"github.com/snivilised/jaywalk/src/prism/widgets/clock"
)

func (m Model) renderSummary(b *strings.Builder) {
	borderStyle := m.theme.BorderStyle

	elapsedSecs := 0
	var elapsed time.Duration
	if m.done {
		elapsed = m.elapsed
		elapsedSecs = int(m.elapsed.Seconds())
	} else if !m.start.IsZero() {
		elapsed = time.Since(m.start)
		elapsedSecs = int(elapsed.Seconds())
	}

	var files, dirs, errors int
	if m.realMode {
		files = m.files
		dirs = m.dirs
		errors = m.errors
	} else {
		files = elapsedSecs*23 + 5
		dirs = elapsedSecs*7 + 2
		errors = elapsedSecs / 15
	}

	pct := m.percent
	barView := m.progress.ViewAs(float64(m.percent) / 100.0)

	fileIcon := m.theme.TreeIcons[contract.TreeIconFile]
	fileLabel := m.theme.SummaryLabelStyle.Render(fileIcon + " files:")
	fileValue := m.theme.SummaryValueStyle.Render(fmt.Sprintf("%4d", files))
	seg1 := " " + fileLabel + " " + fileValue + " "

	dirIcon := m.theme.TreeIcons[contract.TreeIconDirectory]
	dirLabel := m.theme.SummaryLabelStyle.Render(dirIcon + " dirs:")
	dirValue := m.theme.SummaryValueStyle.Render(fmt.Sprintf("%3d", dirs))
	seg2 := " " + dirLabel + " " + dirValue + " "

	errIcon := m.theme.TreeIcons[contract.TreeIconError]
	errLabel := m.theme.ErrorStyle.Render(errIcon + " errors:")
	errValue := m.theme.SummaryValueStyle.Render(fmt.Sprintf("%3d", errors))
	seg3 := " " + errLabel + " " + errValue + " "

	pctLabel := m.theme.ProgressStyle.Render(fmt.Sprintf("%3d%%", pct))
	seg4 := " " + barView + "  " + pctLabel + " "

	seg5 := ""
	if m.done {
		if m.errMsg != "" {
			label := "❌ Failed: " + m.errMsg
			if m.errors > 1 {
				label += fmt.Sprintf(" (+%d more)", m.errors-1)
			}
			seg5 = " " + m.theme.ErrorStyle.Render(label) + " "
		} else {
			seg5 = " " + m.theme.ProgressStyle.Render("✔ complete") + " "
		}
	}

	elapsedIcon := m.theme.TreeIcons[contract.TreeIconElapsed]
	elapsedLabel := m.theme.SummaryLabelStyle.Render(elapsedIcon + " elapsed:")
	elapsedStr := clock.FormatDuration(elapsed)
	elapsedValue := m.theme.SummaryValueStyle.Render(elapsedStr)
	elapsedText := " " + elapsedLabel + " " + elapsedValue + " "

	row := layout.NewRow(m.width-4).
		Caps(borderStyle.Render("│ "), borderStyle.Render(" │")).
		Content(seg1).
		Content(borderStyle.Render("│")).
		Content(seg2).
		Content(borderStyle.Render("│")).
		Content(seg3).
		Content(borderStyle.Render("│")).
		Content(seg4).
		Content(borderStyle.Render("│")).
		Content(seg5).
		RightContent(elapsedText)

	row.RenderTo(b)
	b.WriteString("\n")

	N := max(0, m.width-7)
	b.WriteString(borderStyle.Render(
		"╰─..★." + strings.Repeat("─", N) + "╯",
	))
	if m.done {
		b.WriteString("\n")
		b.WriteString(m.theme.MutedStyle.Render(" • press space to exit"))
	}
}
