package ui

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/agenor/pref"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/highway"
	"github.com/snivilised/jaywalk/src/prism/movies"
)

// HighwayConfig is declared in registry.go alongside the other view
// configs. The definition lives there so all ViewConfig
// implementations are co-located with LoadConfig and New.

type highwayPresenter struct {
	program    *tea.Program
	done       chan struct{}
	cfg        HighwayConfig
	noW        uint
	maxDepth   uint
	totalFiles uint
	totalDirs  uint
	theme      contract.Theme
	noRecurse  bool

	// Header info fields for OvertureMsg
	cascadeDisplay string // "🔒", "depth:N", or ""
	filesGlob      string // raw pattern when --files-glob used
	filesRegex     string // raw pattern when --files-regex used
	dirsGlob       string // raw pattern when --dirs-glob used
	dirsRegex      string // raw pattern when --dirs-regex used
	fileTypeMode   string // "glob" or "regex" for files, default "glob"
	dirTypeMode    string // "glob" or "regex" for dirs, default "glob"

	// Job emoji pool state — populated from config, random emoji picked per job arrival.
	jobEmojiPool []string
}

func newHighwayPresenter(palette contract.Palette, hCfg HighwayConfig) (report.Presenter, error) {
	movies.RegisterAll()
	theme, err := contract.NewTheme(palette, os.Stdout)
	if err != nil {
		return nil, err
	}
	return &highwayPresenter{cfg: hCfg, theme: theme}, nil
}

func (h *highwayPresenter) OnTraversalOptions(o *pref.Options) {
	// Read concurrency settings from options - this is structural configuration, not cascade state.
	h.noW = o.Concurrency.NoW
}

func subscriptionLabelFor(s enums.Subscription) string {
	switch s {
	case enums.SubscribeUndefined:
		return "undefined"
	case enums.SubscribeFiles:
		return "files only"
	case enums.SubscribeDirectories:
		return "folders only"
	case enums.SubscribeDirectoriesWithFiles:
		return "directories w/ files"
	case enums.SubscribeUniversal:
		return "files and folders"
	default:
		return "files and folders"
	}
}

// initJobEmojiPool loads the job emoji pool from config or defaults.
// A random emoji is picked from the pool on each job arrival.
func (h *highwayPresenter) initJobEmojiPool() {
	jobEmojis := strings.Fields(h.cfg.JobPool)
	if len(jobEmojis) == 0 {
		jobEmojis = defaultJobEmojiPool
	}
	h.jobEmojiPool = make([]string, len(jobEmojis))
	copy(h.jobEmojiPool, jobEmojis)
}

func (h *highwayPresenter) SetMaxDepth(maxDepth uint) {
	h.maxDepth = maxDepth
}

// OnBegin creates the bubbletea model and starts the program.
func (h *highwayPresenter) OnBegin(e *report.BeginEvent) {
	h.initJobEmojiPool()

	// todo: derive noRecurse from e.CascadeDisplay in OnBegin or pass via options
	lanes := BuildHighwayLanes(h.cfg, h.noW)
	model := highway.NewModel(lanes, highwayTickRate, e.Root, h.maxDepth, h.theme, h.noRecurse)
	h.program = tea.NewProgram(model)

	// Initialize done channel before starting the goroutine that closes it.
	// This prevents "close of nil channel" panic when Ctrl+C interrupts early.
	if h.done == nil {
		h.done = make(chan struct{})
	}

	go func() {
		_, _ = h.program.Run()
		close(h.done)
	}()

	// Monitor for premature bubbletea exit (e.g., user pressed Ctrl-C).
	// When the program exits before OnComplete is called, cancel the
	// traversal context so the navigator stops and saves resume state.
	if e.Cancel != nil {
		go func() {
			<-h.done
			e.Cancel()
		}()
	}

	h.program.Send(highway.OvertureMsg{
		Root:              e.Root,
		Caption:           e.Caption,
		SubscriptionLabel: subscriptionLabelFor(e.Subscription),
		StartedAt:         e.StartedAt,
		DateFormat:        e.DateFormat,
		ActionName:        e.ActionName,
		PipelineName:      e.PipelineName,

		// Header info for filter widgets and cascade display
		CascadeDisplay: h.cascadeDisplay,
		FilesGlob:      h.filesGlob,
		FilesRegex:     h.filesRegex,
		DirsGlob:       h.dirsGlob,
		DirsRegex:      h.dirsRegex,
		FileTypeMode:   h.fileTypeMode,
		DirTypeMode:    h.dirTypeMode,
	})

	if h.totalFiles > 0 || h.totalDirs > 0 {
		h.program.Send(highway.CensusMsg{
			TotalFiles: h.totalFiles,
			TotalDirs:  h.totalDirs,
			MaxDepth:   h.maxDepth,
		})
	}
}

func (h *highwayPresenter) NeedsPeerInfo() bool {
	return true
}

func (h *highwayPresenter) OnPeerInfoBegin(files, dirs uint, _ map[string]*core.PeerInfo) {
	h.totalFiles = files
	h.totalDirs = dirs
}

func (h *highwayPresenter) OnPeerInfoEnd() {}

