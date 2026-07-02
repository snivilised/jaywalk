package contract

import (
	"image/color"
	"math"

	"github.com/snivilised/jaywalk/src/agenor/enums"
)

// applyCurve applies the curve shaping function to a linear parameter t
// in [0,1]. The zero value (CurveKindLinear) is a passthrough.
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

// applyEasing applies the easing distribution function to a parameter t
// in [0,1]. The zero value (EasingKindUniform) is a passthrough.
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

// easedT applies curve shaping then easing distribution to a linear
// parameter t in [0,1]. Either parameter may be zero value, in which
// case that stage is a no-op passthrough.
func easedT(t float64, curve enums.CurveKind, easing enums.EasingKind) float64 {
	t = applyCurve(t, curve)
	t = applyEasing(t, easing)
	return t
}

// InterpolateBetween creates a gradient of 'steps' colours between hiCol and loCol
// by interpolating each RGB channel independently using the given curve and easing.
// If steps is 0 or negative, it uses DefaultStepCount().
func InterpolateBetween(hiCol, loCol color.Color, steps int, curve enums.CurveKind, easing enums.EasingKind) []Color {
	if steps <= 0 {
		steps = DefaultStepCount()
	}

	steps = max(steps, 2) // ensure at least 2 steps for a visible gradient

	hiR, hiG, hiB, _ := hiCol.RGBA()
	loR, loG, loB, _ := loCol.RGBA()

	// Convert to float for interpolation (already in 0-255 range)
	rStart := float64(hiR >> 8)
	gStart := float64(hiG >> 8)
	bStart := float64(hiB >> 8)

	rEnd := float64(loR >> 8)
	gEnd := float64(loG >> 8)
	bEnd := float64(loB >> 8)

	gradient := make([]Color, steps)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1) // 0.0 to 1.0
		t = easedT(t, curve, easing)

		r := rStart + (rEnd-rStart)*t
		g := gStart + (gEnd-gStart)*t
		b := bStart + (bEnd-bStart)*t

		gradient[i] = Color{
			R: uint8(math.Max(0, math.Min(255, r))),
			G: uint8(math.Max(0, math.Min(255, g))),
			B: uint8(math.Max(0, math.Min(255, b))),
		}
	}

	return gradient
}

// InterpolateBetweenRGBA creates a gradient of 'steps' colours between
// hi and lo RGB values using the given curve and easing.
// Returns nil if steps <= 0.
func InterpolateBetweenRGBA(
	hiR, hiG, hiB, loR, loG, loB uint8,
	steps int,
	curve enums.CurveKind,
	easing enums.EasingKind,
) []Color {
	steps = max(steps, 2)

	gradient := make([]Color, steps)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)
		t = easedT(t, curve, easing)

		r := uint8(math.Max(0, math.Min(255, float64(hiR)+(float64(loR)-float64(hiR))*t)))
		g := uint8(math.Max(0, math.Min(255, float64(hiG)+(float64(loG)-float64(hiG))*t)))
		b := uint8(math.Max(0, math.Min(255, float64(hiB)+(float64(loB)-float64(hiB))*t)))

		gradient[i] = Color{R: r, G: g, B: b}
	}

	return gradient
}

// DefaultStepCount returns the default number of gradient steps when not specified.
// It's based on typical animation window sizes (6-10 characters).
func DefaultStepCount() int {
	return 8 // standard 8-step gradient
}
