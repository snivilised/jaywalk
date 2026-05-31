package cascade

import (
	"charm.land/lipgloss/v2"
)

// Styles defines the styles used to render the cascade display widget.
type Styles struct {
	// HeaderStyle is applied to the cascade display text.
	HeaderStyle lipgloss.Style
}

// Render renders the cascade display widget (lock emoji or depth value).
// If cascade is empty, it returns an empty string.
func Render(cascade string, styles Styles) string {
	if cascade == "" {
		return ""
	}
	return styles.HeaderStyle.Render(cascade)
}
