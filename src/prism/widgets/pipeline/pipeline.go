package pipeline

import (
	"charm.land/lipgloss/v2"
)

// Styles defines the styles used to render the pipeline info widget.
type Styles struct {
	// PipelineStyle is applied to the pipeline information text.
	PipelineStyle lipgloss.Style
}

// Render renders the pipeline information widget.
// If pipelineName is empty, it returns an empty string.
// The pipeline name is displayed with decorative dashes and a bullet point.
func Render(pipelineName string, styles Styles) string {
	if pipelineName == "" {
		return ""
	}
	return styles.PipelineStyle.Render("─── [ • via pipeline '" + pipelineName + "' ] ───")
}
