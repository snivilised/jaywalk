package widget

import (
	"charm.land/lipgloss/v2"
)

// PipelineStyles defines the styles used to render the pipeline info widget.
type PipelineStyles struct {
	// PipelineStyle is applied to the pipeline information text.
	PipelineStyle lipgloss.Style
}

// Pipeline renders the pipeline information widget.
// If pipelineName is empty, it returns an empty string.
// The pipeline name is displayed with decorative dashes and a bullet point.
func Pipeline(pipelineName string, styles PipelineStyles) string {
	if pipelineName == "" {
		return ""
	}
	return styles.PipelineStyle.Render("─── [ • via pipeline '" + pipelineName + "' ] ───")
}
