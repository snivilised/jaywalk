package widget

import (
	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

type NodePathStyles struct {
	DirStyle  lipgloss.Style
	FileStyle lipgloss.Style
	MutedStyle lipgloss.Style
	TreeIcons contract.TreeIcons
}

type NodePathConfig struct {
	Path     string
	IsDir    bool
	Label    string
	MaxWidth int
}

func RenderNodePath(cfg NodePathConfig, styles NodePathStyles) (icon, styled string) {
	if cfg.Path != "" {
		icon = styles.TreeIcons[contract.TreeIconFile]
		if cfg.IsDir {
			icon = styles.TreeIcons[contract.TreeIconDirectory]
		}
	}

	var displayPath string
	if cfg.Path != "" {
		displayPath = cfg.Path
		if cfg.IsDir {
			displayPath += "/"
		}
	} else {
		displayPath = cfg.Label
	}

	pathWidth := lipgloss.Width(displayPath)
	if pathWidth > cfg.MaxWidth {
		keepWidth := cfg.MaxWidth - 3
		runes := []rune(displayPath)
		width := 0
		start := len(runes)
		for i := len(runes) - 1; i >= 0; i-- {
			charWidth := lipgloss.Width(string(runes[i]))
			if width+charWidth > keepWidth {
				break
			}
			width += charWidth
			start = i
		}
		displayPath = "..." + string(runes[start:])
	}

	if cfg.Path != "" {
		if cfg.IsDir {
			styled = styles.DirStyle.Render(displayPath)
		} else {
			styled = styles.FileStyle.Render(displayPath)
		}
	} else {
		styled = styles.MutedStyle.Render(displayPath)
	}

	return icon, styled
}
