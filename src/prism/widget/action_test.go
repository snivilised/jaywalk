package widget_test

import (
	"errors"

	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/widget"
)

var _ = Describe("Action", func() {
	It("renders error when present", func() {
		out := ansi.Strip(widget.RenderAction(widget.ActionConfig{
			Error: errors.New("permission denied"),
		}, widget.ActionStyles{}))
		Expect(out).To(Equal(" ! permission denied"))
	})

	It("renders action name when no error", func() {
		out := ansi.Strip(widget.RenderAction(widget.ActionConfig{
			ActionName: "copy",
		}, widget.ActionStyles{}))
		Expect(out).To(Equal(" • via copy"))
	})

	It("renders pipeline name when no error or action", func() {
		out := ansi.Strip(widget.RenderAction(widget.ActionConfig{
			PipelineName: "build",
		}, widget.ActionStyles{}))
		Expect(out).To(Equal(" • via build"))
	})

	It("returns empty when no fields set", func() {
		out := widget.RenderAction(widget.ActionConfig{}, widget.ActionStyles{})
		Expect(out).To(BeEmpty())
	})

	It("prioritises error over action name", func() {
		out := ansi.Strip(widget.RenderAction(widget.ActionConfig{
			Error:      errors.New("fail"),
			ActionName: "copy",
		}, widget.ActionStyles{}))
		Expect(out).To(Equal(" ! fail"))
	})
})
