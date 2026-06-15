package contract

// Position values for the "place above or below the panel" concept
// shared by the highway view's flags row and the banner. Used
// universally by prism/widgets/*, prism/highway and app/ui so that
// the YAML config vocabulary, the model state and the widget-level
// constants cannot drift apart.
//
// Position is a strongly-typed concept; field types in messages and
// configs remain string. The constants below are the only place
// these literal values are defined.
const (
	// PositionTop places the element above the panel.
	PositionTop = "top"

	// PositionBottom places the element below the panel.
	PositionBottom = "bottom"
)
