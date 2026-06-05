package flow

import "github.com/snivilised/jaywalk/src/prism/contract"

// LineRenderer is the optional, higher-fidelity capability exposed by
// the flow renderer. Views that need to capture each rendered Motif
// as a string (eg the porthole view's content buffer) depend on this
// interface rather than the base contract.Renderer.
//
// Implementations of contract.Renderer that also implement
// LineRenderer return the fully-styled line (including the trailing
// newline) from RenderLine. The base Show method is implemented in
// terms of RenderLine so the two paths always produce the same
// bytes.
type LineRenderer interface {
	RenderLine(motif contract.Motif) string
}
