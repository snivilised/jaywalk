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
	"github.com/snivilised/jaywalk/src/prism"
	"github.com/snivilised/jaywalk/src/prism/traffic"
	"github.com/snivilised/jaywalk/src/prism/traffic/highway"
)

// HighwayConfig holds configuration for the highway bubbletea view.
type HighwayConfig struct {
	// Pool is a space-separated list of emoji runes for decoration.
	Pool string

	// Separator between emoji and content info (default: " ").
	Separator string

	// SpinnerNames lists the spinner types to use for each lane, looked up
	// Categories expand via traffic.SpinnerCategories; individual spinner names
	// are registered in traffic.SpinnerNames.
	// When empty, buildHighwayLanes falls back to defaults.
	SpinnerNames []string

	// Labels for each lane, paired with SpinnerNames. When empty, defaults
	// are used.
	Labels []string

	// Overrides maps spinner name to interval in milliseconds.
	// When set, the lane's animation advances at this rate instead of the
	// global tick rate. Multiple lanes sharing the same spinner name each
	// get the same interval behaviour.
	Overrides map[string]int

	// AnimationGradient specifies the gradient to apply to frame animations.
	// Optional; nil or empty means use default styling without gradients.
	// When set, uses highlights.gradients configuration from palette (hi/lo endpoints).
	AnimationGradient string // name of gradient defined in theme palette (DEPRECATED - unused)
}

type highwayPresenter struct {
	program    *tea.Program
	done       chan struct{}
	cfg        HighwayConfig
	noW        uint
	maxDepth   uint
	totalFiles uint
	totalDirs  uint
	theme      prism.Theme
	noRecurse  bool
}

func newHighwayPresenter(palette prism.Palette, cfg HighwayConfig) (report.Presenter, error) {
	traffic.RegisterAll()
	theme, err := prism.NewTheme(palette, os.Stdout)
	if err != nil {
		return nil, err
	}
	return &highwayPresenter{cfg: cfg, theme: theme}, nil
}

func (h *highwayPresenter) OnTraversalOptions(o *pref.Options) {
	h.noW = o.Concurrency.NoW
	h.noRecurse = o.Behaviours.Cascade.NoRecurse
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

func (h *highwayPresenter) OnBegin(e *report.BeginEvent) {
	h.done = make(chan struct{})

	lanes := BuildHighwayLanes(h.cfg, h.noW)
	model := highway.NewModel(lanes, highwayTickRate, e.Root, h.maxDepth, h.theme, h.noRecurse)
	h.program = tea.NewProgram(model)

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

func (h *highwayPresenter) SetMaxDepth(maxDepth uint) {
	h.maxDepth = maxDepth
}

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
// The gradient is retrieved from the theme's HighlightsComponents for the
// highway-animation component using component-based lookup. This ensures
// gradients configured in themes are properly applied to animation frames.
func (h *highwayPresenter) sendMotif(path, name string, isDir bool, depth uint,
	actionName, pipelineName, commandOutput, executionString string, dryRun bool, err error) {
	select {
	case <-h.done:
		return
	default:
	}

	defer func() { _ = recover() }()

	var grad *prism.ResolvedGradient
	// Retrieve gradient by component name lookup (not direct gradient name).
	// GradientFor handles the component → gradient name → resolved gradient chain.
	g, has := h.theme.GradientFor(prism.GradientComponentHighwayAnimation)
	if has && g.Steps > 0 {
		grad = &prism.ResolvedGradient{Steps: g.Steps, Hi: g.Hi, Lo: g.Lo}
	}

	h.program.Send(highway.MotifMsg{
		Data: highway.MotifData{
			Path:            path,
			Name:            name,
			IsDir:           isDir,
			Depth:           depth,
			ActionName:      actionName,
			PipelineName:    pipelineName,
			CommandOutput:   commandOutput,
			ExecutionString: executionString,
			DryRun:          dryRun,
			Err:             err,
			Gradient:        grad,
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
	emojis := strings.Fields(cfg.Pool)
	if len(emojis) == 0 {
		emojis = defaultEmojiPool
	}

	deck := make([]string, len(emojis))
	copy(deck, emojis)
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	names := traffic.ExpandNames(cfg.SpinnerNames)
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
		def, ok := traffic.Lookup(name)
		if !ok {
			def, _ = traffic.Lookup(traffic.SpinnerTypeDefault)
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
			// HighlightGradient will be populated by the model when MotifMsg is received
			HighlightGradient: nil,
		}
	}
	return lanes
}
