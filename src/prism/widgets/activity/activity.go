package activity

import (
	"charm.land/lipgloss/v2"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
)

type Styles struct {
	FrameStyle lipgloss.Style
}

type Config struct {
	Content string
}

type Effect struct {
	Gradient *contract.ResolvedGradient
	State    *effects.GradientState
}

func Render(cfg Config,
	styles Styles,
	effect Effect,
) string {
	if cfg.Content == "" {
		return ""
	}

	if effect.Gradient != nil && effect.State != nil {
		// Strip outer ┃ bars from film-strip and bounce frames so the
		// gradient doesn't sweep through them. The bars are re-added
		// with the gradient's Hi (left) and Lo (right) colours.

		inner, withBars := stripOuterBars(cfg.Content)
		gradientRuns := effects.ApplyGradient(
			*effect.Gradient,
			inner,
			effect.State,
		)

		if gradientRuns != nil {
			styledFrame := effects.ApplyGradientStyled(gradientRuns)
			if withBars {
				leftBarStyle := lipgloss.NewStyle().Foreground(effect.Gradient.Hi)
				rightBarStyle := lipgloss.NewStyle().Foreground(effect.Gradient.Lo)
				return leftBarStyle.Render("┃") +
					styledFrame +
					rightBarStyle.Render("┃")
			}
			return styles.FrameStyle.Render(styledFrame)
		}
	}

	return styles.FrameStyle.Render(cfg.Content)
}

// stripOuterBars checks if content is wrapped in ┃...┃ and returns the
// inner portion. Returns hasBars=false when no outer bars are detected.
func stripOuterBars(content string) (inner string, hasBars bool) {
	runes := []rune(content)
	if len(runes) >= 2 && runes[0] == '┃' && runes[len(runes)-1] == '┃' {
		return string(runes[1 : len(runes)-1]), true
	}
	return content, false
}
