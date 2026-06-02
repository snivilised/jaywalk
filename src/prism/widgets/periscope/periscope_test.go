//go:build !race

package periscope_test

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
	"github.com/snivilised/jaywalk/src/prism/widgets/periscope"
	lab "github.com/snivilised/jaywalk/test/laboratory"
)

var _ = Describe("Periscope.Render", func() {
	Describe("basic rendering", func() {
		It("renders all empty squares at fill 0", func() {
			out := ansi.Strip(periscope.Render(periscope.Config{
				Width: 5,
				Fill:  0,
			}, periscope.Styles{}, periscope.Effect{}))
			Expect(out).To(Equal("◻◻◻◻◻"))
		})

		It("renders all filled squares at fill == width", func() {
			out := ansi.Strip(periscope.Render(periscope.Config{
				Width: 5,
				Fill:  5,
			}, periscope.Styles{}, periscope.Effect{}))
			Expect(out).To(Equal("◼◼◼◼◼"))
		})

		It("renders partial fill correctly", func() {
			out := ansi.Strip(periscope.Render(periscope.Config{
				Width: 5,
				Fill:  3,
			}, periscope.Styles{}, periscope.Effect{}))
			Expect(out).To(Equal("◼◼◼◻◻"))
		})

		It("clamps fill at width", func() {
			out := ansi.Strip(periscope.Render(periscope.Config{
				Width: 5,
				Fill:  7,
			}, periscope.Styles{}, periscope.Effect{}))
			Expect(out).To(Equal("◼◼◼◼◼"))
		})

		It("handles width 0", func() {
			out := ansi.Strip(periscope.Render(periscope.Config{
				Width: 0,
				Fill:  0,
			}, periscope.Styles{}, periscope.Effect{}))
			Expect(out).To(BeEmpty())
		})

		It("wraps each segment in its own style", func() {
			out := periscope.Render(periscope.Config{
				Width: 3,
				Fill:  1,
			}, periscope.Styles{}, periscope.Effect{})
			stripped := ansi.Strip(out)
			Expect(stripped).To(Equal("◼◻◻"))
		})
	})

	Describe("gradient application", func() {
		var (
			hi     color.Color
			lo     color.Color
			grad   *contract.ResolvedGradient
			state  *effects.GradientState
			styles periscope.Styles
		)

		BeforeEach(func() {
			hi = color.RGBA{R: 255, G: 0, B: 0, A: 255}
			lo = color.RGBA{R: 0, G: 0, B: 255, A: 255}
			grad = &contract.ResolvedGradient{
				Steps:   4,
				Hi:      hi,
				Lo:      lo,
				Animate: true,
			}
			state = effects.NewGradientState()
			state.TotalSteps = 4
			styles = periscope.Styles{
				Filled: lipgloss.NewStyle(),
				Empty:  lipgloss.NewStyle(),
			}
		})

		Context("animated gradient", func() {
			It("applies gradient across entire bar width (filled + empty)", func() {
				result := periscope.Render(periscope.Config{
					Width: 5,
					Fill:  3,
				}, styles, periscope.Effect{
					Gradient: grad,
					State:    state,
				})

				stripped := lab.StripANSI(result)
				Expect(stripped).To(Equal("◼◼◼◻◻"))
				ansiCount := strings.Count(result, "\x1b[38;2;")
				Expect(ansiCount).To(Equal(5))
			})

			It("interpolates colours from Hi towards Lo", func() {
				result := periscope.Render(periscope.Config{
					Width: 4,
					Fill:  2,
				}, styles, periscope.Effect{
					Gradient: grad,
					State:    state,
				})

				Expect(result).To(ContainSubstring("\x1b[38;2;255;0;0m"))
				Expect(result).To(ContainSubstring("\x1b[38;2;0;0;255m"))
			})

			It("falls back to plain styles when State is nil", func() {
				result := periscope.Render(periscope.Config{
					Width: 5,
					Fill:  3,
				}, styles, periscope.Effect{
					Gradient: grad,
					State:    nil,
				})

				stripped := lab.StripANSI(result)
				Expect(stripped).To(Equal("◼◼◼◻◻"))
				Expect(result).NotTo(ContainSubstring("\x1b[38;2;"))
			})

			It("falls back to plain styles when fill is 0", func() {
				result := periscope.Render(periscope.Config{
					Width: 5,
					Fill:  0,
				}, styles, periscope.Effect{
					Gradient: grad,
					State:    state,
				})

				stripped := lab.StripANSI(result)
				Expect(stripped).To(Equal("◻◻◻◻◻"))
				Expect(result).NotTo(ContainSubstring("\x1b[38;2;"))
			})

			It("advances gradient state across successive renders", func() {
				cfg := periscope.Config{Width: 4, Fill: 2}
				effect := periscope.Effect{Gradient: grad, State: state}

				first := periscope.Render(cfg, styles, effect)
				offsetBefore := state.Offset
				state.Update(1)
				second := periscope.Render(cfg, styles, effect)

				Expect(state.Offset).To(Equal(offsetBefore + 1))
				Expect(lab.StripANSI(first)).To(Equal(lab.StripANSI(second)))
				Expect(first).NotTo(Equal(second))
			})
		})

		Context("static gradient (Animate=false)", func() {
			It("applies static gradient across entire bar width without GradientState", func() {
				staticGrad := &contract.ResolvedGradient{
					Steps:   4,
					Hi:      hi,
					Lo:      lo,
					Animate: false,
				}

				result := periscope.Render(periscope.Config{
					Width: 5,
					Fill:  3,
				}, styles, periscope.Effect{
					Gradient: staticGrad,
					State:    nil,
				})

				stripped := lab.StripANSI(result)
				Expect(stripped).To(Equal("◼◼◼◻◻"))
				ansiCount := strings.Count(result, "\x1b[38;2;")
				Expect(ansiCount).To(Equal(5))
			})

			It("renders empty bar with gradient when fill=0", func() {
				staticGrad := &contract.ResolvedGradient{
					Steps:   4,
					Hi:      hi,
					Lo:      lo,
					Animate: false,
				}

				result := periscope.Render(periscope.Config{
					Width: 5,
					Fill:  0,
				}, styles, periscope.Effect{
					Gradient: staticGrad,
					State:    nil,
				})

				stripped := lab.StripANSI(result)
				Expect(stripped).To(Equal("◻◻◻◻◻"))
				Expect(result).To(ContainSubstring("\x1b[38;2;"))
			})
		})
	})
})
