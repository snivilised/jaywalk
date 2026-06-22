package controller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"sync"

	"github.com/snivilised/jaywalk/src/agenor"
	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/agenor/pref"
	"github.com/snivilised/jaywalk/src/app/bedrock"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/app/shell"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

// Coordinator coordinates the layers between the command adapters and
// the agenor traversal engine. It is the single place that constructs
// pref.Facade values and calls into agenor via the agenor.Scenario on
// the request. It never imports cobra, mamba, or the command package.
//
// Dependency direction: command -> controller -> agenor
type Coordinator struct {
	config        *bedrock.Config
	locate        shell.LocateFunc
	exec          shell.ExecuteFunc
	rush          string
	forestBuilder pref.BuildForest
	actionRegexes map[string]*regexp.Regexp
	adminPath     string
	logger        *slog.Logger

	// poolExec is the shell pool executor that submits commands to the
	// pants worker pool asynchronously. Non-nil only when concurrent mode
	// is active (IsConcurrent && !DryRun && action/pipeline configured).
	// When non-nil, handleServant uses poolExec.Post for async submission
	// instead of the synchronous c.exec path.
	poolExec *shellPoolExecutor

	// workerStateTracker tracks per-worker activity state and notifies
	// the UI presenter when each worker completes a job. Created when
	// the shell pool executor is wired up; nil otherwise (non-concurrent
	// traversals, dry runs, tests).
	workerStateTracker *WorkerStateTracker

	// lastWorkerID is the execution ID from the most recent shell command
	// execution, formatted as "<worker-id>-<work-tag>-<job-id>".
	// Populated by the shell pool executor after each Execute call, read
	// by executeAction to propagate WorkerID into the event chain. Empty
	// when no command has been executed yet.
	lastWorkerID string
}

// AdminPath returns the configured admin path for resume state files.
func (c *Coordinator) AdminPath() string {
	return c.adminPath
}

// Logger returns the configured structured logger.
func (c *Coordinator) Logger() *slog.Logger {
	return c.logger
}

// CoordinatorOption is a functional option for Coordinator.
type CoordinatorOption func(*Coordinator)

// WithLocate overrides the LocateFunc used during PreFlight to validate
// whether action executables are invokable. The default is the
// platform-appropriate function returned by shell.Detect(). Use this
// in tests to inject a stub without spawning real subprocesses.
func WithLocate(fn shell.LocateFunc) CoordinatorOption {
	return func(c *Coordinator) {
		c.locate = fn
	}
}

// WithExec defines the platform-appropriate function for executing commands.
func WithExec(fn shell.ExecuteFunc) CoordinatorOption {
	return func(c *Coordinator) {
		c.exec = fn
	}
}

// WithShell defines the shell executable used by sprint shell pools.
func WithShell(command string) CoordinatorOption {
	return func(c *Coordinator) {
		c.rush = command
	}
}

// WithAdminPath sets the path for admin/resume state files.
func WithAdminPath(path string) CoordinatorOption {
	return func(c *Coordinator) {
		c.adminPath = path
	}
}

// WithLogger sets the structured logger used by the application.
func WithLogger(logger *slog.Logger) CoordinatorOption {
	return func(c *Coordinator) {
		c.logger = logger
	}
}

// WithForest allows injection of a pref.BuildForest, which is used to construct
// the file systems during traversal. This is primarily intended for testing,
// where a stubbed file system can be used to simulate various scenarios without
// relying on the real file system. In production, the Coordinator will use the
// default file system builder if this option is not provided.
func WithForest(forestBuilder pref.BuildForest) CoordinatorOption {
	return func(c *Coordinator) {
		c.forestBuilder = forestBuilder
	}
}

// New returns a ready-to-use Coordinator. config must not be nil.
func New(config *bedrock.Config, opts ...CoordinatorOption) *Coordinator {
	actionRegexes := make(map[string]*regexp.Regexp)
	if config != nil && config.Raw.Actions != nil {
		for name, action := range config.Raw.Actions {
			if action.Capture != "" {
				if re, err := regexp.Compile(action.Capture); err == nil {
					actionRegexes[name] = re
				}
			}
		}
	}

	coord := &Coordinator{
		config: config,
		locate: func(name string) (string, error) {
			return exec.LookPath(name)
		},
		// Production wiring replaces this with shell.Detect().Execute via
		// WithExec. Keeping a failing default makes missing wiring explicit.
		exec: func(ctx context.Context, cmdStr string) ([]byte, error) {
			return nil, errors.New("exec func not defined")
		},
		rush:          "sh",
		actionRegexes: actionRegexes,
	}

	for _, o := range opts {
		o(coord)
	}

	return coord
}

