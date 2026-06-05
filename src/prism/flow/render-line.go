package flow

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

// RenderLine produces the rendered string for a single node given its data.
// The returned string always ends with a trailing newline. branchStack is
// the ancestor continuation state carried across calls; pass nil for the
// first call.
//
// Pipeline steps set isPipelineStep=true and pass visualDepth (the parent
// node depth + 1). The isLastStep flag determines the branch glyph for
// the final step (└── vs ├──).
//
// bodyWidth, when > 0, right-justifies the landing strip within the given
// visible width. Pass 0 to keep the legacy inline behaviour where the
// strip is rendered immediately after the action/pipeline name.
func RenderLine(
	path, name string,
	isDir bool,
	depth uint,
	actionName, pipelineName string,
	commandOutput, executionString string,
	dryRun bool,
	err error,
	isLast bool,
	isPipelineStep bool,
	isLastStep bool,
	visualDepth uint,
	branchStack []bool,
	bodyWidth uint,
	theme contract.Theme,
	activityFrame string,
) RenderLineResult {
	// For pipeline steps the visual depth is one deeper than the parent
	// node and the branch icon uses isLastStep (not isLast).
	prefixDepth := visualDepth
	prefixIsLast := isLast
	if isPipelineStep {
		prefixIsLast = isLastStep && !isDir
	}

	prefix := buildBranchPrefix(prefixDepth, prefixIsLast, branchStack, theme.TreeIcons)
	branchStyle := theme.BranchStyle

	prefixWidth := lipglossWidth(prefix, branchStyle)

	var nameStr string

	switch {
	case isPipelineStep:
		nameStr = renderStep(actionName, commandOutput, executionString, dryRun, err, bodyWidth, prefixWidth, theme, activityFrame)

	case err != nil:
		nameStr = theme.ErrorStyle.Render(
			fmt.Sprintf("! %s  %s", name, err.Error()),
		)

	case isDir:
		nameStr = renderDir(name, actionName, pipelineName, commandOutput, executionString, dryRun, bodyWidth, prefixWidth, theme, activityFrame)

	default:
		nameStr = renderFile(name, actionName, pipelineName, commandOutput, executionString, dryRun, bodyWidth, prefixWidth, theme, activityFrame)
	}

	var b strings.Builder
	b.WriteString(branchStyle.Render(prefix))
	b.WriteString(nameStr)
	b.WriteByte('\n')

	return RenderLineResult{
		Line:        b.String(),
		BranchStack: updateBranchStack(visualDepth, prefixIsLast, branchStack),
	}
}

func renderStep(actionName, commandOutput, executionString string, dryRun bool, err error, bodyWidth, prefixWidth uint, theme contract.Theme, activityFrame string) string {
	var b strings.Builder

	if err != nil {
		b.WriteString(theme.ErrorStyle.Render(
			fmt.Sprintf("! %s  %s", actionName, err.Error()),
		))
	} else {
		action := theme.ActionStyle.Render("  • via " + actionName)
		b.WriteString(action)
		consumed := prefixWidth + lipglossWidth(action, theme.ActionStyle)
		if activityFrame != "" {
			b.WriteString(activityFrame)
			consumed += lipglossWidth(activityFrame, lipgloss.Style{})
		}
		b.WriteString(landing.Render(landing.Config{
			CommandOutput:   commandOutput,
			ExecutionString: executionString,
			DryRun:          dryRun,
			Width:           rightJustifyWidth(bodyWidth, consumed),
		}, landing.Styles{
			BranchStyle:       theme.BranchStyle,
			LandingStripStyle: theme.LandingStripStyle,
		}))
	}

	return b.String()
}

func buildBranchPrefix(depth uint, isLast bool, branchStack []bool, treeIcons contract.TreeIcons) string {
	if depth == 0 {
		return ""
	}

	vertW := len([]rune(treeIcons[contract.TreeIconBranchVertical]))
	indentW := len([]rune(treeIcons[contract.TreeIconBranchIndent]))
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

func renderDir(name, actionName, pipelineName, commandOutput, executionString string, dryRun bool, bodyWidth, prefixWidth uint, theme contract.Theme, activityFrame string) string {
	var b strings.Builder

	itemLabel := buildItemLabel(name, true, theme.TreeIcons)
	dirLabel := theme.DirStyle.Render(itemLabel)
	b.WriteString(dirLabel)
	b.WriteString(renderTask(actionName, pipelineName, commandOutput, executionString, dryRun, bodyWidth, prefixWidth+lipglossWidth(dirLabel, theme.DirStyle), theme, activityFrame))

	return b.String()
}

func renderFile(name, actionName, pipelineName, commandOutput, executionString string, dryRun bool, bodyWidth, prefixWidth uint, theme contract.Theme, activityFrame string) string {
	var b strings.Builder

	itemLabel := buildItemLabel(name, false, theme.TreeIcons)
	fileLabel := theme.FileStyle.Render(itemLabel)
	b.WriteString(fileLabel)
	b.WriteString(renderTask(actionName, pipelineName, commandOutput, executionString, dryRun, bodyWidth, prefixWidth+lipglossWidth(fileLabel, theme.FileStyle), theme, activityFrame))

	return b.String()
}

func renderTask(actionName, pipelineName string, commandOutput, executionString string, dryRun bool, bodyWidth, prefixWidth uint, theme contract.Theme, activityFrame string) string {
	var b strings.Builder

	if actionName != "" {
		action := theme.ActionStyle.Render("  • via " + actionName)
		b.WriteString(action)
		prefixWidth += lipglossWidth(action, theme.ActionStyle)
	} else if pipelineName != "" {
		pipeline := theme.PipelineStyle.Render("  • via " + pipelineName)
		b.WriteString(pipeline)
		prefixWidth += lipglossWidth(pipeline, theme.PipelineStyle)
	}

	if activityFrame != "" {
		b.WriteString(activityFrame)
		prefixWidth += lipglossWidth(activityFrame, lipgloss.Style{})
	}

	b.WriteString(landing.Render(landing.Config{
		CommandOutput:   commandOutput,
		ExecutionString: executionString,
		DryRun:          dryRun,
		Width:           rightJustifyWidth(bodyWidth, prefixWidth),
	}, landing.Styles{
		BranchStyle:       theme.BranchStyle,
		LandingStripStyle: theme.LandingStripStyle,
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
