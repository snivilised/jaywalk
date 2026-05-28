package contract

import (
	"fmt"
	"image/color"
	"io"

	"charm.land/lipgloss/v2"
)

func defaultTreeIcons() TreeIcons {
	return TreeIcons{
		TreeIconRoot:           "✻",
		TreeIconDirectory:      "📁",
		TreeIconFile:           "🔖",
		TreeIconElapsed:        "⏰",
		TreeIconSkipped:        "⛔️",
		TreeIconError:          "🚫",
		TreeIconBranchVertical: "│",
		TreeIconBranchJoint:    "├── ",
		TreeIconBranchLast:     "└── ",
		TreeIconBranchIndent:   "  ",
	}
}

// Theme holds all lipgloss styles used by a renderer. Constructed once
// via NewTheme and injected into the renderer. No package-level style
// variables exist anywhere in prism.
type Theme struct {
	// DirStyle is applied to directory name text in Show output.
	DirStyle lipgloss.Style

	// FileStyle is applied to file name text in Show output.
	FileStyle lipgloss.Style

	// RootStyle is applied to the root node text in tree style output.
	RootStyle lipgloss.Style

	// BranchStyle is applied to tree branch characters in tree style output.
	BranchStyle lipgloss.Style

	// ActionStyle is applied to action names shown alongside nodes.
	ActionStyle lipgloss.Style

	// PipelineStyle is applied to pipeline names shown alongside nodes.
	PipelineStyle lipgloss.Style

	// LandingStripStyle is applied to the landing strip content (execution string or output).
	LandingStripStyle lipgloss.Style

	// SkippedStyle is applied to nodes whose action was skipped.
	SkippedStyle lipgloss.Style

	// BoxStyle is the outer container style for the closing summary.
	BoxStyle lipgloss.Style

	// SummaryLabelStyle is applied to label text inside the summary.
	SummaryLabelStyle lipgloss.Style

	// SummaryValueStyle is applied to value text inside the summary.
	SummaryValueStyle lipgloss.Style

	// ErrorStyle is applied to error entries in Show output and the summary.
	ErrorStyle lipgloss.Style

	// MutedStyle is applied to secondary or de-emphasised text.
	MutedStyle lipgloss.Style

	// ProgressStyle is applied to progress indicators.
	ProgressStyle lipgloss.Style

	// WorkerStyle is applied to active concurrent worker indicators.
	WorkerStyle lipgloss.Style

	// WorkerIdleStyle is applied to idle concurrent worker indicators.
	WorkerIdleStyle lipgloss.Style

	// LaneHeaderStyle is applied to per-worker lane identity headers.
	LaneHeaderStyle lipgloss.Style

	// TreeIcons are the glyphs used to render the branch tree in linear
	// views such as the linear renderer.
	TreeIcons TreeIcons

	// HeaderStyle is the style of the title text in the highway view.
	HeaderStyle lipgloss.Style

	// FrameStyle is the style of animation frames in the highway view.
	FrameStyle lipgloss.Style

	// BorderStyle is the style of box-drawing characters in the highway view.
	BorderStyle lipgloss.Style

	// BarFilledStyle is the style of filled square-bar glyphs.
	BarFilledStyle lipgloss.Style

	// BarEmptyStyle is the style of empty square-bar glyphs.
	BarEmptyStyle lipgloss.Style

	// HighlightGradients holds resolved colour endpoints for animation
	// overlays. Keyed by gradient name from the theme file. Only populated
	// when the user has configured highlights.gradients in their theme.
	HighlightGradients map[string]ResolvedGradient

	// GradientCaches caches the pre-computed interpolated colour steps per
	// gradient name, keyed by gradient name. Empty means compute on-demand
	// using InterpolateBetween(). Only used for highway animation overlays.
	GradientCaches map[string][]color.Color `json:"-"`

	// HighlightsComponents maps a component name (e.g. "activity-control")
	// to a gradient name. Populated from highlights.components in the
	// theme file. A component with no entry falls back to its static style.
	HighlightsComponents map[string]string
}

