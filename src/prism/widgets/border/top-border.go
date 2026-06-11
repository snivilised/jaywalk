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

// RenderTop renders the top border line with corner decorations and optional content.
// The content is centered between the left and right border segments.
// If content is empty, the border is rendered without content or brackets.
// If the content is too long, it is truncated with an ellipsis.
func RenderTop(content string, width int, styles Styles) string {
	contentWidth := lipgloss.Width(content)
	maxContentWidth := width - 13

	if content == "" {
		N := max(0, width-7)
		return styles.CornerStyle.Render(
			contract.Static.Borders.TopLeftCorner+strings.Repeat("─", N)+contract.Static.Borders.TopRight,
		) + "\n"
	}

	if contentWidth > maxContentWidth {
		keep := max(0, maxContentWidth-3)
		content = contract.Ellipses + content[contentWidth-keep:]
		contentWidth = maxContentWidth
	}

	avail := max(2, width-contentWidth-11)
	L := avail / 2
	R := avail - L

	return styles.CornerStyle.Render(contract.Static.Borders.TopLeftCorner+strings.Repeat("─", L)+"[ ") +
		styles.PathStyle.Render(content) +
		styles.CornerStyle.Render(" ]"+strings.Repeat("─", R)+contract.Static.Borders.TopRight) + "\n"
}

func RenderBottom(width int, styles Styles) string {
	N := max(0, width-7)
	return styles.BorderStyle.Render(
		contract.Static.Borders.BottomLeft + strings.Repeat("─", N) + contract.Static.Borders.BottomRightCorner,
	)
}
