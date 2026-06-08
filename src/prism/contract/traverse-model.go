package contract

import (
	"strings"
	"time"
)

// TraverseModel holds the common properties shared between the
// highway and porthole bubbletea models. Embed this struct to
// eliminate field duplication for traversal chrome state.
type TraverseModel struct {
	Width             int
	Start             time.Time
	TickRate          time.Duration
	RootPath          string
	Done              bool
	Errors            int
	Elapsed           time.Duration
	ErrMsg            string
	PipelineName      string
	MaxDepth          uint
	NoRecurse         bool
	SubscriptionLabel string
	StartedAt         time.Time
	Caption           string
	DateFormat        string
	Theme             Theme
	Header            HeaderInfo
	FlagsRowPosition  string
}

// WriteSeparator emits a horizontal border line used to frame the
// legend section on both sides. The legend widget itself is
// layout-agnostic; the surrounding borders are the view's concern.
func (m *TraverseModel) WriteSeparator(b *strings.Builder) {
	dashes := strings.Repeat("─", max(0, m.Width-2))
	b.WriteString(m.Theme.BorderStyle.Render("├" + dashes + "┤"))
	b.WriteString("\n")
}

// ApplyOverture populates the traverse model fields from an
// OvertureMsg. Call this from each model's OvertureMsg handler
// after storing view-specific fields.
func (m *TraverseModel) ApplyOverture(msg *OvertureMsg) {
	m.RootPath = msg.Root
	m.SubscriptionLabel = msg.SubscriptionLabel
	m.StartedAt = msg.StartedAt
	m.Caption = msg.Caption
	m.DateFormat = msg.DateFormat
	m.PipelineName = msg.PipelineName
	m.Header = msg.Header

	m.FlagsRowPosition = msg.FlagsRowPosition
	if m.FlagsRowPosition != PositionTop && m.FlagsRowPosition != PositionBottom {
		m.FlagsRowPosition = PositionBottom
	}
}

// ApplyCompletion populates done/errMsg/errors/elapsed from a
// CompleteMsg. Call this from each model's CompleteMsg handler.
func (m *TraverseModel) ApplyCompletion(errs []error, elapsed time.Duration) {
	m.Done = true
	m.ErrMsg = ""
	m.Errors = len(errs)
	m.Elapsed = elapsed
	for _, e := range errs {
		if e != nil {
			m.ErrMsg = e.Error()
			break
		}
	}
}
