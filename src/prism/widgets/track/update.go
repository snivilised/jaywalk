package track

import (
	tea "charm.land/bubbletea/v2"

	"github.com/snivilised/jaywalk/src/prism/effects"
)

// Update handles widget-owned messages. The widget does not own a
// ticker: TickMsg is supplied by the highway root on every global
// tick. MotifMsg applies data to the current lane, increments the
// deduped files/dirs counters and rotates the lane index.
// CompleteMsg is a flush signal that clears the counted map.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case WidthMsg:
		m.width = msg.Width
		return m, nil

	case TickMsg:
		// Advance each lane's frame counter independently.
		// Lanes with a skip factor > 0 (set via IntervalMs
		// override) only advance their tick every N global
		// ticks, producing a visibly slower animation. Lanes
		// with skip factor 0 advance every tick (full speed).
		for i := range m.lanes {
			if m.skip != nil && i < len(m.skip) && m.skip[i] > 0 {
				m.lanes[i].skipCounter++
				if m.lanes[i].skipCounter >= m.skip[i] {
					m.lanes[i].skipCounter = 0
					m.lanes[i].tick++
				}
			} else {
				m.lanes[i].tick++
			}

			// Advance gradient state for lanes with configured
			// gradients. The GradientState.Offset tracks the
			// current position in the gradient array;
			// ApplyGradient uses this offset to interpolate
			// characters from Hi to Lo. The view reads
			// GradientState to colour the frame.
			if m.lanes[i].HighlightGradient != nil {
				windowSize := m.lanes[i].WindowSize()
				if windowSize <= 0 {
					windowSize = 4
				}
				m.lanes[i].GradientState.Update(windowSize)
			}

			if m.lanes[i].PeriscopeGradient != nil {
				windowSize := m.lanes[i].WindowSize()
				if windowSize <= 0 {
					windowSize = 4
				}
				m.lanes[i].PeriscopeGradientState.Update(windowSize)
			}
		}
		return m, nil

	case MotifMsg:
		return m.applyMotifData(msg), nil

	case CensusMsg:
		if msg.MaxDepth > m.maxDepth {
			m.maxDepth = msg.MaxDepth
		}
		return m, nil

	case CompleteMsg:
		// Flush signal. Clear the dedup map so any further
		// MotifMsg is a no-op dedup-wise. (We intentionally do
		// NOT add a guard that drops motifs received after
		// CompleteMsg, to preserve the pre-refactor behaviour
		// where late motifs were still applied to the lane.)
		m.counted = make(map[string]bool)
		return m, nil

	default:
		return m, nil
	}
}

// applyMotifData is the per-MotifMsg body. Extracted from the
// former highway.Model.applyMotifData so the Update switch arm
// stays readable. The dedup is on the path: a second motif with
// the same path increments neither files nor dirs but still
// applies its data to the current lane and rotates the index.
func (m Model) applyMotifData(msg MotifMsg) Model {
	data := msg.Data
	if m.counted == nil {
		m.counted = make(map[string]bool)
	}
	if !m.counted[data.Path] {
		m.counted[data.Path] = true
		if data.IsDir {
			m.dirs++
		} else {
			m.files++
		}
	}
	if len(m.lanes) > 0 {
		idx := m.currentLaneIdx
		m.lanes[idx].JobEmoji = data.JobEmoji
		m.lanes[idx].Path = data.Path
		m.lanes[idx].Name = data.Name
		m.lanes[idx].IsDir = data.IsDir
		m.lanes[idx].Depth = data.Depth
		m.lanes[idx].ActionName = data.ActionName
		m.lanes[idx].PipelineName = data.PipelineName
		m.lanes[idx].CommandOutput = data.CommandOutput
		m.lanes[idx].ExecutionString = data.ExecutionString
		m.lanes[idx].DryRun = data.DryRun
		m.lanes[idx].Err = data.Err

		// Copy gradient from message to lane if provided. The
		// gradient is a ResolvedGradient{Steps, Hi, Lo} from
		// the theme palette. It holds colour endpoint info;
		// the view applies it via ApplyGradient.
		if data.Gradient != nil {
			m.lanes[idx].HighlightGradient = data.Gradient
			if m.lanes[idx].GradientState == nil {
				m.lanes[idx].GradientState = effects.NewGradientState()
			}
			m.lanes[idx].GradientState.TotalSteps = data.Gradient.Steps
		}
		if data.PeriscopeGradient != nil {
			m.lanes[idx].PeriscopeGradient = data.PeriscopeGradient
			if m.lanes[idx].PeriscopeGradientState == nil {
				m.lanes[idx].PeriscopeGradientState = effects.NewGradientState()
			}
			m.lanes[idx].PeriscopeGradientState.TotalSteps = data.PeriscopeGradient.Steps
		}
		m.currentLaneIdx = (m.currentLaneIdx + 1) % len(m.lanes)
	}
	return m
}
