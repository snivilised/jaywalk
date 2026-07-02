package periscope

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
)

type Styles struct {
	Filled lipgloss.Style
	Empty  lipgloss.Style
}

type Config struct {
	Width  int
	Fill   int
	Styles Styles
}

type Effect struct {
	Gradient *contract.ResolvedGradient
	State    *effects.GradientState
}

func Render(cfg Config, styles Styles, effect Effect) string {
	fill := min(max(cfg.Fill, 0), cfg.Width)
	width := max(cfg.Width, 0)

	if effect.Gradient != nil {
		barContent := strings.Repeat("◼", fill) + strings.Repeat("◻", width-fill)
		if effect.Gradient.Animate {
			if effect.State != nil && fill > 0 {
				runs := effects.ApplyGradient(
					*effect.Gradient,
					barContent,
					effect.State,
				)
				if runs != nil {
					return effects.ApplyGradientStyled(runs)
				}
			}
		} else {
			runs := effects.ApplyGradientStatic(
				*effect.Gradient,
				barContent,
				effect.Gradient.Steps,
			)
			if runs != nil {
				return effects.ApplyGradientStyled(runs)
			}
		}
	}

	filled := styles.Filled.Render(strings.Repeat("◼", fill))
	empty := styles.Empty.Render(strings.Repeat("◻", width-fill))
	return filled + empty
}
