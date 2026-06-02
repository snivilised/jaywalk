package sampler

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/prism/contract"

	"charm.land/lipgloss/v2"
)

var _ = Describe("Sampler.Render", func() {
	noStyles := Styles{
		LabelStyle: lipgloss.NewStyle(),
		ValueStyle: lipgloss.NewStyle(),
	}

	It("returns empty when no sampler fields active", func() {
		result := Render(0, 0, false, noStyles)
		Expect(result).To(Equal(""))
	})

	It("renders the 🐌 emoji when --last is set", func() {
		result := Render(0, 0, true, noStyles)
		Expect(result).To(Equal(contract.Static.Emoji.Snail))
	})

	It("renders #files when numFiles is positive", func() {
		result := Render(10, 0, false, noStyles)
		Expect(result).To(Equal("#files: 10"))
	})

	It("renders #dirs when numFolders is positive", func() {
		result := Render(0, 5, false, noStyles)
		Expect(result).To(Equal("#dirs: 5"))
	})

	It("places 🐌 before the count items", func() {
		result := Render(10, 5, true, noStyles)
		Expect(result).To(Equal("🐌 | #files: 10 | #dirs: 5"))
	})

	It("places #files before #dirs when 🐌 is absent", func() {
		result := Render(10, 5, false, noStyles)
		Expect(result).To(Equal("#files: 10 | #dirs: 5"))
	})

	It("treats a zero value as unset for counts", func() {
		result := Render(0, 5, true, noStyles)
		Expect(result).To(Equal("🐌 | #dirs: 5"))
	})

	It("applies LabelStyle to labels and ValueStyle to numeric values", func() {
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
		valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
		styles := Styles{
			LabelStyle: labelStyle,
			ValueStyle: valueStyle,
		}
		expect := labelStyle.Render(contract.Static.Emoji.Snail) + " | " +
			labelStyle.Render("#files") + ": " + valueStyle.Render("7") + " | " +
			labelStyle.Render("#dirs") + ": " + valueStyle.Render("3")
		result := Render(7, 3, true, styles)
		Expect(result).To(Equal(expect))
	})

	It("applies LabelStyle to the 🐌 indicator (it has no value)", func() {
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
		styles := Styles{
			LabelStyle: labelStyle,
			ValueStyle: lipgloss.NewStyle(),
		}
		result := Render(0, 0, true, styles)
		Expect(result).To(Equal(labelStyle.Render(contract.Static.Emoji.Snail)))
	})
})
