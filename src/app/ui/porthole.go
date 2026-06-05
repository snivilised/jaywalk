package ui

import (
	"math/rand"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/term"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/pref"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/flow"
	"github.com/snivilised/jaywalk/src/prism/movies"
	"github.com/snivilised/jaywalk/src/prism/scroll"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
)

// portholePresenter is the presenter for the porthole bubbletea view.
// It translates report events into scroll.ContentLineMsg and sends them
// to the model via the program. The view renders a single stream of
// content lines in a viewport with optional scrollbar gutter.
type portholePresenter struct {
	program     *tea.Program
	done        chan struct{}
	cfg         PortholeConfig
	noW         uint
	maxDepth    uint
	totalFiles  uint
	totalDirs   uint
	theme       contract.Theme
	noRecurse   bool
	branchStack []bool
	// bodyWidth is the usable column count inside the left/right
	// borders and the scrollbar gutter. It is used to right-justify
	// the landing strip produced by flow.RenderLine.
	bodyWidth uint

	header contract.HeaderInfo

	bannerInfo banner.Info
}

func newPortholePresenter(palette contract.Palette, pCfg PortholeConfig) (report.Presenter, error) {
	theme, err := contract.NewTheme(palette, os.Stdout)
	if err != nil {
		return nil, err
	}
	return &portholePresenter{cfg: pCfg, theme: theme}, nil
}

func (p *portholePresenter) OnTraversalOptions(o *pref.Options) {
	p.noW = o.Concurrency.NoW
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

	model := scroll.NewModel(e.Root, p.maxDepth, p.theme, p.noRecurse)
	model.SetWindowSizeCallback(func(bw uint) {
		p.bodyWidth = bw
	})

	// Pick a random animation for the activity widget.
	movies.RegisterAll()
	names := movies.ExpandNames(p.cfg.SpinnerNames)
	if len(names) == 0 {
		names = []string{movies.SpinnerTypeWave}
	}
	name := names[rand.Intn(len(names))] //nolint:gosec // selecting a random animation, not for security
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

	p.program.Send(scroll.OvertureMsg{
		Root:              e.Root,
		Caption:           e.Caption,
		SubscriptionLabel: subscriptionLabelFor(e.Subscription),
		StartedAt:         e.StartedAt,
		DateFormat:        e.DateFormat,
		PipelineName:      e.PipelineName,

		Header: p.header,

		Banner: p.bannerInfo,

		FlagsRowPosition: contract.PositionBottom,
	})
}

func (p *portholePresenter) NeedsPeerInfo() bool {
	return true
}

func (p *portholePresenter) OnPeerInfoBegin(files, dirs uint, _ map[string]*core.PeerInfo) {
	p.totalFiles = files
	p.totalDirs = dirs
}

func (p *portholePresenter) OnPeerInfoEnd() {}

func (p *portholePresenter) OnNodeEvent(e *report.NeutralEvent) {
	depth := uint(e.Node.Extension.Depth)
	visualDepth := uint(e.Node.VisualDepth())
	result := flow.RenderLine(
		e.Node.Path,
		e.Node.Extension.Name,
		e.Node.IsDirectory(),
		depth,
		"", "", "", "", false, nil,
		e.IsLast, false, false, visualDepth,
		p.branchStack,
		p.bodyWidth,
		p.theme,
		"",
	)
	p.branchStack = result.BranchStack
	p.program.Send(scroll.ContentLineMsg{
		Line: result.Line,
		Params: scroll.RenderParams{
			Path:        e.Node.Path,
			Name:        e.Node.Extension.Name,
			IsDir:       e.Node.IsDirectory(),
			Depth:       depth,
			IsLast:      e.IsLast,
			VisualDepth: visualDepth,
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
	result := flow.RenderLine(
		e.Node.Path,
		e.Node.Extension.Name,
		e.Node.IsDirectory(),
		depth,
		e.Name, "", e.CommandOutput, e.ExecutionString, e.DryRun, e.Err,
		e.IsLast, e.IsPipelineStep, isLastStep, visualDepth,
		p.branchStack,
		p.bodyWidth,
		p.theme,
		"",
	)
	p.branchStack = result.BranchStack
	p.program.Send(scroll.ContentLineMsg{
		Line: result.Line,
		Params: scroll.RenderParams{
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
	result := flow.RenderLine(
		e.Node.Path,
		e.Node.Extension.Name,
		e.Node.IsDirectory(),
		depth,
		"", e.Name, "", "", false, e.Err,
		e.IsLast, false, false, visualDepth,
		p.branchStack,
		p.bodyWidth,
		p.theme,
		"",
	)
	p.branchStack = result.BranchStack
	p.program.Send(scroll.ContentLineMsg{
		Line: result.Line,
		Params: scroll.RenderParams{
			Path:         e.Node.Path,
			Name:         e.Node.Extension.Name,
			IsDir:        e.Node.IsDirectory(),
			Depth:        depth,
			PipelineName: e.Name,
			Err:          e.Err,
			IsLast:       e.IsLast,
			VisualDepth:  visualDepth,
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
	result := flow.RenderLine(
		e.Node.Path,
		e.Node.Extension.Name,
		e.Node.IsDirectory(),
		depth,
		e.Name, "", "", "", false, nil,
		e.IsLast, e.IsPipelineStep, isLastStep, visualDepth,
		p.branchStack,
		p.bodyWidth,
		p.theme,
		"",
	)
	p.branchStack = result.BranchStack
	p.program.Send(scroll.ContentLineMsg{
		Line: result.Line,
		Params: scroll.RenderParams{
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

	p.program.Send(scroll.CompleteMsg{
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
	grad, _ := p.theme.GradientFor(contract.GradientComponentBanner)

	return banner.Info{
		Disable:  false,
		Position: contract.PositionTop,
		Justify:  banner.JustifyRight,
		Width:    80,
		Aspects: banner.Aspects{
			Orientation: banner.OrientationHorizontal,
			Banding:     banner.BandingWithout,
			Unity:       banner.UnityUnified,
			FixedEnd:    banner.FixedEndUnfixed,
		},
		Gradient: &grad,
		State:    nil, // state is managed by the model's ticker
		Tick:     500 * time.Millisecond,
	}
}
