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
		m.files, m.dirs, m.errors, m.skipped =
			msg.Files, msg.Dirs, msg.Errors, msg.Skipped
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
		m = m.recomputePercent()
		cmd := m.inner.SetPercent(float64(m.percent) / 100.0)
		return m, cmd

	case DoneMsg:
		m.done = msg.Done
		m.isDone = msg.IsDone
		if msg.Err != "" {
			m.errMsg = msg.Err
		}
		// At completion the percent is known exactly: 100,
		// overriding any recompute-derived value (which may be
		// < 100 if real navigation is still in progress when
		// DoneMsg arrives, or > 100 if the CensusMsg preview
		// undercounted the true file count).
		if msg.IsDone {
			m.percent = 100
		} else {
			m = m.recomputePercent()
		}
		cmd := m.inner.SetPercent(float64(m.percent) / 100.0)
		return m, cmd

	case IncDoneMsg:
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
// estimate: if real navigation visits more files than previewed,
// done > total and the raw ratio would exceed 100. DoneMsg with
// IsDone=true sets percent to 100 unconditionally on top of
// whatever recompute produces, so "100% reached" at completion
// is preserved as a stable signal.
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
