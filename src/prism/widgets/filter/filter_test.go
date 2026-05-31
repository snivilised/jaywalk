package filter

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"charm.land/lipgloss/v2"
)

var _ = Describe("Filter.Render", func() {
	It("returns empty when no filters active", func() {
		styles := Styles{InfoStyle: lipgloss.NewStyle()}
		result := Render("", "", "", "", "", "", styles)
		Expect(result).To(Equal(""))
	})

	It("shows files glob before files regex", func() {
		styles := Styles{InfoStyle: lipgloss.NewStyle()}
		filesGlob := "*.go"
		filesRegex := `.*\.go`
		expect := " └─ [ files-glob:*.go ]"
		result := Render(filesGlob, filesRegex, "", "", "", "", styles)
		Expect(result).To(Equal(expect))
	})

	It("shows dirs glob before dirs regex", func() {
		styles := Styles{InfoStyle: lipgloss.NewStyle()}
		dirsGlob := "src/*"
		dirsRegex := "src/.*"
		expect := " └─ [ dirs-glob:src/* ]"
		result := Render("", "", dirsGlob, dirsRegex, "", "", styles)
		Expect(result).To(Equal(expect))
	})

	It("handles multiple filter types", func() {
		styles := Styles{InfoStyle: lipgloss.NewStyle()}
		filesGlob := "*.go"
		dirsRegex := "src/.*"
		expect := " └─ [ files-glob:*.go, dirs-regex:src/.* ]"
		result := Render(filesGlob, "", "", dirsRegex, "", "", styles)
		Expect(result).To(Equal(expect))
	})

	It("applies InfoStyle", func() {
		infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
		styles := Styles{InfoStyle: infoStyle}
		filesGlob := "*.js"
		expect := " └─ [ files-glob:*.js ]"
		result := Render(filesGlob, "", "", "", "", "", styles)
		Expect(result).To(Equal(infoStyle.Render(expect)))
	})
})
