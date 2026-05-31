package activity

import (
	"charm.land/lipgloss/v2"
)

type Styles struct {
	FrameStyle lipgloss.Style
}

type Config struct {
	Content string
}

func Render(cfg Config, styles Styles) string {
	if cfg.Content == "" {
		return ""
	}
	return styles.FrameStyle.Render(cfg.Content)
}
