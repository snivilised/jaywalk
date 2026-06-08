// package flow implements the linear view renderer and presenter.
// The linear view is a simple, vertically scrolling output format with one styled
// line per node. It is the default view for jaywalk and is designed for
// maximum compatibility with a wide range of terminal environments.
//
// The linear renderer is implemented in this package and registered as the
// factory for prism.LinearView. The linear presenter is implemented in the
// app/ui package and wraps the linear renderer to translate report events
// into prism.Motif calls.
//
// The linear view supports custom tree icons via the palette's TreeIcons map.
// These icons are applied to the linear renderer via the WithIcons option at
// construction time. Custom icons override the defaults provided by prism.
package linear
