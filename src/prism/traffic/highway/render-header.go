package highway

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/layout"
)

func (m Model) renderHeader(b *strings.Builder) {
	borderStyle := m.theme.BorderStyle
	hStyle := m.theme.HeaderStyle

	dashes := strings.Repeat("─", max(0, m.width-2))
	pathDisplay := m.rootPath
	if pathDisplay == "" {
		pathDisplay = "."
	}
	pathWidth := lipgloss.Width(pathDisplay)

	maxPathWidth := m.width - 13
	if pathWidth > maxPathWidth {
		keep := max(0, maxPathWidth-3)
		pathDisplay = "..." + pathDisplay[lipgloss.Width(pathDisplay)-keep:]
		pathWidth = maxPathWidth
	}

	avail := max(2, m.width-pathWidth-11)
	L := avail / 2
	R := avail - L

	b.WriteString(borderStyle.Render("╭" + strings.Repeat("─", L) + "[ "))
	b.WriteString(m.theme.RootStyle.Render(pathDisplay))
	b.WriteString(borderStyle.Render(
		" ]" + strings.Repeat("─", R) + ".★..─╮",
	))
	b.WriteString("\n")

	// --- header line: "Processing   files and folders  -  date  [pipeline]" ---

	header := hStyle.Render("Processing")

	var infoPart string
	if m.subscriptionLabel != "" && !m.startedAt.IsZero() {
		dateFmt := m.dateFormat
		if dateFmt == "" {
			dateFmt = "Mon, 02 Jan 2006 15:04:05 MST"
		}
		infoStr := fmt.Sprintf("  %s  -  %s", m.subscriptionLabel,
			m.startedAt.Format(dateFmt))
		infoPart = m.theme.SummaryValueStyle.Render(infoStr)
	}

	middle := header + infoPart

	row := layout.NewRow(m.width - 4).
		Caps(borderStyle.Render("│ "), borderStyle.Render(" │")).
		Content(middle)

	if m.pipelineName != "" {
		pipelineInd := m.theme.PipelineStyle.Render(
			fmt.Sprintf("─── [ • via pipeline '%s' ] ───", m.pipelineName),
		)
		row.RightContent(pipelineInd)
	}

	row.RenderTo(b)
	b.WriteString("\n")

	b.WriteString(borderStyle.Render("├" + dashes + "┤"))
	b.WriteString("\n")
}
