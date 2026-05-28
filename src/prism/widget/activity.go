package widget

import (
	"charm.land/lipgloss/v2"
)

type ActivityStyles struct {
	FrameStyle lipgloss.Style
}

type ActivityConfig struct {
	Content string
}

func RenderActivity(cfg ActivityConfig, styles ActivityStyles) string {
	if cfg.Content == "" {
		return ""
	}
	return styles.FrameStyle.Render(cfg.Content)
}
