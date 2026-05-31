package highway

import (
	"strings"

	"github.com/snivilised/jaywalk/src/prism/layout"
	"github.com/snivilised/jaywalk/src/prism/widget"
)

func (m Model) renderHeader(b *strings.Builder) {
	borderStyle := m.theme.BorderStyle
	headerStyle := m.theme.HeaderStyle
	summaryValueStyle := m.theme.SummaryValueStyle
	pipelineStyle := m.theme.PipelineStyle

	dashes := strings.Repeat("─", max(0, m.width-2))

	// Render Top Border
	topBorderWidget := widget.TopBorder(m.rootPath, m.width, widget.TopBorderStyles{
		BorderStyle: borderStyle,
		PathStyle:   m.theme.RootStyle,
		CornerStyle: borderStyle,
	})
	b.WriteString(topBorderWidget)

	// Render Date/Time and Cascade
	dateTimeWidget := widget.DateTime(m.subscriptionLabel, m.startedAt, m.dateFormat, widget.DateTimeStyles{
		InfoStyle: summaryValueStyle,
	})
	cascadeWidget := widget.Cascade(m.CascadeDisplay, widget.CascadeStyles{
		HeaderStyle: headerStyle,
	})
	var infoPart string
	if dateTimeWidget != "" {
		infoPart = dateTimeWidget
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
	pipelineWidget := widget.Pipeline(m.pipelineName, widget.PipelineStyles{
		PipelineStyle: pipelineStyle,
	})
	if pipelineWidget != "" {
		row.RightContent(pipelineWidget)
	}

	// Render Filter Info
	filterWidget := widget.Filter(m.FilesGlob, m.FilesRegex, m.DirsGlob, m.DirsRegex,
		m.FileTypeMode, m.DirTypeMode, widget.FilterStyles{
			InfoStyle: summaryValueStyle,
		})
	if filterWidget != "" {
		row.RightContent(filterWidget)
	}

	row.RenderTo(b)
	b.WriteString("\n")

	b.WriteString(borderStyle.Render("├" + dashes + "┤"))
	b.WriteString("\n")
}