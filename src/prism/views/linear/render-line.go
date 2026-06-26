package linear

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/landing"
	"github.com/snivilised/jaywalk/src/third/lo"
)

// RenderLineResult holds the rendered line and the updated branch stack
// for use by the next call. Callers must pass the returned BranchStack
// back into the next RenderLine call to maintain correct tree glyphs.
type RenderLineResult struct {
	Line        string
	BranchStack []bool
}

// LineParams holds all parameters needed for RenderLine.
type LineParams struct {
	contract.NodeParams
	contract.RenderParams
	BranchStack []bool
}

// renderParams bundles the rendering-specific subset of LineParams
// that the internal renderDir, renderFile, renderTask and renderStep
// functions need. It embeds NodeParams so the node data is available
// without repeating the 14 fields in each function signature.
type renderParams struct {
	contract.NodeParams
	contract.RenderParams
	PrefixWidth uint
}

// RenderLine produces the rendered string for a single node given its data.
// The returned string always ends with a trailing newline. BranchStack is
// the ancestor continuation state carried across calls; pass nil for the
// first call.
//
// Pipeline steps set IsPipelineStep=true and pass VisualDepth (the parent
// node depth + 1). The IsLastStep flag determines the branch glyph for
// the final step (└── vs ├──).
//
// BodyWidth, when > 0, right-justifies the landing strip within the given
// visible width. Pass 0 to keep the legacy inline behaviour where the
// strip is rendered immediately after the action/pipeline name.
func RenderLine(p LineParams) RenderLineResult {
	// For pipeline steps the visual depth is one deeper than the parent
	// node and the branch icon uses IsLastStep (not IsLast).
	prefixDepth := p.VisualDepth
	prefixIsLast := p.IsLast
	if p.IsPipelineStep {
		prefixIsLast = p.IsLastStep && !p.IsDir
	}

	prefix := buildBranchPrefix(prefixDepth, prefixIsLast, p.BranchStack, p.Theme.TreeIcons)
	branchStyle := p.Theme.BranchStyle

	prefixWidth := lipglossWidth(prefix, branchStyle)

	rp := renderParams{
		NodeParams:   p.NodeParams,
		RenderParams: p.RenderParams,
		PrefixWidth:  prefixWidth,
	}

	var nameStr string

	switch {
	case p.IsPipelineStep:
		nameStr = renderStep(rp)

	case p.Err != nil:
		nameStr = p.Theme.ErrorStyle.Render(
			fmt.Sprintf("! %s  %s", p.Name, p.Err.Error()),
		)

	case p.IsDir:
		nameStr = renderDir(rp)

	default:
		nameStr = renderFile(rp)
	}

	var b strings.Builder
	b.WriteString(branchStyle.Render(prefix))
	b.WriteString(nameStr)
	b.WriteByte('\n')

	return RenderLineResult{
		Line:        b.String(),
		BranchStack: updateBranchStack(p.VisualDepth, prefixIsLast, p.BranchStack),
	}
}

func renderStep(rp renderParams) string {
	var b strings.Builder

	if rp.Err != nil {
		b.WriteString(rp.Theme.ErrorStyle.Render(
			fmt.Sprintf("! %s  %s", rp.ActionName, rp.Err.Error()),
		))
	} else {
		action := rp.Theme.ActionStyle.Render("  • via " + rp.ActionName)
		b.WriteString(action)
		consumed := rp.PrefixWidth + lipglossWidth(action, rp.Theme.ActionStyle)
		if rp.ActivityFrame != "" {
			b.WriteString(rp.ActivityFrame)
			consumed += lipglossWidth(rp.ActivityFrame, lipgloss.Style{})
		}
		b.WriteString(landing.Render(landing.Config{
			CommandOutput:   rp.CommandOutput,
			ExecutionString: rp.ExecutionString,
			DryRun:          rp.DryRun,
			Width:           rightJustifyWidth(rp.BodyWidth, consumed),
		}, landing.Styles{
			BranchStyle:       rp.Theme.BranchStyle,
			LandingStripStyle: rp.Theme.LandingStripStyle,
		}))
	}

	return b.String()
}

func buildBranchPrefix(depth uint, isLast bool, branchStack []bool, treeIcons contract.TreeIcons) string {
	if depth == 0 {
		return ""
	}

	vertW := lipgloss.Width(treeIcons[contract.TreeIconBranchVertical])
	indentW := lipgloss.Width(treeIcons[contract.TreeIconBranchIndent])
	colW := vertW + indentW

	var b strings.Builder
	for level := 1; level < int(depth); level++ { //nolint:gosec // depth is always a small traversal depth
		if level-1 < len(branchStack) && branchStack[level-1] {
			b.WriteString(treeIcons[contract.TreeIconBranchVertical])
			b.WriteString(treeIcons[contract.TreeIconBranchIndent])
		} else {
			b.WriteString(strings.Repeat(" ", colW))
		}
	}

	branchIcon := lo.Ternary(isLast,
		contract.TreeIconBranchLast,
		contract.TreeIconBranchJoint,
	)
	b.WriteString(treeIcons[branchIcon])

	return b.String()
}

