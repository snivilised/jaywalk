package widget

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"charm.land/lipgloss/v2"
)

var _ = Describe("Cascade", func() {
	It("returns empty string for empty cascade", func() {
		styles := CascadeStyles{HeaderStyle: lipgloss.NewStyle()}
		result := Cascade("", styles)
		Expect(result).To(Equal(""))
	})

	It("returns styled cascade value", func() {
		headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		styles := CascadeStyles{HeaderStyle: headerStyle}
		cascade := "🔒"
		result := Cascade(cascade, styles)
		Expect(result).To(Equal(headerStyle.Render(cascade)))
	})

	It("returns styled depth value", func() {
		headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		styles := CascadeStyles{HeaderStyle: headerStyle}
		cascade := "depth:5"
		result := Cascade(cascade, styles)
		Expect(result).To(Equal(headerStyle.Render(cascade)))
	})
})
