package filter

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"charm.land/lipgloss/v2"
)

var _ = Describe("Filter.Render", func() {
	noStyles := Styles{
		LabelStyle: lipgloss.NewStyle(),
		ValueStyle: lipgloss.NewStyle(),
	}

	It("returns empty when no filters active", func() {
		result := Render("", "", "", "", "", "", noStyles)
		Expect(result).To(Equal(""))
	})

	It("shows files glob before files regex", func() {
		filesGlob := "*.go"
		filesRegex := `.*\.go`
		expect := "files glob: *.go"
		result := Render(filesGlob, filesRegex, "", "", "", "", noStyles)
		Expect(result).To(Equal(expect))
	})

	It("shows dirs glob before dirs regex", func() {
		dirsGlob := "src/*"
		dirsRegex := "src/.*"
		expect := "dirs glob: src/*"
		result := Render("", "", dirsGlob, dirsRegex, "", "", noStyles)
		Expect(result).To(Equal(expect))
	})

	It("handles multiple filter types", func() {
		filesGlob := "*.go"
		dirsRegex := "src/.*"
		expect := "files glob: *.go | dirs regex: src/.*"
		result := Render(filesGlob, "", "", dirsRegex, "", "", noStyles)
		Expect(result).To(Equal(expect))
	})

	It("uses spaces in labels (not CLI dash form)", func() {
		filesGlob := "*.js"
		result := Render(filesGlob, "", "", "", "", "", noStyles)
		Expect(result).To(Equal("files glob: *.js"))
		Expect(result).NotTo(ContainSubstring("files-glob"))
		Expect(result).NotTo(ContainSubstring("dirs-glob"))
	})

	It("applies LabelStyle to labels and ValueStyle to values", func() {
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
		valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
		styles := Styles{
			LabelStyle: labelStyle,
			ValueStyle: valueStyle,
		}
		filesGlob := "*.js"
		expect := labelStyle.Render("files glob") + ": " + valueStyle.Render("*.js")
		result := Render(filesGlob, "", "", "", "", "", styles)
		Expect(result).To(Equal(expect))
	})

	It("preserves the separator between label and value (uncoloured)", func() {
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
		valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
		styles := Styles{
			LabelStyle: labelStyle,
			ValueStyle: valueStyle,
		}
		result := Render("*.js", "", "", "", "", "", styles)
		Expect(result).To(ContainSubstring(": "))
		// The ": " between label and value is intentionally uncoloured
		// so that the boundary between them is visually clean.
	})
})
