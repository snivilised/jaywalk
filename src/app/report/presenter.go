package report

import (
	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/agenor/pref"
)

// Presenter is the interface all display implementations satisfy.
// It is purely reactive - all methods are event notifications. The
// implementation decides how to render each event; no formatting logic
// lives outside the implementing type.
//
// Presenter is defined here, alongside the event types it handles, so
// that both the ui package and the prism package can depend on report
// without creating a circular dependency between them.
type Presenter interface {
	// OnTraversalOptions is called with the resolved options for a traversal,
	// allowing the presenter to configure itself based on the options. This is
	// called before OnBegin, so the presenter can use the options to influence
	// how it renders the beginning of the traversal.
	OnTraversalOptions(o *pref.Options)

	// OnBegin is called once before any traversal events, with the
	// opening metadata. Implementations should use this to render
	// any opening banner or header.
	OnBegin(e *BeginEvent)

	// OnNodeEvent is called per node visit when no action or pipeline
	// is configured.
	OnNodeEvent(e *NeutralEvent)

	// OnActionEvent is called when a configured action has been executed
	// against a node.
	OnActionEvent(e *ActionEvent)

	// OnPipelineEvent is called when a configured pipeline has been
	// executed against a node.
	OnPipelineEvent(e *PipelineEvent)

	// OnSkipEvent is called when an action is skipped for a node because
	// a placeholder in the action's cmd string resolved to a path at or
	// above the traversal root.
	OnSkipEvent(e *SkipEvent)

	// OnComplete is called once at the end of a traversal with the full
	// structured outcome.
	OnComplete(t *Traversal)

	// OnWorkerState is called when a pool worker's activity state changes.
	// workerID is the pool-assigned goroutine ID, formatted as "W#N".
	OnWorkerState(state enums.WorkerState, workerID string)
}

// PeerAware is an optional interface that a Presenter can implement to
// opt into the peer info facility. When a view implements PeerAware,
// the coordinator runs a preview traversal before the live pass to build
// the PeerInfoMap, which provides correct IsLast
// for every node regardless of filtering or sampling.
// Views that do not need peer info simply do not implement this interface
// and are entirely unaffected by the peer info machinery.
type PeerAware interface {
	// NeedsPeerInfo reports whether this view requires peer position data.
	// Returning true causes the coordinator to run a preview traversal.
	NeedsPeerInfo() bool

	// OnPeerInfoBegin is called after the preview traversal completes,
	// with the total file and directory counts collected during the
	// preview. Views can use these counts to display a progress indicator
	// during the live traversal.
	OnPeerInfoBegin(files, dirs uint, peerInfoMap map[string]*core.PeerInfo)

	// OnPeerInfoEnd is called when the live traversal completes, allowing
	// the view to tear down any progress indicator it displayed.
	OnPeerInfoEnd()
}
