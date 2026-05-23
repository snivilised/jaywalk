package widget_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/prism/widget"
)

var _ = Describe("FormatDuration", func() {
	DescribeTable("magnitude-based formatting",
		func(duration time.Duration, expected string) {
			Expect(widget.FormatDuration(duration)).To(Equal(expected))
		},
		Entry("milliseconds", 123*time.Millisecond, "123ms"),
		Entry("seconds", 5*time.Second, "5s"),
		Entry("minutes and seconds", 1*time.Minute+2*time.Second, "1m2s"),
		Entry("full duration", 2*time.Hour+3*time.Minute+4*time.Second, "2h3m4s"),
		Entry("zero duration", time.Duration(0), "0ms"),
		Entry("sub-millisecond", 500*time.Microsecond, "0ms"),
		Entry("exactly one second", time.Second, "1s"),
		Entry("exactly one minute", time.Minute, "1m0s"),
		Entry("long duration", 25*time.Hour, "25h0m0s"),
	)
})
