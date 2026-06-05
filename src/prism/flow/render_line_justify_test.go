package flow

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// TestRenderLine_RightJustifiedLanding is a regression test for the
// requirement that the porthole view's landing strip be aligned to
// the right edge of the body, not pressed against the action name
// next to the node. When bodyWidth is supplied to RenderLine, the
// landing strip is preceded by enough whitespace to push it to the
// right edge of the body.
func TestRenderLine_RightJustifiedLanding(t *testing.T) {
	th := buildFlowTestTheme(t)

	cases := []struct {
		name      string
		bodyWidth uint
		// expected minimum visible width of the rendered line
		minVisible int
		// expected exact visible width of the rendered line
		// (it must equal bodyWidth when the strip is right-justified)
		wantExactVisible int
	}{
		{name: "bodyWidth=80", bodyWidth: 80, minVisible: 80, wantExactVisible: 80},
		{name: "bodyWidth=120", bodyWidth: 120, minVisible: 120, wantExactVisible: 120},
		{name: "bodyWidth=40 (small)", bodyWidth: 40, minVisible: 40, wantExactVisible: 40},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := RenderLine(
				"a/b", "bedrock", true, 1,
				"boo", "", "sleep 3.00s", "", false, nil,
				true, false, false, 1, nil, c.bodyWidth, th, "",
			)
			w := lipgloss.Width(res.Line)
			if w != c.wantExactVisible {
				t.Errorf("RenderLine visible width = %d, want %d\nraw: %q",
					w, c.wantExactVisible, res.Line)
			}
			// The strip must contain [sleep 3.00s] when ANSI is
			// stripped. Use ansi.Strip to ignore the surrounding
			// colour codes applied by BranchStyle and
			// LandingStripStyle.
			plain := ansi.Strip(res.Line)
			if !strings.Contains(plain, "[sleep 3.00s]") {
				t.Errorf("RenderLine output missing landing strip\nplain: %q", plain)
			}
		})
	}
}

// TestRenderLine_NoJustifyWhenBodyWidthZero confirms that the legacy
// inline behaviour is preserved when bodyWidth=0 (the linear renderer
// uses this path).
func TestRenderLine_NoJustifyWhenBodyWidthZero(t *testing.T) {
	th := buildFlowTestTheme(t)

	res := RenderLine(
		"a/b", "bedrock", true, 1,
		"boo", "", "sleep 3.00s", "", false, nil,
		true, false, false, 1, nil, 0, th, "",
	)
	// Inline behaviour: the strip should appear immediately after the
	// action name with a single space. The line should be much shorter
	// than the body width because no padding is inserted.
	if lipgloss.Width(res.Line) > 50 {
		t.Errorf("expected inline (short) line when bodyWidth=0, got visible %d",
			lipgloss.Width(res.Line))
	}
	plain := ansi.Strip(res.Line)
	if !strings.Contains(plain, "  • via boo [sleep 3.00s]") {
		t.Errorf("expected inline strip directly after action, got %q", plain)
	}
}

// buildFlowTestTheme returns a minimal theme with the styles needed
// by RenderLine. We avoid using contract.NewTheme because the test
// runs in environments where stdout may be a non-tty writer.
func buildFlowTestTheme(t *testing.T) contract.Theme {
	t.Helper()
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
