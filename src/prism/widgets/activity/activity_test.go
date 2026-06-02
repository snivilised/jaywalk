//go:build !race

package activity_test

import (
	"image/color"

	"charm.land/lipgloss/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
	"github.com/snivilised/jaywalk/src/prism/widgets/activity"
	lab "github.com/snivilised/jaywalk/test/laboratory"
)

var _ = Describe("Activity.Render", func() {
	Describe("basic rendering", func() {
		It("renders frame content with style", func() {
			out := activity.Render(activity.Config{
				Content: "⠁⠂⠄⡀",
			}, activity.Styles{},
				activity.Effect{},
			)
			Expect(out).NotTo(BeEmpty())
		})

		It("returns empty when content is empty", func() {
			out := activity.Render(activity.Config{}, activity.Styles{}, activity.Effect{})
			Expect(out).To(BeEmpty())
		})

		It("renders plain content (no gradient) within FrameStyle", func() {
			frameStyle := lipgloss.NewStyle().Bold(true)
			out := activity.Render(activity.Config{
				Content: "⠁⠂",
			}, activity.Styles{FrameStyle: frameStyle},
				activity.Effect{},
			)
			Expect(lab.StripANSI(out)).To(Equal("⠁⠂"))
		})
	})

	Describe("gradient application", func() {
		var (
			hi   color.Color
			lo   color.Color
			grad *contract.ResolvedGradient
			st   *effects.GradientState
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
			st = effects.NewGradientState()
			st.TotalSteps = 4
		})

		It("applies gradient to content without outer bars", func() {
			out := activity.Render(activity.Config{
				Content: "⠁⠂⠄⡀",
			}, activity.Styles{}, activity.Effect{
				Gradient: grad,
				State:    st,
			})

			Expect(out).NotTo(BeEmpty())
			Expect(lab.StripANSI(out)).To(Equal("⠁⠂⠄⡀"))
			Expect(out).To(ContainSubstring("\x1b[38;2;"))
		})

		It("interpolates colours from Hi towards Lo across the inner content", func() {
			out := activity.Render(activity.Config{
				Content: "⠁⠂⠄⡀",
			}, activity.Styles{}, activity.Effect{
				Gradient: grad,
				State:    st,
			})

			Expect(out).To(ContainSubstring("\x1b[38;2;255;0;0m"))
			Expect(out).To(ContainSubstring("\x1b[38;2;0;0;255m"))
		})

		It("strips outer ┃ bars and re-adds them with Hi/Lo colours", func() {
			out := activity.Render(activity.Config{
				Content: "┃⠁⠂⠄⡀┃",
			}, activity.Styles{}, activity.Effect{
				Gradient: grad,
				State:    st,
			})

			Expect(out).NotTo(BeEmpty())

			stripped := lab.StripANSI(out)
			Expect(stripped).To(HavePrefix("┃"))
			Expect(stripped).To(HaveSuffix("┃"))
			Expect(stripped).To(ContainSubstring("⠁⠂⠄⡀"))

			Expect(out).To(ContainSubstring("\x1b[38;2;255;0;0m"))
			Expect(out).To(ContainSubstring("\x1b[38;2;0;0;255m"))
		})

		It("falls back to plain rendering when Gradient is set but State is nil", func() {
			frameStyle := lipgloss.NewStyle()
			out := activity.Render(activity.Config{
				Content: "⠁⠂",
			}, activity.Styles{FrameStyle: frameStyle}, activity.Effect{
				Gradient: grad,
				State:    nil,
			})

			Expect(lab.StripANSI(out)).To(Equal("⠁⠂"))
			Expect(out).NotTo(ContainSubstring("\x1b[38;2;"))
		})

		It("falls back to plain rendering when State is set but Gradient is nil", func() {
			out := activity.Render(activity.Config{
				Content: "⠁⠂",
			}, activity.Styles{}, activity.Effect{
				Gradient: nil,
				State:    st,
			})

			Expect(lab.StripANSI(out)).To(Equal("⠁⠂"))
			Expect(out).NotTo(ContainSubstring("\x1b[38;2;"))
		})

		It("still returns empty when content is empty even with gradient set", func() {
			out := activity.Render(activity.Config{}, activity.Styles{}, activity.Effect{
				Gradient: grad,
				State:    st,
			})

			Expect(out).To(BeEmpty())
		})

		It("advances gradient state across successive renders", func() {
			cfg := activity.Config{Content: "⠁⠂⠄⡀"}
			effect := activity.Effect{Gradient: grad, State: st}

			first := activity.Render(cfg, activity.Styles{}, effect)
			offsetBefore := st.Offset
			st.Update(1)
			second := activity.Render(cfg, activity.Styles{}, effect)

			Expect(st.Offset).To(Equal(offsetBefore + 1))
			Expect(lab.StripANSI(first)).To(Equal(lab.StripANSI(second)))
			Expect(first).NotTo(Equal(second))
		})
	})
})
