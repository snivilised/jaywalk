package highway

const (
	// LaneBarWidth is the width of the square depth bar rendered
	// in each lane of the highway view.
	LaneBarWidth = 10

	// SpinnerNameWidth is the fixed width allocated for the spinner
	// name column in each lane. Names shorter than this are right-padded
	// so all following columns stay aligned.
	SpinnerNameWidth = 18
)

// Flags row placement - mirrors ui.FlagsRowPosition*.
const (
	// FlagsRowPositionTop places the flags row immediately after the top
	// border and before the first highway lane.
	FlagsRowPositionTop = "top"

	// FlagsRowPositionBottom places the flags row immediately above the
	// status line and below the last highway lane (the default).
	FlagsRowPositionBottom = "bottom"
)
