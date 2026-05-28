package widget

import (
	"strings"

	"charm.land/lipgloss/v2"
)

type PeriscopeStyles struct {
	Filled lipgloss.Style
	Empty  lipgloss.Style
}

type PeriscopeConfig struct {
	Width  int
	Fill   int
	Styles PeriscopeStyles
}

func RenderPeriscope(cfg PeriscopeConfig) string {
	fill := min(max(cfg.Fill, 0), cfg.Width)

	filled := cfg.Styles.Filled.Render(strings.Repeat("◼", fill))
	empty := cfg.Styles.Empty.Render(strings.Repeat("◻", cfg.Width-fill))
	return filled + empty
}
