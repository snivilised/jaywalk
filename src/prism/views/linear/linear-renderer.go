// Package flow contains the linear renderer implementation and its
// view-specific options.
//
// Dependency rule: flow imports contract (shared types) and is imported
// by prism root.
package linear

import (
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
	"github.com/snivilised/jaywalk/src/prism/layout"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
	"github.com/snivilised/jaywalk/src/prism/widgets/border"
	"github.com/snivilised/jaywalk/src/prism/widgets/intro"
	"github.com/snivilised/jaywalk/src/prism/widgets/landing"
	"github.com/snivilised/jaywalk/src/prism/widgets/status"
	"github.com/snivilised/jaywalk/src/third/lo"
)

// renderer is the linear scrolling view. Output is written immediately as
// events arrive - no internal buffering.
type renderer struct {
	theme  contract.Theme
	writer io.Writer

	// width is the terminal width used for full-width banner rendering.
	width int

	// banner holds the ANSI shadow banner configuration from the Overture.
	// Nil when the banner is disabled or not configured.
	banner *contract.BannerInfo

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
	dateFmt := overture.DateFormat
	if dateFmt == "" {
		dateFmt = time.RFC1123
	}

	// Store banner config for use in Begin (top) and End (bottom)
	r.banner = overture.Banner

	var b strings.Builder

	// Render ANSI banner at top if position is "top"
	if r.banner != nil && !r.banner.Disable && r.banner.Position == contract.PositionTop {
		r.renderAnsiBanner(&b)
	}

	// Render Top Border with path in brackets (uses contract.Static.Borders)
	topBorderContent := border.RenderTop(overture.Root, r.width, border.Styles{
		BorderStyle: r.theme.BorderStyle,
		PathStyle:   r.theme.RootStyle,
		CornerStyle: r.theme.BorderStyle,
	})
	b.WriteString(topBorderContent)

	// Build caption for intro widget
	caption := overture.Caption
	if overture.Kind == contract.ResumeNavigation && overture.ResumeFrom != "" {
		caption += fmt.Sprintf("  -  from: %s", overture.ResumeFrom)
	}

	// Render intro line using intro widget
	introContent := intro.Render(caption, overture.StartedAt, dateFmt, intro.Styles{
		InfoStyle: r.theme.SummaryValueStyle,
	})

	// Render header text
	header := r.theme.HeaderStyle.Render("jay")
	middle := header + introContent

	// Render row with border caps
	row := layout.NewRow(r.width-4).
		Caps(
			r.theme.BorderStyle.Render("│ "),
			r.theme.BorderStyle.Render(" │"),
		).
		Content(middle)
	row.RenderTo(&b)
	b.WriteString("\n")

	// Render bottom border using contract.Static.Borders
	bottomDashes := strings.Repeat("─", max(0, r.width-7))
	b.WriteString(r.theme.BorderStyle.Render(
		contract.Static.Borders.BottomLeft + bottomDashes + contract.Static.Borders.BottomRightCorner,
	))
	b.WriteString("\n")

	_, _ = lipgloss.Fprintln(r.writer, b.String())
}

// renderAnsiBanner renders the static ANSI shadow banner. The gradient is
// applied based on the randomized aspects but does not animate (Offset stays at 0).
func (r *renderer) renderAnsiBanner(b *strings.Builder) {
	if r.banner == nil || r.banner.Gradient == nil {
		return
	}

	// Create a static gradient state (Offset=0, never animated)
	state := effects.NewGradientState()
	state.TotalSteps = r.banner.Gradient.Steps

	// Convert contract.BannerAspects to banner.Aspects
	aspects := banner.Aspects{
		Orientation: banner.Orientation(r.banner.Aspects.Orientation),
		Banding:     banner.Banding(r.banner.Aspects.Banding),
		Unity:       banner.Unity(r.banner.Aspects.Unity),
		FixedEnd:    banner.FixedEnd(r.banner.Aspects.FixedEnd),
	}

	out := banner.Render(banner.Config{
		Width:   r.width,
		Justify: r.banner.Justify,
	}, banner.Styles{}, banner.Effect{
		Gradient: r.banner.Gradient,
		State:    state,
		Aspects:  aspects,
	})
	if out != "" {
		b.WriteString(out)
	}
}

// Show renders a single Motif immediately to the output writer.
func (r *renderer) Show(motif contract.Motif) {
	line := r.RenderLine(motif)
	_, _ = lipgloss.Fprint(r.writer, line)
}

