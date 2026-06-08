// package highway implements the highway view renderer and presenter.
// The highway view is highly animated view, depicting the concurrent traversal of
// multiple paths as lanes on a highway. It is designed to be visually engaging and
// to make it easy to track the progress of multiple paths at once.
//
// The highway renderer is implemented in this package and registered as the
// factory for prism.highwayView. The highway presenter is implemented in the
// app/ui package and wraps the highway renderer to translate report events
// into prism.Motif calls.
//
// The highway view supports custom tree icons via the palette's TreeIcons map.
// These icons are applied to the highway renderer via the WithIcons option at
// construction time. Custom icons override the defaults provided by prism.
package highway
