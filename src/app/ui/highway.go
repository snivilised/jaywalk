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
	// Pool is a space-separated list of emoji runes for lane decoration.
	Pool string

	// Separator between emoji and content info (default: " ").
	Separator string

	// SpinnerNames lists the spinner types to use for each lane, looked up
	// from prism/traffic (e.g. "film-strip", "pulse", "spinner").
	// When empty, buildHighwayLanes falls back to defaults.
	SpinnerNames []string

	// Labels for each lane, paired with SpinnerNames. When empty, defaults
	// are used.
	Labels []string
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
}

func newHighwayPresenter(palette prism.Palette, cfg HighwayConfig) (report.Presenter, error) {
	theme, err := prism.NewTheme(palette, os.Stdout)
	if err != nil {
		return nil, err
	}
	return &highwayPresenter{cfg: cfg, theme: theme}, nil
}

func (h *highwayPresenter) OnTraversalOptions(o *pref.Options) {
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

func (h *highwayPresenter) OnBegin(e *report.BeginEvent) {
	h.done = make(chan struct{})

	lanes := buildHighwayLanes(h.cfg, h.noW)
	model := highway.NewModel(lanes, highwayTickRate, e.Root, h.maxDepth, h.theme)
	h.program = tea.NewProgram(model)

	go func() {
		_, _ = h.program.Run()
		close(h.done)
	}()

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

func (h *highwayPresenter) sendMotif(path, name string, isDir bool, depth uint,
	actionName, pipelineName, commandOutput, executionString string, dryRun bool, err error,
) {
	select {
	case <-h.done:
		return
	default:
	}

	defer func() { _ = recover() }()
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

func buildHighwayLanes(cfg HighwayConfig, now uint) []highway.Lane {
	emojis := strings.Fields(cfg.Pool)
	if len(emojis) == 0 {
		emojis = defaultEmojiPool
	}

	deck := make([]string, len(emojis))
	copy(deck, emojis)
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	names := cfg.SpinnerNames
	labels := cfg.Labels

	numLanes := len(names)
	if numLanes == 0 {
		numLanes = int(now) //nolint:gosec // ok
		if numLanes < 1 {
			numLanes = defaultLaneCount
		}
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
		lanes[i] = highway.Lane{
			Emoji:     deck[i%len(deck)],
			Label:     label,
			FrameFunc: def.Frames,
		}
	}
	return lanes
}
