package traffic

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SpinnerFrames", func() {
	DescribeTable("film strip frames",
		func(tick int, expected string) {
			Expect(filmStripFrame(tick)).To(Equal(expected))
		},
		Entry("tick 0", 0, "┃▓░░░░░░┃"),
		Entry("tick 1", 1, "┃▓▓░░░░░┃"),
		Entry("tick 6", 6, "┃▓▓▓▓▓▓▓┃"),
		Entry("tick 7", 7, "┃▓▓▓▓▓▓▓┃"),
		Entry("tick 8", 8, "┃▓▓▓▓▓▓░┃"),
		Entry("tick 14", 14, "┃▓░░░░░░┃"),
	)

	DescribeTable("pulse frames",
		func(tick int, contains string) {
			frame := pulseFrame(tick)
			Expect(frame).To(HavePrefix(contains))
		},
		Entry("tick 0 starts with 1 full", 0, "█"),
		Entry("tick 4 starts with 5 full", 4, "█████"),
		Entry("tick 7 starts with 8 full", 7, "████████"),
		Entry("tick 8 starts with 7 full", 8, "███████"),
		Entry("tick 15 starts with empty", 15, "█"),
	)

	DescribeTable("spinner frames",
		func(tick int, expected string) {
			Expect(spinnerFrame(tick)).To(Equal(expected))
		},
		Entry("tick 0", 0, "|"),
		Entry("tick 1", 1, "/"),
		Entry("tick 2", 2, "-"),
		Entry("tick 3", 3, "\\"),
		Entry("tick 4 wraps", 4, "|"),
	)
})
