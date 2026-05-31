package movies

import (
	"strings"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

var intensityChars = []string{" ", "░", "▒", "▓", "█"}

func newClassicWaveform() contract.FrameFunc {
	chars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	return func(tick int) string {
		phase := tick % 16
		var out strings.Builder
		for i := range 7 {
			val := (phase + i) % len(chars)
			out.WriteString(chars[val])
		}
		return out.String()
	}
}

func newParticleDrift() contract.FrameFunc {
	particles := []string{" ", "·", "∙", "○"}
	return func(tick int) string {
		seed := tick % 20
		var out strings.Builder
		for i := range 7 {
			idx := (seed + i*3) % len(particles)
			out.WriteString(particles[idx])
		}
		return out.String()
	}
}

func newPulsingRings() contract.FrameFunc {
	rings := []string{" ", "○", "◌", "◍", "●"}
	return func(tick int) string {
		pos := tick % 10
		var out strings.Builder
		for i := range 5 {
			idx := (pos + i*2) % len(rings)
			out.WriteString(rings[idx])
		}
		return out.String()
	}
}

func newASCIILandscape() contract.FrameFunc {
	terrain := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	return func(tick int) string {
		offset := tick % 32
		var out strings.Builder
		for i := range 8 {
			idx := (offset + i*4) % len(terrain)
			out.WriteString(terrain[idx])
		}
		return out.String()
	}
}

func newMatrixRain() contract.FrameFunc {
	chars := []string{"╲", "│", "╱"}
	return func(tick int) string {
		pos := tick % 12
		var out strings.Builder
		for i := range 6 {
			idx := (pos + i*2) % len(chars)
			out.WriteString(chars[idx])
		}
		return out.String()
	}
}

func newGradientFlow() contract.FrameFunc {
	return func(tick int) string {
		pos := tick % 20
		var out strings.Builder
		for i := range 8 {
			idx := (pos + i*3) % len(intensityChars)
			out.WriteString(intensityChars[idx])
		}
		return out.String()
	}
}

func newBreathingCircles() contract.FrameFunc {
	cycle := []string{" ", "·", "○", "●", "◉"}
	return func(tick int) string {
		pos := tick % 12
		var out strings.Builder
		for i := range 5 {
			idx := (pos + i*2) % len(cycle)
			out.WriteString(cycle[idx])
		}
		return out.String()
	}
}

func newNetworkGraph() contract.FrameFunc {
	nodes := []string{" ", "o", "○", "●"}
	return func(tick int) string {
		pos := tick % 16
		var out strings.Builder
		for i := range 7 {
			idx := (pos + i*3) % len(nodes)
			out.WriteString(nodes[idx])
			if i < 6 {
				out.WriteString("─")
			}
		}
		return out.String()
	}
}
