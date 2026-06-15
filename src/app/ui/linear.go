package ui

import (
	"math/rand/v2"
	"os"
	"sync"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/agenor/pref"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
	"github.com/snivilised/jaywalk/src/third/lo"
)

// linearPresenter is the linearPresenter-view display implementation. It translates
// report events into contract.Motif calls and delegates all formatting
// and output to the contract.Renderer. It contains no formatting logic.
//
// Safe for concurrent use - all renderer calls are serialised through
// a mutex so interleaved output from the sprint command's worker pool is
// avoided.
type linearPresenter struct {
	mux          sync.Mutex
	renderer     contract.Renderer
	kind         contract.NavigationKind // remembered from OnBegin for use in OnComplete
	subscription enums.Subscription
	lastParent   string
	peerInfo     map[string]*core.PeerInfo
	renderedDirs map[string]bool
	cfg          LinearConfig
	theme        contract.Theme
}

func (l *linearPresenter) OnTraversalOptions(o *pref.Options) {
	o.View.Peer.IsActive = true
}

// OnBegin translates the BeginEvent into a contract.Overture and calls
// renderer.Begin to render the opening banner.
func (l *linearPresenter) OnBegin(e *report.BeginEvent) {
	l.mux.Lock()
	defer l.mux.Unlock()

	kind := lo.Ternary(e.IsPrime,
		contract.PrimeNavigation,
		contract.ResumeNavigation,
	)

	l.kind = kind
	l.subscription = e.Subscription
	l.lastParent = ""
	l.renderedDirs = make(map[string]bool)

	// Build banner info for the renderer
	bannerInfo := l.buildBannerInfo()

	l.renderer.Begin(contract.Overture{
		Root:       e.Root,
		Caption:    e.Caption,
		StartedAt:  e.StartedAt,
		Kind:       kind,
		ResumeFrom: e.ResumeFrom,
		DateFormat: e.DateFormat,
		Banner:     bannerInfo,
	})
}

// OnNodeEvent translates a neutral node visit into a contract.Motif.
// Depth is sourced from node.Extension.Depth as provided by agenor.
func (l *linearPresenter) OnNodeEvent(e *report.NeutralEvent) {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.ensureParentRendered(e.Node)
	l.renderer.Show(contract.Motif{
		Path:        e.Node.Path,
		Name:        e.Node.Extension.Name,
		IsDir:       e.Node.IsDirectory(),
		Depth:       e.Node.Extension.Depth,
		VisualDepth: e.Node.VisualDepth(),
		IsLast:      e.IsLast,
	})
}

func (l *linearPresenter) OnActionEvent(e *report.ActionEvent) {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.ensureParentRendered(e.Node)
	l.renderer.Show(contract.Motif{
		Path:            e.Node.Path,
		Name:            e.Node.Extension.Name,
		IsDir:           e.Node.IsDirectory(),
		Depth:           e.Node.Extension.Depth,
		VisualDepth:     lo.Ternary(e.IsPipelineStep, e.Node.VisualDepth()+1, e.Node.VisualDepth()),
		ActionName:      e.Name,
		ExecutionString: e.ExecutionString,
		CommandOutput:   e.CommandOutput,
		DryRun:          e.DryRun,
		Err:             e.Err,
		IsLast:          e.IsLast,
		IsPipelineStep:  e.IsPipelineStep,
		IsLastStep:      e.IsLastStep,
	})
}

func (l *linearPresenter) OnPipelineEvent(e *report.PipelineEvent) {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.ensureParentRendered(e.Node)
	l.renderer.Show(contract.Motif{
		Path:             e.Node.Path,
		Name:             e.Node.Extension.Name,
		IsDir:            e.Node.IsDirectory(),
		Depth:            e.Node.Extension.Depth,
		VisualDepth:      e.Node.VisualDepth(),
		PipelineName:     e.Name,
		ExecutionString:  e.ExecutionString,
		CommandOutput:    e.CommandOutput,
		DryRun:           e.DryRun,
		Err:              e.Err,
		IsLast:           e.IsLast,
		IsPipelineHeader: e.IsPipelineHeader,
	})
}

// OnSkipEvent translates a skip event into a contract.Motif flagged as
// skipped so the renderer can apply warning styling.
func (l *linearPresenter) OnSkipEvent(e *report.SkipEvent) {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.ensureParentRendered(e.Node)
	l.renderer.Show(contract.Motif{
		Path:           e.Node.Path,
		Name:           e.Node.Extension.Name,
		IsDir:          e.Node.IsDirectory(),
		Depth:          e.Node.Extension.Depth,
		VisualDepth:    lo.Ternary(e.IsPipelineStep, e.Node.VisualDepth()+1, e.Node.VisualDepth()),
		ActionName:     e.Name,
		Skipped:        true,
		Placeholder:    e.Placeholder,
		ResolvedPath:   e.ResolvedPath,
		IsLast:         e.IsLast,
		IsPipelineStep: e.IsPipelineStep,
		IsLastStep:     e.IsLastStep,
	})
}

