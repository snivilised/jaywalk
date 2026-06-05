package status

import (
	"time"

	bp "charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// defaultInnerFill is the bubbles progress fill used when the host
// does not call WithTheme. It is intentionally the same value the
// previous highway model hard-coded ("#B9FBC0") so the rendered
// appearance is preserved when no theme is supplied.
const defaultInnerFill = "#B9FBC0"

// defaultInnerWidth is the initial width (in cells) of the bubbles
// progress bar. Width is overwritten on receipt of WidthMsg.
const defaultInnerWidth = 10

// Model is the bubbletea Model for the status widget. It owns the
// state required to render the closing summary row of the highway
// view: counts, elapsed, the embedded bubbles progress bar and the
// appearance configuration. Implements tea.Model.
type Model struct {
	// counts
	files   int
	dirs    int
	errors  int
	skipped int

	// timing
	elapsed time.Duration

	// progress (embedded bubbles v2 progress). Drives the spring animation.
	inner    bp.Model
	width    int
	percent  int
	total    int
	done     int
	hasTotal bool

	// completion
	isDone bool
	errMsg string

	// appearance
	styles Styles
	fields FieldSelectors
}

// Option configures Model at construction time.
type Option func(*Model)

// WithStyles sets the lipgloss styles used to render the row. When
// also calling WithTheme, WithTheme wins for fields it populates.
func WithStyles(s Styles) Option {
	return func(m *Model) { m.styles = s }
}

// WithFields sets the segment visibility selectors. When also
// calling WithTheme, WithTheme wins for fields it populates.
func WithFields(f FieldSelectors) Option {
	return func(m *Model) { m.fields = f }
}

// WithWidth sets the initial row width and the embedded bubbles
// progress bar width. Updated again on receipt of WidthMsg.
func WithWidth(w int) Option {
	return func(m *Model) {
		m.width = w
		m.inner.SetWidth(w)
	}
}

// WithTheme populates Styles and the bubbles progress fill colour
// from a resolved theme. This is the recommended way to construct
// the widget from the highway view.
func WithTheme(t contract.Theme) Option {
	return func(m *Model) {
		m.styles = Styles{
			TreeIcons:         t.TreeIcons,
			SummaryLabelStyle: t.SummaryLabelStyle,
			SummaryValueStyle: t.SummaryValueStyle,
			ErrorStyle:        t.ErrorStyle,
			ProgressStyle:     t.ProgressStyle,
			BorderStyle:       t.BorderStyle,
			MutedStyle:        t.MutedStyle,
		}
		if fg := t.ProgressStyle.GetForeground(); fg != nil {
			m.inner.FullColor = fg
		}
	}
}

// New constructs a Model with the supplied options. The default
// embedded bubbles progress bar uses defaultInnerFill and
// defaultInnerWidth; both can be overridden via options.
func New(opts ...Option) Model {
	m := Model{
		inner: bp.New(
			bp.WithColors(lipgloss.Color(defaultInnerFill)),
			bp.WithoutPercentage(),
			bp.WithWidth(defaultInnerWidth),
		),
		width: defaultInnerWidth,
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// Init returns nil. The spring starts lazily on the first
// SetPercent call (issued from Update on a re-target message);
// there is no initial cmd to schedule.
func (m Model) Init() tea.Cmd { return nil }

// IsDone reports whether the widget has been marked as completed
// (i.e. a DoneMsg with IsDone=true was processed). The highway
// view uses this to decide whether to render the "press space to
// exit" footer.
func (m Model) IsDone() bool { return m.isDone }

// Height returns the number of terminal rows the rendered status
// row occupies. The status widget is a single line; this is a
// constant. Host views (highway, porthole) consult this to budget
// vertical space for the status row without re-rendering.
func (m Model) Height() int { return 1 }
