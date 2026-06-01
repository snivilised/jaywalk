package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/snivilised/jaywalk/src/app/bedrock"
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
// Flags row placement
// ---------------------------------------------------------------------------

const (
	// FlagsRowPositionTop places the flags row immediately after the top
	// border and before the first highway lane.
	FlagsRowPositionTop = "top"

	// FlagsRowPositionBottom places the flags row immediately above the
	// status line and below the last highway lane (the default).
	FlagsRowPositionBottom = "bottom"

	// flagsRowPositionDefault is the value used when the configured value
	// is empty or unrecognised.
	flagsRowPositionDefault = FlagsRowPositionBottom
)

// ---------------------------------------------------------------------------
// Polymorphic view configuration
// ---------------------------------------------------------------------------
//
// Every view owns its own concrete config type. Callers outside this
// package see only the sealed ViewConfig interface, so a future view
// can be added without changing the signature of New — the caller
// passes whatever config the selected view's loader produced.
//
// The set of implementations is closed: the unexported isViewConfig
// method cannot be implemented by code outside this package, so the
// compiler enforces that LoadConfig and New only see values produced
// by this package.

// ViewConfig is the polymorphic configuration for a view. Concrete
// implementations are LinearConfig and HighwayConfig (and any future
// view's own type). Callers obtain a value by calling LoadConfig and
// pass it unchanged to New.
type ViewConfig interface {
	isViewConfig()
}

// LinearConfig is the configuration for the linear view. The view
// currently has no per-instance settings beyond the palette, so this
// is a zero-sized placeholder that exists for symmetry with views
// that do have settings.
type LinearConfig struct{}

// isViewConfig seals the implementation set.
func (LinearConfig) isViewConfig() {}

// HighwayConfig is the configuration for the highway view. Every
// field is honoured by the highway presenter; the comment on each
// field describes the source. AnimationGradient is computed from
// the palette inside LoadConfig, not exposed as a flag.
type HighwayConfig struct {
	// WorkerPool is a space-separated list of emoji runes for worker/lane decoration.
	WorkerPool string

	// JobPool is a space-separated list of emoji runes for job decoration.
	JobPool string

	// Separator between emoji and content info (default: " ").
	Separator string

	// SpinnerNames lists the spinner types to use for each lane, looked up
	// Categories expand via movies.SpinnerCategories; individual spinner names
	// are registered in movies.SpinnerNames.
	// When empty, buildHighwayLanes falls back to defaults.
	SpinnerNames []string

	// Labels for each lane, paired with SpinnerNames. When empty, defaults
	// are used.
	Labels []string

	// Overrides maps spinner name to interval in milliseconds.
	// When set, the lane's animation advances at this rate instead of the
	// global tick rate. Multiple lanes sharing the same spinner name each
	// get the same interval behaviour.
	Overrides map[string]int

	// AnimationGradient is the name of a gradient defined in the
	// palette, applied to frame animations. Computed from the palette
	// by LoadConfig; carried on the value so the presenter does not
	// have to look it up again.
	AnimationGradient string

	// FlagsRowPosition is the placement of the supplementary flags row
	// within the highway view. Allowed values are "top" and "bottom";
	// an unrecognised value is normalised to "bottom" at load time
	// (see loadHighwayConfig).
	FlagsRowPosition string
}

// isViewConfig seals the implementation set.
func (HighwayConfig) isViewConfig() {}

// ViewConfigSource loads a named view's raw on-disk configuration
// (typically jay.ui.yml). *bedrock.ViewConfigLoader satisfies this
// implicitly; declaring the interface here keeps the ui package
// decoupled from bedrock's concrete loader type.
type ViewConfigSource interface {
	Load(viewName string, dest any) error
}

// LoadConfig returns the ViewConfig that corresponds to the given
// mode. The source provides the raw, on-disk representation; this
// function translates it into the view's own config type. The
// palette is used to derive per-palette values such as the highway
// animation gradient.
//
// Adding a new view means: (1) define a new ViewConfig
// implementation in this package, (2) add a case here that maps the
// mode to a constructor, and (3) add a case in New that constructs
// the presenter. The view's caller (typically Bootstrap) needs no
// changes — it always calls LoadConfig and New with the mode and
// the returned value.
func LoadConfig(mode string, source ViewConfigSource, palette contract.Palette) (ViewConfig, error) {
	if mode == "" {
		mode = ModeDefault
	}

	switch mode {
	case ModeLinear:
		return LinearConfig{}, nil

	case ModeHighway:
		return loadHighwayConfig(source, palette)

	default:
		return nil, fmt.Errorf(
			"unknown display mode %q (valid modes: %s)",
			mode,
			strings.Join(availableModes(), ", "),
		)
	}
}

func loadHighwayConfig(source ViewConfigSource, palette contract.Palette) (ViewConfig, error) {
	var raw bedrock.HighwayConfig
	if source != nil {
		if err := source.Load("highway", &raw); err != nil {
			return nil, err
		}
	}

	overrides := make(map[string]int)
	for name, cfg := range raw.AnimationData.Spinners.Override {
		if cfg != nil && cfg.Interval > 0 {
			overrides[name] = cfg.Interval
		}
	}

	flagsRowPosition := normaliseFlagsRowPosition(raw.FlagsRowPosition)

	return HighwayConfig{
		WorkerPool:        raw.WorkerPool,
		JobPool:           raw.JobPool,
		Separator:         raw.Separator,
		SpinnerNames:      raw.AnimationData.Spinners.Enabled,
		AnimationGradient: nameFromPalette(palette, contract.GradientComponentActivity),
		Overrides:         overrides,
		FlagsRowPosition:  flagsRowPosition,
	}, nil
}

// normaliseFlagsRowPosition validates the configured flags row position.
// Empty values are treated as unset (use default). Unrecognised values
// also resolve to the default; a warning is written to stderr so the
// user is informed but the application is not aborted.
func normaliseFlagsRowPosition(raw string) string {
	if raw == "" {
		return flagsRowPositionDefault
	}

	switch raw {
	case FlagsRowPositionTop, FlagsRowPositionBottom:
		return raw
	default:
		fmt.Fprintf(os.Stderr,
			"warning: ui.highway.flags-row-position: unrecognised value %q, defaulting to %q\n",
			raw, flagsRowPositionDefault,
		)
		return flagsRowPositionDefault
	}
}

// nameFromPalette looks up the gradient name for a named component
// inside the palette. Mirrors bedrock.ThemeLoader.NameFromPalette
// so the ui package does not have to depend on the theme loader.
func nameFromPalette(palette contract.Palette, componentName string) string {
	components := palette.Highlights.Components
	if len(components) == 0 {
		return ""
	}
	name, ok := components[componentName]
	if !ok {
		return ""
	}
	return name
}

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

// New returns the Presenter for the requested mode, constructed
// with the given palette and the polymorphic view config returned
// by LoadConfig. Only the selected view is instantiated; other views
// are not created.
//
// cfg must be the ViewConfig that corresponds to mode. Passing the
// wrong config type for the mode is reported as an error rather
// than a panic, so a future view can be wired in without the
// compiler enforcing the match (the type-assertion is intentionally
// fail-safe).
func New(mode string, palette contract.Palette, cfg ViewConfig) (report.Presenter, error) {
	if mode == "" {
		mode = ModeDefault
	}

	switch mode {
	case ModeLinear:
		return newLinearPresenter(palette)

	case ModeHighway:
		hCfg, ok := cfg.(HighwayConfig)
		if !ok {
			return nil, fmt.Errorf(
				"ui.New: highway mode requires HighwayConfig, got %T", cfg,
			)
		}
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
