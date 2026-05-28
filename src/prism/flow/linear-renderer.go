// Package flow contains the linear renderer implementation and its
// view-specific options.
//
// Dependency rule: flow imports contract (shared types) and is imported
// by prism root. Parent depends on child; no import cycles.
package flow

import (
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widget"
	"github.com/snivilised/jaywalk/src/third/lo"
)

// renderer is the linear scrolling view. Output is written immediately as
// events arrive - no internal buffering.
type renderer struct {
	theme  contract.Theme
	writer io.Writer

	// treeIcons holds configured tree glyphs from the resolved theme/options.
	treeIcons contract.TreeIcons

	// branchStack tracks ancestor continuation state for tree branch rendering.
	branchStack []bool

	previousDepth  core.TraversalDepth
	previousIsLast bool
}

// Begin renders the opening banner using the Overture metadata. The banner
// adapts to indicate prime or resume traversal.
func (r *renderer) Begin(overture contract.Overture) {
	title := lo.Ternary(overture.Kind == contract.ResumeNavigation,
		fmt.Sprintf("  jay  resuming %s", overture.Root),
		fmt.Sprintf("  jay  %s", overture.Root),
	)

	dateFmt := overture.DateFormat
	if dateFmt == "" {
		dateFmt = time.RFC1123
	}

	caption := fmt.Sprintf("  %s  -  %s",
		overture.Caption,
		overture.StartedAt.Format(dateFmt),
	)

	if overture.Kind == contract.ResumeNavigation && overture.ResumeFrom != "" {
		caption += fmt.Sprintf("  -  from: %s", overture.ResumeFrom)
	}

	box := r.theme.BoxStyle.
		MarginTop(0).
		Render(
			r.theme.SummaryLabelStyle.Width(0).Render(title) +
				"\n" +
				r.theme.SummaryValueStyle.Width(0).Render(caption),
		)

	_, _ = lipgloss.Fprintln(r.writer, box)
}

// Show renders a single Motif immediately to the output writer.
func (r *renderer) Show(motif contract.Motif) {
	prefix := r.branchPrefix(motif)
	depth := r.theme.BranchStyle.Render(prefix)

	var name string

	switch {
	case motif.Err != nil:
		name = r.theme.ErrorStyle.Render(
			fmt.Sprintf("! %s  %s", motif.Name, motif.Err.Error()),
		)

	case motif.Skipped:
		name = r.renderSkipped(motif)

	case motif.IsPipelineStep:
		name = r.renderStep(motif)

	case motif.Depth == 0:
		name = r.renderRoot(motif)

	case motif.IsDir:
		name = r.renderDir(motif)

	default:
		name = r.renderFile(motif)
	}

	_, _ = lipgloss.Fprintf(r.writer, "%s%s\n", depth, name)

	r.updateBranchStack(motif)
}

func (r *renderer) itemLabel(motif contract.Motif) string {
	icon := r.treeIcons[contract.TreeIconFile]
	if motif.IsDir {
		icon = r.treeIcons[contract.TreeIconDirectory]
	}

	label := ""
	if icon != "" {
		label = icon + " "
	}
	label += motif.Name
	if motif.IsDir {
		label += "/"
	}

	return label
}

func (r *renderer) renderSkipped(motif contract.Motif) string {
	var b strings.Builder

	itemName := "~ " + motif.Name
	if motif.IsDir {
		itemName += "/"
		b.WriteString(r.theme.DirStyle.Render(itemName))
	} else {
		b.WriteString(r.theme.FileStyle.Render(itemName))
	}

	skipReason := fmt.Sprintf("  [skipped: %s -> %s]",
		motif.Placeholder,
		motif.ResolvedPath,
	)
	b.WriteString(r.theme.SkippedStyle.Render(skipReason))

	return b.String()
}

func (r *renderer) renderRoot(motif contract.Motif) string {
	var b strings.Builder

	icon := r.treeIcons[contract.TreeIconRoot]
	if icon != "" {
		b.WriteString(icon)
		b.WriteString(" ")
	}

	b.WriteString(motif.Name)
	if motif.IsDir {
		b.WriteString("/")
	}

	return r.theme.RootStyle.Render(b.String())
}

func (r *renderer) renderDir(motif contract.Motif) string {
	var b strings.Builder

	b.WriteString(r.theme.DirStyle.Render(r.itemLabel(motif)))
	b.WriteString(r.renderActionOrPipeline(motif))

	return b.String()
}

func (r *renderer) renderFile(motif contract.Motif) string {
	var b strings.Builder

	b.WriteString(r.theme.FileStyle.Render(r.itemLabel(motif)))
	b.WriteString(r.renderActionOrPipeline(motif))

	return b.String()
}

func (r *renderer) renderStep(motif contract.Motif) string {
	var b strings.Builder

	if motif.Err != nil {
		b.WriteString(r.theme.ErrorStyle.Render(
			fmt.Sprintf("! %s  %s", motif.ActionName, motif.Err.Error()),
		))
	} else if motif.Skipped {
		skipReason := fmt.Sprintf("  • via %s  [skipped: %s -> %s]",
			motif.ActionName,
			motif.Placeholder,
			motif.ResolvedPath,
		)
		b.WriteString(r.theme.SkippedStyle.Render(skipReason))
	} else {
		b.WriteString(r.theme.ActionStyle.Render("  • via " + motif.ActionName))
		b.WriteString(r.renderExecutionInfo(motif))
	}

	return b.String()
}

