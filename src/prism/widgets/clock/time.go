package clock

import (
	"fmt"
	"time"
)

const (
	millisecondsDurationFormat   = "%dms"
	secondsDurationFormat        = "%ds"
	minutesSecondsDurationFormat = "%dm%ds"
)

// FormatDuration returns a display string for the given duration based on
// its magnitude.
func FormatDuration(d time.Duration) string {
	switch {
	case d < time.Second:
		return FormatDurationMilliseconds(d)
	case d < time.Minute:
		return FormatDurationSeconds(d)
	case d < time.Hour:
		return FormatDurationMinutesSeconds(d)
	default:
		return FormatDurationFull(d)
	}
}

// FormatDurationMilliseconds renders durations shorter than one second
// using milliseconds only.
func FormatDurationMilliseconds(d time.Duration) string {
	return fmt.Sprintf(millisecondsDurationFormat, d.Milliseconds())
}

// FormatDurationSeconds renders durations shorter than one minute using
// seconds only.
func FormatDurationSeconds(d time.Duration) string {
	return fmt.Sprintf(secondsDurationFormat, int(d.Seconds()))
}

// FormatDurationMinutesSeconds renders durations shorter than one hour using
// a minute:second display format.
func FormatDurationMinutesSeconds(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf(minutesSecondsDurationFormat, m, s)
}

// FormatDurationFull renders durations using the standard Go duration string.
func FormatDurationFull(d time.Duration) string {
	return d.String()
}
