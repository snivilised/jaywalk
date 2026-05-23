package highway

import (
	"strings"

	"charm.land/bubbletea/v2"
)

func (m Model) View() tea.View {
	var b strings.Builder

	m.renderHeader(&b)
	m.renderLanes(&b)
	m.renderSummary(&b)

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}
