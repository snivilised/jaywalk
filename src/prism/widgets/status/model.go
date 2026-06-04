package status

import (
	"fmt"
	"image/color"
	"time"

	tea "charm.land/bubbletea/v2"
	bp "github.com/charmbracelet/bubbles/progress"

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
// view: counts, elapsed, the embedded bubbles progress bar (used
// for its ViewAs renderer) and the appearance configuration.
// Implements tea.Model.
type Model struct {
	// counts
	files   int
	dirs    int
	errors  int
	skipped int

	// timing
	elapsed time.Duration

	// progress (embedded bubbles progress + bookkeeping).
	// The bubbles model is used purely for its static ViewAs
	// renderer; the spring animation is intentionally disabled
	// because the bubbles library uses bubbletea v1 cmd types
	// that are incompatible with the project's bubbletea v2.
	// TODO(phase-2): adopt bubbles v2 progress (or re-implement
	// the bar with our own lipgloss rendering) so the bar can
	// smoothly animate between percent values. Track in
	// issue: progress-animation-revival.
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
		m.inner.Width = w
	}
}

// WithTheme populates Styles and the bubbles progress fill colour
// from a resolved theme. This is the recommended way to construct
// the widget from the highway view. The fill colour is extracted
// from theme.ProgressStyle's foreground via a small hex conversion
// because bubbles' WithSolidFill takes a string.
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
			m.inner.FullColor = colorToHex(fg)
		}
	}
}

// New constructs a Model with the supplied options. The default
// embedded bubbles progress bar uses defaultInnerFill and
// defaultInnerWidth; both can be overridden via options.
func New(opts ...Option) Model {
	m := Model{
		inner: bp.New(
			bp.WithSolidFill(defaultInnerFill),
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

// Init returns nil. There is no per-widget cmd to schedule: the
// progress bar is rendered statically from m.percent in View, so
// no animation driver is needed.
func (m Model) Init() tea.Cmd { return nil }

// IsDone reports whether the widget has been marked as completed
// (i.e. a DoneMsg with IsDone=true was processed). The highway
// view uses this to decide whether to render the "press space to
// exit" footer.
func (m Model) IsDone() bool { return m.isDone }

// colorToHex converts a color.Color to the "#RRGGBB" string form
// expected by bubbles' WithSolidFill. Alpha is ignored. This is
// the only hex conversion in the status package; if other widgets
// need the same conversion, hoist it to src/internal/third.
// TODO(shared-utility): when a second widget needs hex
// conversion, move colorToHex to src/internal/third/colour.
func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X",
		uint8(r>>8), //nolint:gosec // safe: 16-bit channel >> 8 fits in 8 bits
		uint8(g>>8), //nolint:gosec
		uint8(b>>8), //nolint:gosec
	)
}