// GradientFor looks up the resolved gradient for the given component name.
// Returns the gradient and true if a gradient is configured; returns a zero
// ResolvedGradient and false if the component has no mapping or the
// referenced gradient does not exist. Callers fall back to their static
// style when false.
func (t Theme) GradientFor(component string) (ResolvedGradient, bool) {
	gradName, ok := t.HighlightsComponents[component]
	if !ok {
		return ResolvedGradient{}, false
	}
	grad, ok := t.HighlightGradients[gradName]
	if !ok {
		return ResolvedGradient{}, false
	}
	return grad, true
}

// NewTheme constructs a Theme from the given Palette. The colour profile
// is detected from the environment via colorprofile.Detect so that
// lipgloss can select the best available colour tier (TrueColor,
// ANSI-256, ANSI-16, or none) at style construction time.
//
// Returns an error if any palette entry contains an unrecognised ANSI-16
// colour name. Bootstrap should treat this as a startup failure.
func NewTheme(palette Palette, writer io.Writer) (Theme, error) {
	// Detect the terminal colour profile from the writer's environment.
	// This determines whether TrueColor, ANSI-256, ANSI-16, or plain
	// text output is used. colorprofile.Detect handles NO_COLOR, isatty,
	// COLORTERM and TERM inspection automatically.

	resolve := func(sc SemanticColour, field string) (color.Color, error) {
		ansi, ansi256, trueCol, err := sc.Resolve()
		if err != nil {
			return nil, fmt.Errorf("palette.%s: %w", field, err)
		}

		switch {
		case trueCol != nil:
			return trueCol, nil
		case ansi256 != nil:
			return ansi256, nil
		default:
			return ansi, nil
		}
	}

	dir, err := resolve(palette.Directory, "directory")
	if err != nil {
		return Theme{}, err
	}

	file, err := resolve(palette.File, "file")
	if err != nil {
		return Theme{}, err
	}

	root, err := resolve(palette.Root, "root")
	if err != nil {
		return Theme{}, err
	}

	branch, err := resolve(palette.Branch, "branch")
	if err != nil {
		return Theme{}, err
	}

	action, err := resolve(palette.Action, "action")
	if err != nil {
		return Theme{}, err
	}

	pipeline, err := resolve(palette.Pipeline, "pipeline")
	if err != nil {
		return Theme{}, err
	}

	landingStrip, err := resolve(palette.LandingStrip, "landing-strip")
	if err != nil {
		return Theme{}, err
	}

	skipped, err := resolve(palette.Skipped, "skipped")
	if err != nil {
		return Theme{}, err
	}

	errorCol, err := resolve(palette.Error, "error")
	if err != nil {
		return Theme{}, err
	}

	muted, err := resolve(palette.Muted, "muted")
	if err != nil {
		return Theme{}, err
	}

	progress, err := resolve(palette.Progress, "progress")
	if err != nil {
		return Theme{}, err
	}

	BoxBorder, err := resolve(palette.BoxBorder, "box-border")
	if err != nil {
		return Theme{}, err
	}

	summaryLabel, err := resolve(palette.SummaryLabel, "summary-label")
	if err != nil {
		return Theme{}, err
	}

	summaryValue, err := resolve(palette.SummaryValue, "summary-value")
	if err != nil {
		return Theme{}, err
	}

	worker, err := resolve(palette.Worker, "worker")
	if err != nil {
		return Theme{}, err
	}

	workerIdle, err := resolve(palette.WorkerIdle, "worker-idle")
	if err != nil {
		return Theme{}, err
	}

	laneHeader, err := resolve(palette.LaneHeader, "lane-header")
	if err != nil {
		return Theme{}, err
	}

	header, err := resolve(palette.Header, "header")
	if err != nil {
		return Theme{}, err
	}

	frame, err := resolve(palette.Frame, "frame")
	if err != nil {
		return Theme{}, err
	}

	border, err := resolve(palette.Border, "border")
	if err != nil {
		return Theme{}, err
	}

	barFilled, err := resolve(palette.BarFilled, "bar-filled")
	if err != nil {
		return Theme{}, err
	}

	barEmpty, err := resolve(palette.BarEmpty, "bar-empty")
	if err != nil {
		return Theme{}, err
	}

	treeIcons := make(TreeIcons)
	for k, v := range defaultTreeIcons() {
		treeIcons[k] = v
	}
	for k, v := range palette.TreeIcons {
		if v != "" {
			treeIcons[k] = v
		}
	}

	// Resolve gradients (silently skip entries with neither hi nor lo).
	highlightGradients := make(map[string]ResolvedGradient)
	highlightCaches := make(map[string][]color.Color)
	for name, gd := range palette.Highlights.Gradients {
		if gd.Hi == nil && gd.Lo == nil {
			continue
		}
		var hiCol, loCol color.Color
		if gd.Hi != nil {
			hiCol, err = resolve(*gd.Hi, "highlights.gradients."+name+".hi")
			if err != nil {
				return Theme{}, err
			}
		}
		if gd.Lo != nil {
			loCol, err = resolve(*gd.Lo, "highlights.gradients."+name+".lo")
			if err != nil {
				return Theme{}, err
			}
		}
		if gd.Hi == nil && gd.Lo != nil {
			hiCol = deriveBrighter(loCol)
		} else if gd.Hi != nil && gd.Lo == nil {
			loCol = deriveDimmed(hiCol)
		}
		steps := gd.Steps
		if steps <= 0 {
			steps = DefaultStepCount() // default to 8-step gradient when not specified
		}
		highlightGradients[name] = ResolvedGradient{
			Steps: steps,
			Hi:    hiCol,
			Lo:    loCol,
		}
		// Pre-compute gradient steps and cache them (on-demand when empty)
		if steps > 0 {
			cached := InterpolateBetween(hiCol, loCol, steps)
			highlightCaches[name] = cached
		} else {
			// Cache nil to indicate compute on-demand during rendering
			highlightCaches[name] = nil
		}
	}
	// Store component-to-gradient bindings (may be empty).
	highlightsComponents := make(map[string]string)
	for k, v := range palette.Highlights.Components {
		highlightsComponents[k] = v
	}

	return Theme{
		DirStyle: lipgloss.NewStyle().
			Foreground(dir).
			Bold(true),

		FileStyle: lipgloss.NewStyle().
			Foreground(file),

		ActionStyle: lipgloss.NewStyle().
			Foreground(action),

		PipelineStyle: lipgloss.NewStyle().
			Foreground(pipeline),

		LandingStripStyle: lipgloss.NewStyle().
			Foreground(landingStrip),

		SkippedStyle: lipgloss.NewStyle().
			Foreground(skipped).
			Faint(true),

		BoxStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BoxBorder).
			Padding(0, 2).
			MarginTop(1),

		SummaryLabelStyle: lipgloss.NewStyle().
			Foreground(summaryLabel).
			Bold(true).
			Width(16),

		SummaryValueStyle: lipgloss.NewStyle().
			Foreground(summaryValue),

		ErrorStyle: lipgloss.NewStyle().
			Foreground(errorCol).
			Bold(true),

		MutedStyle: lipgloss.NewStyle().
			Foreground(muted).
			Faint(true),

		ProgressStyle: lipgloss.NewStyle().
			Foreground(progress),

		WorkerStyle: lipgloss.NewStyle().
			Foreground(worker).
			Bold(true),

		WorkerIdleStyle: lipgloss.NewStyle().
			Foreground(workerIdle).
			Faint(true),

		LaneHeaderStyle: lipgloss.NewStyle().
			Foreground(laneHeader).
			Bold(true),

		RootStyle: lipgloss.NewStyle().
			Foreground(root),

		BranchStyle: lipgloss.NewStyle().
			Foreground(branch),

		HeaderStyle: lipgloss.NewStyle().
			Foreground(header).
			Bold(true),

		FrameStyle: lipgloss.NewStyle().
			Foreground(frame),

		BorderStyle: lipgloss.NewStyle().
			Foreground(border),

		BarFilledStyle: lipgloss.NewStyle().
			Foreground(barFilled),

		BarEmptyStyle: lipgloss.NewStyle().
			Foreground(barEmpty),

		TreeIcons:            treeIcons,
		HighlightGradients:   highlightGradients,
		GradientCaches:       highlightCaches,
		HighlightsComponents: highlightsComponents,
	}, nil
}
