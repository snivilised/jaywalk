//go:build !race

package highway

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// Colour is a simple RGBA colour type for in-package gradient interpolation.
// This mirrors image/color.RGBA but without alpha support (alpha always 255).
type Colour struct {
	R, G, B uint8
}

// DefaultStepCount returns the default number of gradient steps.
// Delegates to contract.DefaultStepCount to avoid duplication.
func DefaultStepCount() int {
	return contract.DefaultStepCount()
}

// GradientState holds in-memory state for a single lane's gradient animation.
type GradientState struct {
	Offset     int // Current position in gradient steps array (0 to len-1)
	Direction  int // 1 = forward, -1 = backward; starts as 1
	TotalSteps int // Length of gradient steps array
	stepsArray []Colour
}

// NewGradientState creates a new gradient state for a lane.
func NewGradientState() *GradientState {
	return &GradientState{
		Offset:     0,
		Direction:  1,
		TotalSteps: 0,
	}
}

// SetSteps configures the gradient steps array for this state.
func (s *GradientState) SetSteps(steps []Colour) {
	s.stepsArray = steps
	s.TotalSteps = len(steps)
	if s.TotalSteps == 0 {
		s.TotalSteps = DefaultStepCount()
	}
}

// Update advances the gradient state for one animation tick.
func (s *GradientState) Update(windowSize int) {
	if s.TotalSteps <= 0 || len(s.stepsArray) == 0 {
		return
	}

	lastIdx := s.TotalSteps - 1
	s.Offset += windowSize * s.Direction

	if s.Offset >= lastIdx {
		s.Offset = lastIdx
		s.Direction = -1
	} else if s.Offset <= 0 {
		s.Offset = 0
		s.Direction = 1
	}
}

// Reset resets the gradient state to initial position.
func (s *GradientState) Reset() {
	s.Offset = 0
	s.Direction = 1
	s.TotalSteps = 0
	s.stepsArray = nil
}

// GetEffectiveIndex returns the gradient index for character at position pos.
// Uses modular wrapping so the gradient scrolls through the content rather
// than clamping to the last colour. The Offset acts as a phase shift that
// advances each tick via Update(), creating a pulsing sweep effect.
func (s *GradientState) GetEffectiveIndex(pos int) int {
	if len(s.stepsArray) == 0 {
		return -1
	}

	idx := (s.Offset + pos) % len(s.stepsArray)
	if idx < 0 {
		idx += len(s.stepsArray)
	}

	return idx
}

// ApplyGradient applies a colour gradient from Hi to Lo across an animation frame.
// It converts prism color.Color (from image/color package) to our Colour struct,
// interpolates per-character using the gradient state's current offset, and returns
// a slice where each element contains a rune and its interpolated colour.
// The caller must use ApplyGradientStyled() to convert this to a lipgloss-compatible string.
//
//nolint:all
func ApplyGradient(hiCol, loCol color.Color, frameContent string, state *GradientState) []RunWithColour {
	hiR, hiG, hiB, _ := hiCol.RGBA()
	loR, loG, loB, _ := loCol.RGBA()

	steps := InterpolateBetweenRGBA(
		uint8(hiR>>8),
		uint8(hiG>>8),
		uint8(hiB>>8),
		uint8(loR>>8),
		uint8(loG>>8),
		uint8(loB>>8),
		state.TotalSteps,
	)

	if len(steps) == 0 {
		return nil // caller should fall back to plain rendering
	}

	// Lazily configure the steps array on the state if not already populated.
	// This ensures GetEffectiveIndex and Update work correctly on subsequent ticks.
	if len(state.stepsArray) == 0 {
		state.SetSteps(steps)
	}

	var result []RunWithColour
	runeSlice := []rune(frameContent)
	for i := 0; i < len(runeSlice); i++ {
		r := runeSlice[i]
		stepIdx := state.GetEffectiveIndex(i)

		// Safety check against -1 index lookups
		if stepIdx < 0 || stepIdx >= len(steps) {
			stepIdx = 0
		}

		result = append(result, RunWithColour{
			Rune: r,
			Colour: Colour{
				R: steps[stepIdx].R,
				G: steps[stepIdx].G,
				B: steps[stepIdx].B,
			},
		})
	}

	return result
}

// ApplyGradientStatic applies a colour gradient from Hi to Lo across content
// without animation. Each character gets a colour based on its position in the
// content, distributed evenly across the gradient steps. Unlike ApplyGradient,
// this does not use GradientState and produces a fixed, non-animated output.
//
//nolint:all
func ApplyGradientStatic(hiCol, loCol color.Color, content string, totalSteps int) []RunWithColour {
	hiR, hiG, hiB, _ := hiCol.RGBA()
	loR, loG, loB, _ := loCol.RGBA()

	steps := InterpolateBetweenRGBA(
		uint8(hiR>>8),
		uint8(hiG>>8),
		uint8(hiB>>8),
		uint8(loR>>8),
		uint8(loG>>8),
		uint8(loB>>8),
		totalSteps,
	)

	if len(steps) == 0 {
		return nil
	}

	runeSlice := []rune(content)
	result := make([]RunWithColour, len(runeSlice))
	lastStep := len(steps) - 1
	maxPos := max(len(runeSlice)-1, 1)

	for i, r := range runeSlice {
		idx := i * lastStep / maxPos
		if idx > lastStep {
			idx = lastStep
		}
		result[i] = RunWithColour{
			Rune: r,
			Colour: Colour{
				R: steps[idx].R,
				G: steps[idx].G,
				B: steps[idx].B,
			},
		}
	}

	return result
}

// RunWithColour represents a rune with its gradient-interpolated colour.
type RunWithColour struct {
	Rune   rune
	Colour Colour
}

// ApplyGradientStyled converts the gradient run slice into a lipgloss-compatible string.
// Each character is rendered with its interpolated foreground colour using ANSI escape codes.
// The function builds an ANSI-style string where each non-whitespace character gets
// its own RGB foreground colour set before rendering.
func ApplyGradientStyled(runs []RunWithColour) string {
	if len(runs) == 0 {
		return ""
	}

	var result strings.Builder

	for _, rc := range runs {
		r := rc.Rune

		// Skip whitespace - keep original formatting
		if r == ' ' || r == '\t' || r == '\n' {
			result.WriteByte(byte(r))
			continue
		}

		// Apply ANSI escape code with RGB foreground colour using standard decimal values
		ansiColor := fmt.Sprintf("\x1b[38;2;%d;%d;%dm", rc.Colour.R, rc.Colour.G, rc.Colour.B)
		result.WriteString(ansiColor)
		result.WriteRune(r)
		result.WriteString("\x1b[0m") // reset
	}

	return result.String()
}

// InterpolateBetweenRGBA creates a gradient of 'steps' colours between hi and lo RGB values.
// Returns nil if steps <= 0. Each colour has alpha=255 (opaque).
func InterpolateBetweenRGBA(hiR, hiG, hiB, loR, loG, loB uint8, steps int) []Colour {
	steps = max(steps, 2)

	gradient := make([]Colour, steps)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)

		r := uint8(math.Max(0, math.Min(255, float64(hiR)+(float64(loR)-float64(hiR))*t)))
		g := uint8(math.Max(0, math.Min(255, float64(hiG)+(float64(loG)-float64(hiG))*t)))
		b := uint8(math.Max(0, math.Min(255, float64(hiB)+(float64(loB)-float64(hiB))*t)))

		gradient[i] = Colour{R: r, G: g, B: b}
	}

	return gradient
}
