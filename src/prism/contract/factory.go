package contract

import (
	"fmt"
	"io"
)

type rendererFactory func(Palette, io.Writer) (Renderer, error)

var factories = map[ViewKind]rendererFactory{}

// RegisterFactory allows view-specific packages to register their constructor
// without prism importing those packages directly.
func RegisterFactory(kind ViewKind, factory func(Palette, io.Writer) (Renderer, error)) {
	factories[kind] = factory
}

// New constructs a Renderer for the requested view kind using the given
// Palette. The Writer writer is the output destination; pass os.Stdout for
// production use or a bytes.Buffer in tests.
//
// Returns an error if the Palette contains unrecognised colour names,
// which would indicate a malformed user theme file. Bootstrap should
// treat this as a startup failure.
func New(kind ViewKind, palette Palette, writer io.Writer) (Renderer, error) {
	factory := factories[kind]

	switch {
	case factory != nil:
		renderer, err := factory(palette, writer)
		if err != nil {
			return nil, fmt.Errorf("contract.New: %w", err)
		}
		return renderer, nil
	default:
		fallback := factories[LinearView]
		if fallback == nil {
			return nil, fmt.Errorf("contract.New: no renderer factory registered for view %q", kind)
		}
		renderer, err := fallback(palette, writer)
		if err != nil {
			return nil, fmt.Errorf("contract.New: %w", err)
		}
		return renderer, nil
	}
}
