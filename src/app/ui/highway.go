package ui

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
	"github.com/snivilised/jaywalk/src/prism/movies"
	"github.com/snivilised/jaywalk/src/prism/views/highway"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
	"github.com/snivilised/jaywalk/src/prism/widgets/track"
)

// HighwayConfig is declared in registry.go alongside the other view
// configs. The definition lives there so all view config types are
// co-located with LoadConfig and New.

type highwayPresenter struct {
	presenter
	cfg HighwayConfig

	// Job emoji pool state - populated from config, random emoji picked per job arrival.
	jobEmojiPool []string
}

func newHighwayPresenter(palette contract.Palette, hCfg HighwayConfig) (report.Presenter, error) {
	movies.RegisterAll()
	theme, err := contract.NewTheme(palette, os.Stdout)
	if err != nil {
		return nil, err
	}
	return &highwayPresenter{presenter: presenter{theme: theme}, cfg: hCfg}, nil
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

// OnBegin creates the bubbletea model and starts the program.
func (h *highwayPresenter) OnBegin(e *report.BeginEvent) {
	h.initJobEmojiPool()

	// Cache the header info for introspection and for use below when
	// building the OvertureMsg.
	h.header = e.Header

	// Derive noRecurse from cascade display. The two are mutually exclusive
	// at the flag-binding layer, so only one of 🔒/depth can be present.
	h.noRecurse = strings.Contains(h.header.CascadeDisplay, contract.Static.Emoji.Padlock)

	// Build the banner info once at startup. The random aspects are
	// frozen here so the banner reads the same on every render.
	h.bannerInfo = h.buildBannerInfo()

	lanes := BuildHighwayLanes(h.cfg, h.noW)
	model := highway.NewModel(contract.NewModelParams{
		RootPath:  e.Root,
		MaxDepth:  h.maxDepth,
		Theme:     h.theme,
		NoRecurse: h.noRecurse,
	}, lanes, highwayTickRate)
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
		OvertureMsg: contract.OvertureMsg{
			Root:              e.Root,
			Caption:           e.Caption,
			SubscriptionLabel: subscriptionLabelFor(e.Subscription),
			StartedAt:         e.StartedAt,
			DateFormat:        e.DateFormat,
			PipelineName:      e.PipelineName,
			Header:            h.header,
			FlagsRowPosition:  h.cfg.FlagsRowPosition,
		},
		ActionName: e.ActionName,
		Banner:     h.bannerInfo,
	})

	if h.totalFiles > 0 || h.totalDirs > 0 {
		h.program.Send(contract.CensusMsg{
			TotalFiles: h.totalFiles,
			TotalDirs:  h.totalDirs,
			MaxDepth:   h.maxDepth,
		})
	}
}

func (h *highwayPresenter) OnNodeEvent(e *report.NeutralEvent) {
	h.sendMotif(contract.MotifMsg{
		Path:  e.Node.Path,
		Name:  e.Node.Extension.Name,
		IsDir: e.Node.IsDirectory(),
		Depth: uint(e.Node.Extension.Depth),
	})
}

func (h *highwayPresenter) OnActionEvent(e *report.ActionEvent) {
	h.sendMotif(contract.MotifMsg{
		WorkerID:        e.WorkerID,
		Path:            e.Node.Path,
		Name:            e.Node.Extension.Name,
		IsDir:           e.Node.IsDirectory(),
		Depth:           uint(e.Node.Extension.Depth),
		ActionName:      e.Name,
		CommandOutput:   e.CommandOutput,
		ExecutionString: e.ExecutionString,
		DryRun:          e.DryRun,
		Err:             e.Err,
	})
}

func (h *highwayPresenter) OnPipelineEvent(e *report.PipelineEvent) {
	h.sendMotif(contract.MotifMsg{
		WorkerID:        e.WorkerID,
		Path:            e.Node.Path,
		Name:            e.Node.Extension.Name,
		IsDir:           e.Node.IsDirectory(),
		Depth:           uint(e.Node.Extension.Depth),
		PipelineName:    e.Name,
		CommandOutput:   e.CommandOutput,
		ExecutionString: e.ExecutionString,
		DryRun:          e.DryRun,
		Err:             e.Err,
	})
}

