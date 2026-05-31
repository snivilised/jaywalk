package widget

import (
	"charm.land/lipgloss/v2"
)

// CascadeStyles defines the styles used to render the cascade display widget.
type CascadeStyles struct {
	// HeaderStyle is applied to the cascade display text.
	HeaderStyle lipgloss.Style
}

// Cascade renders the cascade display widget (lock emoji or depth value).
// If cascade is empty, it returns an empty string.
func Cascade(cascade string, styles CascadeStyles) string {
	if cascade == "" {
		return ""
	}
	return styles.HeaderStyle.Render(cascade)
}
