package highway

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
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
		pathDisplay = contract.Ellipses + pathDisplay[lipgloss.Width(pathDisplay)-keep:]
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

		// NEW: cascade display (padlock or depth) for no-recurse/depth flags
		cascadeWidget := renderCascadeDisplay(m)
		if cascadeWidget != "" {
			infoPart = infoPart + " │" + hStyle.Render(cascadeWidget)
		}
	}

	middle := header + infoPart

	row := layout.NewRow(m.width-4).
		Caps(borderStyle.Render("│ "), borderStyle.Render(" │")).
		Content(middle)

	if m.pipelineName != "" {
		pipelineInd := m.theme.PipelineStyle.Render(
			fmt.Sprintf("─── [ • via pipeline '%s' ] ───", m.pipelineName),
		)
		row.RightContent(pipelineInd)
	}

	// NEW: filter info widget for glob/regex flags
	filterWidget := renderFilterInfo(m)
	if filterWidget != "" {
		row.RightContent(filterWidget)
	}

	row.RenderTo(b)
	b.WriteString("\n")

	b.WriteString(borderStyle.Render("├" + dashes + "┤"))
	b.WriteString("\n")
}

// renderCascadeDisplay renders either padlock emoji, depth value, or empty string
// for the cascade behaviour indicator (no-recurse vs depth flags are mutually exclusive)
func renderCascadeDisplay(m Model) string {
	if m.CascadeDisplay == "" {
		return ""
	}
	// NoRecuse takes precedence over depth display, but here we trust CascadeDisplay field
	return m.CascadeDisplay
}

// renderFilterInfo renders active filter flags as [flag:value,...]
// Only displays filters that have non-empty patterns, respecting glob/regex modes per flag type
func renderFilterInfo(m Model) string {
	if m.FilesGlob == "" && m.FilesRegex == "" &&
		m.DirsGlob == "" && m.DirsRegex == "" {
		return "" // no filters to display
	}

	var parts []string

	// Files filter - show whichever is active (glob takes precedence over regex if both set)
	if m.FilesGlob != "" {
		parts = append(parts, fmt.Sprintf("files-glob:%s", m.FilesGlob))
	} else if m.FilesRegex != "" {
		parts = append(parts, fmt.Sprintf("files-regex:%s", m.FilesRegex))
	}

	// Dirs filter - same precedence logic
	if m.DirsGlob != "" {
		parts = append(parts, fmt.Sprintf("dirs-glob:%s", m.DirsGlob))
	} else if m.DirsRegex != "" {
		parts = append(parts, fmt.Sprintf("dirs-regex:%s", m.DirsRegex))
	}

	if len(parts) == 0 {
		return ""
	}

	return " └─ [ " + strings.Join(parts, ", ") + " ]"
}
