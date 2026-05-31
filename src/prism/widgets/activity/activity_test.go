package activity_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/widgets/activity"
)

var _ = Describe("Activity.Render", func() {
	It("renders frame content with style", func() {
		out := activity.Render(activity.Config{
			Content: "⠁⠂⠄⡀",
		}, activity.Styles{})
		Expect(out).NotTo(BeEmpty())
	})

	It("returns empty when content is empty", func() {
		out := activity.Render(activity.Config{}, activity.Styles{})
		Expect(out).To(BeEmpty())
	})
})
