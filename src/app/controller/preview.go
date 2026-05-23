package controller

import (
	"context"
	"fmt"

	"github.com/snivilised/jaywalk/src/agenor"
	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/agenor/pref"
)

// PeerInfoMap maps a node path to its resolved peer info. It is built
// during the preview traversal and consumed during the live traversal
// so that IsLast is correct for every node regardless
// of filtering or sampling.
type PeerInfoMap map[string]*core.PeerInfo

// previewEntry holds the path and peer info for a node during the
// preview traversal.
type previewEntry struct {
	path string
	info *core.PeerInfo
}

// previewBuffer replicates the single-node delay logic, operating
// entirely within the preview traversal callback. It has no dependency
// on the guardian chain or mediator.
type previewBuffer struct {
	pending    *previewEntry
	dirPending map[core.TraversalDepth]*previewEntry
	result     PeerInfoMap
}

func newPreviewBuffer() *previewBuffer {
	return &previewBuffer{
		dirPending: make(map[core.TraversalDepth]*previewEntry),
		result:     make(PeerInfoMap),
	}
}

func (b *previewBuffer) visit(node *core.Node) {
	entry := &previewEntry{
		path: node.Path,
		info: &core.PeerInfo{
			IsLast: false,
		},
	}

	if node.IsDirectory() {
		if b.pending != nil {
			b.result[b.pending.path] = b.pending.info
			b.pending = nil
		}

		// Flush the previous directory at this same depth as non-last
		// before storing the incoming one. It remains in dirPending only
		// if ascend has not yet fired for it, meaning it was not the last
		// peer among its siblings.
		if prev, ok := b.dirPending[node.Extension.Depth]; ok {
			b.result[prev.path] = prev.info
		}

		b.dirPending[node.Extension.Depth] = entry

		return
	}

	if b.pending != nil {
		b.result[b.pending.path] = b.pending.info
	}

	b.pending = entry
}

func (b *previewBuffer) ascend(node *core.Node) {
	depth := node.Extension.Depth

	// The directory being ascended from lives at depth+1. Only the
	// last sibling directory remains in dirPending at that depth -
	// all preceding siblings were flushed as non-last in visit.
	childDepth := depth + 1
	if dir, ok := b.dirPending[childDepth]; ok {
		dir.info.IsLast = true
		b.result[dir.path] = dir.info
		delete(b.dirPending, childDepth)
	}

	if b.pending != nil {
		b.pending.info.IsLast = true
		b.result[b.pending.path] = b.pending.info
		b.pending = nil
	}
}

func (b *previewBuffer) finalise() PeerInfoMap {
	fmt.Println("🔥 DEBUG: previewBuffer finalised with peer info map: 🔥")
	if b.pending != nil {
		b.pending.info.IsLast = true
		b.result[b.pending.path] = b.pending.info
		b.pending = nil
	}

	for _, dir := range b.dirPending {
		dir.info.IsLast = true
		b.result[dir.path] = dir.info
	}

	b.dirPending = nil

	return b.result
}

// buildPeerInfoMap runs a preview traversal using SlowPrime with the
// same settings as the live traversal. It returns a PeerInfoMap, the
// constructed *pref.Options for reuse by the live pass via pref.Using.O,
// the TraverseResult from the preview pass for use by PeerAware
// presenters, and the maximum traversal depth encountered.
func buildPeerInfoMap(
	ctx context.Context,
	req *PrimeRequest,
	settings []pref.Option,
) (PeerInfoMap, *pref.Options, core.TraverseResult, uint, error) {
	fmt.Println("🦋 DEBUG: buildPeerInfoMap: building peer info map with preview traversal ... 🦋")
	buf := newPreviewBuffer()

	var options *pref.Options
	var maxDepth uint

	facade := &pref.Using{
		Subscription: enums.SubscribeUniversal,
		Head: pref.Head{
			Handler: func(servant agenor.Servant) error {
				buf.visit(servant.Node())
				depth := uint(servant.Node().Extension.Depth)
				if depth > maxDepth {
					maxDepth = depth
				}
				return nil
			},
			GetForest: req.GetForest,
		},
		Tree: req.Tree,
	}

	previewSettings := append([]pref.Option{}, settings...)
	previewSettings = append(previewSettings,
		func(o *pref.Options) error {
			options = o
			o.Events.Ascend.On(buf.ascend)
			return nil
		},
	)

	result, err := agenor.Walk().Configure().Extent(
		agenor.Prime(facade, previewSettings...),
	).Navigate(ctx)

	if err != nil {
		return nil, nil, nil, 0, err
	}

	return buf.finalise(), options, result, maxDepth, nil
}
