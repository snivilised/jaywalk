package widget_test

import (
	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/widget"
)

var _ = Describe("LandingStrip", func() {
	It("renders command output in brackets", func() {
		out := ansi.Strip(widget.RenderLandingStrip(widget.LandingStripConfig{
			CommandOutput: "rm -rf /tmp",
		}, widget.LandingStripStyles{}))
		Expect(out).To(Equal(" [rm -rf /tmp]"))
	})

	It("renders execution string when dry-run", func() {
		out := ansi.Strip(widget.RenderLandingStrip(widget.LandingStripConfig{
			ExecutionString: "mv a b",
			DryRun:          true,
		}, widget.LandingStripStyles{}))
		Expect(out).To(Equal(" [mv a b]"))
	})

	It("prefers command output over execution string in normal mode", func() {
		out := ansi.Strip(widget.RenderLandingStrip(widget.LandingStripConfig{
			CommandOutput:   "real output",
			ExecutionString: "would have run",
			DryRun:          false,
		}, widget.LandingStripStyles{}))
		Expect(out).To(Equal(" [real output]"))
	})

	It("returns empty when both content strings are empty", func() {
		out := widget.RenderLandingStrip(widget.LandingStripConfig{}, widget.LandingStripStyles{})
		Expect(out).To(BeEmpty())
	})

	It("renders skipped icon when provided", func() {
		out := ansi.Strip(widget.RenderLandingStrip(widget.LandingStripConfig{
			SkippedIcon: "⛔️",
		}, widget.LandingStripStyles{}))
		Expect(out).To(Equal(" [ ⛔️]"))
	})

	It("renders skipped icon alongside content", func() {
		out := ansi.Strip(widget.RenderLandingStrip(widget.LandingStripConfig{
			CommandOutput: "already done",
			SkippedIcon:   "⛔️",
		}, widget.LandingStripStyles{}))
		Expect(out).To(Equal(" [ ⛔️already done]"))
	})
})
