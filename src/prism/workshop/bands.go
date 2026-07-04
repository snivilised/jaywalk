package workshop

import (
	"fmt"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

// brailleHeights maps a normalised [0,1] value to a
// braille character whose dot pattern corresponds to
// the vertical position (bottom to top).
var brailleHeights = []string{
	"⣀", // 0.0   bottom
	"⡠", // 0.25
	"⠤", // 0.5   middle
	"⠊", // 0.75
	"⠈", // 1.0   top
}

// BandsVisualiser renders a flat gradient bar with a
// braille curve indicator row above it showing the
// easing curve shape.
type BandsVisualiser struct {
	Width int
}

// Name returns the display name for this visualiser.
func (v *BandsVisualiser) Name() string {
	return "bands"
}

// Render produces two rows of output:
//   - Row 1: braille curve indicator tracing the easing
//     curve shape.
//   - Row 2: flat gradient bar with evenly distributed
//     gradient steps.
func (v *BandsVisualiser) Render(
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

	// Row 1: braille curve indicator
	mutedColor := "#888888"
	mutedStyle := lipgloss.NewStyle().Foreground(
		lipgloss.Color(mutedColor),
	)

	for col := 0; col < v.Width; col++ {
		t := float64(col) / float64(v.Width)
		y := easedCurve(t, curve, easing)
		ch := brailleForHeight(y)
		b.WriteString(mutedStyle.Render(ch))
	}

	b.WriteString("\n")

	// Row 2: gradient bar
	styles := make([]lipgloss.Style, n)
	for i, c := range steps {
		styles[i] = lipgloss.NewStyle().Foreground(
			lipgloss.Color(
				fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B),
			),
		)
	}

	for col := 0; col < v.Width; col++ {
		t := float64(col) / float64(v.Width)
		stepIdx := stepForPosition(t, n, curve)

		b.WriteString(styles[stepIdx].Render("█"))
	}

	return b.String()
}

// brailleForHeight maps a value in [0,1] to a braille
// character. 0 maps to the bottom row, 1 maps to the
// top row.
func brailleForHeight(y float64) string {
	n := len(brailleHeights)
	if y <= 0 {
		return brailleHeights[0]
	}
	if y >= 1 {
		return brailleHeights[n-1]
	}
	idx := max(int(math.Round(y*float64(n-1))), 0)
	if idx >= n {
		idx = n - 1
	}
	return brailleHeights[idx]
}

// easedCurve applies the curve shaping then easing
// distribution to a linear parameter t in [0,1].
func easedCurve(
	t float64,
	curve enums.CurveKind,
	easing enums.EasingKind,
) float64 {
	t = applyCurve(t, curve)
	t = applyEasing(t, easing)
	return t
}

// applyCurve applies the curve shaping function to a
// linear parameter t in [0,1].
func applyCurve(t float64, curve enums.CurveKind) float64 {
	switch curve { //nolint:exhaustive // ok
	case enums.CurveKindSine:
		return (1 - math.Cos(math.Pi*t)) / 2
	case enums.CurveKindQuadraticIn:
		return t * t
	case enums.CurveKindQuadraticOut:
		return 1 - (1-t)*(1-t)
	case enums.CurveKindCubic:
		return 3*t*t - 2*t*t*t
	default:
		return t
	}
}

// applyEasing applies the easing distribution function
// to a parameter t in [0,1].
func applyEasing(t float64, easing enums.EasingKind) float64 {
	switch easing { //nolint:exhaustive // ok
	case enums.EasingKindEaseIn:
		return t * t
	case enums.EasingKindEaseOut:
		return 1 - (1-t)*(1-t)
	case enums.EasingKindEaseInOut:
		return (1 - math.Cos(math.Pi*t)) / 2
	default:
		return t
	}
}
