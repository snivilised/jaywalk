package highway

import (
	"strings"

	"github.com/snivilised/jaywalk/src/prism/layout"
	"github.com/snivilised/jaywalk/src/prism/widgets/border"
	"github.com/snivilised/jaywalk/src/prism/widgets/cascade"
	"github.com/snivilised/jaywalk/src/prism/widgets/filter"
	"github.com/snivilised/jaywalk/src/prism/widgets/intro"
	"github.com/snivilised/jaywalk/src/prism/widgets/pipeline"
)

func (m Model) renderHeader(b *strings.Builder) {
	borderStyle := m.theme.BorderStyle
	headerStyle := m.theme.HeaderStyle
	summaryValueStyle := m.theme.SummaryValueStyle
	pipelineStyle := m.theme.PipelineStyle

	dashes := strings.Repeat("─", max(0, m.width-2))

	// Render Top Border
	topBorderContent := border.RenderTop(m.rootPath, m.width, border.TopStyles{
		BorderStyle: borderStyle,
		PathStyle:   m.theme.RootStyle,
		CornerStyle: borderStyle,
	})
	b.WriteString(topBorderContent)

	// Render Date/Time and Cascade
	introContent := intro.Render(m.subscriptionLabel, m.startedAt, m.dateFormat, intro.Styles{
		InfoStyle: summaryValueStyle,
	})
	cascadeWidget := cascade.Render(m.CascadeDisplay, cascade.Styles{
		HeaderStyle: headerStyle,
	})

	var infoPart string
	if introContent != "" {
		infoPart = introContent
		if cascadeWidget != "" {
			infoPart = infoPart + " │" + cascadeWidget
		}
	} else if cascadeWidget != "" {
		infoPart = cascadeWidget
	}

	// Render header text
	header := headerStyle.Render("Processing")
	middle := header + infoPart

	row := layout.NewRow(m.width-4).
		Caps(borderStyle.Render("│ "), borderStyle.Render(" │")).
		Content(middle)

	// Render Pipeline Info
	pipelineContent := pipeline.Render(m.pipelineName, pipeline.Styles{
		PipelineStyle: pipelineStyle,
	})
	if pipelineContent != "" {
		row.RightContent(pipelineContent)
	}

	// Render Filter Info
	filterContent := filter.Render(m.FilesGlob, m.FilesRegex, m.DirsGlob, m.DirsRegex,
		m.FileTypeMode, m.DirTypeMode, filter.Styles{
			InfoStyle: summaryValueStyle,
		})
	if filterContent != "" {
		row.RightContent(filterContent)
	}

	row.RenderTo(b)
	b.WriteString("\n")

	b.WriteString(borderStyle.Render("├" + dashes + "┤"))
	b.WriteString("\n")
}
