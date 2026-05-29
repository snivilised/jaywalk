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
		fmt.Println("🦄 DEBUG: Coordinator.ExecutePrime: peer aware ... 🦄")
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

	fmt.Println("🦁 DEBUG: Coordinator.ExecutePrime: executing live traversal only ... 🦁")
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
	cascadeDisplay, filesGlob, filesRegex, dirsGlob, dirsRegex, fileTypeMode, dirTypeMode :=
		extractHeaderInfo(req)

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

		// Header info fields
		CascadeDisplay: cascadeDisplay,
		FilesGlob:      filesGlob,
		FilesRegex:     filesRegex,
		DirsGlob:       dirsGlob,
		DirsRegex:      dirsRegex,
		FileTypeMode:   fileTypeMode,
		DirTypeMode:    dirTypeMode,
	})

	result, err := req.Scenario(facade, req.Settings...).Navigate(ctx)

	traversal.Err = err
	if result != nil {
		traversal.FilesVisited = result.Metrics().Count(enums.MetricNoFilesInvoked)
		traversal.DirsVisited = result.Metrics().Count(enums.MetricNoDirectoriesInvoked)
		traversal.Elapsed = result.Session().Elapsed()
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

	pool := newShellPoolExecutor(manifold)

	previousExec := c.exec
	c.exec = func(_ context.Context, command string) ([]byte, error) {
		return pool.Execute(ctx, command)
	}

	return func() {
		c.exec = previousExec
		pool.closeAll()
	}, nil
}

// extractHeaderInfo extracts cascade display (padlock or depth) and filter flag info from the Options
// that were passed through BuildTraversalSettings. This preserves display information
// even though the pref.Options struct doesn't directly expose the original flags.
func extractHeaderInfo(req *Request) (cascade, filesGlob, filesRegex, dirsGlob, dirsRegex string, fileType, dirType string) {
	options := &pref.Options{}

	// Extract options from settings by running all option functions
	for _, setting := range req.Settings {
		if setting == nil {
			continue
		}
		if err := setting(options); err != nil {
			return "", "", "", "", "", "", "" // error case - return empty
		}
	}

	// Extract cascade display (no-recurse or depth) from behaviours
	if options.Behaviours.Cascade.NoRecurse {
		cascade = "🔒"
	} else if options.Behaviours.Cascade.Depth > 0 {
		cascade = fmt.Sprintf("depth:%d", options.Behaviours.Cascade.Depth)
	}

	// Extract filter info from poly filter definitions
	if options.Filter.Node != nil && options.Filter.Node.Poly != nil {
		fileDef := options.Filter.Node.Poly.File
		dirDef := options.Filter.Node.Poly.Directory

		// Determine file pattern and type
		if fileDef.Type == enums.FilterTypeGlobEx && fileDef.Pattern != "" {
			filesGlob = fileDef.Pattern
		} else if fileDef.Type == enums.FilterTypeRegex && fileDef.Pattern != "" {
			filesRegex = fileDef.Pattern
		}

		// Determine directory pattern and type
		if dirDef.Type == enums.FilterTypeGlob && dirDef.Pattern != "" {
			dirsGlob = dirDef.Pattern
		} else if dirDef.Type == enums.FilterTypeRegex && dirDef.Pattern != "" {
			dirsRegex = dirDef.Pattern
		}

		// Determine filter types based on precedence (regex takes priority when both exist)
		fileType = determineFileType(fileDef.Type, fileDef.Pattern != "")
		dirType = determineFileType(dirDef.Type, dirDef.Pattern != "")
	}

	return cascade, filesGlob, filesRegex, dirsGlob, dirsRegex, fileType, dirType
}

// determineFileType returns "regex" if pattern is a regex, otherwise "glob".
// This helps distinguish display format but defaults to glob for backwards compatibility.
func determineFileType(filterType enums.FilterType, hasPattern bool) string {
	if filterType == enums.FilterTypeRegex && hasPattern {
		return "regex"
	}
	return "glob" // default
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
