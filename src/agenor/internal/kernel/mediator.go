package kernel

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/agenor/internal/enclave"
	"github.com/snivilised/jaywalk/src/agenor/internal/level"
	"github.com/snivilised/jaywalk/src/agenor/life"
	"github.com/snivilised/jaywalk/src/agenor/pref"
	"github.com/snivilised/jaywalk/src/agenor/stock"
)

// mediator controls traversal, sends notifications and emits
// life-cycle events
type mediator struct {
	tree         string
	subscription enums.Subscription
	facade       pref.Facade
	impl         NavigatorImpl
	guardian     *guardian
	periscope    *level.Periscope
	o            *pref.Options
	resources    *enclave.Resources
	metrics      core.Metrics
	order        []enums.Role
}

// NewMediator creates new Mediator
func NewMediator(inception *Inception,
	sealer enclave.GuardianSealer,
) (enclave.Mediator, error) {
	o := inception.Harvest.Options()
	facade := inception.Facade
	resources := inception.Resources
	impl, err := newImpl(o, inception)

	metrics := resources.Supervisor.Many(
		enums.MetricNoFilesInvoked,
		enums.MetricNoDirectoriesInvoked,
		enums.MetricNoChildFilesFound,
	)

	return &mediator{
		tree:         inception.NavigationTree(),
		subscription: inception.Subscription,
		facade:       facade,
		impl:         impl,
		guardian: newGuardian(&guardianInfo{
			subscription: inception.Subscription,
			client:       facade.Client(),
			master:       sealer,
			metrics:      metrics,
		}),
		periscope: level.New(),
		o:         o,
		resources: resources,
		metrics:   metrics,
	}, err
}

// Decorate adds a decorator to the invocation chain. The order of decoration is
// important, as it determines the order in which the decorators are invoked.
// For example, if a filter is decorated before a master sealer, then the filter
// will be invoked before the master sealer, which would prevent fastward resume
// from being used.
func (m *mediator) Decorate(link enclave.Link) error {
	return m.guardian.Decorate(link)
}

// Unwind removes the most recently added decorator from the invocation chain. This is
// used to allow decorators to be removed from the chain, such as when a filter is no
// longer needed.
func (m *mediator) Unwind(role enums.Role) error {
	return m.guardian.Unwind(role)
}

// Swap replaces the underlying client handler with the provided decorator. This
// is used to allow the guardian to be decorated with different handlers, such as
// filters or a master sealer.
func (m *mediator) Swap(decorator core.Client) {
	m.guardian.Swap(decorator)
}

// Invoke executes the chain which may or may not end up resulting in
// the invocation of the client's callback, depending on the contents
// of the chain.
func (m *mediator) Invoke(servant core.Servant,
	inspection enclave.Inspection,
) error {
	return m.guardian.Invoke(servant, inspection)
}

// Poke invokes the client directly, bypassing the guardian chain.
// Used by plugins that need to deliver a different servant than the
// one currently being processed, such as the peer plugin which
// delivers the previously buffered node rather than the current one.
func (m *mediator) Poke(servant core.Servant) error {
	return m.guardian.Poke(servant)
}

// Flush is the same as invoke but is invoked as a result of the peer plugin
// flushing its buffer, which means it needs to accept the inspection as well,
// which is used to enrich the peer info that is emitted with the flush. This
// is used to allow the peer plugin to flush its buffer with the enriched
// peer info, which is necessary for the peer plugin to function correctly.
// When active, this should only happen once per traversal.
func (m *mediator) Flush(servant core.Servant,
	inspection enclave.Inspection,
) error {
	fmt.Println("🔆🔆🔆 DEBUG: mediator.Flush 🔆🔆🔆")
	return m.guardian.Invoke(servant, inspection)
}

// Arrange allows the guardian to arrange the active and order for a role. This is used to
// allow the guardian to arrange the active and order for a role, which is necessary
// for fastward resume. When resuming, the active and order for the master sealer role
// need to be arranged in a specific way to ensure that the invocation chain is
// correctly reconstructed.
func (m *mediator) Arrange(active, order []enums.Role) {
	m.order = order
	m.guardian.arrange(active, order)
}

