package border

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// Styles defines the styles used to render the top border widget.
type Styles struct {
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
func RenderTop(rootPath string, width int, styles Styles) string {
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

	return styles.CornerStyle.Render(contract.Static.Borders.TopLeftCorner+strings.Repeat("─", L)+"[ ") +
		styles.PathStyle.Render(rootPath) +
		styles.CornerStyle.Render(" ]"+strings.Repeat("─", R)+contract.Static.Borders.TopRight) + "\n"
}

func RenderBottom(width int, styles Styles) string {
	N := max(0, width-7)
	return styles.BorderStyle.Render(
		contract.Static.Borders.BottomLeft + strings.Repeat("─", N) + contract.Static.Borders.BottomRightCorner,
	)
}
