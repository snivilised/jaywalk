package status

import (
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
)

// Update handles widget-owned messages. Messages that re-target the
// progress bar (PercentMsg, TotalMsg, IncDoneMsg, DoneMsg, ResetMsg)
// drive the embedded spring via SetPercent and return its first-frame
// cmd. FrameMsg is forwarded to the inner model so the spring can
// advance toward equilibrium. Any unknown message is ignored and
// returns a nil cmd.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case WidthMsg:
		// Only update the row width. The embedded progress bar's
		// width is intentionally NOT resized on terminal resize
		// because the bar is meant to stay a small fixed glyph
		// cluster (10 cells by default) - matching the previous
		// behaviour where the bubbles progress Model's width
		// was set once at construction and never touched again.
		m.width = msg.Width
		return m, nil

	case CountsMsg:
		// CountsMsg now comes from a single source (the view model's
		// live tracking), so the values are always monotonically
		// increasing during navigation and match the final values at
		// completion. Direct assignment replaces the previous max()
		// hack that masked a disagreement between MotifMsg-based and
		// agenor-metric-based counts.
		m.files = msg.Files
		m.dirs = msg.Dirs
		m.errors = msg.Errors
		m.skipped = msg.Skipped
		return m, nil

	case ElapsedMsg:
		m.elapsed = msg.Elapsed
		return m, nil

	case PercentMsg:
		m.percent = clamp(msg.Percent, 0, 100)
		cmd := m.inner.SetPercent(float64(m.percent) / 100.0)
		return m, cmd

	case TotalMsg:
		m.total, m.hasTotal = msg.Total, true
		m.totalFiles = msg.TotalFiles
		m.totalDirs = msg.TotalDirs
		m = m.recomputePercent()
		cmd := m.inner.SetPercent(float64(m.percent) / 100.0)
		return m, cmd

	case DoneMsg:
		m.done = msg.Done
		m.isDone = msg.IsDone
		if msg.Err != "" {
			m.errMsg = msg.Err
		}
		// The isDone flag gates the "✔ complete" message, but
		// the message itself only renders when percent >= 100
		// (error path aside). This lets the bar fill up during
		// navigation; "✔ complete" appears only when the last
		// worker actually stops AND progress has reached 100%.
		m = m.recomputePercent()
		cmd := m.inner.SetPercent(float64(m.percent) / 100.0)
		return m, cmd

	case IncDoneMsg:
		if m.isDone {
			// Already completed — a late IncDoneMsg must not
			// overwrite the 100% that DoneMsg set. This can
			// happen when a worker dispatches a MotifMsg after
			// the traversal has already completed (race between
			// the last worker's message and CompleteMsg).
			return m, nil
		}
		n := msg.N
		if n == 0 {
			n = 1
		}
		m.done += n
		m = m.recomputePercent()
		cmd := m.inner.SetPercent(float64(m.percent) / 100.0)
		return m, cmd

	case ResetMsg:
		m.percent, m.done, m.total, m.hasTotal = 0, 0, 0, false
		m.isDone = false
		m.errMsg = ""
		cmd := m.inner.SetPercent(0.0)
		return m, cmd

	case progress.FrameMsg:
		// Forward animation frames to the inner spring. The
		// inner Update returns the next-frame cmd while the
		// spring is still moving and nil once it has reached
		// equilibrium - we propagate either verbatim.
		inner, cmd := m.inner.Update(msg)
		m.inner = inner
		return m, cmd

	default:
		return m, nil
	}
}

// recomputePercent derives m.percent from the done/total ratio.
// Returns 0 if no total has been set (TotalMsg not yet seen) so
// the caller does not have to guard. The result is clamped to
// 100 because the CensusMsg-supplied total is a preview
// estimate: if real navigation visits more items than previewed,
// done > total and the raw ratio would exceed 100.
//
// DoneMsg with IsDone=true does NOT force percent to 100; the
// "✔ complete" message and the progress bar are independent.
// The bar tracks the done/total ratio throughout, so it can
// reach 100% naturally when done == total during navigation.
//
// TODO(progress-indicator-design): consider switching the label
// to a count display ("X / Y") during navigation so the
// denominator is always visible. This would avoid the rare
// premature-100% case where the preview undercounts the true
// file count. For now we render the percent label as before.
func (m Model) recomputePercent() Model {
	if !m.hasTotal || m.total <= 0 {
		return m
	}
	pct := (m.done * 100) / m.total
	m.percent = clamp(pct, 0, 100)
	return m
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
