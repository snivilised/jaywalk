package track

import (
	tea "charm.land/bubbletea/v2"

	"github.com/snivilised/jaywalk/src/agenor/enums"
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
		// Advance each lane's frame counter independently, but only
		// for lanes whose worker is currently working. Idle lanes
		// freeze their animation frame and gradient position until
		// their worker becomes active again.
		for i := range m.lanes {
			if m.lanes[i].State != enums.WorkerStateWorking {
				continue
			}

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
		// Flush signal. Do NOT clear the dedup map — late
		// MotifMsg that arrive after CompleteMsg must still be
		// de-duped against previously-seen paths so the track
		// child's file/dir counters stay frozen. Without this,
		// a MotifMsg for an already-counted path increments
		// the counter a second time, making the internal count
		// inconsistent with the status widget's frozen display.
		// MotifMsg are still forwarded to lanes (via the m.Done
		// guard in highway model.go) for rendering; only the
		// counter increment is suppressed by the intact map.

		// Freeze all lanes: the traversal is done, so no more
		// MotifMsg will arrive to re-activate any lane. Setting
		// every lane to Idle stops their tick advance so the
		// animation frame freezes on the last displayed frame.
		for i := range m.lanes {
			m.lanes[i].State = enums.WorkerStateIdle
		}
		return m, nil

	case WorkerStateMsg:
		if msg.LaneID >= 0 && msg.LaneID < len(m.lanes) {
			m.lanes[msg.LaneID].State = msg.State
		}
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
		// Route the motif to the lane that corresponds to the
		// pool worker that produced this result. The lane index
		// is derived from worker-id minus 1 (pants assigns
		// consecutive IDs starting at 1), giving a 0-based index
		// that preserves numeric order.
		idx := WorkerIndex(msg.WorkerID) - 1
		if idx < 0 || idx >= len(m.lanes) {
			idx = 0
		}
		// Expand visible lanes when a previously unseen worker ID is
		// observed. This reveals the lane for the new pool worker.
		if _, seen := m.seenWorkers[msg.WorkerID]; !seen {
			m.seenWorkers[msg.WorkerID] = struct{}{}
			if needed := idx + 1; needed > m.visibleCount {
				if needed > len(m.lanes) {
					needed = len(m.lanes)
				}
				m.visibleCount = needed
			}
		}

		m.lanes[idx].WorkerID = msg.WorkerID
		m.lanes[idx].State = enums.WorkerStateWorking
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
	}
	return m
}
