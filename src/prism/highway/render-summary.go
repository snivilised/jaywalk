package highway

import (
	"strings"
	"time"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/status"
)

func (m Model) renderSummary(b *strings.Builder) {
	elapsedSecs := 0
	var elapsed time.Duration
	if m.done {
		elapsed = m.elapsed
		elapsedSecs = int(m.elapsed.Seconds())
	} else if !m.start.IsZero() {
		elapsed = time.Since(m.start)
		elapsedSecs = int(elapsed.Seconds())
	}

	var files, dirs, errors int
	if m.realMode {
		files = m.files
		dirs = m.dirs
		errors = m.errors
	} else {
		files = elapsedSecs*23 + 5
		dirs = elapsedSecs*7 + 2
		errors = elapsedSecs / 15
	}

	barView := m.progress.ViewAs(float64(m.percent) / 100.0)

	statusRow := status.Render(status.Config{
		Files:        files,
		Dirs:         dirs,
		Errors:       errors,
		Elapsed:      elapsed,
		Percent:      m.percent,
		IsDone:       m.done,
		ErrMsg:       m.errMsg,
		ProgressView: barView,
	}, m.statusStyles, m.statusFields, m.width)

	b.WriteString(statusRow)
	b.WriteString("\n")

	N := max(0, m.width-7)
	b.WriteString(m.theme.BorderStyle.Render(
		contract.Static.Borders.BottomLeft + strings.Repeat("─", N) + contract.Static.Borders.BottomRightCorner,
	))

	if m.done {
		b.WriteString("\n")
		b.WriteString(m.theme.MutedStyle.Render(" • press space to exit"))
	}
}
