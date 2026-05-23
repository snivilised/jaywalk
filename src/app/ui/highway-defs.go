package ui

import "time"

// buildHighwayLanes defaults used when the user config is incomplete.
const (
	defaultLaneCount = 4
	highwayTickRate  = 50 * time.Millisecond
)

var (
	defaultEmojiPool = []string{"🔍", "⚙️", "🔄", "📡"}
	highwaySpinnerTypes = []string{
		"film-strip",
		"pulse",
		"spinner",
	}
	defaultLabels = []string{"Navigating", "Processing", "Spinner", "Auxiliary"}
)
