package periscope_test

import (
	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/prism/widgets/periscope"
)

var _ = Describe("Periscope.Render", func() {
	It("renders all empty squares at fill 0", func() {
		out := ansi.Strip(periscope.Render(periscope.Config{
			Width: 5,
			Fill:  0,
		}))
		Expect(out).To(Equal("◻◻◻◻◻"))
	})

	It("renders all filled squares at fill == width", func() {
		out := ansi.Strip(periscope.Render(periscope.Config{
			Width: 5,
			Fill:  5,
		}))
		Expect(out).To(Equal("◼◼◼◼◼"))
	})

	It("renders partial fill correctly", func() {
		out := ansi.Strip(periscope.Render(periscope.Config{
			Width: 5,
			Fill:  3,
		}))
		Expect(out).To(Equal("◼◼◼◻◻"))
	})

	It("clamps fill at width", func() {
		out := ansi.Strip(periscope.Render(periscope.Config{
			Width: 5,
			Fill:  7,
		}))
		Expect(out).To(Equal("◼◼◼◼◼"))
	})

	It("handles width 0", func() {
		out := ansi.Strip(periscope.Render(periscope.Config{
			Width: 0,
			Fill:  0,
		}))
		Expect(out).To(BeEmpty())
	})

	It("wraps each segment in its own style", func() {
		out := periscope.Render(periscope.Config{
			Width: 3,
			Fill:  1,
		})
		stripped := ansi.Strip(out)
		Expect(stripped).To(Equal("◼◻◻"))
	})
})
