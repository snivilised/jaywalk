package widget_test

import (
	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/prism/widget"
)

var _ = Describe("Periscope", func() {
	It("renders all empty squares at fill 0", func() {
		out := ansi.Strip(widget.RenderPeriscope(widget.PeriscopeConfig{
			Width: 5,
			Fill:  0,
		}))
		Expect(out).To(Equal("◻◻◻◻◻"))
	})

	It("renders all filled squares at fill == width", func() {
		out := ansi.Strip(widget.RenderPeriscope(widget.PeriscopeConfig{
			Width: 5,
			Fill:  5,
		}))
		Expect(out).To(Equal("◼◼◼◼◼"))
	})

	It("renders partial fill correctly", func() {
		out := ansi.Strip(widget.RenderPeriscope(widget.PeriscopeConfig{
			Width: 5,
			Fill:  3,
		}))
		Expect(out).To(Equal("◼◼◼◻◻"))
	})

	It("clamps fill at width", func() {
		out := ansi.Strip(widget.RenderPeriscope(widget.PeriscopeConfig{
			Width: 5,
			Fill:  7,
		}))
		Expect(out).To(Equal("◼◼◼◼◼"))
	})

	It("handles width 0", func() {
		out := ansi.Strip(widget.RenderPeriscope(widget.PeriscopeConfig{
			Width: 0,
			Fill:  0,
		}))
		Expect(out).To(BeEmpty())
	})

	It("wraps each segment in its own style", func() {
		out := widget.RenderPeriscope(widget.PeriscopeConfig{
			Width: 3,
			Fill:  1,
		})
		stripped := ansi.Strip(out)
		Expect(stripped).To(Equal("◼◻◻"))
	})
})
