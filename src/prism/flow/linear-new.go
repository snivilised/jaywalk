package flow

import (
	"fmt"
	"io"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// New constructs a linear renderer for linear-style output.
func New(palette contract.Palette, writer io.Writer, opts ...LinearOption) (contract.Renderer, error) {
	theme, err := contract.NewTheme(palette, writer)
	if err != nil {
		return nil, fmt.Errorf("flow.New: %w", err)
	}

	r := &renderer{
		theme:     theme,
		writer:    writer,
		treeIcons: theme.TreeIcons,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}
