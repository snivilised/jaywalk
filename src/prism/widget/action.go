package widget

import (
	"charm.land/lipgloss/v2"
)

type ActionStyles struct {
	ErrorStyle    lipgloss.Style
	ActionStyle   lipgloss.Style
	PipelineStyle lipgloss.Style
}

type ActionConfig struct {
	Error        error
	ActionName   string
	PipelineName string
}

func RenderAction(cfg ActionConfig, styles ActionStyles) string {
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
