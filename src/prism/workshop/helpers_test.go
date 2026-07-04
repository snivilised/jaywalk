package workshop_test

import (
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"strings"
)

func makeSteps(n int) []contract.Color {
	steps := make([]contract.Color, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(max(n-1, 1))
		steps[i] = contract.Color{
			R: uint8(255 * (1 - t)),
			G: uint8(0),
			B: uint8(255 * t),
		}
	}
	return steps
}

type stepsEntry struct {
	Count     int
	Curve     enums.CurveKind
	Easing    enums.EasingKind
	AnimFrame int
}

func containsAny(s string, chars []string) bool {
	for _, ch := range chars {
		if strings.Contains(s, ch) {
			return true
		}
	}
	return false
}
