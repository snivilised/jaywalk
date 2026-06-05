package landing_test

import (
	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/prism/widgets/landing"
)

var _ = Describe("LandingStrip", func() {
	It("renders command output in brackets", func() {
		out := ansi.Strip(landing.Render(landing.Config{
			CommandOutput: "rm -rf /tmp",
		}, landing.Styles{}))
		Expect(out).To(Equal(" [rm -rf /tmp]"))
	})

	It("renders execution string when dry-run", func() {
		out := ansi.Strip(landing.Render(landing.Config{
			ExecutionString: "mv a b",
			DryRun:          true,
		}, landing.Styles{}))
		Expect(out).To(Equal(" [mv a b]"))
	})

	It("prefers command output over execution string in normal mode", func() {
		out := ansi.Strip(landing.Render(landing.Config{
			CommandOutput:   "real output",
			ExecutionString: "would have run",
			DryRun:          false,
		}, landing.Styles{}))
		Expect(out).To(Equal(" [real output]"))
	})

	It("returns empty when both content strings are empty", func() {
		out := landing.Render(landing.Config{}, landing.Styles{})
		Expect(out).To(BeEmpty())
	})

	It("renders skipped icon when provided", func() {
		out := ansi.Strip(landing.Render(landing.Config{
			SkippedIcon: "⛔️",
		}, landing.Styles{}))
		Expect(out).To(Equal(" [ ⛔️]"))
	})

	It("renders skipped icon alongside content", func() {
		out := ansi.Strip(landing.Render(landing.Config{
			CommandOutput: "already done",
			SkippedIcon:   "⛔️",
		}, landing.Styles{}))
		Expect(out).To(Equal(" [ ⛔️already done]"))
	})

	Context("right-justification via Width", func() {
		It("does not pad when Width is zero", func() {
			out := ansi.Strip(landing.Render(landing.Config{
				CommandOutput: "x",
				Width:         0,
			}, landing.Styles{}))
			Expect(out).To(Equal(" [x]"))
		})

		It("pads with leading spaces to align to Width", func() {
			out := ansi.Strip(landing.Render(landing.Config{
				CommandOutput: "x",
				Width:         10,
			}, landing.Styles{}))
			// strip " [x]" is 4 chars; pad to 10 = 6 leading spaces.
			// Confirm the strip is at the right edge by checking
			// the trailing chars and total visible length.
			Expect(out).To(HaveLen(10))
			Expect(out).To(ContainSubstring(" [x]"))
			Expect(out[len(out)-4:]).To(Equal(" [x]"))
		})

		It("does not over-pad when strip already at or beyond Width", func() {
			out := ansi.Strip(landing.Render(landing.Config{
				CommandOutput: "abcdefghij",
				Width:         5,
			}, landing.Styles{}))
			// strip " [abcdefghij]" is 13 chars; width is 5, no padding
			Expect(out).To(Equal(" [abcdefghij]"))
		})

		It("Width=0 with command output renders inline (no padding)", func() {
			out := ansi.Strip(landing.Render(landing.Config{
				CommandOutput: "y",
				Width:         -1,
			}, landing.Styles{}))
			Expect(out).To(Equal(" [y]"))
		})
	})
})
