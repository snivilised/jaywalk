package traffic

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
)

type SpinnerDef struct {
	Frames   func(tick int) string
	Interval time.Duration
}

func filmStripFrame(tick int) string {
	const width = 9
	total := 2 * (width - 2)
	pos := tick % total
	if pos >= width-2 {
		pos = total - pos - 1
	}
	filled := strings.Repeat("▓", pos+1)
	empty := strings.Repeat("░", width-2-(pos+1))
	return "┃" + filled + empty + "┃"
}

func spaceFilledFrame(tick int) string {
	const (
		steps = 8
		full  = "█"
		empty = "░"
	)
	total := 2 * steps
	pos := tick % total
	if pos >= steps {
		pos = total - pos - 1
	}
	return strings.Repeat(full, pos+1) + strings.Repeat(empty, steps-(pos+1))
}

func spinnerFrame(tick int) string {
	return spinner.Line.Frames[tick%len(spinner.Line.Frames)]
}

var builtinSpinners = map[string]SpinnerDef{
	"film-strip":   {Frames: filmStripFrame, Interval: 100 * time.Millisecond},
	"space-filled": {Frames: spaceFilledFrame, Interval: 80 * time.Millisecond},
	"spinner":      {Frames: spinnerFrame, Interval: 100 * time.Millisecond},
}

func Lookup(name string) (SpinnerDef, bool) {
	def, ok := builtinSpinners[name]
	return def, ok
}


