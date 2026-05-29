package contract

// FrameFunc generates the animation frame string for a tick-driven spinner.
type FrameFunc func(tick int) string

const (
	Ellipses = "…"
)
