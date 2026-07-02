package contract

import (
	"fmt"
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/snivilised/jaywalk/src/agenor/enums"
)

// ansi16Names maps human-friendly colour names to their ANSI-16 number
// strings. These are the only names accepted in theme files for the
// ansi16 tier. ANSI colour numbers are stable and will never change.
var ansi16Names = map[string]string{
	"black":          "0",
	"red":            "1",
	"green":          "2",
	"yellow":         "3",
	"blue":           "4",
	"magenta":        "5",
	"cyan":           "6",
	"white":          "7",
	"bright-black":   "8",
	"bright-red":     "9",
	"bright-green":   "10",
	"bright-yellow":  "11",
	"bright-blue":    "12",
	"bright-magenta": "13",
	"bright-cyan":    "14",
	"bright-white":   "15",
}

// ResolveANSI16 converts a colour name or raw number string into a
// color.Color suitable for the ANSI-16 tier. In lipgloss v2, Color is
// a function (not a type) that returns color.Color from image/color.
// Names are resolved via ansi16Names. Raw number strings ("0"-"15")
// are passed through. An empty string produces a nil color.Color.
// Returns an error for unrecognised names.
func ResolveANSI16(s string) (color.Color, error) {
	if s == "" {
		return nil, nil
	}

	// Accept colour names.
	if n, ok := ansi16Names[s]; ok {
		return lipgloss.Color(n), nil
	}

	// Accept raw number strings "0"-"15" directly.
	for _, known := range ansi16Names {
		if s == known {
			return lipgloss.Color(s), nil
		}
	}

	return nil, fmt.Errorf(
		"unknown ansi16 colour %q - use a name (e.g. \"cyan\") or number (\"6\")",
		s,
	)
}

// SemanticColour holds colour specifications for all three terminal
// capability tiers. At theme construction time each SemanticColour is
// resolved and passed to lipgloss.Complete() which selects the best
// available tier at render time.
//
// Fields use mapstructure tags so that bedrock can decode theme YAML
// files with kebab-case keys into this struct directly.
type SemanticColour struct {
	// ANSI16 is the colour name or number string for ANSI-16 terminals.
	// Accepts names ("cyan", "bright-red") or numbers ("6", "9").
	// When set, it respects the user's terminal theme colour assignments.
	ANSI16 string `mapstructure:"ansi16" yaml:"ansi16"`

	// ANSI256 is the number string for ANSI-256 terminals ("0"-"255").
	ANSI256 string `mapstructure:"ansi256" yaml:"ansi256"`

	// TrueColor is the hex colour string for TrueColor terminals ("#RRGGBB").
	TrueColor string `mapstructure:"true-color" yaml:"true-color"`
}

// Resolve converts a SemanticColour into three color.Color values
// representing the ANSI-16, ANSI-256, and TrueColor tiers respectively.
// Returns an error if the ansi16 field contains an unrecognised name.
// Callers pass the three returned values to lipgloss.Complete(profile).
func (sc SemanticColour) Resolve() (ansi, ansi256, trueColor color.Color, err error) {
	ansi, err = ResolveANSI16(sc.ANSI16)
	if err != nil {
		return nil, nil, nil, err
	}

	if sc.ANSI256 != "" {
		ansi256 = lipgloss.Color(sc.ANSI256)
	}

	if sc.TrueColor != "" {
		trueColor = lipgloss.Color(sc.TrueColor)
	}

	return ansi, ansi256, trueColor, nil
}

// ---------------------------------------------------------------------------
// Gradients (animation colour overlays)
// ---------------------------------------------------------------------------

// GradientDef defines a colour gradient used by animation rendering.
// At least one of Hi or Lo must be set. When only one is set, the missing
// end is derived by dimming (missing lo) or brightening (missing hi) the
// resolved colour, creating a single-colour fade.
type GradientDef struct {
	// Steps is the number of colour stops in the gradient (including
	// both endpoints). The renderer interpolates this many colours
	// between Hi and Lo. When 0, the renderer picks a default based
	// on the animation frame count.
	Steps int             `mapstructure:"steps,omitempty" yaml:"steps,omitempty"`
	Hi    *SemanticColour `mapstructure:"hi,omitempty" yaml:"hi,omitempty"`
	Lo    *SemanticColour `mapstructure:"lo,omitempty" yaml:"lo,omitempty"`

	// Animate controls whether the gradient sweeps over time (true) or
	// is applied statically (false). When nil, defaults to true.
	Animate *bool `mapstructure:"animate,omitempty" yaml:"animate,omitempty"`

	// Curve controls the interpolation shape between Hi and Lo.
	// Omit to use linear interpolation (default).
	Curve enums.CurveKind `mapstructure:"curve,omitempty" yaml:"curve,omitempty"`

	// Easing controls step distribution along the curve.
	// Omit for uniform distribution (default).
	Easing enums.EasingKind `mapstructure:"easing,omitempty" yaml:"easing,omitempty"`
}

