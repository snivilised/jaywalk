package contract

// Renderer is the rendering abstraction for prism views. All view kinds
// implement this interface. Callers depend on Renderer, never on a
// concrete view type.
type Renderer interface {
	// Begin is called once before any traversal events.
	Begin(overture Overture)

	// Show is called for each item encountered during traversal.
	Show(motif Motif)

	// End is called once when traversal completes.
	End(summary Summary)
}
