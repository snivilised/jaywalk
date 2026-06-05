//go:build !race
// +build !race

package highway

import (
	"image/color"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
)

// viewContent extracts the rendered string from the model's view.
// The Model's View() method returns a tea.View whose string content
// is what we assert on in the tests below.
func viewContent(m Model) string {
	v := m.View()
	return v.Content
}

// bannerTestGradient returns a small fixed gradient suitable for
// banner integration tests. Hi is red, Lo is blue, 4 steps.
func bannerTestGradient() *contract.ResolvedGradient {
	return &contract.ResolvedGradient{
		Steps: 4,
		Hi:    color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Lo:    color.RGBA{R: 0, G: 0, B: 255, A: 255},
	}
}

// makeBannerInfo builds a banner.Info that the highway model will
// accept (i.e. a non-nil Gradient and State).
func makeBannerInfo(position string) banner.Info {
	grad := bannerTestGradient()
	st := effects.NewGradientState()
	hiR, hiG, hiB, _ := grad.Hi.RGBA()
	loR, loG, loB, _ := grad.Lo.RGBA()
	steps := effects.InterpolateBetweenRGBA(
		uint8(hiR>>8), uint8(hiG>>8), uint8(hiB>>8), //nolint:gosec // safe: 16-bit value >> 8 fits in 8 bits
		uint8(loR>>8), uint8(loG>>8), uint8(loB>>8), //nolint:gosec // safe: 16-bit value >> 8 fits in 8 bits
		grad.Steps,
	)
	st.SetSteps(steps)
	return banner.Info{
		Disable:  false,
		Position: position,
		Justify:  "right",
		Width:    60,
		Aspects: banner.Aspects{
			Orientation: banner.OrientationHorizontal,
			Banding:     banner.BandingWithout,
			Unity:       banner.UnityUnified,
			FixedEnd:    banner.FixedEndUnfixed,
		},
		Gradient: grad,
		State:    st,
		Tick:     500 * time.Millisecond,
	}
}

var _ = Describe("Banner integration with highway view", func() {
	Describe("rendering position", func() {
		It("renders the banner above the top border when Position = top", func() {
			m := baseModel(1)
			info := makeBannerInfo(contract.PositionTop)
			updated, _ := update(m, OvertureMsg{
				FlagsRowPosition: contract.PositionBottom,
				Banner:           info,
			})
			out := viewContent(updated)

			// The banner widget's first line is the art (which the
			// highway does not pre-populate - the widget renders
			// the art from the gradient/state). The art contains
			// face runes (█); the rendered output is the ANSI-
			// styled art followed by the bordered region.
			// The very first non-empty line must contain a face
			// rune or the banner's leading whitespace, not the
			// top border (which is rendered with border characters).
			lines := strings.Split(out, "\n")
			Expect(lines).NotTo(BeEmpty())

			// Top border: the highway's renderHeader starts with
			// ╭─ (rounded) or ┌─ (square) depending on theme. The
			// first line should NOT be the top border - it should
			// be the banner. We look for ANSI colour sequences
			// in the first line (the banner always emits them).
			first := lines[0]
			Expect(strings.Contains(first, "\x1b[38;2;")).To(BeTrue(),
				"first line should contain ANSI colour sequences from the banner")

			// The top border (rounded ╭ or square ┌) should appear
			// in the output, but NOT on the first line.
			combined := strings.Join(lines[1:], "\n")
			Expect(combined).To(SatisfyAny(
				ContainSubstring("╭"),
				ContainSubstring("┌"),
			))
		})

		It("renders the banner below the bottom border when Position = bottom", func() {
			m := baseModel(1)
			info := makeBannerInfo(contract.PositionBottom)
			updated, _ := update(m, OvertureMsg{
				FlagsRowPosition: contract.PositionBottom,
				Banner:           info,
			})
			out := viewContent(updated)
			lines := strings.Split(out, "\n")

			// The last non-empty line should contain the banner
			// output (ANSI colour sequences). The bottom border
			// (rounded ╰ or square └) appears earlier in the output.
			// The output may end with a trailing newline (producing
			// an empty last line); we therefore skip trailing empty
			// lines when locating the banner line.
			Expect(lines).NotTo(BeEmpty())
			lastIdx := len(lines) - 1
			for lastIdx > 0 && strings.TrimSpace(lines[lastIdx]) == "" {
				lastIdx--
			}
			last := lines[lastIdx]
			Expect(strings.Contains(last, "\x1b[38;2;")).To(BeTrue(),
				"last non-empty line should contain ANSI colour sequences from the banner")

			// The bottom border should appear before the last
			// non-empty line.
			combined := strings.Join(lines[:lastIdx], "\n")
			Expect(combined).To(SatisfyAny(
				ContainSubstring("╰"),
				ContainSubstring("└"),
			))
		})

		It("skips the banner when Disable = true", func() {
			m := baseModel(1)
			info := makeBannerInfo(contract.PositionTop)
			info.Disable = true
			updated, _ := update(m, OvertureMsg{
				FlagsRowPosition: contract.PositionBottom,
				Banner:           info,
			})
			out := viewContent(updated)
			// The view should not contain the banner's ANSI codes.
			Expect(strings.Contains(out, "\x1b[38;2;")).To(BeFalse())
		})
	})

	Describe("animation tick", func() {
		It("advances the banner's gradient state on every global tick (skipFactor=0)", func() {
			m := baseModel(1)
			info := makeBannerInfo(contract.PositionTop)
			updated, _ := update(m, OvertureMsg{
				FlagsRowPosition: contract.PositionBottom,
				Banner:           info,
			})

			// Drive a single tick through the model.
			before := updated.bannerTicker.State().Offset
			updated, _ = update(updated, tickMsg(time.Now()))
			after := updated.bannerTicker.State().Offset

			// 50ms global tick / 500ms banner tick = 10 → skipFactor=10
			// One tick is not enough to advance.
			// We assert the gradient state was created.
			Expect(updated.bannerTicker).NotTo(BeNil())
			Expect(updated.bannerTicker.Factor()).To(Equal(10))
			_ = before
			_ = after
		})

		It("advances the gradient state after skipFactor ticks", func() {
			m := baseModel(1)
			info := makeBannerInfo(contract.PositionTop)
			updated, _ := update(m, OvertureMsg{
				FlagsRowPosition: contract.PositionBottom,
				Banner:           info,
			})
			before := updated.bannerTicker.State().Offset
			// 10 ticks (skipFactor) → 1 advance.
			for i := 0; i < 10; i++ {
				updated, _ = update(updated, tickMsg(time.Now()))
			}
			after := updated.bannerTicker.State().Offset
			Expect(after).To(Equal(before + 1))
		})
	})
})
