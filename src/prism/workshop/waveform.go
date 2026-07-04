package workshop

import (
	"fmt"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

var heightChars = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

// WaveformVisualiser renders a sine wave of
// height-varying block characters whose colour follows
// the gradient sweep. A flat sweep bar beneath provides
// the accurate colour reference.
type WaveformVisualiser struct {
	Width int
}

// Name returns the display name for this visualiser.
func (v *WaveformVisualiser) Name() string {
	return "waveform"
}

// Render produces two rows of output:
//   - Row 1: waveform of height-varying block characters
//     coloured by gradient step.
//   - Row 2: flat sweep bar with evenly distributed
//     gradient steps.
func (v *WaveformVisualiser) Render(
	steps []contract.Color,
	curve enums.CurveKind,
	easing enums.EasingKind,
	animFrame int,
) string {
	if len(steps) == 0 || v.Width <= 0 {
		return ""
	}

	var b strings.Builder
	n := len(steps)

	// The curve type affects the colour mapping so that
	// higher curve values bunch colours toward one end.
	// We apply the curve shape when mapping column
	// position to step index.
	freq := 3.0
	speed := animationSpeed(easing)

	// Row 1: waveform
	for col := 0; col < v.Width; col++ {
		t := float64(col) / float64(v.Width)
		phase := t * 2 * math.Pi * freq
		phase += float64(animFrame) * speed * 0.1
		h := math.Sin(phase)

		idx := int((h + 1) / 2 * 7)
		if idx < 0 {
			idx = 0
		}
		if idx > 7 {
			idx = 7
		}

		stepIdx := stepForPosition(t, n, curve)
		c := steps[stepIdx]
		style := lipgloss.NewStyle().Foreground(
			lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)),
		)
		b.WriteString(style.Render(heightChars[idx]))
	}

	b.WriteString("\n")

	// Row 2: flat sweep bar
	for col := 0; col < v.Width; col++ {
		t := float64(col) / float64(v.Width)
		stepIdx := stepForPosition(t, n, curve)

		c := steps[stepIdx]
		style := lipgloss.NewStyle().Foreground(
			lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)),
		)
		b.WriteString(style.Render("█"))
	}

	return b.String()
}

// stepForPosition maps a fractional position [0,1] to a
// gradient step index, applying the curve shape so that
// non-linear curves bunch steps toward one end.
func stepForPosition(t float64, n int, curve enums.CurveKind) int {
	curved := applyCurve(t, curve)
	stepIdx := int(math.Round(curved * float64(n-1)))
	if stepIdx >= n {
		stepIdx = n - 1
	}
	return stepIdx
}

// animationSpeed returns a multiplier based on the easing
// preset. EaseIn accelerates, EaseOut decelerates,
// EaseInOut oscillates.
func animationSpeed(easing enums.EasingKind) float64 {
	switch easing { //nolint:exhaustive // ok
	case enums.EasingKindEaseIn:
		return 1.5
	case enums.EasingKindEaseOut:
		return 0.5
	case enums.EasingKindEaseInOut:
		return 1.0 + math.Sin(float64(0))
	default:
		return 1.0
	}
}
