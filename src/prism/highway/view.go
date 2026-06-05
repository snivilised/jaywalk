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
	if m.bannerInfo.Position == banner.PositionTop {
		m.writeBanner(&b)
	}

	m.renderHeader(&b)
	if m.FlagsRowPosition == FlagsRowPositionTop {
		m.renderFlagsRow(&b)
	}

	// Lane rows + separators come from the track child. The track
	// widget is a self-contained bubbletea Model that owns its
	// own styling and rendering.
	b.WriteString(m.track.View().Content)

	if m.FlagsRowPosition == FlagsRowPositionBottom || m.FlagsRowPosition == "" {
		m.renderFlagsRow(&b)
	}
	m.renderSummary(&b)

	// Banner at bottom: rendered OUTSIDE the bordered region, below
	// the summary. The summary's last line is the bottom border
	// (no trailing newline), so we emit a separator newline to
	// keep the banner from overwriting the border.
	if m.bannerInfo.Position == banner.PositionBottom {
		b.WriteByte('\n')
		m.writeBanner(&b)
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
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