// OnComplete translates the Traversal outcome into a contract.Summary and
// calls renderer.End to render the closing summary box. Kind is carried
// from OnBegin so the summary labels correctly for resume traversals.
func (l *linearPresenter) OnComplete(traversal *report.Traversal) {
	l.mux.Lock()
	defer l.mux.Unlock()

	errs := []error{}
	if traversal.Err != nil {
		errs = append(errs, traversal.Err)
	}

	l.renderer.End(contract.Summary{
		FilesVisited: traversal.FilesVisited,
		DirsVisited:  traversal.DirsVisited,
		Skipped:      traversal.ActionsSkipped.Value(),
		Elapsed:      traversal.Elapsed,
		Errors:       errs,
		Kind:         l.kind,
	})
}

func (l *linearPresenter) OnWorkerState(_ enums.WorkerState, _ string) {
	// No-op: linear view has no per-worker animation to control.
}

// NeedsPeerInfo reports whether this view requires peer position data.
// Returning true causes the coordinator to run a preview traversal.
func (l *linearPresenter) NeedsPeerInfo() bool {
	return true
}

// OnPeerInfoBegin is called after the preview traversal completes,
// with the total file and directory counts collected during the
// preview. Views can use these counts to display a progress indicator
// during the live traversal.
func (l *linearPresenter) OnPeerInfoBegin(files, dirs uint, peerInfoMap map[string]*core.PeerInfo) {
	l.peerInfo = peerInfoMap
}

// OnPeerInfoEnd is called when the live traversal completes, allowing
// the view to tear down any progress indicator it displayed.
func (l *linearPresenter) OnPeerInfoEnd() {
}

func (l *linearPresenter) ensureParentRendered(node *core.Node) {
	if l.subscription != enums.SubscribeFiles {
		return
	}

	if node.Parent == nil || node.Parent.Path == "" {
		return
	}

	// find all unrendered ancestors
	ancestors := []*core.Node{}
	curr := node.Parent
	for curr != nil {
		if l.renderedDirs[curr.Path] {
			break
		}
		ancestors = append(ancestors, curr)
		curr = curr.Parent
	}

	// render them in reverse order (top-down)
	for i := len(ancestors) - 1; i >= 0; i-- {
		p := ancestors[i]
		isLast := false
		if l.peerInfo != nil {
			if info, ok := l.peerInfo[p.Path]; ok {
				isLast = info.IsLast
			}
		}

		l.renderer.Show(contract.Motif{
			Path:        p.Path,
			Name:        p.Extension.Name,
			IsDir:       true,
			Depth:       p.Extension.Depth,
			VisualDepth: p.VisualDepth(),
			IsLast:      isLast,
		})
		l.renderedDirs[p.Path] = true
		l.lastParent = p.Path
	}
}

// buildBannerInfo assembles the per-session BannerInfo for the linear
// renderer. The random aspect selection is performed exactly once here
// so the aspects are frozen for the duration of the process. The gradient
// endpoints are resolved from the theme's "banner-control" component,
// falling back to no gradient when the binding is absent.
func (l *linearPresenter) buildBannerInfo() *contract.BannerInfo {
	info := &contract.BannerInfo{
		Disable:  l.cfg.Banner.Disable,
		Position: l.cfg.FlagsRowPosition,
		Justify:  l.cfg.Banner.Justify,
	}

	if l.cfg.Banner.Disable {
		return info
	}

	// Resolve the gradient from the theme
	var grad *contract.ResolvedGradient
	if g, ok := l.theme.GradientFor(contract.GradientComponentBanner); ok && g.Steps > 0 {
		steps := g.Steps
		if l.cfg.Banner.StepsOverride > 0 {
			steps = l.cfg.Banner.StepsOverride
		}
		grad = &contract.ResolvedGradient{Steps: steps, Hi: g.Hi, Lo: g.Lo}
	}
	info.Gradient = grad

	// Pick the random aspects ONCE here, not per-render
	rng := rand.New(rand.NewPCG(uint64(os.Getpid()), uint64(core.Now().UnixNano()))) //nolint:gosec // non-security
	aspects := banner.RandomiseAspects(rng)
	info.Aspects = contract.BannerAspects{
		Orientation: int(aspects.Orientation),
		Banding:     int(aspects.Banding),
		Unity:       int(aspects.Unity),
		FixedEnd:    int(aspects.FixedEnd),
	}

	return info
}

// NewLinearWithRenderer constructs a linear presenter backed by the
// given renderer. Intended for use in tests only - production code
// constructs linear via the ui registry using New(). This allows a
// spy or stub renderer to be injected without going through contract.New
// and without requiring a real terminal.
func NewLinearWithRenderer(r contract.Renderer) report.Presenter {
	return &linearPresenter{renderer: r}
}
