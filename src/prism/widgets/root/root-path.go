package root

import (
	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// Styles defines the styles used to render the root path widget.
type Styles struct {
	// PathStyle is applied to the root path text.
	PathStyle lipgloss.Style
}

// Render renders the root path with truncation if needed.
// If the path is empty, it defaults to ".".
// If the path exceeds maxWidth, it is truncated with an ellipsis.
func Render(path string, maxWidth int, styles Styles) string {
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
