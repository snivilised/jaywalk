package highway

type Lane struct {
	Emoji           string
	Label           string
	FrameFunc       func(tick int) string
	Path            string
	Name            string
	IsDir           bool
	Depth           uint
	ActionName      string
	PipelineName    string
	CommandOutput   string
	ExecutionString string
	DryRun          bool
	Err             error

	tick int
}