// ExecutePrime runs a fresh directory traversal using the scenario
// provided on the request. When the presenter implements PeerAware
// and NeedsPeerInfo returns true, a preview traversal is run first
// to build the PeerInfoMap and collect node counts for the progress
// indicator. The live traversal reuses the options built during the
// preview pass.
func (c *Coordinator) ExecutePrime(ctx context.Context, req *PrimeRequest) error {
	req.Root = req.Tree

	traversal := &report.Traversal{}
	view, isPeerAware := req.UI.(report.PeerAware)

	if isPeerAware && view.NeedsPeerInfo() {
		// Execute the preview traversal to build the PeerInfoMap and collect.
		peerInfoMap, builtOptions, result, maxDepth, err := buildPeerInfoMap(
			ctx, req, req.Settings,
		)
		if err != nil {
			return err
		}

		filesCount := result.Metrics().Count(enums.MetricNoFilesInvoked)
		dirsCount := result.Metrics().Count(enums.MetricNoDirectoriesInvoked)

		view.OnPeerInfoBegin(
			uint(filesCount), // NB: casting these to MetricValue causes a rendering
			uint(dirsCount),  // problem with the last entry in the tree
			peerInfoMap,
		)

		if md, ok := view.(interface{ SetMaxDepth(uint) }); ok {
			md.SetMaxDepth(maxDepth)
		}

		// The preview captured builtOptions without the coordinator's
		// logger and adminPath. Apply them now so they are not lost
		// when buildOptions short-circuits on using.O != nil.
		if c.logger != nil {
			_ = pref.WithLogger(c.logger)(builtOptions)
		}
		if c.adminPath != "" {
			_ = pref.WithAdminPath(c.adminPath)(builtOptions)
		}

		facade := &pref.Using{
			Subscription: req.Subscription,
			Head: pref.Head{
				Handler: func(servant agenor.Servant) error {
					return c.handleServant(ctx, servant, &req.Request, traversal, peerInfoMap)
				},
				GetForest: c.forestBuilder,
			},
			Tree: req.Tree,
			O:    builtOptions,
		}

		// Execute the live traversal with the peer info map and options from the
		// preview pass.
		err = c.execute(ctx, &req.Request, facade, traversal, true, "")
		view.OnPeerInfoEnd()

		return err
	}

	facade := &pref.Using{
		Subscription: req.Subscription,
		Head: pref.Head{
			Handler: func(servant agenor.Servant) error {
				return c.handleServant(ctx, servant, &req.Request, traversal, nil)
			},
			GetForest: c.forestBuilder,
		},
		Tree: req.Tree,
	}

	// Execute the live traversal without peer info.
	return c.execute(ctx, &req.Request, facade, traversal, true, "")
}

// ExecuteResume resumes an interrupted traversal. Peer info is not
// currently supported for resume - this will be addressed in a
// dedicated issue.
func (c *Coordinator) ExecuteResume(ctx context.Context, req *ResumeRequest) error {
	traversal := &report.Traversal{}

	// TODO: implement peer info support for resume traversals.

	facade := &pref.Relic{
		Head: pref.Head{
			Handler: func(servant agenor.Servant) error {
				return c.handleServant(ctx, servant, &req.Request, traversal, nil)
			},
		},
		Strategy: req.Strategy,
	}

	// Execute the live traversal without peer info.
	return c.execute(ctx, &req.Request, facade, traversal, false, req.ResumeFrom)
}

