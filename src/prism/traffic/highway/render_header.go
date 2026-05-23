package highway

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

func (m Model) renderHeader(b *strings.Builder) {
	borderStyle := m.theme.BorderStyle
	hStyle := m.theme.HeaderStyle

	dashes := strings.Repeat("─", maxInt(0, m.width-2))
	pathDisplay := m.rootPath
	if pathDisplay == "" {
		pathDisplay = "."
	}
	pathWidth := lipgloss.Width(pathDisplay)

	maxPathWidth := m.width - 13
	if pathWidth > maxPathWidth {
		keep := maxInt(0, maxPathWidth-3)
		pathDisplay = "..." + pathDisplay[lipgloss.Width(pathDisplay)-keep:]
		pathWidth = maxPathWidth
	}

	avail := maxInt(2, m.width-pathWidth-11)
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

	var headerContent string
	if m.pipelineName != "" {
		middle := header + infoPart

		pipelineInd := m.theme.PipelineStyle.Render(
			fmt.Sprintf("─── [ • via pipeline '%s' ] ───", m.pipelineName),
		)
		availSpace := maxInt(1,
			m.width-4-lipgloss.Width(middle)-lipgloss.Width(pipelineInd),
		)
		headerContent = lipgloss.JoinHorizontal(lipgloss.Left,
			middle,
			strings.Repeat(" ", availSpace),
			pipelineInd,
		)
	} else {
		middle := header + infoPart
		headerContent = lipgloss.JoinHorizontal(lipgloss.Left,
			middle,
			strings.Repeat(" ", maxInt(1, m.width-4-lipgloss.Width(middle))),
		)
	}

	b.WriteString(borderStyle.Render("│ "))
	b.WriteString(headerContent)
	b.WriteString(borderStyle.Render(" │"))
	b.WriteString("\n")

	b.WriteString(borderStyle.Render("├" + dashes + "┤"))
	b.WriteString("\n")
}
