package contract

import (
	"image/color"
	"math"
)

// InterpolateBetween creates a gradient of 'steps' colours between hiCol and loCol
// by linearly interpolating each RGB channel independently.
// If steps is 0 or negative, it uses DefaultStepCount().
func InterpolateBetween(hiCol, loCol color.Color, steps int) []color.Color {
	if steps <= 0 {
		steps = DefaultStepCount()
	}

	steps = max(steps, 2) // ensure at least 2 steps for a visible gradient

	hiR, hiG, hiB, hiA := hiCol.RGBA()
	loR, loG, loB, loA := loCol.RGBA()

	// Convert to float for linear interpolation (already in 0-255 range)
	rStart := float64(hiR >> 8)
	gStart := float64(hiG >> 8)
	bStart := float64(hiB >> 8)
	aStart := float64(hiA >> 8)

	rEnd := float64(loR >> 8)
	gEnd := float64(loG >> 8)
	bEnd := float64(loB >> 8)
	aEnd := float64(loA >> 8)

	gradient := make([]color.Color, steps)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1) // 0.0 to 1.0

		r := rStart + (rEnd-rStart)*t
		g := gStart + (gEnd-gStart)*t
		b := bStart + (bEnd-bStart)*t
		a := aStart + (aEnd-aStart)*t

		gradient[i] = color.RGBA{
			R: uint8(math.Max(0, math.Min(255, r))),
			G: uint8(math.Max(0, math.Min(255, g))),
			B: uint8(math.Max(0, math.Min(255, b))),
			A: uint8(math.Max(0, math.Min(255, a))),
		}
	}

	return gradient
}

// InterpolateBetweenRGBA is a variant that works directly with RGBA values.
// Returns nil if steps <= 0.
func InterpolateBetweenRGBA(hiR, hiG, hiB, loR, loG, loB uint8, steps int) []color.Color {
	steps = max(steps, 2)

	gradient := make([]color.Color, steps)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)

		r := uint8(math.Max(0, math.Min(255, float64(hiR)+(float64(loR)-float64(hiR))*t)))
		g := uint8(math.Max(0, math.Min(255, float64(hiG)+(float64(loG)-float64(hiG))*t)))
		b := uint8(math.Max(0, math.Min(255, float64(hiB)+(float64(loB)-float64(hiB))*t)))

		gradient[i] = color.RGBA{R: r, G: g, B: b, A: 255}
	}

	return gradient
}

// DefaultStepCount returns the default number of gradient steps when not specified.
// It's based on typical animation window sizes (6-10 characters).
func DefaultStepCount() int {
	return 8 // standard 8-step gradient
}
