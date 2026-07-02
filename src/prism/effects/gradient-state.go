//go:build !race

package effects

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

// GradientState holds in-memory state for a single lane's gradient animation.
type GradientState struct {
	Offset     int // Current position in gradient steps array (0 to len-1)
	Direction  int // 1 = forward, -1 = backward; starts as 1
	TotalSteps int // Length of gradient steps array
	stepsArray []contract.Color
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
func (s *GradientState) SetSteps(steps []contract.Color) {
	s.stepsArray = steps
	s.TotalSteps = len(steps)
	if s.TotalSteps == 0 {
		s.TotalSteps = contract.DefaultStepCount()
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
//
// The offset is treated as the phase of a triangle wave whose period
// is 2 * (n - 1). The phase for this rune is (offset + pos); mapped
// into the gradient index space this means:
//
//	phase 0..n-1        → index 0..n-1       (forward sweep)
//	phase n..2*(n-1)    → index n-2..0       (reverse sweep)
//
// This produces a smooth Hi → Lo → Hi ping-pong with no sharp
// wrap-around at the boundary (which would otherwise be a hard
// jump from step[n-1] back to step[0]). The pos argument is a
// per-rune phase shift that lets the gradient scroll across the
// content while preserving the bounce.
func (s *GradientState) GetEffectiveIndex(pos int) int {
	n := len(s.stepsArray)
	if n == 0 {
		return -1
	}
	if n == 1 {
		return 0
	}

	period := 2 * (n - 1)
	phase := (s.Offset + pos) % period
	if phase < 0 {
		phase += period
	}
	if phase < n {
		return phase
	}
	return period - phase
}

// ApplyGradient applies a colour gradient from Hi to Lo across an animation frame.
// It converts prism color.Color (from image/color package) to our Colour struct,
// interpolates per-character using the gradient state's current offset, and returns
// a slice where each element contains a rune and its interpolated colour.
// The caller must use ApplyGradientStyled() to convert this to a lipgloss-compatible string.
//
//nolint:all
func ApplyGradient(gradient contract.ResolvedGradient, frameContent string, state *GradientState) []RunWithColor {
	hiR, hiG, hiB, _ := gradient.Hi.RGBA()
	loR, loG, loB, _ := gradient.Lo.RGBA()

	steps := contract.InterpolateBetweenRGBA(
		uint8(hiR>>8),
		uint8(hiG>>8),
		uint8(hiB>>8),
		uint8(loR>>8),
		uint8(loG>>8),
		uint8(loB>>8),
		state.TotalSteps,
		gradient.Curve,
		gradient.Easing,
	)

	if len(steps) == 0 {
		return nil // caller should fall back to plain rendering
	}

	// Lazily configure the steps array on the state if not already populated.
	// This ensures GetEffectiveIndex and Update work correctly on subsequent ticks.
	if len(state.stepsArray) == 0 {
		state.SetSteps(steps)
	}

	var result []RunWithColor
	runeSlice := []rune(frameContent)
	for i := range runeSlice {
		r := runeSlice[i]
		stepIdx := state.GetEffectiveIndex(i)

		// Safety check against -1 index lookups
		if stepIdx < 0 || stepIdx >= len(steps) {
			stepIdx = 0
		}

		result = append(result, RunWithColor{
			Rune: r,
			Color: contract.Color{
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
func ApplyGradientStatic(gradient contract.ResolvedGradient, content string, totalSteps int) []RunWithColor {
	hiR, hiG, hiB, _ := gradient.Hi.RGBA()
	loR, loG, loB, _ := gradient.Lo.RGBA()

	steps := contract.InterpolateBetweenRGBA(
		uint8(hiR>>8),
		uint8(hiG>>8),
		uint8(hiB>>8),
		uint8(loR>>8),
		uint8(loG>>8),
		uint8(loB>>8),
		totalSteps,
		gradient.Curve,
		gradient.Easing,
	)

	if len(steps) == 0 {
		return nil
	}

	runeSlice := []rune(content)
	result := make([]RunWithColor, len(runeSlice))
	lastStep := len(steps) - 1
	maxPos := max(len(runeSlice)-1, 1)

	for i, r := range runeSlice {
		idx := min(i*lastStep/maxPos, lastStep)
		result[i] = RunWithColor{
			Rune: r,
			Color: contract.Color{
				R: steps[idx].R,
				G: steps[idx].G,
				B: steps[idx].B,
			},
		}
	}

	return result
}

// RunWithColor represents a rune with its gradient-interpolated colour.
type RunWithColor struct {
	Rune  rune
	Color contract.Color
}

// ApplyGradientStyled converts the gradient run slice into a lipgloss-compatible string.
// Each character is rendered with its interpolated foreground colour using ANSI escape codes.
// The function builds an ANSI-style string where each non-whitespace character gets
// its own RGB foreground colour set before rendering.
func ApplyGradientStyled(runs []RunWithColor) string {
	if len(runs) == 0 {
		return ""
	}

	var b strings.Builder

	for _, rc := range runs {
		r := rc.Rune

		if r == ' ' || r == '\t' || r == '\n' {
			b.WriteRune(r)
			continue
		}

		style := lipgloss.NewStyle().Foreground(
			lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", rc.Color.R, rc.Color.G, rc.Color.B)),
		)

		b.WriteString(style.Render(string(r)))
	}

	return b.String()
}
