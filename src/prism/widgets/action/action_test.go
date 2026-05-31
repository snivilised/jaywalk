package action_test

import (
	"errors"

	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/prism/widgets/action"
)

var _ = Describe("Action.Render", func() {
	It("renders error when present", func() {
		out := ansi.Strip(action.Render(action.Config{
			Error: errors.New("permission denied"),
		}, action.Styles{}))
		Expect(out).To(Equal(" ! permission denied"))
	})

	It("renders action name when no error", func() {
		out := ansi.Strip(action.Render(action.Config{
			ActionName: "copy",
		}, action.Styles{}))
		Expect(out).To(Equal(" • via copy"))
	})

	It("renders pipeline name when no error or action", func() {
		out := ansi.Strip(action.Render(action.Config{
			PipelineName: "build",
		}, action.Styles{}))
		Expect(out).To(Equal(" • via build"))
	})

	It("returns empty when no fields set", func() {
		out := action.Render(action.Config{}, action.Styles{})
		Expect(out).To(BeEmpty())
	})

	It("prioritises error over action name", func() {
		out := ansi.Strip(action.Render(action.Config{
			Error:      errors.New("fail"),
			ActionName: "copy",
		}, action.Styles{}))
		Expect(out).To(Equal(" ! fail"))
	})
})
