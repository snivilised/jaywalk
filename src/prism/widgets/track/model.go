package track

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// defaultTickMs is the global tick rate used when WithTickRate is
// not supplied. Matches the implicit 50ms default the highway model
// used previously.
const defaultTickMs = 50

// LaneBarWidth is the width of the square depth bar rendered in
// each lane of the highway view. Mirrors the constant previously
// declared in the highway package.
const LaneBarWidth = 10

// SpinnerNameWidth is the fixed width allocated for the spinner
// name column in each lane. Names shorter than this are
// right-padded so all following columns stay aligned. Mirrors the
// constant previously declared in the highway package.
const SpinnerNameWidth = 24

// Model is the bubbletea Model for the track widget. It owns the
// per-lane animation state, the per-tick advance, the motif-data
// application, the deduplication map, and the per-lane rendering.
// The widget does NOT own its ticker; the highway root drives
// tea.Tick and forwards TickMsg to Update.
type Model struct {
	lanes     []Lane
	skip      []int
	width     int
	counted   map[string]bool
	files     int
	dirs      int
	maxDepth  uint
	noRecurse bool
	tickRate  time.Duration
	styles    Styles

	// visibleCount controls how many lanes are rendered. Starts at 1
	// and grows as unique pool WorkerIDs are observed, up to the total
	// number of pre-allocated lanes (len(lanes)).
	visibleCount int

	// seenWorkers tracks which pool WorkerIDs have been observed. Each
	// new WorkerID triggers a potential visibleCount expansion.
	seenWorkers map[string]struct{}
}

// Option configures Model at construction time. Mirrors the
// status.Option pattern.
type Option func(*Model)

// WithLanes sets the initial lane slice and computes the per-lane
// skip factors based on the currently-supplied tick rate. The
// order of options matters: WithTickRate should appear before
// WithLanes if both are used. When WithLanes is called more than
// once, the most recent call wins and skip is recomputed.
func WithLanes(lanes []Lane) Option {
	return func(m *Model) {
		m.lanes = lanes
		m.skip = initLaneSkip(lanes, m.tickRate)
		m.counted = make(map[string]bool)
		m.seenWorkers = make(map[string]struct{})

		// Start with a single visible lane. The view will render up
		// to visibleCount lanes; remaining lanes are revealed as new
		// pool WorkerIDs are observed.
		m.visibleCount = 1
		if m.visibleCount > len(lanes) {
			m.visibleCount = len(lanes)
		}
	}
}

// WithTickRate sets the global tick rate used for skip-factor
// computation. A zero or negative value falls back to 50ms.
func WithTickRate(rate time.Duration) Option {
	return func(m *Model) {
		if rate <= 0 {
			rate = defaultTickMs * time.Millisecond
		}
		m.tickRate = rate
		// Recompute skip if lanes are already set; the caller may
		// also have supplied them before the rate.
		if m.lanes != nil {
			m.skip = initLaneSkip(m.lanes, m.tickRate)
		}
	}
}

// WithMaxDepth sets the maximum tree depth observed during the
// preview pass. Used by the periscope bar fill formula.
func WithMaxDepth(d uint) Option {
	return func(m *Model) { m.maxDepth = d }
}

// WithNoRecurse sets the no-recurse flag. When true the periscope
// bar renders as a single filled glyph (no animation, no depth
// fill calculation).
func WithNoRecurse(b bool) Option {
	return func(m *Model) { m.noRecurse = b }
}

// WithWidth sets the initial layout width. Updated again on
// receipt of WidthMsg.
func WithWidth(w int) Option {
	return func(m *Model) { m.width = w }
}

// WithStyles sets the lipgloss styles used at render time. When
// also calling WithTheme, WithTheme wins for fields it populates.
func WithStyles(s Styles) Option {
	return func(m *Model) { m.styles = s }
}

// WithTheme populates Styles from a resolved theme. This is the
// recommended way to construct the widget from the highway view.
func WithTheme(t contract.Theme) Option {
	return func(m *Model) {
		m.styles = Styles{
			BarFilledStyle:    t.BarFilledStyle,
			BarEmptyStyle:     t.BarEmptyStyle,
			ErrorStyle:        t.ErrorStyle,
			ActionStyle:       t.ActionStyle,
			PipelineStyle:     t.PipelineStyle,
			DirStyle:          t.DirStyle,
			FileStyle:         t.FileStyle,
			MutedStyle:        t.MutedStyle,
			TreeIcons:         t.TreeIcons,
			FrameStyle:        t.FrameStyle,
			BorderStyle:       t.BorderStyle,
			BranchStyle:       t.BranchStyle,
			LandingStripStyle: t.LandingStripStyle,
			IdleStyle:         t.WorkerIdleStyle,
			WorkingStyle:      t.WorkerStyle,
		}
	}
}

// New constructs a Model with the supplied options. The default
// tick rate is 50ms; the default width is 0 (caller is expected to
// send a WidthMsg before View).
func New(opts ...Option) Model {
	m := Model{
		tickRate: defaultTickMs * time.Millisecond,
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// Init returns nil. The highway root drives tea.Tick; the widget
// does not start its own timer. This mirrors the status widget
// pattern.
func (m Model) Init() tea.Cmd { return nil }

// Files returns the current files count. Set by MotifMsg
// processing (one increment per unique non-dir path).
func (m Model) Files() int { return m.files }

// Dirs returns the current directories count. Set by MotifMsg
// processing (one increment per unique dir path).
func (m Model) Dirs() int { return m.dirs }

// Lanes returns the current lane slice. Read-only accessor for
// tests and callers that need to inspect lane state directly.
func (m Model) Lanes() []Lane { return m.lanes }

// Tick returns the current frame tick of the lane at idx. The
// tick advances on every TickMsg for lanes with IntervalMs=0
// (skip factor 0) and on every Nth tick for lanes with a skip
// factor of N. Exposed for tests.
func (m Model) Tick(idx int) int {
	if idx < 0 || idx >= len(m.lanes) {
		return 0
	}
	return m.lanes[idx].tick
}

// MaxDepth returns the maximum tree depth observed via CensusMsg.
func (m Model) MaxDepth() uint { return m.maxDepth }

// VisibleCount returns the number of lanes currently rendered. Exposed
// for tests and the highway model to inspect expansion state.
func (m Model) VisibleCount() int { return m.visibleCount }

// initLaneSkip computes the per-lane skip factor from each lane's
// IntervalMs. The skip factor = IntervalMs / tickRate (in ms). A
// lane with no override (IntervalMs=0) gets factor 0 - it advances
// every tick. A lane with IntervalMs=5000 at 50ms tick rate gets
// factor 100, advancing one frame every 100 ticks (every 5
// seconds). Moved verbatim from highway.initLaneSkip.
func initLaneSkip(lanes []Lane, tickRate time.Duration) []int {
	factors := make([]int, len(lanes))
	tickMs := int(tickRate.Milliseconds())
	if tickMs == 0 {
		tickMs = defaultTickMs
	}
	for i, lane := range lanes {
		if lane.IntervalMs > 0 {
			factors[i] = lane.IntervalMs / tickMs
			if factors[i] < 1 {
				factors[i] = 1
			}
		}
	}
	return factors
}
