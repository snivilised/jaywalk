package banner

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/contract/ansi"
	"github.com/snivilised/jaywalk/src/prism/effects"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// Config carries the inputs to Render. Art is the multi-line ASCII art
// to display. Width is the terminal width used by justification.
// Justify selects the horizontal alignment of the banner within Width.
type Config struct {
	Art     string
	Width   int
	Justify string
}

// Styles is reserved for future use; the banner derives its appearance
// from the gradient colour sweep alone. The widget is theme-driven via
// the existing palette/components pipeline, so no lipgloss.Style
// fields are required.
type Styles struct{}

// Effect bundles the gradient colour endpoints, the per-widget
// animation state and the frozen aspect selection made at startup.
type Effect struct {
	Gradient *contract.ResolvedGradient
	State    *effects.GradientState
	Aspects  Aspects
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// Render produces the ANSI-styled banner string.
//
// Behavioural rules:
//   - Returns "" when Art and DefaultArt are both empty (which is
//     only possible if the caller has explicitly passed an empty
//     Art; the default is the JAYWALK banner from banner-defs.go).
//   - When Gradient or State is nil the art is returned without
//     ANSI colour codes (graceful fallback, mirroring activity.Render).
//   - Spaces and newlines are passed through untouched so the art's
//     own alignment is preserved.
//   - Face characters (█) and shadow characters (anything else
//     non-space) are styled per the Aspects selection.
//   - Leading and trailing newlines on the art are trimmed so the
//     source can be written in a human-readable form (e.g. with a
//     newline immediately after the opening backtick of a raw
//     string literal). The output is always terminated with a
//     single newline so the banner sits on its own lines.
func Render(cfg Config, _ Styles, effect Effect) string {
	art := cfg.Art
	if art == "" {
		art = DefaultArt
	}
	if art == "" {
		return ""
	}

	// Trim leading and trailing newlines so the source can use
	// idiomatic raw-string formatting (newline after opening
	// backtick, newline before closing backtick). This leaves the
	// actual line content (including any leading spaces) intact.
	art = strings.Trim(art, "\n")
	if art == "" {
		return ""
	}

	if effect.Gradient == nil || effect.State == nil {
		return art + "\n"
	}

	steps := paletteSteps(effect.Gradient, effect.State)
	if len(steps) == 0 {
		return art + "\n"
	}

	lines := strings.Split(art, "\n")
	pads := applyJustification(lines, cfg.Width, cfg.Justify)

	// posOfLine[i] is the global rune position of the first rune of
	// the i-th line in the un-justified art. Used for Horizontal
	// orientation.
	posOfLine := buildPosOfLine(lines)

	// rowOfGlobalPos[pos] is the row index of the rune at global
	// position pos. Used for Vertical orientation.
	rowOfGlobalPos := buildRowOfGlobalPos(lines)

	var b strings.Builder
	for i, line := range lines {
		b.WriteString(pads[i])
		writeColouredLine(&b, line, posOfLine[i], rowOfGlobalPos, effect, steps)
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	// Trailing newline so the banner sits on its own lines and
	// does not concatenate with whatever the host view renders
	// above (top banner) or below (bottom banner).
	b.WriteByte('\n')
	return b.String()
}

// ---------------------------------------------------------------------------
// Position maps
// ---------------------------------------------------------------------------

// buildPosOfLine returns, for each line, the global rune index of its
// first rune. The cumulative offset between lines includes the newline
// rune so the gradient sweep is continuous across line breaks.
func buildPosOfLine(lines []string) []int {
	pos := make([]int, len(lines))
	cursor := 0
	for i, line := range lines {
		pos[i] = cursor
		cursor += len([]rune(line)) + 1 // +1 for '\n'
	}
	return pos
}

// buildRowOfGlobalPos returns a slice mapping the global rune index in
// the un-justified art to the row (line number) it belongs to. The
// slice has one entry per rune.
func buildRowOfGlobalPos(lines []string) []int {
	total := 0
	for _, l := range lines {
		total += len([]rune(l))
	}
	rows := make([]int, total)
	cursor := 0
	for lineIdx, line := range lines {
		n := len([]rune(line))
		for j := 0; j < n; j++ {
			rows[cursor+j] = lineIdx
		}
		cursor += n
	}
	return rows
}

// ---------------------------------------------------------------------------
// Per-rune colour resolution
// ---------------------------------------------------------------------------

// colourPolicy describes how a non-whitespace rune is coloured.
type colourPolicy struct {
	fixed bool // true => use a fixed endpoint, ignore gradient sweep
	hi    bool // when fixed: true => Hi, false => Lo
}

// policyFor returns the colour policy for a non-whitespace rune given
// the active aspects.
func policyFor(isFace bool, a Aspects) colourPolicy {
	switch a.Unity {
	case UnityUnified:
		return colourPolicy{}
	case UnityGradientFace:
		if isFace {
			return colourPolicy{} // face sweeps
		}
		return colourPolicy{fixed: true, hi: a.FixedEnd != FixedEndLo}
	case UnityShadowFace:
		if !isFace {
			return colourPolicy{} // shadow sweeps
		}
		return colourPolicy{fixed: true, hi: a.FixedEnd != FixedEndLo}
	}
	return colourPolicy{}
}

// runeColour returns the RGB colour for the rune at global position
// pos in the art, based on its face/shadow class and the active unity
// aspect.
func runeColour(pos int, r rune, effect Effect, steps []contract.Color, rows []int) contract.Color {
	isFace := r == faceRune
	p := policyFor(isFace, effect.Aspects)
	if p.fixed {
		c := effect.Gradient.Lo
		if p.hi {
			c = effect.Gradient.Hi
		}
		return rgbOf(c)
	}
	// Gradient path: index into the step palette via the state, using
	// the orientation aspect to decide whether to index by column or
	// row.
	var stepIdx int
	switch effect.Aspects.Orientation {
	case OrientationVertical:
		row := 0
		if pos >= 0 && pos < len(rows) {
			row = rows[pos]
		}
		stepIdx = effect.State.GetEffectiveIndex(row)
	case OrientationHorizontal:
		stepIdx = effect.State.GetEffectiveIndex(pos)
	}
	if stepIdx < 0 || stepIdx >= len(steps) {
		stepIdx = 0
	}
	return steps[stepIdx]
}

// ---------------------------------------------------------------------------
// Line rendering
// ---------------------------------------------------------------------------

// writeColouredLine emits the ANSI-styled version of one line. lineStart
// is the global rune index of the first character of the line in the
// un-justified art, used by Horizontal orientation.
func writeColouredLine(b *strings.Builder, line string, lineStart int, rows []int, effect Effect, steps []contract.Color) {
	runes := []rune(line)
	for localIdx, r := range runes {
		if r == ' ' || r == '\n' || r == '\t' {
			b.WriteRune(r)
			continue
		}
		pos := lineStart + localIdx
		col := runeColour(pos, r, effect, steps, rows)
		emitAnsiRune(b, r, col)
	}
}

// ---------------------------------------------------------------------------
// Justification
// ---------------------------------------------------------------------------

// applyJustification returns the per-line leading-space padding
// required to align each line under the configured Justify mode. The
// banner's intrinsic leading spaces are preserved; the padding is
// appended in front of them.
//
//	right: pad = max(0, Width - lineWidth)
//	left:   pad = 0
//	center: pad = max(0, (Width - lineWidth) / 2)
func applyJustification(lines []string, width int, justify string) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		lineWidth := lipgloss.Width(line)
		var pad int
		switch justify {
		case JustifyLeft:
			pad = 0
		case JustifyCenter:
			if width > lineWidth {
				pad = (width - lineWidth) / 2
			}
		default: // JustifyRight (and unknown)
			if width > lineWidth {
				pad = width - lineWidth
			}
		}
		if pad > 0 {
			out[i] = strings.Repeat(" ", pad)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Gradient step palette
// ---------------------------------------------------------------------------

// paletteSteps returns the interpolated step array for the given
// gradient, lazily populating the state's steps array the first time it
// is called. This mirrors the lazy population in effects.ApplyGradient
// (see gradient-state.go:108-112).
func paletteSteps(g *contract.ResolvedGradient, st *effects.GradientState) []contract.Color {
	if g == nil || st == nil {
		return nil
	}
	if st.TotalSteps <= 0 {
		st.TotalSteps = g.Steps
	}
	if st.TotalSteps <= 0 {
		st.TotalSteps = contract.DefaultStepCount()
	}
	hiR, hiG, hiB, _ := g.Hi.RGBA()
	loR, loG, loB, _ := g.Lo.RGBA()
	steps := contract.InterpolateBetweenRGBA(
		uint8(hiR>>8), uint8(hiG>>8), uint8(hiB>>8), //nolint:gosec // safe: 16-bit value >> 8 fits in 8 bits
		uint8(loR>>8), uint8(loG>>8), uint8(loB>>8), //nolint:gosec // safe: 16-bit value >> 8 fits in 8 bits
		st.TotalSteps,
		g.Curve, g.Easing,
	)
	if len(steps) > 0 {
		st.SetSteps(steps)
	}
	return steps
}

// rgbOf converts a color.Color to an 8-bit-per-channel contract.Color.
func rgbOf(c color.Color) contract.Color {
	if c == nil {
		return contract.Color{}
	}
	r, g, b, _ := c.RGBA()
	return contract.Color{
		R: uint8(r >> 8), //nolint:gosec
		G: uint8(g >> 8), //nolint:gosec
		B: uint8(b >> 8), //nolint:gosec
	}
}

// ---------------------------------------------------------------------------
// ANSI emission
// ---------------------------------------------------------------------------

// emitAnsiRune writes a single coloured rune to b using 24-bit ANSI
// escape codes. The escape sequence mirrors the one used by
// effects.ApplyGradientStyled (see gradient-state.go:212) so the
// output is compatible with every other widget.
func emitAnsiRune(b *strings.Builder, r rune, col contract.Color) {
	_, _ = fmt.Fprint(b, ansi.EscapedTrueColor(col.R, col.G, col.B))
	b.WriteRune(r)
	b.WriteString("\x1b[0m")
}
