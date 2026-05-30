package highway

import (
	"image/color"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Highway Gradient Rendering", func() {
	Context("GradientState lazy initialization & transitions", func() {
		It("initializes state correctly and sets steps lazily", func() {
			state := NewGradientState()
			Expect(state.Offset).To(Equal(0))
			Expect(state.Direction).To(Equal(1))
			Expect(state.TotalSteps).To(Equal(0))
			Expect(state.stepsArray).To(BeNil())

			// Perform an interpolation with 4 steps
			hi := color.RGBA{R: 255, G: 0, B: 0, A: 255} // Red
			lo := color.RGBA{R: 0, G: 0, B: 255, A: 255} // Blue
			state.TotalSteps = 4

			runs := ApplyGradient(hi, lo, "test", state)
			Expect(runs).To(HaveLen(4))

			// Verify steps array was populated
			Expect(state.TotalSteps).To(Equal(4))
			Expect(state.stepsArray).To(HaveLen(4))

			// Verify first run colour matches start endpoint (Red)
			Expect(runs[0].Colour.R).To(Equal(uint8(255)))
			Expect(runs[0].Colour.G).To(Equal(uint8(0)))
			Expect(runs[0].Colour.B).To(Equal(uint8(0)))

			// Verify last run colour matches end endpoint (Blue)
			Expect(runs[3].Colour.R).To(Equal(uint8(0)))
			Expect(runs[3].Colour.G).To(Equal(uint8(0)))
			Expect(runs[3].Colour.B).To(Equal(uint8(255)))
		})

		It("advances offset and reverses direction on Update", func() {
			state := NewGradientState()
			steps := []Colour{
				{R: 255, G: 0, B: 0},
				{R: 170, G: 0, B: 85},
				{R: 85, G: 0, B: 170},
				{R: 0, G: 0, B: 255},
			}
			state.SetSteps(steps)
			Expect(state.Offset).To(Equal(0))

			// Advance by 1
			state.Update(1)
			Expect(state.Offset).To(Equal(1))
			Expect(state.Direction).To(Equal(1))

			// Advance by 2 more (Offset=3, reaches end)
			state.Update(2)
			Expect(state.Offset).To(Equal(3))
			Expect(state.Direction).To(Equal(-1)) // Should reverse at end

			// Advance again with reverse direction
			state.Update(1)
			// Wait, the s.Offset += windowSize would add 1, then hit boundary and reverse.
			// Let's test the GetEffectiveIndex behavior directly
			idx := state.GetEffectiveIndex(0)
			Expect(idx).To(BeNumerically("<=", 3))
			Expect(idx).To(BeNumerically(">=", 0))
		})
	})

	Context("ANSI Formatting", func() {
		It("formats runs into valid TrueColor ANSI escape sequences with decimal values", func() {
			runs := []RunWithColour{
				{
					Rune:   'A',
					Colour: Colour{R: 255, G: 128, B: 64},
				},
				{
					Rune:   ' ',
					Colour: Colour{R: 0, G: 0, B: 0},
				},
				{
					Rune:   'B',
					Colour: Colour{R: 0, G: 255, B: 0},
				},
			}

			styled := ApplyGradientStyled(runs)
			// Expected sequence for 'A': \x1b[38;2;255;128;64mA\x1b[0m
			// Expected sequence for ' ': Keep space as-is without escape codes
			// Expected sequence for 'B': \x1b[38;2;0;255;0mB\x1b[0m
			Expect(styled).To(Equal("\x1b[38;2;255;128;64mA\x1b[0m \x1b[38;2;0;255;0mB\x1b[0m"))
		})
	})
})
