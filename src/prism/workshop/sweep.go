package workshop

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

// SweepVisualiser renders a plain animated horizontal
// sweep bar. Each character position is coloured with
// the corresponding gradient step. The animFrame offset
// shifts which step maps to the left edge, producing a
// scrolling animation.
type SweepVisualiser struct {
	Width int
}

// Name returns the display name for this visualiser.
func (v *SweepVisualiser) Name() string {
	return "sweep"
}

// Render produces the terminal string for the sweep
// bar. steps is the pre-interpolated colour slice.
// animFrame shifts the starting step for animation.
func (v *SweepVisualiser) Render(
	steps []contract.Color,
	_ enums.CurveKind,
	_ enums.EasingKind,
	animFrame int,
) string {
	if len(steps) == 0 || v.Width <= 0 {
		return ""
	}

	var b strings.Builder
	n := len(steps)

	styles := make([]lipgloss.Style, n)
	for i, c := range steps {
		styles[i] = lipgloss.NewStyle().Foreground(
			lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)),
		)
	}

	for col := 0; col < v.Width; col++ {
		stepIdx := (col + animFrame) % n
		if stepIdx < 0 {
			stepIdx += n
		}
		b.WriteString(styles[stepIdx].Render("█"))
	}

	return b.String()
}