func (h *highwayPresenter) OnNodeEvent(e *report.NeutralEvent) {
	h.sendMotif(e.Node.Path, e.Node.Extension.Name, e.Node.IsDirectory(),
		uint(e.Node.Extension.Depth), "", "", "", "", false, nil)
}

func (h *highwayPresenter) OnActionEvent(e *report.ActionEvent) {
	h.sendMotif(e.Node.Path, e.Node.Extension.Name, e.Node.IsDirectory(),
		uint(e.Node.Extension.Depth), e.Name, "", e.CommandOutput, e.ExecutionString, e.DryRun, e.Err)
}

func (h *highwayPresenter) OnPipelineEvent(e *report.PipelineEvent) {
	h.sendMotif(e.Node.Path, e.Node.Extension.Name, e.Node.IsDirectory(),
		uint(e.Node.Extension.Depth), "", e.Name, e.CommandOutput, e.ExecutionString, e.DryRun, e.Err)
}

func (h *highwayPresenter) OnSkipEvent(e *report.SkipEvent) {
	h.sendMotif(e.Node.Path, e.Node.Extension.Name, e.Node.IsDirectory(),
		uint(e.Node.Extension.Depth), e.Name, "", "", "", false, nil)
}

// sendMotif sends a motif message with optional gradient overlay.
// The gradient is retrieved from the theme's HighlightsComponents using
// component-based lookup. This ensures gradients configured in themes
// are properly applied to animation frames.
func (h *highwayPresenter) sendMotif(path, name string, isDir bool, depth uint,
	actionName, pipelineName, commandOutput, executionString string, dryRun bool, err error) {
	select {
	case <-h.done:
		return
	default:
	}

	defer func() { _ = recover() }()

	// Pick a random job emoji from the pool for each job arrival.
	jobEmoji := h.jobEmojiPool[rand.IntN(len(h.jobEmojiPool))] //nolint:gosec // non-security random

	var grad *contract.ResolvedGradient
	// Retrieve gradient by component name lookup (not direct gradient name).
	g, has := h.theme.GradientFor(contract.GradientComponentActivity)
	if has && g.Steps > 0 {
		grad = &contract.ResolvedGradient{Steps: g.Steps, Hi: g.Hi, Lo: g.Lo}
	}

	var periscopeGrad *contract.ResolvedGradient
	pg, hasPG := h.theme.GradientFor(contract.GradientComponentPeriscope)
	if hasPG && pg.Steps > 0 {
		periscopeGrad = &contract.ResolvedGradient{Steps: pg.Steps, Hi: pg.Hi, Lo: pg.Lo}
	}

	h.program.Send(highway.MotifMsg{
		Data: highway.MotifData{
			Path:              path,
			Name:              name,
			IsDir:             isDir,
			Depth:             depth,
			ActionName:        actionName,
			PipelineName:      pipelineName,
			CommandOutput:     commandOutput,
			ExecutionString:   executionString,
			DryRun:            dryRun,
			Err:               err,
			JobEmoji:          jobEmoji,
			Gradient:          grad,
			PeriscopeGradient: periscopeGrad,
		},
	})
}

func (h *highwayPresenter) OnComplete(t *report.Traversal) {
	select {
	case <-h.done:
		return
	default:
	}

	defer func() { _ = recover() }()

	var errs []error
	if t.Err != nil {
		errs = []error{t.Err}
	}

	h.program.Send(highway.CompleteMsg{
		Files:   int(t.FilesVisited),
		Dirs:    int(t.DirsVisited),
		Errs:    errs,
		Elapsed: t.Elapsed,
	})
	<-h.done
}

func BuildHighwayLanes(cfg HighwayConfig, now uint) []highway.Lane {
	// Static worker emoji allocation: shuffle once, assign sequentially per lane.
	emojis := strings.Fields(cfg.WorkerPool)
	if len(emojis) == 0 {
		emojis = defaultWorkerEmojiPool
	}

	deck := make([]string, len(emojis))
	copy(deck, emojis)
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	names := movies.ExpandNames(cfg.SpinnerNames)
	labels := cfg.Labels

	numLanes := int(now) //nolint:gosec // ok
	if numLanes < 1 {
		numLanes = defaultLaneCount
	}

	lanes := make([]highway.Lane, numLanes)
	for i := 0; i < numLanes; i++ {
		var name string
		if len(names) > 0 {
			name = names[i%len(names)]
		} else {
			name = highwaySpinnerTypes[i%len(highwaySpinnerTypes)]
		}
		var label string
		if len(labels) > 0 {
			label = labels[i%len(labels)]
		} else if i < len(defaultLabels) {
			label = defaultLabels[i]
		} else {
			label = fmt.Sprintf("Worker %d", i+1)
		}
		def, ok := movies.Lookup(name)
		if !ok {
			def, _ = movies.Lookup(movies.SpinnerTypeDefault)
		}
		interval := cfg.Overrides[name]
		if interval < 1 {
			interval = 0
		}

		lanes[i] = highway.Lane{
			Emoji:       deck[i%len(deck)],
			Label:       label,
			FrameFn:     def.Frames,
			SpinnerName: name,
			IntervalMs:  interval,
			// JobEmoji is populated dynamically by the model when MotifMsg is received
			HighlightGradient: nil,
		}
	}
	return lanes
}
