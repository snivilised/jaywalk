package periscope

import (
	"strings"

	"charm.land/lipgloss/v2"
)

type Styles struct {
	Filled lipgloss.Style
	Empty  lipgloss.Style
}

type Config struct {
	Width  int
	Fill   int
	Styles Styles
}

func Render(cfg Config) string {
	fill := min(max(cfg.Fill, 0), cfg.Width)

	filled := cfg.Styles.Filled.Render(strings.Repeat("◼", fill))
	empty := cfg.Styles.Empty.Render(strings.Repeat("◻", cfg.Width-fill))
	return filled + empty
}
