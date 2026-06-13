package ui

import (
	"math/rand/v2"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/term"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
	"github.com/snivilised/jaywalk/src/prism/movies"
	"github.com/snivilised/jaywalk/src/prism/views/linear"
	"github.com/snivilised/jaywalk/src/prism/views/porthole"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
)

// portholePresenter is the presenter for the porthole bubbletea view.
// It translates report events into scroll.ContentLineMsg and sends them
// to the model via the program. The view renders a single stream of
// content lines in a viewport with optional scrollbar gutter.
type portholePresenter struct {
	presenter
	cfg PortholeConfig

	branchStack []bool
	// bodyWidth is the usable column count inside the left/right
	// borders and the scrollbar gutter. It is used to right-justify
	// the landing strip produced by linear.RenderLine.
	bodyWidth uint
}

func newPortholePresenter(palette contract.Palette, pCfg PortholeConfig) (report.Presenter, error) {
	theme, err := contract.NewTheme(palette, os.Stdout)
	if err != nil {
		return nil, err
	}
	return &portholePresenter{presenter: presenter{theme: theme}, cfg: pCfg}, nil
}

// OnBegin creates the bubbletea model and starts the program.
func (p *portholePresenter) OnBegin(e *report.BeginEvent) {
	// Cache the header info for introspection and for use below when
	// building the OvertureMsg.
	p.header = e.Header

	// Derive noRecurse from cascade display. The two are mutually exclusive
	// at the flag-binding layer, so only one of 🔒/depth can be present.
	p.noRecurse = strings.Contains(p.header.CascadeDisplay, contract.Static.Emoji.Padlock)

	p.bannerInfo = p.buildBannerInfo()

	// Query the terminal size so landing strips can be right-justified
	// to the same column the model's viewport will use. The same
	// formula (window - gutter - 2 borders) is used in the view's
	// renderBody, so the strip always lands at the right edge.
	if w, _, err := term.GetSize(0); err == nil && w > 0 {
		bw := w - 1 - 2
		if bw > 0 {
			p.bodyWidth = uint(bw)
		}
	}
	if p.bodyWidth == 0 {
		p.bodyWidth = 80
	}

	model := porthole.NewModel(contract.NewModelParams{
		RootPath:  e.Root,
		MaxDepth:  p.maxDepth,
		Theme:     p.theme,
		NoRecurse: p.noRecurse,
	})
	model.SetWindowSizeCallback(func(bw uint) {
		p.bodyWidth = bw
	})

	// Pick a random animation for the activity widget.
	movies.RegisterAll()
	names := movies.ExpandNames(p.cfg.SpinnerNames)
	if len(names) == 0 {
		names = []string{movies.SpinnerTypeWave}
	}
	name := names[rand.IntN(len(names))] //nolint:gosec // selecting a random animation, not for security
	if def, ok := movies.Lookup(name); ok {
		var grad *contract.ResolvedGradient
		if g, has := p.theme.GradientFor(contract.GradientComponentActivity); has && g.Steps > 0 {
			grad = &contract.ResolvedGradient{Steps: g.Steps, Hi: g.Hi, Lo: g.Lo}
		}
		model.SetActivity(def.Frames, grad)
	}

	p.program = tea.NewProgram(&model)

	// Initialize done channel before starting the goroutine that closes it.
	if p.done == nil {
		p.done = make(chan struct{})
	}

	go func() {
		_, _ = p.program.Run()
		close(p.done)
	}()

	// Monitor for premature bubbletea exit (e.g., user pressed Ctrl+C).
	if e.Cancel != nil {
		go func() {
			<-p.done
			e.Cancel()
		}()
	}

	p.program.Send(porthole.OvertureMsg{
		OvertureMsg: contract.OvertureMsg{
			Root:              e.Root,
			Caption:           e.Caption,
			SubscriptionLabel: subscriptionLabelFor(e.Subscription),
			StartedAt:         e.StartedAt,
			DateFormat:        e.DateFormat,
			PipelineName:      e.PipelineName,
			Header:            p.header,
			FlagsRowPosition:  contract.PositionBottom,
		},
		Banner: p.bannerInfo,
	})

	if p.totalFiles > 0 || p.totalDirs > 0 {
		p.program.Send(porthole.CensusMsg{
			TotalFiles: p.totalFiles,
			TotalDirs:  p.totalDirs,
			MaxDepth:   p.maxDepth,
		})
	}
}

