package traffic

import "strings"

var intensityChars = []string{" ", "░", "▒", "▓", "█"}

func newClassicWaveform() func(tick int) string {
	chars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	return func(tick int) string {
		phase := tick % 16
		var out strings.Builder
		for i := 0; i < 7; i++ {
			val := (phase + i) % len(chars)
			out.WriteString(chars[val])
		}
		return out.String()
	}
}

func newParticleDrift() func(tick int) string {
	particles := []string{" ", "·", "∙", "○"}
	return func(tick int) string {
		seed := tick % 20
		var out strings.Builder
		for i := 0; i < 7; i++ {
			idx := (seed + i*3) % len(particles)
			out.WriteString(particles[idx])
		}
		return out.String()
	}
}

func newPulsingRings() func(tick int) string {
	rings := []string{" ", "○", "◌", "◍", "●"}
	return func(tick int) string {
		pos := tick % 10
		var out strings.Builder
		for i := 0; i < 5; i++ {
			idx := (pos + i*2) % len(rings)
			out.WriteString(rings[idx])
		}
		return out.String()
	}
}

func newASCIILandscape() func(tick int) string {
	terrain := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	return func(tick int) string {
		offset := tick % 32
		var out strings.Builder
		for i := 0; i < 8; i++ {
			idx := (offset + i*4) % len(terrain)
			out.WriteString(terrain[idx])
		}
		return out.String()
	}
}

func newMatrixRain() func(tick int) string {
	chars := []string{"╲", "│", "╱"}
	return func(tick int) string {
		pos := tick % 12
		var out strings.Builder
		for i := 0; i < 6; i++ {
			idx := (pos + i*2) % len(chars)
			out.WriteString(chars[idx])
		}
		return out.String()
	}
}

func newGradientFlow() func(tick int) string {
	return func(tick int) string {
		pos := tick % 20
		var out strings.Builder
		for i := 0; i < 8; i++ {
			idx := (pos + i*3) % len(intensityChars)
			out.WriteString(intensityChars[idx])
		}
		return out.String()
	}
}

func newBreathingCircles() func(tick int) string {
	cycle := []string{" ", "·", "○", "●", "◉"}
	return func(tick int) string {
		pos := tick % 12
		var out strings.Builder
		for i := 0; i < 5; i++ {
			idx := (pos + i*2) % len(cycle)
			out.WriteString(cycle[idx])
		}
		return out.String()
	}
}

func newNetworkGraph() func(tick int) string {
	nodes := []string{" ", "o", "○", "●"}
	return func(tick int) string {
		pos := tick % 16
		var out strings.Builder
		for i := 0; i < 7; i++ {
			idx := (pos + i*3) % len(nodes)
			out.WriteString(nodes[idx])
			if i < 6 {
				out.WriteString("─")
			}
		}
		return out.String()
	}
}
