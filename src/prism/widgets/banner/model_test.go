package banner_test

import (
	"image/color"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
)

// makeModelInfo builds a populated Info suitable for unit tests.
// Disable is false, Gradient and State are non-nil, and the aspects
// pick a fixed (non-random) value so tests are deterministic.
func makeModelInfo(width int) banner.Info {
	grad := &contract.ResolvedGradient{
		Steps: 4,
		Hi:    color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Lo:    color.RGBA{R: 0, G: 0, B: 255, A: 255},
	}
	st := effects.NewGradientState()
	st.TotalSteps = grad.Steps
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
		Position: contract.PositionTop,
		Justify:  banner.JustifyRight,
		Width:    width,
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

var _ = Describe("NewModel", func() {
	It("captures info from WithInfo", func() {
		info := makeModelInfo(60)
		m := banner.NewModel(banner.WithInfo(info))
		// Disabled checks Gradient and State are wired up.
		Expect(m.Disabled()).To(BeFalse())
	})

	It("uses WithWidth override when set", func() {
		info := makeModelInfo(40)
		m := banner.NewModel(
			banner.WithInfo(info),
			banner.WithWidth(120),
		)
		// The rendered output's padding is derived from width; the
		// simplest way to assert the override took effect is via
		// a regression comparison with a direct Render call below.
		out := m.View()
		Expect(out).NotTo(BeEmpty())
		// A width-120 render should differ from a width-40 render.
		// (We assert only that the model is not stuck on info.Width.)
		Expect(m).NotTo(Equal(banner.NewModel(banner.WithInfo(info))))
	})

	It("falls back to info.Width when WithWidth not supplied", func() {
		info := makeModelInfo(40)
		m1 := banner.NewModel(banner.WithInfo(info))
		m2 := banner.NewModel(banner.WithInfo(info), banner.WithWidth(40))
		// Both should produce the same output.
		Expect(m1.View()).To(Equal(m2.View()))
	})
})

var _ = Describe("Disabled", func() {
	It("returns true when info.Disable", func() {
		info := makeModelInfo(60)
		info.Disable = true
		m := banner.NewModel(banner.WithInfo(info))
		Expect(m.Disabled()).To(BeTrue())
	})

	It("returns true when info.Gradient is nil", func() {
		info := makeModelInfo(60)
		info.Gradient = nil
		m := banner.NewModel(banner.WithInfo(info))
		Expect(m.Disabled()).To(BeTrue())
	})

	It("returns true when info.State is nil", func() {
		info := makeModelInfo(60)
		info.State = nil
		m := banner.NewModel(banner.WithInfo(info))
		Expect(m.Disabled()).To(BeTrue())
	})

	It("returns false when all three are set", func() {
		info := makeModelInfo(60)
		m := banner.NewModel(banner.WithInfo(info))
		Expect(m.Disabled()).To(BeFalse())
	})
})

var _ = Describe("View", func() {
	It("returns '' when Disabled", func() {
		info := makeModelInfo(60)
		info.Disable = true
		m := banner.NewModel(banner.WithInfo(info))
		Expect(m.View()).To(BeEmpty())
	})

	It("calls Render with the captured width, justify, gradient, state, and aspects", func() {
		info := makeModelInfo(60)
		m := banner.NewModel(banner.WithInfo(info), banner.WithWidth(60))
		out := m.View()
		Expect(out).NotTo(BeEmpty())
		// Render always emits ANSI codes when gradient/state are set.
		Expect(out).To(ContainSubstring("\x1b[38;2;"))
	})

	It("produces the same output as a direct Render call", func() {
		// Regression: NewModel must be a transparent wrapper around
		// the existing Render function. Construct both with the same
		// inputs and assert byte equality.
		info := makeModelInfo(60)
		m := banner.NewModel(banner.WithInfo(info), banner.WithWidth(60))

		directOut := banner.Render(
			banner.Config{Width: 60, Justify: info.Justify},
			banner.Styles{},
			banner.Effect{
				Gradient: info.Gradient,
				State:    info.State,
				Aspects:  info.Aspects,
			},
		)

		Expect(m.View()).To(Equal(directOut))
	})
})
