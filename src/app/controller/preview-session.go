package controller

import (
	"sync/atomic"
	"time"

	"github.com/snivilised/jaywalk/src/agenor/core"
)

// PreviewSession is a lightweight session implementation for the tweak
// preview traversal. It implements core.Session but strips out all
// persistence, resume, and fault handling. Timing methods return zero
// values — they are not meaningful for a disposable preview session.
// Ctrl-C exits cleanly via context cancellation in the caller — no
// navigation state is saved.
type PreviewSession struct {
	done atomic.Bool
}

// NewPreviewSession creates a PreviewSession with IsComplete() == false.
func NewPreviewSession() *PreviewSession {
	return &PreviewSession{}
}

// IsComplete returns true when MarkComplete has been called.
func (s *PreviewSession) IsComplete() bool {
	return s.done.Load()
}

// StartedAt returns the zero time. Not meaningful for preview sessions.
func (s *PreviewSession) StartedAt() time.Time {
	return time.Time{}
}

// Elapsed returns 0. Not meaningful for preview sessions.
func (s *PreviewSession) Elapsed() time.Duration {
	return 0
}

// MarkComplete sets the session as completed. Called by the tweak
// coordinator when a preview traversal finishes normally. On Ctrl-C
// the context cancels the traversal and MarkComplete is never called,
// so IsComplete remains false.
func (s *PreviewSession) MarkComplete() {
	s.done.Store(true)
}

// compile-time check: *PreviewSession satisfies core.Session.
var _ core.Session = (*PreviewSession)(nil)
