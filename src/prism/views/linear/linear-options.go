package linear

// Option configures linear renderer behavior at construction time.
type Option func(*renderer)

// WithIcons configures tree glyphs and item icons for linear renderers.
// The map keys are the standard token names used by prism tree rendering.
func WithIcons(icons map[string]string) Option {
	return func(r *renderer) {
		if r.treeIcons == nil {
			r.treeIcons = make(map[string]string)
		}

		for key, value := range icons {
			if value != "" {
				r.treeIcons[key] = value
			}
		}
	}
}
