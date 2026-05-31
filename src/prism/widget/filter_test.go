package widget

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"charm.land/lipgloss/v2"
)

var _ = Describe("Filter", func() {
	It("returns empty when no filters active", func() {
		styles := FilterStyles{InfoStyle: lipgloss.NewStyle()}
		result := Filter("", "", "", "", "", "", styles)
		Expect(result).To(Equal(""))
	})

	It("shows files glob before files regex", func() {
		styles := FilterStyles{InfoStyle: lipgloss.NewStyle()}
		filesGlob := "*.go"
		filesRegex := `.*\.go`
		expect := " └─ [ files-glob:*.go ]"
		result := Filter(filesGlob, filesRegex, "", "", "", "", styles)
		Expect(result).To(Equal(expect))
	})

	It("shows dirs glob before dirs regex", func() {
		styles := FilterStyles{InfoStyle: lipgloss.NewStyle()}
		dirsGlob := "src/*"
		dirsRegex := "src/.*"
		expect := " └─ [ dirs-glob:src/* ]"
		result := Filter("", "", dirsGlob, dirsRegex, "", "", styles)
		Expect(result).To(Equal(expect))
	})

	It("handles multiple filter types", func() {
		styles := FilterStyles{InfoStyle: lipgloss.NewStyle()}
		filesGlob := "*.go"
		dirsRegex := "src/.*"
		expect := " └─ [ files-glob:*.go, dirs-regex:src/.* ]"
		result := Filter(filesGlob, "", "", dirsRegex, "", "", styles)
		Expect(result).To(Equal(expect))
	})

	It("applies InfoStyle", func() {
		infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
		styles := FilterStyles{InfoStyle: infoStyle}
		filesGlob := "*.js"
		expect := " └─ [ files-glob:*.js ]"
		result := Filter(filesGlob, "", "", "", "", "", styles)
		Expect(result).To(Equal(infoStyle.Render(expect)))
	})
})