// HighlightsConfig holds gradient definitions and their component bindings.
// Components maps a semantic component name (e.g. "highway-frame") to a
// gradient name defined under Gradients. Multiple components can share the
// same gradient.
type HighlightsConfig struct {
	Gradients  map[string]GradientDef `mapstructure:"gradients,omitempty" yaml:"gradients,omitempty"`
	Components map[string]string      `mapstructure:"components,omitempty" yaml:"components,omitempty"`
}

// ResolvedGradient holds the resolved colour endpoints after theme
// construction. Hi and Lo are the best-available-tier colours from the
// palette's colour profile. They are never both nil.
type ResolvedGradient struct {
	// Steps is the number of colour stops, copied from GradientDef.
	// 0 means the renderer should use its own default.
	Steps int
	Hi    color.Color
	Lo    color.Color

	// Animate controls whether the gradient sweeps over time.
	// True means animate using GradientState; false means static.
	Animate bool

	// Curve and Easing are copied from GradientDef during theme construction.
	Curve  enums.CurveKind
	Easing enums.EasingKind
}

// Gradient component names - used in theme YAML under
// highlights.components to bind a gradient to a rendering component.
const (
	// GradientComponentActivity is the component name for animation frames.
	GradientComponentActivity = "activity-control"

	// GradientComponentPeriscope is the depth bar (formerly squarebar).
	GradientComponentPeriscope = "periscope-control"

	// GradientComponentAction is the action/pipeline/error info area.
	GradientComponentAction = "action-control"

	// GradientComponentNodePath is the node path display area.
	GradientComponentNodePath = "node-path-control"

	// GradientComponentLandingStrip is the execution info at the lane end.
	GradientComponentLandingStrip = "landing-strip-control"

	// GradientComponentBanner is the ANSI shadow banner widget.
	// Themes bind this component to a gradient name to control the
	// colour sweep of the banner's face/shadow characters.
	GradientComponentBanner = "banner-control"
)

// deriveDimmed returns a colour approximately halfway between the given
// colour and black. Used when a gradient defines only one endpoint to
// produce a natural fade. Works on any color.Color by converting to RGBA,
// scaling each channel by 0.5, and returning the clamped result.
// deriveBrighter returns a colour approximately halfway between the given
// colour and white. Used as the inverse of deriveDimmed: when a gradient
// defines only lo (no hi), this produces a natural bright endpoint.
func deriveBrighter(c color.Color) color.Color {
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: clampU8((float64(r>>8) + 255) * 0.5),
		G: clampU8((float64(g>>8) + 255) * 0.5),
		B: clampU8((float64(b>>8) + 255) * 0.5),
		A: uint8(a >> 8), //nolint:gosec // a>>8 is always 0-255
	}
}

// clampU8 clamps a float to the uint8 range [0, 255].
func clampU8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func deriveDimmed(c color.Color) color.Color {
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: clampU8(float64(r>>8) * 0.5),
		G: clampU8(float64(g>>8) * 0.5),
		B: clampU8(float64(b>>8) * 0.5),
		A: uint8(a >> 8), //nolint:gosec // a>>8 is always 0-255
	}
}

