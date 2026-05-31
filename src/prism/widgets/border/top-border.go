package border

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// TopStyles defines the styles used to render the top border widget.
type TopStyles struct {
	// BorderStyle is applied to the border elements.
	BorderStyle lipgloss.Style

	// PathStyle is applied to the root path text.
	PathStyle lipgloss.Style

	// CornerStyle is applied to the corner decorations.
	CornerStyle lipgloss.Style
}

// RenderTop renders the top border line with corner decorations and root path.
// The root path is centered between the left and right border segments.
// If the root path is too long, it is truncated with an ellipsis.
func RenderTop(rootPath string, width int, styles TopStyles) string {
	pathWidth := lipgloss.Width(rootPath)
	maxPathWidth := width - 13
	if pathWidth > maxPathWidth {
		keep := max(0, maxPathWidth-3)
		rootPath = contract.Ellipses + rootPath[pathWidth-keep:]
		pathWidth = maxPathWidth
	}

	avail := max(2, width-pathWidth-11)
	L := avail / 2
	R := avail - L

	return styles.CornerStyle.Render("╭"+strings.Repeat("─", L)+"[ ") +
		styles.PathStyle.Render(rootPath) +
		styles.CornerStyle.Render(" ]"+strings.Repeat("─", R)+".★..─╮") + "\n"
}