func (h *highwayPresenter) OnSkipEvent(e *report.SkipEvent) {
	h.sendMotif(contract.MotifMsg{
		WorkerID:   e.WorkerID,
		Path:       e.Node.Path,
		Name:       e.Node.Extension.Name,
		IsDir:      e.Node.IsDirectory(),
		Depth:      uint(e.Node.Extension.Depth),
		ActionName: e.Name,
	})
}

// sendMotif sends a motif message with optional gradient overlay.
// The gradient is retrieved from the theme's HighlightsComponents using
// component-based lookup. This ensures gradients configured in themes
// are properly applied to animation frames.
func (h *highwayPresenter) sendMotif(msg contract.MotifMsg) {
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

	msg.JobEmoji = jobEmoji
	msg.Gradient = grad
	msg.PeriscopeGradient = periscopeGrad

	h.program.Send(msg)
}

func (h *highwayPresenter) OnWorkerState(state enums.WorkerState, workerID string) {
	select {
	case <-h.done:
		return
	default:
	}

	defer func() { _ = recover() }()

	laneCount := int(h.noW) //nolint:gosec // bounded by config
	if laneCount < 1 {
		laneCount = defaultLaneCount
	}
	laneIndex := track.WorkerIndex(workerID) - 1
	if laneIndex < 0 || laneIndex >= laneCount {
		laneIndex = 0
	}

	h.program.Send(contract.WorkerStateMsg{
		LaneID: laneIndex,
		State:  state,
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

	h.program.Send(contract.CompleteMsg{
		Files:   int(t.FilesVisited),
		Dirs:    int(t.DirsVisited),
		Errs:    errs,
		Elapsed: t.Elapsed,
	})
	<-h.done
}

func BuildHighwayLanes(cfg HighwayConfig, now uint) []track.Lane {
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

	lanes := make([]track.Lane, numLanes)
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

		lanes[i] = track.Lane{
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

// buildBannerInfo assembles the per-session banner.Info for the highway
// model. The random aspect selection is performed exactly once here
// (math/rand/v2, package-level source) so the aspects are frozen for
// the duration of the process. The gradient endpoints are resolved
// from the theme's "banner-control" component, falling back to no
// gradient (which causes the widget to render as plain text) when
// the binding is absent.
//
// When the user has disabled the banner, the returned banner.Info has
// Disable=true and the gradient/state pointers are nil. The highway
// model treats this as "do not render".
func (h *highwayPresenter) buildBannerInfo() banner.Info {
	info := banner.Info{
		Disable:  h.cfg.Banner.Disable,
		Position: h.cfg.Banner.Position,
		Justify:  h.cfg.Banner.Justify,
	}

	if h.cfg.Banner.Disable {
		return info
	}

	// Resolve the gradient from the theme. The "banner-control"
	// component must be bound to a gradient in the theme YAML;
	// when missing the banner renders as plain text.
	//
	// StepsOverride (from ui.highway.banner.steps) lets the user
	// keep sharing the gradient's colour endpoints with other
	// widgets (so the colour scheme stays consistent) but tune
	// the banner's sweep smoothness independently. Zero means
	// "use the gradient's own steps".
	var grad *contract.ResolvedGradient
	if g, ok := h.theme.GradientFor(contract.GradientComponentBanner); ok && g.Steps > 0 {
		steps := g.Steps
		if h.cfg.Banner.StepsOverride > 0 {
			steps = h.cfg.Banner.StepsOverride
		}
		grad = &contract.ResolvedGradient{Steps: steps, Hi: g.Hi, Lo: g.Lo}
	}
	info.Gradient = grad

	if grad != nil {
		st := effects.NewGradientState()
		st.TotalSteps = grad.Steps
		info.State = st
	}

	// Pick the random aspects ONCE here, not per-render. Use the
	// package-level random source (math/rand/v2) which is fine for
	// non-security purposes.
	rng := rand.New(rand.NewPCG(uint64(os.Getpid()), uint64(core.Now().UnixNano()))) //nolint:gosec // non-security
	info.Aspects = banner.RandomiseAspects(rng)

	// Resolve the per-tick interval. Tick in the config is in
	// milliseconds; the model uses a time.Duration.
	tick := h.cfg.Banner.Tick
	if tick <= 0 {
		tick = bannerDefaultTickMs
	}
	info.Tick = time.Duration(tick) * time.Millisecond

	return info
}
