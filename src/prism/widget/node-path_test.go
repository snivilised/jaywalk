package widget_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widget"
)

var _ = Describe("NodePath", Ordered, func() {
	var (
		styles widget.NodePathStyles
	)

	BeforeEach(func() {
		styles = widget.NodePathStyles{
			TreeIcons: contract.TreeIcons{
				contract.TreeIconFile:      "📁",
				contract.TreeIconDirectory: "📄",
			},
		}
	})

	Context("given file path", func() {
		It("returns icon and path text combined", func() {
			result := widget.RenderNodePath(widget.NodePathConfig{
				Path: "/tmp/foo.txt", MaxWidth: 100,
			}, styles)
			Expect(result).To(ContainSubstring("📁"))
			Expect(strings.Contains(result, "/tmp/foo.txt")).To(BeTrue())
		})
	})

	Context("given directory path with IsDir flag", func() {
		It("returns directory icon and path text combined", func() {
			result := widget.RenderNodePath(widget.NodePathConfig{
				Path: "/tmp", IsDir: true, MaxWidth: 100,
			}, styles)
			Expect(result).To(ContainSubstring("📄"))
			Expect(strings.Contains(result, "/tmp/")).To(BeTrue())
		})

		It("handles nested directory paths", func() {
			result := widget.RenderNodePath(widget.NodePathConfig{
				Path: "/var/log", IsDir: true, MaxWidth: 100,
			}, styles)
			Expect(result).To(ContainSubstring("📄 /var/log/"))
		})
	})

	Context("given label when path is empty", func() {
		It("returns label without icon", func() {
			result := widget.RenderNodePath(widget.NodePathConfig{
				Label: "idle worker", MaxWidth: 100,
			}, styles)
			Expect(result).To(Equal("idle worker"))
		})

		It("handles empty label returns empty string", func() {
			result := widget.RenderNodePath(widget.NodePathConfig{
				Label: "", MaxWidth: 100,
			}, styles)
			Expect(result).To(Equal(""))
		})
	})

	Context("given long paths that need truncation", func() {
		It("prefixes with ellipsis and respects max width", func() {
			result := widget.RenderNodePath(widget.NodePathConfig{
				Path: "/very/long/path/to/some/file.txt", MaxWidth: 20,
			}, styles)
			Expect(result).To(ContainSubstring(contract.Ellipses))
		})
	})

	Context("given short paths that fit within max width", func() {
		It("returns full path without truncation", func() {
			result := widget.RenderNodePath(widget.NodePathConfig{
				Path: "/tmp/foo.txt", MaxWidth: 100,
			}, styles)
			Expect(result).To(ContainSubstring("/tmp/foo.txt"))
			Expect(strings.Contains(result, contract.Ellipses)).To(BeFalse())
		})
	})

	Context("given paths with special characters", func() {
		It("handles paths with spaces in directory names", func() {
			result := widget.RenderNodePath(widget.NodePathConfig{
				Path: "/opt/my app/config.ini", MaxWidth: 100,
			}, styles)
			Expect(result).To(ContainSubstring("/opt/my"))
			Expect(result).To(ContainSubstring("app/"))
		})
	})

	Context("empty path and IsDir together", func() {
		It("returns only label with directory icon if both set", func() {
			result := widget.RenderNodePath(widget.NodePathConfig{
				Path: "/var", IsDir: true, MaxWidth: 100,
			}, styles)
			Expect(result).To(ContainSubstring("📄 /var/"))
		})
	})

	Context("path without IsDir flag is treated as file", func() {
		It("returns file icon for paths that are not marked IsDir", func() {
			result := widget.RenderNodePath(widget.NodePathConfig{
				Path: "/etc/passwd", MaxWidth: 100,
			}, styles)
			Expect(result).To(ContainSubstring("📁 /etc/passwd"))
		})
	})

	It("handles root slash path as directory", func() {
		result := widget.RenderNodePath(widget.NodePathConfig{
			Path: "/", IsDir: true, MaxWidth: 100,
		}, styles)
		Expect(result).To(ContainSubstring("📄 /"))
	})
})
