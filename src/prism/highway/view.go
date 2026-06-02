package highway

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
)

func (m Model) View() tea.View {
	var b strings.Builder

	// Banner at top: rendered OUTSIDE the bordered region, above the
	// top border. This is independent of the flags row position
	// because the flags row lives INSIDE the border.
	if m.banner != nil && m.bannerInfo.Position == BannerPositionTop {
		m.bannerInfo.renderTo(&b, m.width)
	}

	m.renderHeader(&b)
	if m.FlagsRowPosition == FlagsRowPositionTop {
		m.renderFlagsRow(&b)
	}
	m.renderLanes(&b)
	if m.FlagsRowPosition == FlagsRowPositionBottom || m.FlagsRowPosition == "" {
		m.renderFlagsRow(&b)
	}
	m.renderSummary(&b)

	// Banner at bottom: rendered OUTSIDE the bordered region, below
	// the summary. The summary's last line is the bottom border
	// (no trailing newline), so we emit a separator newline to
	// keep the banner from overwriting the border.
	if m.banner != nil && m.bannerInfo.Position == BannerPositionBottom {
		b.WriteByte('\n')
		m.bannerInfo.renderTo(&b, m.width)
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

// renderTo invokes the banner widget and writes the result to b.
// Centralising the call here means the view.go body stays purely
// about layout, and the banner widget stays decoupled from the
// highway model.
func (info BannerInfo) renderTo(b *strings.Builder, width int) {
	if info.Gradient == nil || info.State == nil {
		return
	}
	out := banner.Render(banner.Config{
		Width:   width,
		Justify: info.Justify,
	}, banner.Styles{}, banner.Effect{
		Gradient: info.Gradient,
		State:    info.State,
		Aspects: banner.Aspects{
			Orientation: banner.Orientation(info.Aspects.Orientation),
			Banding:     banner.Banding(info.Aspects.Banding),
			Unity:       banner.Unity(info.Aspects.Unity),
			FixedEnd:    banner.FixedEnd(info.Aspects.FixedEnd),
		},
	})
	if out == "" {
		return
	}
	b.WriteString(out)
}
