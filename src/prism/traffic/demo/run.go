package demo

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Run(lanes []Lane, tickRate time.Duration, rootPath string) error {
	model := NewModel(lanes, tickRate, rootPath)

	p := tea.NewProgram(model)
	_, err := p.Run()
	return err
}
