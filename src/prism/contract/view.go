package contract

// ViewKind identifies the rendering view to use. Defined as a typed
// string rather than an iota so that the set remains open - new views
// can be added in future without modifying this file.
type ViewKind string

const (
	// LinearView is a linear scrolling output view rendered with lipgloss.
	LinearView ViewKind = "linear"

	// PortholeView is a bubbletea view with a static header and footer
	// and vertically scrolling content between them.
	PortholeView ViewKind = "porthole"

	// LanesView is a bubbletea view showing parallel lanes of activity,
	// suited to concurrent worker output.
	LanesView ViewKind = "lanes"
)
