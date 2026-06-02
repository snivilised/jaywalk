package cascade

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/prism/contract"

	"charm.land/lipgloss/v2"
)

var _ = Describe("Cascade.Render", func() {
	It("returns empty string for empty cascade", func() {
		styles := Styles{ValueStyle: lipgloss.NewStyle()}
		result := Render("", styles)
		Expect(result).To(Equal(""))
	})

	It("returns styled cascade value", func() {
		valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		styles := Styles{ValueStyle: valueStyle}
		cascade := contract.Static.Emoji.Padlock
		result := Render(cascade, styles)
		Expect(result).To(Equal(valueStyle.Render(cascade)))
	})

	It("returns styled depth value", func() {
		valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		styles := Styles{ValueStyle: valueStyle}
		cascade := "depth:5"
		result := Render(cascade, styles)
		Expect(result).To(Equal(valueStyle.Render(cascade)))
	})
})
