package workshop

import (
	"fmt"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

var bloomChars = []string{" ", "○", "◌", "◎", "●"}

// BloomVisualiser renders a radial gradient bloom. The
// gradient radiates outward from a central point. Each
// ring takes one gradient step. Pulses slowly using the
// easing preset.
type BloomVisualiser struct {
	Width int
}

// Name returns the display name for this visualiser.
func (v *BloomVisualiser) Name() string {
	return "bloom"
}

// Render produces a diamond-shaped radial bloom where
// each concentric ring is coloured with a gradient step
// and the animFrame offset shifts the mapping.
func (v *BloomVisualiser) Render(
	steps []contract.Color,
	_ enums.CurveKind,
	easing enums.EasingKind,
	animFrame int,
) string {
	if len(steps) == 0 || v.Width <= 0 {
		return ""
	}

	n := len(steps)
	rings := min(n, maxRings(v.Width))

	if rings == 0 {
		return ""
	}

	var b strings.Builder

	// Distance multiplier based on easing.
	// Uniform = 1.0, EaseIn pulses inward,
	// EaseOut pulses outward.
	easeFactor := pulseFactor(easing, animFrame)

	rp := ringParams{
		b:          &b,
		steps:      steps,
		n:          n,
		animFrame:  animFrame,
		easeFactor: easeFactor,
		rings:      rings,
	}

	// Upper half of diamond (rings-1 lines)
	for row := 0; row < rings; row++ {
		spaces := rings - row - 1
		b.WriteString(strings.Repeat(" ", spaces))

		for ring := 0; ring <= row; ring++ {
			rp.ring = ring

			writeRing(&rp)
		}
		b.WriteString("\n")
	}

	// Lower half of diamond (rings-1 lines, inverted)
	for row := rings - 2; row >= 0; row-- {
		spaces := rings - row - 1
		b.WriteString(strings.Repeat(" ", spaces))

		for ring := 0; ring <= row; ring++ {
			rp.ring = ring
			writeRing(&rp)
		}
		if row > 0 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

type ringParams struct {
	b          *strings.Builder
	steps      []contract.Color
	ring       int
	n          int
	animFrame  int
	easeFactor float64
	rings      int
}

func writeRing(params *ringParams) {
	stepIdx := ringAt(params.ring, params.n, params.animFrame, params.easeFactor)
	if stepIdx < 0 || stepIdx >= params.n {
		stepIdx = 0
	}
	c := params.steps[stepIdx]
	ch := charForRing(params.ring, params.rings)
	style := lipgloss.NewStyle().Foreground(
		lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)),
	)
	params.b.WriteString(style.Render(ch))
	params.b.WriteString(" ")
}

// maxRings returns the number of rings that fit in the
// given width. Each ring pair takes 2 characters (char +
// space), so at most floor(width/2) rings, capped at
// rings for a diamond shape.
func maxRings(width int) int {
	maxR := width / 2
	if maxR > 8 {
		return 8
	}
	return maxR
}

// ringAt returns the gradient step index for the given
// ring, by cycling through steps and pulsing with
// the animFrame offset.
func ringAt(
	ring, n, animFrame int, easeFactor float64,
) int {
	offset := int(float64(animFrame) * easeFactor)
	return (ring + offset) % n
}

// charForRing maps a ring index to a visual character.
// Outer rings use sparse characters, inner rings use
// denser characters.
func charForRing(ring, total int) string {
	if ring == total-1 {
		return "●"
	}
	// Distribute remaining characters across rings
	idx := (ring * (len(bloomChars) - 2)) / max(total-1, 1)
	if idx >= len(bloomChars)-1 {
		idx = len(bloomChars) - 2
	}
	return bloomChars[idx+1]
}

// pulseFactor returns a multiplier based on the easing
// preset and animation frame. EaseIn accelerates toward
// the center, EaseOut decelerates, EaseInOut oscillates.
func pulseFactor(
	easing enums.EasingKind, animFrame int,
) float64 {
	switch easing { //nolint:exhaustive // ok
	case enums.EasingKindEaseIn:
		return 0.5 + 0.5*math.Sin(
			float64(animFrame)*0.1,
		)
	case enums.EasingKindEaseOut:
		return 0.5 + 0.5*math.Cos(
			float64(animFrame)*0.1,
		)
	case enums.EasingKindEaseInOut:
		return 0.3 + 0.7*math.Sin(
			float64(animFrame)*0.05,
		)
	default:
		return 1.0
	}
}
