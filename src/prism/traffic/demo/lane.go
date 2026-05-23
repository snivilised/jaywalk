package demo

type Lane struct {
	Emoji     string
	Label     string
	FrameFunc func(tick int) string

	tick int
}
