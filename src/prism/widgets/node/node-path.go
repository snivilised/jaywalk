package node

import (
	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

type Styles struct {
	DirStyle   lipgloss.Style
	FileStyle  lipgloss.Style
	MutedStyle lipgloss.Style
	TreeIcons  contract.TreeIcons
}

type Config struct {
	Path     string
	IsDir    bool
	Label    string
	MaxWidth int
}

func Render(cfg Config, styles Styles) string {
	var displayPath string
	if cfg.Path != "" {
		displayPath = cfg.Path
		if cfg.IsDir {
			displayPath += "/"
		}
	} else if cfg.Label != "" {
		displayPath = cfg.Label
	}

	pathWidth := lipgloss.Width(displayPath)
	if pathWidth > 0 && pathWidth > cfg.MaxWidth {
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
		displayPath = contract.Ellipses + string(runes[start:])
	}

	var icon string
	if cfg.Path != "" {
		if cfg.IsDir {
			icon = styles.TreeIcons[contract.TreeIconDirectory]
		} else {
			icon = styles.TreeIcons[contract.TreeIconFile]
		}
	}

	// Combine plain strings first, then apply style once to the full result to preserve proper formatting
	if icon != "" {
		result := icon + " " + displayPath
		if cfg.IsDir {
			return styles.DirStyle.Render(result)
		}
		return styles.FileStyle.Render(result)
	}

	if cfg.Label != "" {
		return styles.MutedStyle.Render(displayPath)
	}
	return displayPath
}
