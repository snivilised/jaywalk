package widget_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/widget"
)

var _ = Describe("Activity", func() {
	It("renders frame content with style", func() {
		out := widget.RenderActivity(widget.ActivityConfig{
			Content: "⠁⠂⠄⡀",
		}, widget.ActivityStyles{})
		Expect(out).NotTo(BeEmpty())
	})

	It("returns empty when content is empty", func() {
		out := widget.RenderActivity(widget.ActivityConfig{}, widget.ActivityStyles{})
		Expect(out).To(BeEmpty())
	})
})