// Read acquires the contents of a directory
func (m *mediator) Read(path string) ([]fs.DirEntry, error) {
	return m.o.Hooks.ReadDirectory.Invoke()(m.resources.Forest.T, path)
}

// Spawn allows the mediator to spawn a new child navigation with the specified
// tree. This is used by the guardian to spawn a new child navigation when a
// new session is started or when a session is resumed, which allows the kernel
// to navigate the specified tree and to perform the necessary actions based on
// the structure of the file system and the options defined for the session.
func (m *mediator) Spawn(ctx context.Context,
	tree string,
) (*enclave.KernelResult, error) {
	return m.impl.Top(ctx, &navigationStatic{
		mediator:     m,
		tree:         tree,
		calc:         m.resources.Forest.T.Calc(),
		subscription: m.subscription,
		magnitude:    m.facade.Magnitude(),
	})
}

// Bridge combines information gleaned from the previous traversal that was
// interrupted, into the resume traversal
func (m *mediator) Bridge(active *core.ActiveState) {
	m.tree = active.Tree
	m.periscope.Offset(active.Depth)
	m.Supervisor().Load(active.Metrics)
}

// Supervisor gets the supervisor from the resources
func (m *mediator) Supervisor() *enclave.Supervisor {
	return m.resources.Supervisor
}

// Navigate performs the traversal through the file system,
// using the provided context for cancellation and timeout control.
func (m *mediator) Navigate(ctx context.Context) (result *enclave.KernelResult, err error) {
	result, err = m.impl.Top(ctx, &navigationStatic{
		mediator:     m,
		tree:         m.tree,
		calc:         m.resources.Forest.T.Calc(),
		subscription: m.subscription,
		magnitude:    m.facade.Magnitude(),
	})

	if m.o != nil {
		if !stock.IsBenignError(err) && err != nil {
			m.o.Monitor.Log.Error(err.Error())
		} else if result != nil {
			m.o.Monitor.Log.Info("navigation completed",
				"files", result.Metrics().Count(enums.MetricNoFilesInvoked),
				"directories", result.Metrics().Count(enums.MetricNoDirectoriesInvoked),
			)
		}
	}

	return result, err
}

// Ignite primes the navigator for traversal and announces
func (m *mediator) Ignite(ignition *enclave.Ignition) {
	m.impl.Ignite(ignition)
	if m.o != nil {
		m.o.Monitor.Log.Info("navigation started", "root", m.tree)
	}

	m.resources.Binder.Controls.Begin.Dispatch()(&life.BeginState{
		Tree: m.tree,
	})
}

// Result retrieves the kernel navigation result
func (m *mediator) Result(ctx context.Context,
) *enclave.KernelResult {
	return m.impl.Result(ctx)
}

// Snooze tbd
func (m *mediator) Snooze(ctx context.Context,
	active *core.ActiveState,
) (result *enclave.KernelResult, err error) {
	m.tree = active.Tree

	return m.impl.Top(ctx, &navigationStatic{
		mediator:     m,
		tree:         active.Tree,
		calc:         m.resources.Forest.T.Calc(),
		subscription: m.subscription,
		magnitude:    m.facade.Magnitude(),
	})
}

func (m *mediator) Bye(result core.TraverseResult) {
	m.resources.Binder.Controls.End.Dispatch()(result)
}

func (m *mediator) descend(node *core.Node) bool {
	if !m.periscope.Descend(m.o.Behaviours.Cascade.Depth) {
		return false
	}

	m.resources.Binder.Controls.Descend.Dispatch()(node)

	return true
}

func (m *mediator) ascend(node *core.Node, permit bool) {
	if permit {
		m.periscope.Ascend()
		m.resources.Binder.Controls.Ascend.Dispatch()(node)
	}
}
