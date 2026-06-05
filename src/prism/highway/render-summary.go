package highway

import (
	"strings"

	"github.com/snivilised/jaywalk/src/prism/widgets/border"
)

// renderSummary writes the status row, the bottom border and the
// optional "press space to exit" footer to b.
//
// The status row is rendered entirely by the child status widget
// (m.status.View().Content). The bottom border is delegated to the
// border widget (border.RenderBottom) so the same chrome code path
// is shared with the linear view and the highway header. The
// "press space to exit" footer reads m.status.IsDone() — the
// widget owns the done flag, not the root.
func (m Model) renderSummary(b *strings.Builder) {
	b.WriteString(m.status.View().Content)
	b.WriteString("\n")

	b.WriteString(border.RenderBottom(m.width, border.Styles{
		BorderStyle: m.theme.BorderStyle,
	}))

	if m.status.IsDone() {
		b.WriteString("\n")
		b.WriteString(m.theme.MutedStyle.Render(" • press space to exit"))
	}
}
