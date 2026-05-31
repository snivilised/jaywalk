package cascade

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"charm.land/lipgloss/v2"
)

var _ = Describe("Cascade.Render", func() {
	It("returns empty string for empty cascade", func() {
		styles := Styles{HeaderStyle: lipgloss.NewStyle()}
		result := Render("", styles)
		Expect(result).To(Equal(""))
	})

	It("returns styled cascade value", func() {
		headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		styles := Styles{HeaderStyle: headerStyle}
		cascade := "🔒"
		result := Render(cascade, styles)
		Expect(result).To(Equal(headerStyle.Render(cascade)))
	})

	It("returns styled depth value", func() {
		headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		styles := Styles{HeaderStyle: headerStyle}
		cascade := "depth:5"
		result := Render(cascade, styles)
		Expect(result).To(Equal(headerStyle.Render(cascade)))
	})
})
