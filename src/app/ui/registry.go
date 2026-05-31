package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/flow"
)

// ---------------------------------------------------------------------------
// Named display modes - legal values for --tui
// ---------------------------------------------------------------------------

const (
	// ModeLinear is the default linear view. One styled line per node
	// via prism's linear renderer with lipgloss formatting.
	ModeLinear = "linear"

	// ModeHighway is a bubbletea view showing parallel lanes of activity,
	// suited to concurrent worker output. Config loaded on-demand from
	// jay.ui.yml when this mode is selected.
	ModeHighway = "highway"

	// ModeDefault is the display used when --tui is not specified.
	ModeDefault = ModeLinear
)

// ---------------------------------------------------------------------------
// View factory functions - on-demand creation
// ---------------------------------------------------------------------------

// newLinearPresenter creates a linear view presenter with the
// given palette. The presenter wraps a prism linear renderer and
// applies the palette's theme settings (colors, icons, styles). Custom
// tree icons from the palette are explicitly applied via WithIcons to
// ensure they override the defaults.
func newLinearPresenter(palette contract.Palette) (report.Presenter, error) {
	renderer, err := flow.New(
		palette,
		os.Stdout,
		flow.WithIcons(palette.TreeIcons),
	)
	if err != nil {
		return nil, err
	}

	return &linear{renderer: renderer}, nil
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// New returns the Presenter for the requested mode, constructed with
// the given palette. Only the selected view is instantiated; other views
// are not created. Returns an error if the mode is unknown or if the
// palette contains unrecognised colour names.
func New(mode string, palette contract.Palette, hCfg HighwayConfig) (report.Presenter, error) {
	if mode == "" {
		mode = ModeDefault
	}

	switch mode {
	case ModeLinear:
		return newLinearPresenter(palette)

	case ModeHighway:
		return newHighwayPresenter(palette, hCfg)

	default:
		return nil, fmt.Errorf(
			"unknown display mode %q (valid modes: %s)",
			mode,
			strings.Join(availableModes(), ", "),
		)
	}
}

// availableModes returns all known mode names, for error messages.
func availableModes() []string {
	return []string{ModeLinear, ModeHighway}
}
