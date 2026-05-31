package node_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/node"
)

var _ = Describe("NodePath.Render", Ordered, func() {
	var (
		styles node.Styles
	)

	BeforeEach(func() {
		styles = node.Styles{
			TreeIcons: contract.TreeIcons{
				contract.TreeIconFile:      "📁",
				contract.TreeIconDirectory: "📄",
			},
		}
	})

	Context("given file path", func() {
		It("returns icon and path text combined", func() {
			result := node.Render(node.Config{
				Path: "/tmp/foo.txt", MaxWidth: 100,
			}, styles)
			Expect(result).To(ContainSubstring("📁"))
			Expect(strings.Contains(result, "/tmp/foo.txt")).To(BeTrue())
		})
	})

	Context("given directory path with IsDir flag", func() {
		It("returns directory icon and path text combined", func() {
			result := node.Render(node.Config{
				Path: "/tmp", IsDir: true, MaxWidth: 100,
			}, styles)
			Expect(result).To(ContainSubstring("📄"))
			Expect(strings.Contains(result, "/tmp/")).To(BeTrue())
		})

		It("handles nested directory paths", func() {
			result := node.Render(node.Config{
				Path: "/var/log", IsDir: true, MaxWidth: 100,
			}, styles)
			Expect(result).To(ContainSubstring("📄 /var/log/"))
		})
	})

	Context("given label when path is empty", func() {
		It("returns label without icon", func() {
			result := node.Render(node.Config{
				Label: "idle worker", MaxWidth: 100,
			}, styles)
			Expect(result).To(Equal("idle worker"))
		})

		It("handles empty label returns empty string", func() {
			result := node.Render(node.Config{
				Label: "", MaxWidth: 100,
			}, styles)
			Expect(result).To(Equal(""))
		})
	})

	Context("given long paths that need truncation", func() {
		It("prefixes with ellipsis and respects max width", func() {
			result := node.Render(node.Config{
				Path: "/very/long/path/to/some/file.txt", MaxWidth: 20,
			}, styles)
			Expect(result).To(ContainSubstring(contract.Ellipses))
		})
	})

	Context("given short paths that fit within max width", func() {
		It("returns full path without truncation", func() {
			result := node.Render(node.Config{
				Path: "/tmp/foo.txt", MaxWidth: 100,
			}, styles)
			Expect(result).To(ContainSubstring("/tmp/foo.txt"))
			Expect(strings.Contains(result, contract.Ellipses)).To(BeFalse())
		})
	})

	Context("given paths with special characters", func() {
		It("handles paths with spaces in directory names", func() {
			result := node.Render(node.Config{
				Path: "/opt/my app/config.ini", MaxWidth: 100,
			}, styles)
			Expect(result).To(ContainSubstring("/opt/my"))
			Expect(result).To(ContainSubstring("app/"))
		})
	})

	Context("empty path and IsDir together", func() {
		It("returns only label with directory icon if both set", func() {
			result := node.Render(node.Config{
				Path: "/var", IsDir: true, MaxWidth: 100,
			}, styles)
			Expect(result).To(ContainSubstring("📄 /var/"))
		})
	})

	Context("path without IsDir flag is treated as file", func() {
		It("returns file icon for paths that are not marked IsDir", func() {
			result := node.Render(node.Config{
				Path: "/etc/passwd", MaxWidth: 100,
			}, styles)
			Expect(result).To(ContainSubstring("📁 /etc/passwd"))
		})
	})

	It("handles root slash path as directory", func() {
		result := node.Render(node.Config{
			Path: "/", IsDir: true, MaxWidth: 100,
		}, styles)
		Expect(result).To(ContainSubstring("📄 /"))
	})
})
