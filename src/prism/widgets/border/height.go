package border

// Height is the number of terminal rows a single border line
// occupies. RenderTop and RenderBottom both render to a single
// terminal row, so this is a constant. Host views (highway,
// porthole) consult this to budget vertical space for the bordered
// region without re-rendering.
const Height = 1