// RenderLine produces the rendered string for a single Motif. The
// returned string always ends with a trailing newline. Callers that
// drive output line-by-line (eg the porthole view's content buffer)
// use this directly; Show is the io.Writer-bound form.
func (r *renderer) RenderLine(motif contract.Motif) string {
	prefix := r.branchPrefix(motif)
	depth := r.theme.BranchStyle.Render(prefix)

	var name string

	switch {
	case motif.Err != nil:
		name = r.theme.ErrorStyle.Render(
			fmt.Sprintf("! %s  %s", motif.Name, motif.Err.Error()),
		)

	case motif.Skipped:
		name = r.skipped(motif)

	case motif.IsPipelineStep:
		name = r.step(motif)

	case motif.Depth == 0:
		name = r.root(motif)

	case motif.IsDir:
		name = r.dir(motif)

	default:
		name = r.file(motif)
	}

	var b strings.Builder
	b.WriteString(depth)
	b.WriteString(name)
	b.WriteByte('\n')

	r.updateBranchStack(motif)
	return b.String()
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

func (r *renderer) skipped(motif contract.Motif) string {
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

func (r *renderer) root(motif contract.Motif) string {
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

func (r *renderer) dir(motif contract.Motif) string {
	var b strings.Builder

	b.WriteString(r.theme.DirStyle.Render(r.itemLabel(motif)))
	b.WriteString(r.task(motif))

	return b.String()
}

func (r *renderer) file(motif contract.Motif) string {
	var b strings.Builder

	b.WriteString(r.theme.FileStyle.Render(r.itemLabel(motif)))
	b.WriteString(r.task(motif))

	return b.String()
}

func (r *renderer) step(motif contract.Motif) string {
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
		b.WriteString(r.execution(motif))
	}

	return b.String()
}

func (r *renderer) task(motif contract.Motif) string {
	var b strings.Builder

	if motif.ActionName != "" {
		b.WriteString(r.theme.ActionStyle.Render("  • via " + motif.ActionName))
		b.WriteString(r.execution(motif))
	} else if motif.PipelineName != "" {
		b.WriteString(r.theme.PipelineStyle.Render("  • via " + motif.PipelineName))
		b.WriteString(r.execution(motif))
	}

	return b.String()
}

func (r *renderer) execution(motif contract.Motif) string {
	skippedIcon := ""
	if motif.Skipped {
		skippedIcon = r.treeIcons[contract.TreeIconSkipped]
	}
	return landing.Render(landing.Config{
		CommandOutput:   motif.CommandOutput,
		ExecutionString: motif.ExecutionString,
		DryRun:          motif.DryRun,
		SkippedIcon:     skippedIcon,
	}, landing.Styles{
		BranchStyle:       r.theme.BranchStyle,
		LandingStripStyle: r.theme.LandingStripStyle,
	})
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

// End renders the closing status row with traversal counts and elapsed time.
func (r *renderer) End(summ contract.Summary) {
	errorCount := len(summ.Errors)

	// Render top border (empty content - just corner decorations)
	topBorder := border.RenderTop("", r.width, border.Styles{
		BorderStyle: r.theme.BorderStyle,
		CornerStyle: r.theme.BorderStyle,
	})
	_, _ = lipgloss.Fprint(r.writer, topBorder)

	statusStyles := status.Styles{
		TreeIcons:         r.treeIcons,
		SummaryLabelStyle: r.theme.SummaryLabelStyle,
		SummaryValueStyle: r.theme.SummaryValueStyle,
		ErrorStyle:        r.theme.ErrorStyle,
		BorderStyle:       r.theme.BorderStyle,
	}

	fields := status.FieldSelectors{
		ShowFiles:    true,
		ShowDirs:     true,
		ShowErrors:   true,
		ShowSkipped:  true,
		ShowProgress: false,
		ShowComplete: false,
		ShowElapsed:  true,
	}

	statusRow := status.Render(status.Config{
		Files:   int(summ.FilesVisited),
		Dirs:    int(summ.DirsVisited),
		Errors:  errorCount,
		Skipped: int(summ.Skipped),
		Elapsed: summ.Elapsed,
	}, statusStyles, fields, r.width)

	_, _ = lipgloss.Fprintln(r.writer, statusRow)

	// Render bottom border
	bottomBorder := border.RenderBottom(r.width, border.Styles{
		BorderStyle: r.theme.BorderStyle,
	})
	_, _ = lipgloss.Fprintln(r.writer, bottomBorder)

	// Render ANSI banner at bottom if position is "bottom"
	if r.banner != nil && !r.banner.Disable && r.banner.Position == contract.PositionBottom {
		var b strings.Builder
		r.renderAnsiBanner(&b)
		_, _ = lipgloss.Fprintln(r.writer, b.String())
	}
}
