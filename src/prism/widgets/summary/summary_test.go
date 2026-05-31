package summary_test

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/prism/widgets/summary"
)

var _ = Describe("Summary", func() {
	It("renders summary box with aligned values across emoji icon widths", func() {
		output := ansi.Strip(summary.Render([]summary.Field{
			{Icon: "🔖", Label: "Files", Value: "55"},
			{Icon: "📁", Label: "Directories", Value: "7"},
			{Icon: "⏱️", Label: "Elapsed", Value: "2ms"},
		}, summary.Styles{
			Box: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(0, 2),
			Default: summary.CellStyles{
				Icon:  lipgloss.NewStyle(),
				Label: lipgloss.NewStyle(),
				Value: lipgloss.NewStyle(),
			},
		}))

		filesEnd := summaryValueEndColumn(output, "Files", "55")

		Expect(summaryValueEndColumn(output, "Directories", "7")).To(Equal(filesEnd),
			"Directories value should align with Files\n%s", filesEnd, output,
		)
		Expect(summaryValueEndColumn(output, "Elapsed", "2ms")).To(Equal(filesEnd),
			"Elapsed value should align with Files\n%s", filesEnd, output,
		)
	})
})

func summaryValueEndColumn(output, label, value string) int {
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, label) {
			continue
		}

		valueIndex := strings.LastIndex(line, value)
		Expect(valueIndex).NotTo(Equal(-1), "expected line %q to contain value %q", line, value)

		return runewidth.StringWidth(line[:valueIndex]) + runewidth.StringWidth(value)
	}

	Fail(fmt.Sprintf("summary label not found: %s", label))

	return 0
}
