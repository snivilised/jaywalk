package enums

import "fmt"

//go:generate stringer -type=ViewKind -linecomment -trimprefix=ViewKind -output view-kind-en-auto.go

// ViewKind identifies the rendering view to use.
type ViewKind uint

const (
	// ViewKindLinear is a linear scrolling output view rendered with lipgloss.
	ViewKindLinear ViewKind = iota // linear

	// ViewKindPorthole is a bubbletea view with a static header and footer
	// and vertically scrolling content between them.
	ViewKindPorthole // porthole

	// ViewKindHighway is a bubbletea view showing parallel lanes of activity,
	// suited to concurrent worker output.
	ViewKindHighway // highway
)

// UnmarshalText implements encoding.TextUnmarshaler so that ViewKind
// can be decoded from string values.
func (k *ViewKind) UnmarshalText(data []byte) error {
	switch string(data) {
	case "linear":
		*k = ViewKindLinear
	case "porthole":
		*k = ViewKindPorthole
	case "highway":
		*k = ViewKindHighway
	default:
		return fmt.Errorf("unknown view kind %q: valid values are linear, porthole, highway", string(data))
	}
	return nil
}
