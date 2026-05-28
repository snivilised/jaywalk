package widget

import (
	"charm.land/lipgloss/v2"
)

type LandingStripStyles struct {
	BranchStyle       lipgloss.Style
	LandingStripStyle lipgloss.Style
}

type LandingStripConfig struct {
	CommandOutput   string
	ExecutionString string
	DryRun          bool
	SkippedIcon     string
}

func RenderLandingStrip(cfg LandingStripConfig, styles LandingStripStyles) string {
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
