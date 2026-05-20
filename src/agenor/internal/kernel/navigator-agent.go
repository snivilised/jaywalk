package kernel

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/internal/enclave"
	"github.com/snivilised/jaywalk/src/agenor/pref"
	"github.com/snivilised/jaywalk/src/agenor/tapable"
	"github.com/snivilised/jaywalk/src/locale"
	"github.com/snivilised/jaywalk/src/third/lo"
)

type readHooks struct {
	read tapable.Hook[core.ReadDirectoryHook, core.ChainReadDirectoryHook]
	sort tapable.Hook[core.SortHook, core.ChainSortHook]
}

type readOptions struct {
	hooks     readHooks
	behaviour *pref.SortBehaviour
}

type agentOptions struct {
	hooks   *tapable.Hooks
	defects *pref.DefectOptions
}

// navigatorAgent does work on behalf of the navigator. The agent performs
// generic tasks that apply to all navigators.
type navigatorAgent struct {
	ao        *agentOptions
	ro        *readOptions
	resources *enclave.Resources
	session   core.Session
	persister author
	magnitude string
}

func (n *navigatorAgent) Ignite(ignition *enclave.Ignition) {
	n.session = ignition.Session
}

func (n *navigatorAgent) top(ctx context.Context,
	ns *navigationStatic,
) (result *enclave.KernelResult, err error) {
	info, ie := n.ao.hooks.QueryStatus.Invoke()(
		ns.mediator.resources.Forest.T, ns.tree,
	)

	err = lo.TernaryF(ie != nil,
		func() error {
			return n.ao.defects.Fault.Accept(&pref.NavigationFault{
				Err:  ie,
				Path: ns.tree,
				Info: info,
			})
		},
		func() error {
			_, te := ns.mediator.impl.Traverse(ctx, ns,
				servant{
					node: core.Top(ns.tree, info),
					peer: nil, // tbd
				},
			)

			return te
		},
	)

	return ns.mediator.impl.Result(ctx), err
}

// Result is the single point at which a Result is constructed. Due to
// the spawn resume strategy, a Result may occur more than once during
// a navigation session. The session knows when completion occurs. Any
// Result that occurs prior to completion are as a result of child
// navigation whose result should be combined in the final Result. This
// is all handled by the strategy.
func (n *navigatorAgent) Result(_ context.Context) *enclave.KernelResult {
	result := enclave.NewResult(n.session,
		n.resources.Supervisor,
		n.session.IsComplete(),
	)

	return result
}

const (
	continueTraversal = true
	skipTraversal     = false
)

// travel is the general recursive navigation function which returns a bool
// indicating whether we continue travelling or not in response to an
// error.
// true: success path; continue/progress
// false: skip (all, dir)
//
// When an error occurs for this node, we return false (skipTraversal) indicating
// a skip. A skip can mean skip the entire navigation process (fs.SkipAll),
// or just skip all remaining sibling nodes in this directory (fs.SkipDir).
func (n *navigatorAgent) travel(ctx context.Context,
	ns *navigationStatic,
	vapour inspection,
) (skip bool, err error) {
	defer func() {
		if data := recover(); data != nil {
			// The tree on the mediator always points to the original tree root
			// requested by the user. The tree on navigation static can be different
			// when a spawn resume is in play; ie, the Spawn is created using a child
			// tree and this is what is set on the navigation static when a sub tree
			// is seeded, but this would be the incorrect tree to persist, rather, we
			// need the real ancestor that is denoted by the mediator's tree.
			//
			to, rescueErr := n.ao.defects.Panic.Rescue(n, &vex{
				data:     data,
				anc:      ns.mediator.tree,
				vap:      vapour,
				catalyst: "panic",
				mag:      n.magnitude,
			}) // wrap???

			// TODO: this needs some serious attention, the underlying panic is not being
			// made apparent in the error that is being returned.
			panicErr, ok := data.(error)
			if !ok {
				panicErr = errors.New("")
			}

			if rescueErr != nil {
				err = locale.NewTraversalNotSavedError(errors.Join(panicErr, rescueErr), to)
			}

			err = lo.TernaryF(rescueErr != nil,
				func() error {
					return locale.NewTraversalNotSavedError(errors.Join(panicErr, rescueErr), to)
				},
				func() error {
					return locale.NewTraversalSavedError(panicErr, to)
				},
			)

			skip = skipTraversal
		}
	}()

	var (
		parent = vapour.Current()
	)

	for _, entry := range vapour.Entries() {
		if ctx.Err() != nil {
			return n.handleInterrupt(ns, vapour, ctx.Err())
		}

		path := filepath.Join(parent.Path, entry.Name())
		info, e := entry.Info()

		// TODO: check sampling; should happen transparently, by plugin

		// TODO: ok for Travel to by-pass mediator?
		//
		if progress, err := ns.mediator.impl.Traverse(
			ctx, ns, servant{
				node: core.New(
					path,
					entry,
					info,
					parent,
					e,
				),
			},
		); !progress {
			if err != nil {
				if errors.Is(err, fs.SkipDir) {
					// The returning of skipTraversal by the child, denotes
					// a skip. So when a child node returns a SkipDir error and
					// skipTraversal, what we're saying is that we want to skip
					// processing all successive siblings but continue traversal.
					// The !progress indicates we're skipping the remaining
					// processing of all of the parent node's remaining children.
					// (see the ✨ below ...)
					//
					return skipTraversal, err
				}

				if ctx.Err() != nil {
					return n.handleInterrupt(ns, vapour, ctx.Err())
				}

				return continueTraversal, err
			}
		} else if err != nil {
			// ✨ ... we skip processing all the remaining children for
			// this node, but still continue the overall traversal.
			//
			switch {
			case errors.Is(err, fs.SkipDir):
				continue
			case errors.Is(err, fs.SkipAll):
				break
			default:
				if ctx.Err() != nil {
					return n.handleInterrupt(ns, vapour, ctx.Err())
				}

				return continueTraversal, err
			}
		}
	}

	if ctx.Err() != nil {
		return n.handleInterrupt(ns, vapour, ctx.Err())
	}

	return continueTraversal, nil
}

// handleInterrupt handles the case where the context is cancelled
// (e.g. by Ctrl-C). It logs the warning and saves the resume state
// via the configured panic handler, without raising a panic.
func (n *navigatorAgent) handleInterrupt(ns *navigationStatic,
	vapour inspection, cause error,
) (skip bool, err error) { //nolint:unparam // always false is ok because of travel signature
	if ns.mediator.o != nil && ns.mediator.o.Monitor.Log != nil {
		ns.mediator.o.Monitor.Log.Warn("navigation terminated by user (ctrl-c)")
	}

	to, rescueErr := n.ao.defects.Panic.Rescue(n, &vex{
		data:     cause,
		anc:      ns.mediator.tree,
		vap:      vapour,
		catalyst: "interrupt",
		mag:      n.magnitude,
	})

	if rescueErr != nil {
		return skipTraversal, locale.NewTraversalNotSavedError(
			errors.Join(cause, rescueErr), to,
		)
	}

	return skipTraversal, locale.NewTraversalSavedError(cause, to)
}

func (n *navigatorAgent) Save(data pref.RescueData) (string, error) {
	if v, ok := data.(vexation); ok {
		return n.persister.write(v)
	}

	// TODO: this should be a proper i18n error; actually, this is not
	// a save failure, its a problem with the vexation data, so perhaps this
	// should be a different error type.
	return "", errors.New("save failed")
}
