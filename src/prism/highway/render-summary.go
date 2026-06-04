package highway

import (
	"strings"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// renderSummary writes the status row and bottom border to b.
// The status row is rendered entirely by the child status widget
// (m.status.View().Content). The bottom border and the optional
// "press space to exit" footer are highway chrome and remain
// here.
//
// The "press space to exit" footer reads m.status.IsDone() — the
// widget owns the done flag, not the root.
func (m Model) renderSummary(b *strings.Builder) {
	b.WriteString(m.status.View().Content)
	b.WriteString("\n")

	N := max(0, m.width-7)
	b.WriteString(m.theme.BorderStyle.Render(
		contract.Static.Borders.BottomLeft + strings.Repeat("─", N) + contract.Static.Borders.BottomRightCorner,
	))

	if m.status.IsDone() {
		b.WriteString("\n")
		b.WriteString(m.theme.MutedStyle.Render(" • press space to exit"))
	}
}