// Palette is the full traversal visual vocabulary for prism. Each field
// represents a distinct visual concept encountered during directory
// traversal. Multiple concepts may map to the same ANSI-16 colour -
// that is intentional, since ANSI-16 has only 16 slots and not all
// concepts appear simultaneously in all views.
//
// Fields use mapstructure tags for YAML theme file decoding.
//
// TODO: It feels like this is defined in the wrong place. Configuration
// concerns are in bedrock. Palette here has been decorated with persistence
// tags, that mean this is be used as a data access object. That feels like
// a pollution of responsibilities. If we have a DAO and a model object,
// then these responsibilities should be split and a translation defined
// between the two.
type Palette struct {
	// --- Traversal nodes ---

	// Directory is the colour of directory names during traversal.
	Directory SemanticColour `mapstructure:"directory" yaml:"directory"`

	// File is the colour of file names during traversal.
	File SemanticColour `mapstructure:"file" yaml:"file"`

	// Root is the colour used to highlight the traversal root path.
	Root SemanticColour `mapstructure:"root" yaml:"root"`

	// Branch is the colour of tree branch characters in tree-style output.
	Branch SemanticColour `mapstructure:"branch" yaml:"branch"`

	// TreeIcons holds optional glyph configuration used by tree-style
	// linear renderers such as the linear view.
	TreeIcons TreeIcons `mapstructure:"tree-icons" yaml:"tree-icons"`

	// --- Execution ---

	// Action is the colour of action names shown alongside nodes.
	Action SemanticColour `mapstructure:"action" yaml:"action"`

	// Pipeline is the colour of pipeline names shown alongside nodes.
	Pipeline SemanticColour `mapstructure:"pipeline" yaml:"pipeline"`

	// LandingStrip is the colour of the landing strip content (execution string or output).
	LandingStrip SemanticColour `mapstructure:"landing-strip" yaml:"landing-strip"`

	// Skipped is the colour of nodes whose action was skipped due to
	// a placeholder breach.
	Skipped SemanticColour `mapstructure:"skipped" yaml:"skipped"`

	// --- Status ---

	// Error is the colour of nodes or actions that produced an error.
	Error SemanticColour `mapstructure:"error" yaml:"error"`

	// Muted is the colour of secondary or de-emphasised information.
	Muted SemanticColour `mapstructure:"muted" yaml:"muted"`

	// Progress is the colour of progress indicators.
	Progress SemanticColour `mapstructure:"progress" yaml:"progress"`

	// --- Summary ---

	// BoxBorder is the colour of the closing summary container border.
	BoxBorder SemanticColour `mapstructure:"box-border" yaml:"box-border"`

	// SummaryLabel is the colour of labels in the closing summary.
	SummaryLabel SemanticColour `mapstructure:"summary-label" yaml:"summary-label"`

	// SummaryValue is the colour of values in the closing summary.
	SummaryValue SemanticColour `mapstructure:"summary-value" yaml:"summary-value"`

	// --- Concurrent views (porthole, lanes) ---

	// Worker is the colour representing an active concurrent worker.
	Worker SemanticColour `mapstructure:"worker" yaml:"worker"`

	// WorkerIdle is the colour representing an idle concurrent worker.
	WorkerIdle SemanticColour `mapstructure:"worker-idle" yaml:"worker-idle"`

	// LaneHeader is the colour of the per-worker lane identity header.
	LaneHeader SemanticColour `mapstructure:"lane-header" yaml:"lane-header"`

	// --- Highway view ---

	// Header is the colour of the title text in the highway view.
	Header SemanticColour `mapstructure:"header" yaml:"header"`

	// Frame is the colour of animation frames (spinners) in the highway view.
	Frame SemanticColour `mapstructure:"frame" yaml:"frame"`

	// Border is the colour of box-drawing characters in the highway view.
	Border SemanticColour `mapstructure:"border" yaml:"border"`

	// BarFilled is the colour of filled square-bar glyphs in the highway view.
	BarFilled SemanticColour `mapstructure:"bar-filled" yaml:"bar-filled"`

	// BarEmpty is the colour of empty square-bar glyphs in the highway view.
	BarEmpty SemanticColour `mapstructure:"bar-empty" yaml:"bar-empty"`

	// Highlights holds named gradients and their component bindings for
	// animation colour overlays. See HighlightsConfig, GradientDef.
	Highlights HighlightsConfig `mapstructure:"highlights,omitempty" yaml:"highlights,omitempty"`
}

// SystemPalette returns the default ANSI-16-only palette. All TrueColor
// and ANSI-256 fields are empty - only the ANSI-16 tier is set, using
// semantic colour names. This palette respects whatever terminal theme
// the user has configured and requires no configuration from the caller.
func SystemPalette() Palette {
	return Palette{
		Directory:    SemanticColour{ANSI16: "cyan"},
		File:         SemanticColour{ANSI16: "white"},
		Root:         SemanticColour{ANSI16: "bright-white"},
		Branch:       SemanticColour{ANSI16: "bright-black"},
		Action:       SemanticColour{ANSI16: "blue"},
		Pipeline:     SemanticColour{ANSI16: "blue"},
		LandingStrip: SemanticColour{ANSI16: "yellow"},
		Skipped:      SemanticColour{ANSI16: "bright-black"},
		Error:        SemanticColour{ANSI16: "red"},
		Muted:        SemanticColour{ANSI16: "bright-black"},
		Progress:     SemanticColour{ANSI16: "green"},
		BoxBorder:    SemanticColour{ANSI16: "magenta"},
		SummaryLabel: SemanticColour{ANSI16: "blue"},
		SummaryValue: SemanticColour{ANSI16: "white"},
		Worker:       SemanticColour{ANSI16: "cyan"},
		WorkerIdle:   SemanticColour{ANSI16: "bright-black"},
		LaneHeader:   SemanticColour{ANSI16: "magenta"},
		Header:       SemanticColour{ANSI16: "bright-white"},
		Frame:        SemanticColour{ANSI16: "blue"},
		Border:       SemanticColour{ANSI16: "bright-black"},
		BarFilled:    SemanticColour{ANSI16: "blue"},
		BarEmpty:     SemanticColour{ANSI16: "bright-black"},
		TreeIcons: TreeIcons{
			TreeIconRoot:           "✻",
			TreeIconDirectory:      "📁",
			TreeIconFile:           "🔖",
			TreeIconElapsed:        "⏰",
			TreeIconSkipped:        "⛔️",
			TreeIconError:          "🚫",
			TreeIconBranchVertical: "│",
			TreeIconBranchJoint:    "├── ",
			TreeIconBranchLast:     "└── ",
			TreeIconBranchIndent:   "  ",
		},
	}
}
