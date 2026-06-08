package linear

import (
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

var _ = Describe("RenderLine justify", func() {
	DescribeTable("right-justified landing strip",
		func(bodyWidth uint, wantExactVisible int) {
			th := buildFlowTestTheme()
			res := RenderLine(LineParams{
				NodeParams: contract.NodeParams{
					Path:          "a/b",
					Name:          "bedrock",
					IsDir:         true,
					Depth:         1,
					ActionName:    "boo",
					CommandOutput: "sleep 3.00s",
					IsLast:        true,
					VisualDepth:   1,
				},
				RenderParams: contract.RenderParams{
					BodyWidth: bodyWidth,
					Theme:     th,
				},
			})
			Expect(lipgloss.Width(res.Line)).To(Equal(wantExactVisible))

			plain := ansi.Strip(res.Line)
			Expect(plain).To(ContainSubstring("[sleep 3.00s]"))
		},
		Entry("bodyWidth=80", uint(80), 80),
		Entry("bodyWidth=120", uint(120), 120),
		Entry("bodyWidth=40 (small)", uint(40), 40),
	)

	It("does not justify when bodyWidth is zero", func() {
		th := buildFlowTestTheme()
		res := RenderLine(LineParams{
			NodeParams: contract.NodeParams{
				Path:          "a/b",
				Name:          "bedrock",
				IsDir:         true,
				Depth:         1,
				ActionName:    "boo",
				CommandOutput: "sleep 3.00s",
				IsLast:        true,
				VisualDepth:   1,
			},
			RenderParams: contract.RenderParams{
				BodyWidth: 0,
				Theme:     th,
			},
		})
		// Inline behaviour: the strip should appear immediately after the
		// action name with a single space. The line should be much shorter
		// than the body width because no padding is inserted.
		Expect(lipgloss.Width(res.Line)).To(BeNumerically("<", 50))

		plain := ansi.Strip(res.Line)
		Expect(plain).To(ContainSubstring("  • via boo [sleep 3.00s]"))
	})
})

// buildFlowTestTheme returns a minimal theme with the styles needed
// by RenderLine. We avoid using contract.NewTheme because the test
// runs in environments where stdout may be a non-tty writer.
func buildFlowTestTheme() contract.Theme {
	return contract.Theme{
		BranchStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		DirStyle:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
		FileStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
		ActionStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
		PipelineStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("13")),
		LandingStripStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("11")),
		ErrorStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("9")),
		TreeIcons: contract.TreeIcons{
			contract.TreeIconBranchVertical: "│",
			contract.TreeIconBranchIndent:   "  ",
			contract.TreeIconBranchJoint:    "├── ",
			contract.TreeIconBranchLast:     "└── ",
		},
	}
}
