package ui

import "time"

// buildHighwayLanes defaults used when the user config is incomplete.
const (
	defaultLaneCount = 4
	highwayTickRate  = 50 * time.Millisecond

	// bannerDefaultTickMs is the fallback tick interval (ms) for
	// the banner's gradient animation. Users can override via
	// ui.highway.banner.tick. The value is in milliseconds to match
	// the underlying config field; the model converts to
	// time.Duration on receipt.
	bannerDefaultTickMs = 500
)

var (
	defaultWorkerEmojiPool = []string{"🔍", "⚙️", "🔄", "📡"}
	defaultJobEmojiPool    = []string{"🍎", "🍊", "🍋", "🍇", "🍓", "🍑", "🍒", "🍌", "🍍", "🥝"}
	highwaySpinnerTypes    = []string{
		"ascii-landscape",
		"barcode",
		"braille-pulse",
		"braille-wave",
		"braille",
		"breathe",
		"breathing-circles",
		"cascade",
		"checkerboard",
		"classic-waveform",
		"columns",
		"diagswipe",
		"dna",
		"dot",
		"fillsweep",
		"globe",
		"gradient-flow",
		"helix",
		"matrix-rain",
		"moon",
		"morse",
		"musical",
		"network-graph",
		"orbit",
		"particle-drift",
		"pulse",
		"pulsing-rings",
		"rain",
		"scan",
		"scanline",
		"snake",
		"sparkle",
		"spinner",
		"starlight",
		"wave",
		"waverows",
	}
	defaultLabels = []string{
		"Navigating", "Processing", "Scanning", "Indexing",
		"Resolving", "Compiling", "Linking", "Testing",
	}
)
