package ui

import (
	"fmt"
	"sync"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/agenor/pref"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/third/lo"
)

// linear is the linear-view display implementation. It translates
// report events into contract.Motif calls and delegates all formatting
// and output to the contract.Renderer. It contains no formatting logic.
//
// Safe for concurrent use - all renderer calls are serialised through
// a mutex so interleaved output from the sprint command's worker pool is
// avoided.
type linear struct {
	mux          sync.Mutex
	renderer     contract.Renderer
	kind         contract.NavigationKind // remembered from OnBegin for use in OnComplete
	subscription enums.Subscription
	lastParent   string
	peerInfo     map[string]*core.PeerInfo
	renderedDirs map[string]bool
}

func (l *linear) OnTraversalOptions(o *pref.Options) {
	o.View.Peer.IsActive = true
}

// OnBegin translates the BeginEvent into a contract.Overture and calls
// renderer.Begin to render the opening banner.
func (l *linear) OnBegin(e *report.BeginEvent) {
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

	l.renderer.Begin(contract.Overture{
		Root:       e.Root,
		Caption:    e.Caption,
		StartedAt:  e.StartedAt,
		Kind:       kind,
		ResumeFrom: e.ResumeFrom,
		DateFormat: e.DateFormat,
	})
}

// OnNodeEvent translates a neutral node visit into a contract.Motif.
// Depth is sourced from node.Extension.Depth as provided by agenor.
func (l *linear) OnNodeEvent(e *report.NeutralEvent) {
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

func (l *linear) OnActionEvent(e *report.ActionEvent) {
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

func (l *linear) OnPipelineEvent(e *report.PipelineEvent) {
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
func (l *linear) OnSkipEvent(e *report.SkipEvent) {
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
func (l *linear) OnComplete(traversal *report.Traversal) {
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

// NeedsPeerInfo reports whether this view requires peer position data.
// Returning true causes the coordinator to run a preview traversal.
func (l *linear) NeedsPeerInfo() bool {
	return true
}

// OnPeerInfoBegin is called after the preview traversal completes,
// with the total file and directory counts collected during the
// preview. Views can use these counts to display a progress indicator
// during the live traversal.
func (l *linear) OnPeerInfoBegin(files, dirs uint, peerInfoMap map[string]*core.PeerInfo) {
	fmt.Printf("🐸 DEBUG: linear.OnPeerInfoBegin (files: %d, dirs:%d) 🐸\n", files, dirs)
	l.peerInfo = peerInfoMap
}

// OnPeerInfoEnd is called when the live traversal completes, allowing
// the view to tear down any progress indicator it displayed.
func (l *linear) OnPeerInfoEnd() {
	fmt.Println("🐸 DEBUG: linear.OnPeerInfoEnd 🐸")
}

func (l *linear) ensureParentRendered(node *core.Node) {
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

// NewLinearWithRenderer constructs a linear presenter backed by the
// given renderer. Intended for use in tests only - production code
// constructs linear via the ui registry using New(). This allows a
// spy or stub renderer to be injected without going through contract.New
// and without requiring a real terminal.
func NewLinearWithRenderer(r contract.Renderer) report.Presenter {
	return &linear{renderer: r}
}
