package workshop_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/prism/workshop"
)

var _ = Describe("BandsVisualiser", func() {
	var v *workshop.BandsVisualiser

	BeforeEach(func() {
		v = &workshop.BandsVisualiser{Width: 40}
	})

	It("returns the name 'bands'", func() {
		Expect(v.Name()).To(Equal("bands"))
	})

	DescribeTable("renders gradient steps",
		func(entry stepsEntry) {
			result := v.Render(
				makeSteps(entry.Count),
				entry.Curve,
				entry.Easing,
				entry.AnimFrame,
			)
			Expect(result).ToNot(BeEmpty())
			Expect(result).To(ContainSubstring("\x1b[38;2;"))
			Expect(result).To(ContainSubstring("\n"))
		},
		Entry("2 steps, linear, uniform, frame 0",
			stepsEntry{2, enums.CurveKindLinear,
				enums.EasingKindUniform, 0}),
		Entry("8 steps, sine, ease-in, frame 10",
			stepsEntry{8, enums.CurveKindSine,
				enums.EasingKindEaseIn, 10}),
		Entry("64 steps, cubic, ease-out, frame 100",
			stepsEntry{64, enums.CurveKindCubic,
				enums.EasingKindEaseOut, 100}),
		Entry("256 steps, q-in, ease-in-out, frame 0",
			stepsEntry{256, enums.CurveKindQuadraticIn,
				enums.EasingKindEaseInOut, 0}),
		Entry("8 steps, q-out, uniform, frame 50",
			stepsEntry{8, enums.CurveKindQuadraticOut,
				enums.EasingKindUniform, 50}),
	)

	DescribeTable("contains braille and block chars",
		func(entry stepsEntry) {
			result := v.Render(
				makeSteps(entry.Count),
				entry.Curve,
				entry.Easing,
				entry.AnimFrame,
			)

			Expect(result).To(Satisfy(
				func(s string) bool {
					return containsAny(s, []string{
						"⣀", "⡠", "⠤", "⠊", "⠈", "█",
					})
				},
			))
		},
		Entry("8 steps, linear",
			stepsEntry{8, enums.CurveKindLinear,
				enums.EasingKindUniform, 0}),
	)

	It("returns empty for zero steps", func() {
		result := v.Render(
			makeSteps(0),
			enums.CurveKindLinear,
			enums.EasingKindUniform,
			0,
		)
		Expect(result).To(BeEmpty())
	})
})
