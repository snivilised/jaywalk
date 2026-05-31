package widget

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

var _ = Describe("RootPath", func() {
	It("returns current directory for empty path", func() {
		styles := RootPathStyles{PathStyle: lipgloss.NewStyle()}
		expect := "."
		result := RootPath("", 10, styles)
		Expect(result).To(Equal(expect))
	})

	It("returns the path unchanged if within max width", func() {
		styles := RootPathStyles{PathStyle: lipgloss.NewStyle()}
		path := "/some/path"
		maxWidth := 20
		result := RootPath(path, maxWidth, styles)
		Expect(result).To(Equal(path))
	})

	It("truncates long paths with ellipsis", func() {
		styles := RootPathStyles{PathStyle: lipgloss.NewStyle()}
		path := "/very/long/path/to/a/deeply/nested/directory"
		maxWidth := 20
		result := RootPath(path, maxWidth, styles)
		// Calculate expected based on the actual logic
		pathWidth := lipgloss.Width(path)
		keep := max(0, maxWidth-3)
		expected := contract.Ellipses + path[pathWidth-keep:]
		Expect(result).To(Equal(expected))
	})

	It("applies PathStyle", func() {
		pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
		styles := RootPathStyles{PathStyle: pathStyle}
		path := "/some/path"
		result := RootPath(path, 20, styles)
		Expect(result).To(Equal(pathStyle.Render(path)))
	})
})