// execute is the shared orchestration path for both prime and resume
// traversals.
func (c *Coordinator) execute(
	parentCtx context.Context,
	req *Request,
	facade pref.Facade,
	traversal *report.Traversal,
	isPrime bool,
	resumeFrom string,
) error {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	if err := c.PreFlight(req); err != nil {
		return err
	}

	// Prepend default settings from Coordinator (admin path, logger)
	// so that user-provided settings override them.
	if c.adminPath != "" {
		req.Settings = append([]pref.Option{pref.WithAdminPath(c.adminPath)}, req.Settings...)
	}
	if c.logger != nil {
		req.Settings = append(
			[]pref.Option{pref.WithLogger(c.logger)},
			req.Settings...,
		)
	}

	closeExec, err := c.useShellPoolExec(ctx, req)
	if err != nil {
		return err
	}
	defer closeExec()

	// Extract header info before creating BeginEvent for consistency with highway model
	headerInfo := extractHeaderInfo(req)

	req.UI.OnBegin(&report.BeginEvent{
		Root:         req.Root,
		Caption:      c.captionFor(req),
		StartedAt:    core.Now(),
		IsPrime:      isPrime,
		ResumeFrom:   resumeFrom,
		Subscription: req.Subscription,
		ActionName:   req.ActionName,
		PipelineName: req.PipelineName,
		DateFormat:   c.config.Mapped.Interaction.DateFormat,
		Cancel:       cancel,

		// Header info: supplementary flag values extracted from
		// resolved traversal options for display in the highway view.
		Header: headerInfo,

		// Position of the flags row within the highway view
		FlagsRowPosition: c.config.Mapped.Highway.FlagsRowPosition,
	})

	result, err := req.Scenario(facade, req.Settings...).Navigate(ctx)

	traversal.Err = err
	if result != nil {
		traversal.FilesVisited = result.Metrics().Count(enums.MetricNoFilesInvoked)
		traversal.DirsVisited = result.Metrics().Count(enums.MetricNoDirectoriesInvoked)
		traversal.Elapsed = result.Session().Elapsed()
	}

	// Wait for the worker pool to finish all pending jobs before signalling
	// UI completion. Without this, CompleteMsg arrives at the UI before all
	// async MotifMsgs have been dispatched, causing the "✔ complete" message
	// to appear while workers are still executing.
	if c.poolExec != nil {
		c.poolExec.closeAll()
		select {
		case <-c.poolExec.done:
		case <-ctx.Done():
		}
	}

	req.UI.OnComplete(traversal)

	return err
}

func (c *Coordinator) useShellPoolExec(
	ctx context.Context,
	req *Request,
) (func(), error) {
	if !req.IsConcurrent || req.DryRun || (req.ActionName == "" && req.PipelineName == "") {
		return func() {}, nil
	}

	shell := c.rush
	if shell == "" {
		shell = "sh"
	}

	// Determine pool size from concurrency config.
	options := pref.DefaultOptions()
	for _, setting := range req.Settings {
		if setting == nil {
			continue
		}
		if err := setting(options); err != nil {
			return nil, err
		}
	}

	if options.Concurrency.NoW < 1 {
		options.Concurrency.NoW = 4
	}
	if options.Concurrency.Input.Size == 0 {
		options.Concurrency.Input.Size = options.Concurrency.NoW
	}
	if options.Concurrency.Output.Size == 0 {
		options.Concurrency.Output.Size = options.Concurrency.NoW
	}

	var wg sync.WaitGroup

	manifold, err := newJayShellPool(
		ctx, shell, &wg,
		options.Concurrency.NoW,
		options.Concurrency.Input.Size,
		options.Concurrency.Output.Size,
		options.Concurrency.Output.CheckCloseInterval,
		options.Concurrency.Output.TimeoutOnSend,
	)
	if err != nil {
		return nil, err
	}

	executor := newShellPoolExecutor(manifold)
	c.poolExec = executor
	c.workerStateTracker = NewWorkerStateTracker(req.UI, options.Concurrency.NoW)

	// Keep the synchronous wrapper for pipelines and non-async callers.
	previousExec := c.exec
	c.exec = func(_ context.Context, command string) ([]byte, error) {
		workerID, output, err := executor.Execute(ctx, command)
		c.lastWorkerID = workerID
		return output, err
	}

	return func() {
		c.poolExec = nil
		c.workerStateTracker = nil
		c.exec = previousExec
	}, nil
}

