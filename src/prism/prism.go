package prism

import (
	"io"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// Dependency rule for prism view packages:
// - root prism re-exports types from contract for backward compatibility;
// - view-specific implementation/options live in dedicated sub-packages
//   (for example prism/flow);
// - root prism imports contract; child packages import contract.
//   This ensures parent packages depend on children, not the other way around.

// Type aliases for backward compatibility.
type (
	ViewKind         = contract.ViewKind
	NavigationKind   = contract.NavigationKind
	SurveyResult     = contract.SurveyResult
	Overture         = contract.Overture
	Motif            = contract.Motif
	Summary          = contract.Summary
	Renderer         = contract.Renderer
	TreeIcons        = contract.TreeIcons
	SemanticColour   = contract.SemanticColour
	Palette          = contract.Palette
	Theme            = contract.Theme
	GradientDef      = contract.GradientDef
	HighlightsConfig = contract.HighlightsConfig
	ResolvedGradient = contract.ResolvedGradient
)

// Re-exported constants.
const (
	LinearView   = contract.LinearView
	PortholeView = contract.PortholeView
	LanesView    = contract.LanesView
)

const (
	PrimeNavigation   = contract.PrimeNavigation
	ResumeNavigation  = contract.ResumeNavigation
)

// Re-exported variables and functions.
var (
	InterpolateBetween      = contract.InterpolateBetween
	InterpolateBetweenRGBA  = contract.InterpolateBetweenRGBA
	DefaultStepCount        = contract.DefaultStepCount
	ResolveANSI16           = contract.ResolveANSI16
	SystemPalette           = contract.SystemPalette
	NewTheme                = contract.NewTheme
	RegisterFactory         = contract.RegisterFactory
)

// Tree icon name constants.
const (
	TreeIconRoot           = contract.TreeIconRoot
	TreeIconDirectory      = contract.TreeIconDirectory
	TreeIconFile           = contract.TreeIconFile
	TreeIconElapsed        = contract.TreeIconElapsed
	TreeIconSkipped        = contract.TreeIconSkipped
	TreeIconError          = contract.TreeIconError
	TreeIconBranchVertical = contract.TreeIconBranchVertical
	TreeIconBranchJoint    = contract.TreeIconBranchJoint
	TreeIconBranchLast     = contract.TreeIconBranchLast
	TreeIconBranchIndent   = contract.TreeIconBranchIndent
)

const (
	GradientComponentActivity         = contract.GradientComponentActivity
	GradientComponentPeriscope        = contract.GradientComponentPeriscope
	GradientComponentAction           = contract.GradientComponentAction
	GradientComponentNodePath         = contract.GradientComponentNodePath
	GradientComponentLandingStrip     = contract.GradientComponentLandingStrip

)

// New constructs a Renderer for the requested view kind. Delegates to
// contract.New for the implementation.
func New(kind ViewKind, palette Palette, writer io.Writer) (Renderer, error) {
	return contract.New(kind, palette, writer)
}
