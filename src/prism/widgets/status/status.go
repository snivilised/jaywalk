package status

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/layout"
)

// Config holds the data for rendering a status row.
type Config struct {
	// Files is the number of files visited.
	Files int

	// Dirs is the number of directories visited.
	Dirs int

	// Errors is the number of errors encountered.
	Errors int

	// Skipped is the number of items skipped.
	Skipped int

	// Elapsed is the time elapsed since traversal started.
	Elapsed time.Duration

	// Percent is the completion percentage (0-100).
	// Only shown when Fields.ShowProgress is true.
	Percent int

	// IsDone indicates whether the traversal is complete.
	IsDone bool

	// ErrMsg is the error message shown when done with errors.
	ErrMsg string

	// ProgressView is the pre-rendered progress bar string.
	// Only shown when Fields.ShowProgress is true.
	ProgressView string
}

// Styles holds the lipgloss styles for the status row.
type Styles struct {
	// TreeIcons are the glyphs used for field icons.
	TreeIcons contract.TreeIcons

	// SummaryLabelStyle is applied to field labels.
	SummaryLabelStyle lipgloss.Style

	// SummaryValueStyle is applied to field values.
	SummaryValueStyle lipgloss.Style

	// ErrorStyle is applied to error-related content.
	ErrorStyle lipgloss.Style

	// ProgressStyle is applied to progress-related content.
	ProgressStyle lipgloss.Style

	// BorderStyle is applied to border characters.
	BorderStyle lipgloss.Style

	// MutedStyle is applied to secondary text.
	MutedStyle lipgloss.Style
}

// FieldSelectors controls which segments are rendered in the status row.
type FieldSelectors struct {
	// ShowFiles enables the files count segment.
	ShowFiles bool

	// ShowDirs enables the directories count segment.
	ShowDirs bool

	// ShowErrors enables the errors count segment.
	ShowErrors bool

	// ShowSkipped enables the skipped count segment.
	ShowSkipped bool

	// ShowProgress enables the progress bar and percentage segment.
	ShowProgress bool

	// ShowComplete enables the complete/failed message segment.
	ShowComplete bool

	// ShowElapsed enables the elapsed time segment.
	ShowElapsed bool
}

// segment represents a single segment in the status row.
type segment struct {
	content   string
	separator bool // true to render a separator after this segment
}

// Render produces the status row string with the specified fields.
func Render(cfg Config, styles Styles, fields FieldSelectors, width int) string {
	borderStyle := styles.BorderStyle

	segments := make([]segment, 0, 9)
	var elapsedContent string

	// Files segment
	if fields.ShowFiles {
		icon := styles.TreeIcons[contract.TreeIconFile]
		label := styles.SummaryLabelStyle.Render(icon + " files:")
		value := styles.SummaryValueStyle.Render(fmt.Sprintf("%4d", cfg.Files))
		segments = append(segments, segment{
			content:   " " + label + " " + value + " ",
			separator: true,
		})
	}

	// Dirs segment
	if fields.ShowDirs {
		icon := styles.TreeIcons[contract.TreeIconDirectory]
		label := styles.SummaryLabelStyle.Render(icon + " dirs:")
		value := styles.SummaryValueStyle.Render(fmt.Sprintf("%3d", cfg.Dirs))
		segments = append(segments, segment{
			content:   " " + label + " " + value + " ",
			separator: true,
		})
	}

	// Errors segment
	if fields.ShowErrors {
		icon := styles.TreeIcons[contract.TreeIconError]
		label := styles.ErrorStyle.Render(icon + " errors:")
		value := styles.SummaryValueStyle.Render(fmt.Sprintf("%3d", cfg.Errors))
		segments = append(segments, segment{
			content:   " " + label + " " + value + " ",
			separator: true,
		})
	}

	// Skipped segment
	if fields.ShowSkipped {
		icon := styles.TreeIcons[contract.TreeIconSkipped]
		label := styles.SummaryLabelStyle.Render(icon + " skipped:")
		value := styles.SummaryValueStyle.Render(fmt.Sprintf("%3d", cfg.Skipped))
		segments = append(segments, segment{
			content:   " " + label + " " + value + " ",
			separator: true,
		})
	}

	// Progress segment
	if fields.ShowProgress && cfg.ProgressView != "" {
		pctLabel := styles.ProgressStyle.Render(fmt.Sprintf("%3d%%", cfg.Percent))
		segments = append(segments, segment{
			content:   " " + cfg.ProgressView + "  " + pctLabel + " ",
			separator: true,
		})
	}

	// Complete/Failed message segment
	if fields.ShowComplete && cfg.IsDone {
		var msg string
		if cfg.ErrMsg != "" {
			label := "❌ Failed: " + cfg.ErrMsg
			if cfg.Errors > 1 {
				label += fmt.Sprintf(" (+%d more)", cfg.Errors-1)
			}
			msg = " " + styles.ErrorStyle.Render(label) + " "
		} else {
			msg = " " + styles.ProgressStyle.Render("✔ complete") + " "
		}
		segments = append(segments, segment{
			content:   msg,
			separator: true,
		})
	}

	// Elapsed segment (always last, right-aligned)
	if fields.ShowElapsed {
		icon := styles.TreeIcons[contract.TreeIconElapsed]
		label := styles.SummaryLabelStyle.Render(icon + " elapsed:")
		elapsedStr := formatDuration(cfg.Elapsed)
		value := styles.SummaryValueStyle.Render(elapsedStr)
		elapsedContent = " " + label + " " + value + " "
	}

	// Build the row using layout.NewRow
	row := layout.NewRow(width-4).
		Caps(borderStyle.Render("│ "), borderStyle.Render(" │"))

	for _, seg := range segments {
		row.Content(seg.content)
		if seg.separator {
			row.Content(borderStyle.Render("│"))
		}
	}

	if elapsedContent != "" {
		row.RightContent(elapsedContent)
	}

	return row.Render()
}

// formatDuration formats a duration as a human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Milliseconds()))
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	mins := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%02ds", mins, secs)
}
