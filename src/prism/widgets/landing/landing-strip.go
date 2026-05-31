package landing

import (
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
	return styles.BranchStyle.Render(" ["+bracket) +
		styles.LandingStripStyle.Render(content) +
		styles.BranchStyle.Render("]")
}
