package widget

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"charm.land/lipgloss/v2"
)

var _ = Describe("Pipeline", func() {
	It("returns empty string for empty pipeline name", func() {
		styles := PipelineStyles{PipelineStyle: lipgloss.NewStyle()}
		result := Pipeline("", styles)
		Expect(result).To(Equal(""))
	})

	It("formats with dashes and bullet point", func() {
		pipelineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
		styles := PipelineStyles{PipelineStyle: pipelineStyle}
		pipelineName := "my-pipeline"
		expect := "─── [ • via pipeline 'my-pipeline' ] ───"
		result := Pipeline(pipelineName, styles)
		Expect(result).To(Equal(pipelineStyle.Render(expect)))
	})

	It("applies PipelineStyle", func() {
		pipelineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
		styles := PipelineStyles{PipelineStyle: pipelineStyle}
		pipelineName := "another-pipeline"
		expect := "─── [ • via pipeline 'another-pipeline' ] ───"
		result := Pipeline(pipelineName, styles)
		Expect(result).To(Equal(pipelineStyle.Render(expect)))
	})
})