func (p *portholePresenter) OnNodeEvent(e *report.NeutralEvent) {
	depth := uint(e.Node.Extension.Depth)
	visualDepth := uint(e.Node.VisualDepth())
	result := linear.RenderLine(linear.LineParams{
		NodeParams: contract.NodeParams{
			Path:        e.Node.Path,
			Name:        e.Node.Extension.Name,
			IsDir:       e.Node.IsDirectory(),
			Depth:       depth,
			IsLast:      e.IsLast,
			VisualDepth: visualDepth,
		},
		RenderParams: contract.RenderParams{
			BodyWidth: p.bodyWidth,
			Theme:     p.theme,
		},
		BranchStack: p.branchStack,
	})
	p.branchStack = result.BranchStack
	p.program.Send(porthole.ContentLineMsg{
		Line: result.Line,
		Params: porthole.RenderParams{
			NodeParams: contract.NodeParams{
				Path:        e.Node.Path,
				Name:        e.Node.Extension.Name,
				IsDir:       e.Node.IsDirectory(),
				Depth:       depth,
				IsLast:      e.IsLast,
				VisualDepth: visualDepth,
			},
		},
		BranchStack: result.BranchStack,
	})
}

func (p *portholePresenter) OnActionEvent(e *report.ActionEvent) {
	depth := uint(e.Node.Extension.Depth)
	visualDepth := uint(e.Node.VisualDepth())
	isLastStep := false
	if e.IsPipelineStep {
		visualDepth++
		isLastStep = e.IsLastStep
	}
	result := linear.RenderLine(linear.LineParams{
		NodeParams: contract.NodeParams{
			Path:            e.Node.Path,
			Name:            e.Node.Extension.Name,
			IsDir:           e.Node.IsDirectory(),
			Depth:           depth,
			ActionName:      e.Name,
			CommandOutput:   e.CommandOutput,
			ExecutionString: e.ExecutionString,
			DryRun:          e.DryRun,
			Err:             e.Err,
			IsLast:          e.IsLast,
			IsPipelineStep:  e.IsPipelineStep,
			IsLastStep:      isLastStep,
			VisualDepth:     visualDepth,
		},
		RenderParams: contract.RenderParams{
			BodyWidth: p.bodyWidth,
			Theme:     p.theme,
		},
		BranchStack: p.branchStack,
	})
	p.branchStack = result.BranchStack
	p.program.Send(porthole.ContentLineMsg{
		Line: result.Line,
		Params: porthole.RenderParams{
			NodeParams: contract.NodeParams{
				Path:            e.Node.Path,
				Name:            e.Node.Extension.Name,
				IsDir:           e.Node.IsDirectory(),
				Depth:           depth,
				ActionName:      e.Name,
				CommandOutput:   e.CommandOutput,
				ExecutionString: e.ExecutionString,
				DryRun:          e.DryRun,
				Err:             e.Err,
				IsLast:          e.IsLast,
				IsPipelineStep:  e.IsPipelineStep,
				IsLastStep:      isLastStep,
				VisualDepth:     visualDepth,
			},
		},
		BranchStack: result.BranchStack,
	})
}

func (p *portholePresenter) OnPipelineEvent(e *report.PipelineEvent) {
	// The pipeline header is rendered by the parent node's OnNodeEvent
	// (which already includes the pipeline name). Skip the header event
	// to avoid duplicating the node in the output.
	if e.IsPipelineHeader {
		return
	}

	// Pipeline-not-found error: render as a standalone error line.
	depth := uint(e.Node.Extension.Depth)
	visualDepth := uint(e.Node.VisualDepth())
	result := linear.RenderLine(linear.LineParams{
		NodeParams: contract.NodeParams{
			Path:         e.Node.Path,
			Name:         e.Node.Extension.Name,
			IsDir:        e.Node.IsDirectory(),
			Depth:        depth,
			PipelineName: e.Name,
			Err:          e.Err,
			IsLast:       e.IsLast,
			VisualDepth:  visualDepth,
		},
		RenderParams: contract.RenderParams{
			BodyWidth: p.bodyWidth,
			Theme:     p.theme,
		},
		BranchStack: p.branchStack,
	})
	p.branchStack = result.BranchStack
	p.program.Send(porthole.ContentLineMsg{
		Line: result.Line,
		Params: porthole.RenderParams{
			NodeParams: contract.NodeParams{
				Path:         e.Node.Path,
				Name:         e.Node.Extension.Name,
				IsDir:        e.Node.IsDirectory(),
				Depth:        depth,
				PipelineName: e.Name,
				Err:          e.Err,
				IsLast:       e.IsLast,
				VisualDepth:  visualDepth,
			},
		},
		BranchStack: result.BranchStack,
	})
}

