//go:build !race

package highway

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widget"
)

var _ = Describe("Periscope Gradient", func() {
	var (
		dummyStyle  lipgloss.Style
		testStyles  widget.PeriscopeStyles
	)

	BeforeEach(func() {
		dummyStyle = lipgloss.NewStyle()
		testStyles = widget.PeriscopeStyles{
			Filled: dummyStyle,
			Empty:  dummyStyle,
		}
	})

	Context("GradientState lazy initialization", func() {
		It("initializes periscope gradient state correctly", func() {
			state := NewGradientState()
			Expect(state.Offset).To(Equal(0))
			Expect(state.Direction).To(Equal(1))
			Expect(state.TotalSteps).To(Equal(0))
		})
	})

	Context("ApplyGradientStatic", func() {
		It("applies gradient across content by position", func() {
			hi := color.RGBA{R: 255, G: 0, B: 0, A: 255}
			lo := color.RGBA{R: 0, G: 0, B: 255, A: 255}
			runs := ApplyGradientStatic(hi, lo, "◼◼◼◼◼", 5)
			Expect(runs).To(HaveLen(5))
			// First char should be close to Hi (red)
			Expect(runs[0].Colour.R).To(BeNumerically(">", runs[4].Colour.R))
			// Last char should be close to Lo (blue)
			Expect(runs[4].Colour.B).To(BeNumerically(">", runs[0].Colour.B))
		})

		It("returns nil for empty content", func() {
			hi := color.RGBA{R: 255, G: 0, B: 0, A: 255}
			lo := color.RGBA{R: 0, G: 0, B: 255, A: 255}
			runs := ApplyGradientStatic(hi, lo, "", 5)
			Expect(runs).To(HaveLen(0))
		})
	})

	Context("renderPeriscopeBar with animated gradient", func() {
		It("applies gradient across entire bar width (filled + empty)", func() {
			model := Model{
				theme: contract.Theme{},
			}
			hi := color.RGBA{R: 255, G: 0, B: 0, A: 255}
			lo := color.RGBA{R: 0, G: 0, B: 255, A: 255}
			lane := Lane{
				Depth: 3,
				PeriscopeGradient: &contract.ResolvedGradient{
					Steps:   4,
					Hi:      hi,
					Lo:      lo,
					Animate: true,
				},
				PeriscopeGradientState: NewGradientState(),
			}
			lane.PeriscopeGradientState.TotalSteps = 4

			model.maxDepth = 5
			result := model.renderPeriscopeBar(lane, 0, 5, testStyles)
			// Both filled (◼) and empty (◻) characters should have gradient ANSI codes
			stripped := stripANSI(result)
			Expect(stripped).To(Equal("◼◼◼◻◻"))
			// Verify ANSI codes appear for all positions
			ansiCount := strings.Count(result, "\x1b[38;2;")
			Expect(ansiCount).To(Equal(5))
		})

		It("falls back to static styles when no periscope gradient is set", func() {
			model := Model{
				theme: contract.Theme{},
			}
			lane := Lane{
				Depth: 2,
			}

			model.maxDepth = 5
			result := model.renderPeriscopeBar(lane, 0, 5, testStyles)
			Expect(result).To(ContainSubstring("◼"))
			Expect(result).To(ContainSubstring("◻"))
		})

		It("renders empty bar when fill is 0", func() {
			model := Model{
				theme: contract.Theme{},
			}

			lane := Lane{
				Depth: 0,
			}

			model.maxDepth = 5
			result := model.renderPeriscopeBar(lane, 0, 5, testStyles)
			Expect(result).To(Equal("◻◻◻◻◻"))
		})
	})

	Context("renderPeriscopeBar with static gradient (Animate=false)", func() {
		It("applies static gradient across entire bar width without GradientState", func() {
			model := Model{
				theme: contract.Theme{},
			}
			hi := color.RGBA{R: 255, G: 0, B: 0, A: 255}
			lo := color.RGBA{R: 0, G: 0, B: 255, A: 255}
			lane := Lane{
				Depth: 3,
				PeriscopeGradient: &contract.ResolvedGradient{
					Steps:   4,
					Hi:      hi,
					Lo:      lo,
					Animate: false,
				},
				// PeriscopeGradientState is nil — static path doesn't use it
			}

			model.maxDepth = 5
			result := model.renderPeriscopeBar(lane, 0, 5, testStyles)
			stripped := stripANSI(result)
			Expect(stripped).To(Equal("◼◼◼◻◻"))
			// Verify ANSI codes appear for all positions
			ansiCount := strings.Count(result, "\x1b[38;2;")
			Expect(ansiCount).To(Equal(5))
		})

		It("renders empty bar with gradient when fill=0", func() {
			model := Model{
				theme: contract.Theme{},
			}
			hi := color.RGBA{R: 255, G: 0, B: 0, A: 255}
			lo := color.RGBA{R: 0, G: 0, B: 255, A: 255}
			lane := Lane{
				Depth: 0,
				PeriscopeGradient: &contract.ResolvedGradient{
					Steps:   4,
					Hi:      hi,
					Lo:      lo,
					Animate: false,
				},
			}

			model.maxDepth = 5
			result := model.renderPeriscopeBar(lane, 0, 5, testStyles)
			stripped := stripANSI(result)
			Expect(stripped).To(Equal("◻◻◻◻◻"))
			// Static gradient applied even when all empty
			Expect(result).To(ContainSubstring("\x1b[38;2;"))
		})
	})

	Context("GradientState advancement", func() {
		It("advances periscope gradient state via Update after ApplyGradient populates stepsArray", func() {
			lane := Lane{
				PeriscopeGradient: &contract.ResolvedGradient{
					Steps: 4,
					Hi:    color.RGBA{R: 255, G: 0, B: 0, A: 255},
					Lo:    color.RGBA{R: 0, G: 0, B: 255, A: 255},
				},
				PeriscopeGradientState: NewGradientState(),
			}
			lane.PeriscopeGradientState.TotalSteps = 4

			ApplyGradient(
				lane.PeriscopeGradient.Hi,
				lane.PeriscopeGradient.Lo,
				"◼◼◼",
				lane.PeriscopeGradientState,
			)

			offsetBefore := lane.PeriscopeGradientState.Offset
			lane.PeriscopeGradientState.Update(1)
			Expect(lane.PeriscopeGradientState.Offset).To(Equal(offsetBefore + 1))
		})
	})
})

// stripANSI removes ANSI escape sequences from a string for test assertions.
func stripANSI(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
}
