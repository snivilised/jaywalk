package highway

import (
	"strings"

	"github.com/snivilised/jaywalk/src/prism/layout"
	"github.com/snivilised/jaywalk/src/prism/widgets/border"
	"github.com/snivilised/jaywalk/src/prism/widgets/intro"
	"github.com/snivilised/jaywalk/src/prism/widgets/pipeline"
)

func (m Model) renderHeader(b *strings.Builder) {
	borderStyle := m.Theme.BorderStyle
	headerStyle := m.Theme.HeaderStyle
	summaryValueStyle := m.Theme.SummaryValueStyle
	pipelineStyle := m.Theme.PipelineStyle

	dashes := strings.Repeat("─", max(0, m.Width-2))

	// Render Top Border
	topBorderContent := border.RenderTop(m.RootPath, m.Width, border.Styles{
		BorderStyle: borderStyle,
		PathStyle:   m.Theme.RootStyle,
		CornerStyle: borderStyle,
	})
	b.WriteString(topBorderContent)

	// Render the intro line. The intro contains primary flags only
	// (subscription, date/time) - the cascade (lock/depth) and filter
	// information have been moved to the flags row renderer to keep all
	// supplementary flag presentation in one place.
	introContent := intro.Render(m.SubscriptionLabel, m.StartedAt, m.DateFormat, intro.Styles{
		InfoStyle: summaryValueStyle,
	})

	// Render header text
	header := headerStyle.Render("Processing")
	middle := header + introContent

	row := layout.NewRow(m.Width-4).
		Caps(borderStyle.Render("│ "), borderStyle.Render(" │")).
		Content(middle)

	// Render Pipeline Info
	pipelineContent := pipeline.Render(m.PipelineName, pipeline.Styles{
		PipelineStyle: pipelineStyle,
	})
	if pipelineContent != "" {
		row.RightContent(pipelineContent)
	}

	row.RenderTo(b)
	b.WriteString("\n")

	b.WriteString(borderStyle.Render("├" + dashes + "┤"))
	b.WriteString("\n")
}
