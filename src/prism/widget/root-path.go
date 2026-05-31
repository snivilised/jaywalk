package widget

import (
	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// RootPathStyles defines the styles used to render the root path widget.
type RootPathStyles struct {
	// PathStyle is applied to the root path text.
	PathStyle lipgloss.Style
}

// RootPath renders the root path with truncation if needed.
// If the path is empty, it defaults to ".".
// If the path exceeds maxWidth, it is truncated with an ellipsis.
func RootPath(path string, maxWidth int, styles RootPathStyles) string {
	if path == "" {
		path = "."
	}

	pathWidth := lipgloss.Width(path)
	if pathWidth > maxWidth {
		keep := max(0, maxWidth-3)
		path = contract.Ellipses + path[pathWidth-keep:]
	}

	return styles.PathStyle.Render(path)
}