// updateBranchStack updates the ancestor continuation state after
// rendering a node. It mirrors the highway renderer's logic.
func updateBranchStack(depth uint, isLast bool, stack []bool) []bool {
	if depth == 0 {
		return nil
	}

	prevDepth := uint(len(stack))
	if depth > prevDepth {
		for d := prevDepth; d < depth; d++ {
			stack = append(stack, !isLast)
		}
	} else if depth < prevDepth {
		stack = stack[:depth]
	}

	if depth > 0 {
		stack[int(depth)-1] = !isLast //nolint:gosec // depth > 0 checked, always a small value
	}

	return stack
}

func renderDir(rp renderParams) string {
	var b strings.Builder

	itemLabel := buildItemLabel(rp.Name, true, rp.Theme.TreeIcons)
	dirLabel := rp.Theme.DirStyle.Render(itemLabel)
	b.WriteString(dirLabel)
	rp.PrefixWidth += lipglossWidth(dirLabel, rp.Theme.DirStyle)
	b.WriteString(renderTask(rp))

	return b.String()
}

func renderFile(rp renderParams) string {
	var b strings.Builder

	itemLabel := buildItemLabel(rp.Name, false, rp.Theme.TreeIcons)
	fileLabel := rp.Theme.FileStyle.Render(itemLabel)
	b.WriteString(fileLabel)
	rp.PrefixWidth += lipglossWidth(fileLabel, rp.Theme.FileStyle)
	b.WriteString(renderTask(rp))

	return b.String()
}

func renderTask(rp renderParams) string {
	var b strings.Builder

	if rp.ActionName != "" {
		action := rp.Theme.ActionStyle.Render("  • via " + rp.ActionName)
		b.WriteString(action)
		rp.PrefixWidth += lipglossWidth(action, rp.Theme.ActionStyle)
	} else if rp.PipelineName != "" {
		pipeline := rp.Theme.PipelineStyle.Render("  • via " + rp.PipelineName)
		b.WriteString(pipeline)
		rp.PrefixWidth += lipglossWidth(pipeline, rp.Theme.PipelineStyle)
	}

	if rp.ActivityFrame != "" {
		b.WriteString(rp.ActivityFrame)
		rp.PrefixWidth += lipglossWidth(rp.ActivityFrame, lipgloss.Style{})
	}

	b.WriteString(landing.Render(landing.Config{
		CommandOutput:   rp.CommandOutput,
		ExecutionString: rp.ExecutionString,
		DryRun:          rp.DryRun,
		Width:           rightJustifyWidth(rp.BodyWidth, rp.PrefixWidth),
	}, landing.Styles{
		BranchStyle:       rp.Theme.BranchStyle,
		LandingStripStyle: rp.Theme.LandingStripStyle,
	}))

	return b.String()
}

// rightJustifyWidth returns the width to pass to landing.Render for
// right-justification, or 0 when justification is not in effect. The
// bodyWidth includes the branch prefix width (which is rendered
// before the item), so the strip needs to be padded to the *remainder*
// of the body after subtracting the prefix.
func rightJustifyWidth(bodyWidth, prefixWidth uint) int {
	if bodyWidth == 0 {
		return 0
	}
	if bodyWidth <= prefixWidth {
		return 0
	}
	return int(bodyWidth - prefixWidth) //nolint:gosec // bodyWidth > prefixWidth checked, result is non-negative
}

// lipglossWidth returns the visible width of an already-styled string
// (i.e. the style is applied, so ANSI codes are stripped before the
// measurement). For an unstyled string the style argument is ignored.
func lipglossWidth(s string, _ lipgloss.Style) uint {
	w := lipgloss.Width(s)
	if w < 0 {
		return 0
	}
	return uint(w)
}

func buildItemLabel(name string, isDir bool, treeIcons contract.TreeIcons) string {
	var label strings.Builder
	if icon := treeIcons[contract.TreeIconFile]; icon != "" && !isDir {
		label.WriteString(icon)
		label.WriteString(" ")
	}
	if icon := treeIcons[contract.TreeIconDirectory]; icon != "" && isDir {
		label.WriteString(icon)
		label.WriteString(" ")
	}
	label.WriteString(name)
	if isDir {
		label.WriteString("/")
	}
	return label.String()
}