func (p *portholePresenter) OnSkipEvent(e *report.SkipEvent) {
	depth := uint(e.Node.Extension.Depth)
	visualDepth := uint(e.Node.VisualDepth())
	isLastStep := false
	if e.IsPipelineStep {
		visualDepth++
		isLastStep = true // skip always terminates the pipeline
	}
	result := linear.RenderLine(linear.LineParams{
		NodeParams: contract.NodeParams{
			Path:           e.Node.Path,
			Name:           e.Node.Extension.Name,
			IsDir:          e.Node.IsDirectory(),
			Depth:          depth,
			ActionName:     e.Name,
			IsLast:         e.IsLast,
			IsPipelineStep: e.IsPipelineStep,
			IsLastStep:     isLastStep,
			VisualDepth:    visualDepth,
		},
		RenderParams: contract.RenderParams{
			BodyWidth: p.bodyWidth,
			Theme:     p.theme,
		},
		BranchStack: p.branchStack,
	})
	p.branchStack = result.BranchStack
	p.program.Send(porthole.ContentLineMsg{
		Line: result.Line,
		Params: porthole.RenderParams{
			NodeParams: contract.NodeParams{
				Path:           e.Node.Path,
				Name:           e.Node.Extension.Name,
				IsDir:          e.Node.IsDirectory(),
				Depth:          depth,
				ActionName:     e.Name,
				IsLast:         e.IsLast,
				IsPipelineStep: e.IsPipelineStep,
				IsLastStep:     isLastStep,
				VisualDepth:    visualDepth,
			},
		},
		BranchStack: result.BranchStack,
	})
}

func (p *portholePresenter) OnComplete(traversal *report.Traversal) {
	// Guard against the bubbletea program having already exited
	// (e.g. via Ctrl+C). Without this select, a closed done
	// channel would cause an immediate, non-blocking receive that
	// falls through.
	select {
	case <-p.done:
		return
	default:
	}

	defer func() { _ = recover() }()

	var errs []error
	if traversal.Err != nil {
		errs = append(errs, traversal.Err)
	}

	p.program.Send(porthole.CompleteMsg{
		Files:   int(traversal.FilesVisited),
		Dirs:    int(traversal.DirsVisited),
		Errs:    errs,
		Elapsed: traversal.Elapsed,
	})

	// Block until the bubbletea program exits. The porthole mirrors
	// the highway view's behaviour: on completion the chrome shows
	// "press space to exit" and the program only terminates when the
	// user presses space (or Ctrl+C). Without this wait, the
	// coordinator's goroutine would return immediately, the cobra
	// command would exit, and the process would be torn down before
	// the user sees the completion state.
	<-p.done
}

func (p *portholePresenter) buildBannerInfo() banner.Info {
	info := banner.Info{
		Disable:  p.cfg.Banner.Disable,
		Position: p.cfg.Banner.Position,
		Justify:  p.cfg.Banner.Justify,
	}

	if p.cfg.Banner.Disable {
		return info
	}

	// Resolve the gradient from the theme. The "banner-control"
	// component must be bound to a gradient in the theme YAML;
	// when missing the banner renders as plain text.
	//
	// StepsOverride (from ui.porthole.banner.steps) lets the user
	// keep sharing the gradient's colour endpoints with other
	// widgets (so the colour scheme stays consistent) but tune
	// the banner's sweep smoothness independently. Zero means
	// "use the gradient's own steps".
	var grad *contract.ResolvedGradient
	if g, ok := p.theme.GradientFor(contract.GradientComponentBanner); ok && g.Steps > 0 {
		steps := g.Steps
		if p.cfg.Banner.StepsOverride > 0 {
			steps = p.cfg.Banner.StepsOverride
		}
		grad = &contract.ResolvedGradient{Steps: steps, Hi: g.Hi, Lo: g.Lo}
	}
	info.Gradient = grad

	if grad != nil {
		st := effects.NewGradientState()
		st.TotalSteps = grad.Steps
		info.State = st
	}

	// Pick the random aspects ONCE here, not per-render. Use the
	// package-level random source (math/rand/v2) which is fine for
	// non-security purposes.
	rng := rand.New(rand.NewPCG(uint64(os.Getpid()), uint64(core.Now().UnixNano()))) //nolint:gosec // non-security
	info.Aspects = banner.RandomiseAspects(rng)

	tick := p.cfg.Banner.Tick
	if tick <= 0 {
		tick = 500 // Default banner tick in ms
	}
	info.Tick = time.Duration(tick) * time.Millisecond

	return info
}