// extractHeaderInfo extracts cascade display (padlock or depth), filter
// flag info and sampler info from the Options that were passed through
// BuildTraversalSettings. This preserves display information even
// though the pref.Options struct doesn't directly expose the original
// flags. The returned HeaderInfo is then embedded in the BeginEvent
// sent to the UI presenter (see contract.HeaderInfo for field semantics).
//
// The scratch Options MUST be initialised via pref.DefaultOptions() (not
// a bare &pref.Options{}), because the Hooks subfields are interface
// types and a zero value would be nil; options such as
// WithHookReadDirectory call methods on those interfaces and would
// panic on a nil itab. useShellPoolExec uses the same pattern.
func extractHeaderInfo(req *Request) contract.HeaderInfo {
	var info contract.HeaderInfo
	options := pref.DefaultOptions()

	// Extract options from settings by running all option functions
	for _, setting := range req.Settings {
		if setting == nil {
			continue
		}
		if err := setting(options); err != nil {
			return contract.HeaderInfo{} // error case - return empty
		}
	}

	// Extract cascade display (no-recurse or depth) from behaviours
	if options.Behaviours.Cascade.NoRecurse {
		info.CascadeDisplay = contract.Static.Emoji.Padlock
	} else if options.Behaviours.Cascade.Depth > 0 {
		info.CascadeDisplay = fmt.Sprintf("depth:%d", options.Behaviours.Cascade.Depth)
	}

	// Extract filter info from poly filter definitions
	if options.Filter.Node != nil && options.Filter.Node.Poly != nil {
		fileDef := options.Filter.Node.Poly.File
		dirDef := options.Filter.Node.Poly.Directory

		// TranslateFilterIntent places BenignNodeFilterDef into the slot
		// the user did not specify (e.g. when the user only sets
		// --files, the directory slot is filled with the benign default
		// "match anything" filter). We must skip those placeholders here
		// so they don't appear in the flags row as if the user had
		// specified them (e.g. "dirs regex: .").
		fileActive := !fileDef.IsBenign() && fileDef.Pattern != ""
		dirActive := !dirDef.IsBenign() && dirDef.Pattern != ""

		// Determine file pattern and type
		if fileActive {
			//nolint:staticcheck // QF1003: if/else chain intentionally
			// avoids a switch because the `exhaustive` linter requires
			// every FilterType enum value to be listed.
			if fileDef.Type == enums.FilterTypeGlobEx {
				info.FilesGlob = fileDef.Pattern
			} else if fileDef.Type == enums.FilterTypeRegex {
				info.FilesRegex = fileDef.Pattern
			}
		}

		// Determine directory pattern and type
		if dirActive {
			//nolint:staticcheck // QF1003: see comment above.
			if dirDef.Type == enums.FilterTypeGlob {
				info.DirsGlob = dirDef.Pattern
			} else if dirDef.Type == enums.FilterTypeRegex {
				info.DirsRegex = dirDef.Pattern
			}
		}

		// Determine filter types based on precedence (regex takes priority
		// when both exist). Defaults to "glob" when the slot is benign
		// (no user-specified filter).
		if fileActive {
			info.FileTypeMode = determineFileType(fileDef.Type)
		} else {
			info.FileTypeMode = filterModeGlob
		}
		if dirActive {
			info.DirTypeMode = determineFileType(dirDef.Type)
		} else {
			info.DirTypeMode = filterModeGlob
		}
	}

	// Extract sampler info (only meaningful when sampling is active)
	if options.Sampling.IsSamplingActive() {
		info.NumFiles = options.Sampling.NoOf.Files
		info.NumFolders = options.Sampling.NoOf.Directories
		info.SampleLast = options.Sampling.InReverse
	}

	return info
}

// filter mode labels used by HeaderInfo.FileTypeMode / HeaderInfo.DirTypeMode
const (
	filterModeGlob  = "glob"
	filterModeRegex = "regex"
)

// determineFileType returns "regex" if the filter is a regex type,
// otherwise "glob". The caller must have already established that the
// filter is user-supplied and not the benign default; this helper only
// inspects the type.
func determineFileType(filterType enums.FilterType) string {
	if filterType == enums.FilterTypeRegex {
		return filterModeRegex
	}
	return filterModeGlob
}
func (c *Coordinator) captionFor(req *Request) string {
	subscription := ""
	switch req.Subscription {
	case enums.SubscribeFiles:
		subscription = "files only"
	case enums.SubscribeDirectories:
		subscription = "folders only"
	case enums.SubscribeDirectoriesWithFiles:
		subscription = "folders only /w files"
	case enums.SubscribeUniversal:
		subscription = "universal"
	case enums.SubscribeUndefined:
		subscription = "undefined"
	default:
		subscription = "files and folders"
	}

	if req.ActionName != "" {
		action, ok := c.config.Raw.Actions[req.ActionName]
		if ok {
			return fmt.Sprintf("%s • via '%s'", subscription, action.Cmd)
		}
	}

	if req.PipelineName != "" {
		return fmt.Sprintf("%s • via pipeline '%s'", subscription, req.PipelineName)
	}

	return subscription
}
