package widget_test

import (
	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widget"
)

var _ = Describe("NodePath", func() {
	var (
		styles widget.NodePathStyles
	)

	BeforeEach(func() {
		styles = widget.NodePathStyles{
			TreeIcons: contract.TreeIcons{
				contract.TreeIconFile:      "📄",
				contract.TreeIconDirectory: "📁",
			},
		}
	})

	It("returns file icon for file paths", func() {
		icon, styled := widget.RenderNodePath(widget.NodePathConfig{
			Path: "/tmp/foo.txt", MaxWidth: 100,
		}, styles)
		Expect(icon).To(Equal("📄"))
		Expect(ansi.Strip(styled)).To(Equal("/tmp/foo.txt"))
	})

	It("returns directory icon for dir paths", func() {
		icon, styled := widget.RenderNodePath(widget.NodePathConfig{
			Path: "/tmp", IsDir: true, MaxWidth: 100,
		}, styles)
		Expect(icon).To(Equal("📁"))
		Expect(ansi.Strip(styled)).To(Equal("/tmp/"))
	})

	It("uses label when path is empty", func() {
		icon, styled := widget.RenderNodePath(widget.NodePathConfig{
			Label: "idle worker", MaxWidth: 100,
		}, styles)
		Expect(icon).To(BeEmpty())
		Expect(ansi.Strip(styled)).To(Equal("idle worker"))
	})

	It("truncates long paths with ellipsis prefix", func() {
		_, styled := widget.RenderNodePath(widget.NodePathConfig{
			Path: "/very/long/path/to/some/file.txt", MaxWidth: 20,
		}, styles)
		trimmed := ansi.Strip(styled)
		Expect(trimmed).To(HavePrefix("..."))
		Expect(len([]rune(trimmed))).To(BeNumerically("<=", 20))
	})

	It("does not truncate when path fits within max width", func() {
		_, styled := widget.RenderNodePath(widget.NodePathConfig{
			Path: "/tmp", MaxWidth: 100,
		}, styles)
		Expect(ansi.Strip(styled)).To(Equal("/tmp"))
	})

	It("returns empty icon when path is empty", func() {
		icon, _ := widget.RenderNodePath(widget.NodePathConfig{
			Label: "idle", MaxWidth: 100,
		}, styles)
		Expect(icon).To(BeEmpty())
	})
})
