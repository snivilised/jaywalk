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
})
