package track

import (
	tea "charm.land/bubbletea/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
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

	case contract.MotifMsg:
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
func (m Model) applyMotifData(msg contract.MotifMsg) Model {
	if m.counted == nil {
		m.counted = make(map[string]bool)
	}
	if !m.counted[msg.Path] {
		m.counted[msg.Path] = true
		if msg.IsDir {
			m.dirs++
		} else {
			m.files++
		}
	}
	if len(m.lanes) > 0 {
		idx := m.currentLaneIdx
		m.lanes[idx].JobEmoji = msg.JobEmoji
		m.lanes[idx].Path = msg.Path
		m.lanes[idx].Name = msg.Name
		m.lanes[idx].IsDir = msg.IsDir
		m.lanes[idx].Depth = msg.Depth
		m.lanes[idx].ActionName = msg.ActionName
		m.lanes[idx].PipelineName = msg.PipelineName
		m.lanes[idx].CommandOutput = msg.CommandOutput
		m.lanes[idx].ExecutionString = msg.ExecutionString
		m.lanes[idx].DryRun = msg.DryRun
		m.lanes[idx].Err = msg.Err

		// Copy gradient from message to lane if provided. The
		// gradient is a ResolvedGradient{Steps, Hi, Lo} from
		// the theme palette. It holds colour endpoint info;
		// the view applies it via ApplyGradient.
		if msg.Gradient != nil {
			m.lanes[idx].HighlightGradient = msg.Gradient
			if m.lanes[idx].GradientState == nil {
				m.lanes[idx].GradientState = effects.NewGradientState()
			}
			m.lanes[idx].GradientState.TotalSteps = msg.Gradient.Steps
		}
		if msg.PeriscopeGradient != nil {
			m.lanes[idx].PeriscopeGradient = msg.PeriscopeGradient
			if m.lanes[idx].PeriscopeGradientState == nil {
				m.lanes[idx].PeriscopeGradientState = effects.NewGradientState()
			}
			m.lanes[idx].PeriscopeGradientState.TotalSteps = msg.PeriscopeGradient.Steps
		}
		m.currentLaneIdx = (m.currentLaneIdx + 1) % len(m.lanes)
	}
	return m
}
