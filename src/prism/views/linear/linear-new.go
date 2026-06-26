package linear

import (
	"fmt"
	"io"

	"github.com/charmbracelet/x/term"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

// New constructs a linear renderer for linear-style output.
func New(palette contract.Palette, writer io.Writer, opts ...Option) (contract.Renderer, error) {
	theme, err := contract.NewTheme(palette, writer)
	if err != nil {
		return nil, fmt.Errorf("linear.New: %w", err)
	}

	width := 104
	if w, _, err := term.GetSize(0); err == nil && w > 0 {
		width = w
	}

	r := &renderer{
		theme:     theme,
		writer:    writer,
		width:     width,
		treeIcons: theme.TreeIcons,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}
