package intro

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
)

// Styles defines the styles used to render the date/time widget.
type Styles struct {
	// InfoStyle is applied to the date/time information text.
	InfoStyle lipgloss.Style
}

// Render renders the date/time information widget.
// If subscriptionLabel is empty or startedAt is zero time, it returns an empty string.
// The date format defaults to "Mon, 02 Jan 2006 15:04:05 MST" if not provided.
func Render(subscriptionLabel string, startedAt time.Time, dateFormat string, styles Styles) string {
	if subscriptionLabel == "" || startedAt.IsZero() {
		return ""
	}

	dateFmt := dateFormat
	if dateFmt == "" {
		dateFmt = "Mon, 02 Jan 2006 15:04:05 MST"
	}

	infoStr := fmt.Sprintf("  %s  -  %s", subscriptionLabel, startedAt.Format(dateFmt))
	return styles.InfoStyle.Render(infoStr)
}
