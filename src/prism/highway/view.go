package highway

import (
	"strings"

	"charm.land/bubbletea/v2"
)

func (m Model) View() tea.View {
	var b strings.Builder

	m.renderHeader(&b)
	if m.FlagsRowPosition == FlagsRowPositionTop {
		m.renderFlagsRow(&b)
	}
	m.renderLanes(&b)
	if m.FlagsRowPosition == FlagsRowPositionBottom || m.FlagsRowPosition == "" {
		m.renderFlagsRow(&b)
	}
	m.renderSummary(&b)

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}
