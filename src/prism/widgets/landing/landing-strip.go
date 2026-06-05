package landing

import (
	"strings"

	"charm.land/lipgloss/v2"
)

type Styles struct {
	BranchStyle       lipgloss.Style
	LandingStripStyle lipgloss.Style
}

type Config struct {
	CommandOutput   string
	ExecutionString string
	DryRun          bool
	SkippedIcon     string
	// Width, when > 0, right-justifies the landing strip within the
	// given visible width. Padding is rendered using blank cells with
	// no extra style so the strip aligns to the right edge of the
	// surrounding body.
	Width int
}

func Render(cfg Config, styles Styles) string {
	content := cfg.CommandOutput
	if cfg.DryRun {
		content = cfg.ExecutionString
	}
	if content == "" && cfg.SkippedIcon == "" {
		return ""
	}
	bracket := ""
	if cfg.SkippedIcon != "" {
		bracket = " " + cfg.SkippedIcon
	}
	strip := styles.BranchStyle.Render(" ["+bracket) +
		styles.LandingStripStyle.Render(content) +
		styles.BranchStyle.Render("]")

	if cfg.Width <= 0 {
		return strip
	}

	visible := lipgloss.Width(strip)
	if visible >= cfg.Width {
		return strip
	}
	return strings.Repeat(" ", cfg.Width-visible) + strip
}
