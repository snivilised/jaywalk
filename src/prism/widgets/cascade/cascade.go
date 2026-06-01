package cascade

import (
	"charm.land/lipgloss/v2"
)

// Styles defines the styles used to render the cascade display widget.
type Styles struct {
	// ValueStyle is applied to the cascade display text. The cascade
	// widget renders a single value (lock emoji or depth value) with
	// no associated label, so the value style is the only style it
	// needs.
	ValueStyle lipgloss.Style
}

// Render renders the cascade display widget (lock emoji or depth value).
// If cascade is empty, it returns an empty string.
func Render(cascade string, styles Styles) string {
	if cascade == "" {
		return ""
	}
	return styles.ValueStyle.Render(cascade)
}
