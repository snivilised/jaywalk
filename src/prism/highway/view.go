package highway

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
	"github.com/snivilised/jaywalk/src/prism/widgets/legend"
)

func (m Model) View() tea.View {
	var b strings.Builder

	// Banner at top: rendered OUTSIDE the bordered region, above the
	// top border. This is independent of the flags row position
	// because the flags row lives INSIDE the border.
	if m.bannerInfo.Position == contract.PositionTop {
		m.writeBanner(&b)
	}

	m.renderHeader(&b)
	if m.FlagsRowPosition == contract.PositionTop && m.hasFlags() {
		m.writeSeparator(&b)
		m.writeLegend(&b)
		m.writeSeparator(&b)
	}

	// Lane rows + separators come from the track child. The track
	// widget is a self-contained bubbletea Model that owns its
	// own styling and rendering.
	b.WriteString(m.track.View().Content)

	if (m.FlagsRowPosition == contract.PositionBottom || m.FlagsRowPosition == "") && m.hasFlags() {
		m.writeSeparator(&b)
		m.writeLegend(&b)
		m.writeSeparator(&b)
	}
	m.renderSummary(&b)

	// Banner at bottom: rendered OUTSIDE the bordered region, below
	// the summary. The summary's last line is the bottom border
	// (no trailing newline), so we emit a separator newline to
	// keep the banner from overwriting the border.
	if m.bannerInfo.Position == contract.PositionBottom {
		b.WriteByte('\n')
		m.writeBanner(&b)
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

// writeSeparator emits a horizontal "├─────┤" border line, used to
// frame the legend section on both sides. The legend widget itself
// is layout-agnostic; the surrounding borders are the view's concern.
func (m Model) writeSeparator(b *strings.Builder) {
	dashes := strings.Repeat("─", max(0, m.width-2))
	b.WriteString(m.theme.BorderStyle.Render("├" + dashes + "┤"))
	b.WriteString("\n")
}

// hasFlags reports whether the legend section will render any
// content. Used to skip emitting the surrounding separator borders
// when there are no active flags, so the layout collapses cleanly.
func (m Model) hasFlags() bool {
	lm := legend.NewModel(
		legend.WithInfo(legend.Info{
			Position: m.FlagsRowPosition,
			Header:   m.header,
		}),
		legend.WithWidth(m.width),
		legend.WithStyles(legend.Styles{
			LabelStyle:  m.theme.SummaryLabelStyle.Width(0),
			ValueStyle:  m.theme.SummaryValueStyle,
			BorderStyle: m.theme.BorderStyle,
		}),
	)
	return lm.Height() > 0
}

// writeBanner constructs a banner.Model on the fly using the
// frozen-per-session bannerInfo plus the current terminal width,
// calls View(), and writes the result to b. The transient model
// pattern matches the user's guidance: maximum reuse of the
// existing Render function, no long-lived banner child on the
// highway root.
//
// The view only calls this when bannerInfo.Position is top or
// bottom; an unknown position value means "no banner" (see the
// Risk note in make-banner-its-own-child-widget.implementation-plan.issue-606.md).
// Disabled is also checked so the highway does not emit ANSI codes
// when the gradient binding is absent.
func (m Model) writeBanner(b *strings.Builder) {
	bm := banner.NewModel(
		banner.WithInfo(m.bannerInfo),
		banner.WithWidth(m.width),
	)
	if bm.Disabled() {
		return
	}
	if out := bm.View(); out != "" {
		b.WriteString(out)
	}
}

// writeLegend constructs a legend.Model on the fly using the
// frozen-per-session headerInfo plus the current terminal width
// and the theme's flags-row styles, calls View(), and writes the
// result to b. The transient model pattern mirrors writeBanner:
// no long-lived legend child on the highway root.
//
// The view only calls this when FlagsRowPosition is top or bottom;
// an unknown / empty value (which the model has already coerced
// to bottom) means "render the row at the bottom". The legend
// widget itself short-circuits to "" when no flag is active, so
// callers do not need to guard.
func (m Model) writeLegend(b *strings.Builder) {
	lm := legend.NewModel(
		legend.WithInfo(legend.Info{
			Position: m.FlagsRowPosition,
			Header:   m.header,
		}),
		legend.WithWidth(m.width),
		legend.WithStyles(legend.Styles{
			LabelStyle:  m.theme.SummaryLabelStyle.Width(0),
			ValueStyle:  m.theme.SummaryValueStyle,
			BorderStyle: m.theme.BorderStyle,
		}),
	)
	if out := lm.View(); out != "" {
		b.WriteString(out)
	}
}