func (r *renderer) renderActionOrPipeline(motif contract.Motif) string {
	var b strings.Builder

	if motif.ActionName != "" {
		b.WriteString(r.theme.ActionStyle.Render("  • via " + motif.ActionName))
		b.WriteString(r.renderExecutionInfo(motif))
	} else if motif.PipelineName != "" {
		b.WriteString(r.theme.PipelineStyle.Render("  • via " + motif.PipelineName))
		b.WriteString(r.renderExecutionInfo(motif))
	}

	return b.String()
}

func (r *renderer) renderExecutionInfo(motif contract.Motif) string {
	skipped := lo.Ternary(motif.Skipped, fmt.Sprintf(" %s", r.treeIcons[contract.TreeIconSkipped]), "")
	content := motif.CommandOutput
	if motif.DryRun {
		content = motif.ExecutionString
	}

	if content == "" && skipped == "" {
		return ""
	}

	// brackets are part of the chrome, so use BranchStyle
	return r.theme.BranchStyle.Render(" ["+skipped) +
		r.theme.LandingStripStyle.Render(content) +
		r.theme.BranchStyle.Render("]")
}

func (r *renderer) branchPrefix(motif contract.Motif) string {
	if motif.VisualDepth == 0 {
		return ""
	}

	var b strings.Builder
	//nolint:gosec // branchStack is only modified by updateBranchStack based on motif.VisualDepth.
	for level := 1; level < int(motif.VisualDepth); level++ {
		if level-1 < len(r.branchStack) && r.branchStack[level-1] {
			b.WriteString(r.treeIcons[contract.TreeIconBranchVertical])
			b.WriteString(r.treeIcons[contract.TreeIconBranchIndent])
		} else {
			b.WriteString(
				strings.Repeat(" ",
					len(r.treeIcons[contract.TreeIconBranchVertical])+len(r.treeIcons[contract.TreeIconBranchIndent]),
				),
			)
		}
	}

	isLast := lo.Ternary(motif.IsPipelineStep, motif.IsLastStep && !motif.IsDir, motif.IsLast)
	branchIcon := lo.Ternary(isLast,
		contract.TreeIconBranchLast,
		contract.TreeIconBranchJoint,
	)
	b.WriteString(r.treeIcons[branchIcon])

	return b.String()
}

func (r *renderer) updateBranchStack(motif contract.Motif) {
	if motif.VisualDepth == 0 {
		r.branchStack = nil
		r.previousDepth = motif.VisualDepth
		r.previousIsLast = motif.IsLast
		return
	}

	isLast := lo.Ternary(motif.IsPipelineStep, motif.IsLastStep && !motif.IsDir, motif.IsLast)
	if motif.VisualDepth > r.previousDepth {
		for d := r.previousDepth; d < motif.VisualDepth; d++ {
			r.branchStack = append(r.branchStack, !isLast)
		}
	} else if motif.VisualDepth < r.previousDepth {
		//nolint:gosec // VisualDepth is verified by navigator bounds
		r.branchStack = r.branchStack[:int(motif.VisualDepth)]
	}

	if motif.VisualDepth > 0 {
		//nolint:gosec // VisualDepth is verified by navigator bounds
		r.branchStack[int(motif.VisualDepth)-1] = !isLast
	}

	r.previousDepth = motif.VisualDepth
	r.previousIsLast = motif.IsLast
}

// End renders the closing summary box with traversal counts and elapsed time.
func (r *renderer) End(summary contract.Summary) {
	fileLabel := "Files"
	dirLabel := "Directories"
	skippedLabel := "Skipped"
	elapsedLabel := "Elapsed"

	if summary.Kind == contract.ResumeNavigation {
		fileLabel = "Files (resumed)"
		dirLabel = "Dirs (resumed)"
	}

	errorCount := len(summary.Errors)

	rows := []widget.SummaryRow{
		r.summaryRow(contract.TreeIconFile, fileLabel, fmt.Sprintf("%d", summary.FilesVisited)),
		r.summaryRow(contract.TreeIconDirectory, dirLabel, fmt.Sprintf("%d", summary.DirsVisited)),
		r.summaryRow(contract.TreeIconSkipped, skippedLabel, fmt.Sprintf("%d", summary.Skipped)),
		r.summaryRow(contract.TreeIconError, "Errors", fmt.Sprintf("%d", errorCount)),
		r.summaryRow(contract.TreeIconElapsed, elapsedLabel, widget.FormatDuration(summary.Elapsed)),
	}

	if errorCount > 0 {
		errorStyles := widget.SummaryCellStyles{
			Icon:  r.theme.ErrorStyle,
			Label: r.theme.ErrorStyle,
			Value: r.theme.ErrorStyle,
		}

		rows[len(rows)-2].Styles = &errorStyles

		for _, err := range summary.Errors {
			rows = append(rows, widget.SummaryRow{
				Label:  err.Error(),
				Styles: &errorStyles,
			})
		}
	}

	box := widget.RenderSummary(rows, widget.SummaryStyles{
		Box: r.theme.BoxStyle,
		Default: widget.SummaryCellStyles{
			Icon:  r.theme.SummaryLabelStyle,
			Label: r.theme.SummaryLabelStyle,
			Value: r.theme.SummaryValueStyle,
		},
	})
	_, _ = lipgloss.Fprintln(r.writer, box)
}

func (r *renderer) summaryRow(iconKey, label, value string) widget.SummaryRow {
	return widget.SummaryRow{
		Icon:  r.treeIcons[iconKey],
		Label: label,
		Value: value,
	}
}
