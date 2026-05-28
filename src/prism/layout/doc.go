// Package layout provides reusable horizontal row-layout primitives
// for terminal UI views. It handles width partitioning between
// a left group (fixed-/content-sized), an optional flexible middle
// segment, and a right group (right-aligned, fixed-/content-sized).
//
// The Row type is a builder that accumulates segments, computes
// the layout, and renders the result. Use it to replace manual
// strings.Repeat padding and lipgloss.Width arithmetic.
//
// Usage:
//
//	row := layout.NewRow(width).
//	    Caps("│ ", " │").
//	    Content(emoji).Gap(2).
//	    Fixed(SpinnerNameWidth, spinner).
//	    Flex(true).Gap(2).
//	    Content(frame).
//	    RightContent(landing)
//
//	pathWidth := row.FlexWidth()
//	row.SetFlexContent(renderPath(pathWidth))
//	row.RenderTo(&b)
package layout
