package contract

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/agenor/enums"
)

var _ = Describe("easedT", func() {
	DescribeTable("boundary values",
		func(curve enums.CurveKind, easing enums.EasingKind, tIn, expected float64) {
			result := easedT(tIn, curve, easing)
			Expect(result).To(BeNumerically("~", expected, 1e-10))
		},
		// t=0 must always return 0 regardless of curve/easing
		Entry("linear/uniform t=0", enums.CurveKindLinear, enums.EasingKindUniform, 0.0, 0.0),
		Entry("sine/ease-in t=0", enums.CurveKindSine, enums.EasingKindEaseIn, 0.0, 0.0),
		Entry("cubic/ease-in-out t=0", enums.CurveKindCubic, enums.EasingKindEaseInOut, 0.0, 0.0),

		// t=1 must always return 1 regardless of curve/easing
		Entry("linear/uniform t=1", enums.CurveKindLinear, enums.EasingKindUniform, 1.0, 1.0),
		Entry("sine/ease-out t=1", enums.CurveKindSine, enums.EasingKindEaseOut, 1.0, 1.0),
		Entry("cubic/ease-in-out t=1", enums.CurveKindCubic, enums.EasingKindEaseInOut, 1.0, 1.0),

		// t=0.5 midpoints (shape-dependent, not exact 0.5)
		Entry("linear/uniform t=0.5", enums.CurveKindLinear, enums.EasingKindUniform, 0.5, 0.5),
		Entry("sine t=0.5", enums.CurveKindSine, enums.EasingKindUniform, 0.5, 0.5),
		Entry("quadratic-in t=0.5", enums.CurveKindQuadraticIn, enums.EasingKindUniform, 0.5, 0.25),
		Entry("quadratic-out t=0.5", enums.CurveKindQuadraticOut, enums.EasingKindUniform, 0.5, 0.75),
		Entry("cubic t=0.5", enums.CurveKindCubic, enums.EasingKindUniform, 0.5, 0.5),

		// zero values are passthrough
		Entry("zero values t=0.3", enums.CurveKind(0), enums.EasingKind(0), 0.3, 0.3),
	)

	DescribeTable("linear passthrough",
		func(tIn float64) {
			result := easedT(tIn, enums.CurveKindLinear, enums.EasingKindUniform)
			Expect(result).To(BeNumerically("~", tIn, 1e-10))
		},
		Entry("t=0", 0.0),
		Entry("t=0.25", 0.25),
		Entry("t=0.5", 0.5),
		Entry("t=0.75", 0.75),
		Entry("t=1", 1.0),
	)
})
