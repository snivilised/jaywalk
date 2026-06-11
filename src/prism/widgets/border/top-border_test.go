package border

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

var _ = Describe("TopBorder", func() {
	It("renders correct border with root path", func() {
		theme, _ := contract.NewTheme(contract.SystemPalette(), nil)
		styles := Styles{
			BorderStyle: theme.BorderStyle,
			PathStyle:   theme.RootStyle,
			CornerStyle: theme.BorderStyle,
		}
		rootPath := "/home/user"
		width := 40

		// Expected calculation based on the logic in TopBorder
		pathWidth := lipgloss.Width(rootPath)
		maxPathWidth := width - 13
		if pathWidth > maxPathWidth {
			keep := max(0, maxPathWidth-3)
			rootPath = contract.Ellipses + rootPath[pathWidth-keep:]
			pathWidth = maxPathWidth
		}
		avail := max(2, width-pathWidth-11)
		L := avail / 2
		R := avail - L

		expect := theme.BorderStyle.Render(contract.Static.Borders.TopLeftCorner+strings.Repeat("─", L)+"[ ") +
			theme.RootStyle.Render(rootPath) +
			theme.BorderStyle.Render(" ]"+strings.Repeat("─", R)+contract.Static.Borders.TopRight) + "\n"

		result := RenderTop(rootPath, width, styles)
		Expect(result).To(Equal(expect))
	})

	It("truncates root path if too long", func() {
		theme, _ := contract.NewTheme(contract.SystemPalette(), nil)
		styles := Styles{
			BorderStyle: theme.BorderStyle,
			PathStyle:   theme.RootStyle,
			CornerStyle: theme.BorderStyle,
		}
		rootPath := "/very/long/path/to/a/deeply/nested/directory"
		width := 40

		// Expected calculation based on the logic in TopBorder
		// Compute truncated path using widget logic
		pathWidth := lipgloss.Width(rootPath)
		maxPathWidth := width - 13
		if pathWidth > maxPathWidth {
			keep := max(0, maxPathWidth-3)
			rootPath = contract.Ellipses + rootPath[pathWidth-keep:]
		}
		truncatedPath := rootPath
		// Recalculate width after truncation
		pathWidth = lipgloss.Width(truncatedPath)

		avail := max(2, width-pathWidth-11)
		L := avail / 2
		R := avail - L

		expect := theme.BorderStyle.Render(contract.Static.Borders.TopLeftCorner+strings.Repeat("─", L)+"[ ") +
			theme.RootStyle.Render(truncatedPath) +
			theme.BorderStyle.Render(" ]"+strings.Repeat("─", R)+contract.Static.Borders.TopRight) + "\n"

		result := RenderTop(rootPath, width, styles)
		Expect(result).To(Equal(expect))
	})

	It("renders border without content when content is empty", func() {
		theme, _ := contract.NewTheme(contract.SystemPalette(), nil)
		styles := Styles{
			BorderStyle: theme.BorderStyle,
			PathStyle:   theme.RootStyle,
			CornerStyle: theme.BorderStyle,
		}
		width := 40

		N := max(0, width-7)
		expect := theme.BorderStyle.Render(
			contract.Static.Borders.TopLeftCorner+strings.Repeat("─", N)+contract.Static.Borders.TopRight,
		) + "\n"

		result := RenderTop("", width, styles)
		Expect(result).To(Equal(expect))
	})
})
