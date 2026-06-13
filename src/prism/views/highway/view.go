package highway

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/views/shared"
	"github.com/snivilised/jaywalk/src/prism/widgets/legend"
)

// Any significant changes to the highway Model/View api should reflected in the
// documentation in docs/bubbletea-view-layout.md

func (m Model) View() tea.View {
	var b strings.Builder

	// Banner at top: rendered OUTSIDE the bordered region, above the
	// top border. This is independent of the flags row position
	// because the flags row lives INSIDE the border.
	if m.bannerInfo.Position == contract.PositionTop {
		shared.WriteBanner(&b, m.bannerInfo, m.Width)
	}

	m.renderHeader(&b)
	if m.FlagsRowPosition == contract.PositionTop && m.hasFlags() {
		m.WriteSeparator(&b)
		m.writeLegend(&b)
		m.WriteSeparator(&b)
	}

	// Lane rows + separators come from the track child. The track
	// widget is a self-contained bubbletea Model that owns its
	// own styling and rendering.
	b.WriteString(m.track.View().Content)

	if (m.FlagsRowPosition == contract.PositionBottom || m.FlagsRowPosition == "") &&
		m.hasFlags() {
		m.WriteSeparator(&b)
		m.writeLegend(&b)
		m.WriteSeparator(&b)
	}
	m.renderSummary(&b)

	// Banner at bottom: rendered OUTSIDE the bordered region, below
	// the summary. The summary's last line is the bottom border
	// (no trailing newline), so we emit a separator newline to
	// keep the banner from overwriting the border.
	if m.bannerInfo.Position == contract.PositionBottom {
		b.WriteByte('\n')
		shared.WriteBanner(&b, m.bannerInfo, m.Width)
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

// hasFlags reports whether the legend section will render any content.
// Used to skip emitting the surrounding separator borders when there
// are no active flags, so the layout collapses cleanly.
func (m Model) hasFlags() bool {
	return shared.LegendHeight(m.legendParams()) > 0
}

// writeLegend renders the flags/legend row into b using the shared
// chrome utility. The row is a no-op when no flag is active.
func (m Model) writeLegend(b *strings.Builder) {
	shared.WriteLegend(b, m.legendParams())
}

// legendParams returns the LegendParams for the current model state.
// Centralised here so hasFlags and writeLegend stay consistent.
func (m Model) legendParams() shared.LegendParams {
	return shared.LegendParams{
		Info: legend.Info{
			Position: m.FlagsRowPosition,
			Header:   m.Header,
		},
		Width: m.Width,
		Styles: legend.Styles{
			LabelStyle:  m.Theme.SummaryLabelStyle.Width(0),
			ValueStyle:  m.Theme.SummaryValueStyle,
			BorderStyle: m.Theme.BorderStyle,
		},
	}
}
