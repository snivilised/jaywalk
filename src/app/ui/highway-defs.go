package ui

import "time"

// buildHighwayLanes defaults used when the user config is incomplete.
const (
	defaultLaneCount = 4
	highwayTickRate  = 50 * time.Millisecond
)

var (
	defaultWorkerEmojiPool = []string{"🔍", "⚙️", "🔄", "📡"}
	defaultJobEmojiPool    = []string{"🍎", "🍊", "🍋", "🍇", "🍓", "🍑", "🍒", "🍌", "🍍", "🥝"}
	highwaySpinnerTypes    = []string{
		"wave",
		"musical",
		"morse",
		"starlight",
		"barcode",
		"spinner",
		"braille",
		"braillewave",
		"dna",
		"scan",
		"rain",
		"scanline",
		"braille-pulse",
		"snake",
		"sparkle",
		"cascade",
		"columns",
		"orbit",
		"breathe",
		"waverows",
		"checkerboard",
		"helix",
		"fillsweep",
		"diagswipe",
		"classic-waveform",
		"particle-drift",
		"pulsing-rings",
		"ascii-landscape",
		"matrix-rain",
		"gradient-flow",
		"breathing-circles",
		"network-graph",
		"dot",
		"pulse",
		"globe",
		"moon",
	}
	defaultLabels = []string{
		"Navigating", "Processing", "Scanning", "Indexing",
		"Resolving", "Compiling", "Linking", "Testing",
	}
)
