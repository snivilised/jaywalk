package action

import (
	"charm.land/lipgloss/v2"
)

type Styles struct {
	ErrorStyle    lipgloss.Style
	ActionStyle   lipgloss.Style
	PipelineStyle lipgloss.Style
}

type Config struct {
	Error        error
	ActionName   string
	PipelineName string
}

func Render(cfg Config, styles Styles) string {
	if cfg.Error != nil {
		return styles.ErrorStyle.Render(" ! " + cfg.Error.Error())
	}
	if cfg.ActionName != "" {
		return styles.ActionStyle.Render(" • via " + cfg.ActionName)
	}
	if cfg.PipelineName != "" {
		return styles.PipelineStyle.Render(" • via " + cfg.PipelineName)
	}
	return ""
}
