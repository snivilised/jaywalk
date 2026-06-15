package bedrock

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/snivilised/jaywalk/src/prism/movies"
	"github.com/snivilised/jaywalk/src/third/lo"
)

// AnimationState holds loaded animation data from config.
type AnimationState struct {
	Data map[string]*SpinnerItemConfig `json:"data,omitempty"`
}

// LoadAnimationData decodes animation data from the Highway configuration.
func LoadAnimationData(config *HighwayConfig) (*AnimationState, error) {
	state := &AnimationState{Data: make(map[string]*SpinnerItemConfig)}
	spinners := &config.AnimationData.Spinners

	if len(spinners.Enabled) == 0 {
		for _, name := range lo.Keys(DefaultSpinnerIntervals) {
			state.Data[name] = &SpinnerItemConfig{Interval: DefaultIntervalFor(name)}
		}
	} else {
		for _, name := range movies.ExpandNames(spinners.Enabled) {
			if override, ok := spinners.Override[name]; ok && override != nil {
				state.Data[name] = &SpinnerItemConfig{
					Interval: override.Interval,
				}
			} else {
				state.Data[name] = &SpinnerItemConfig{Interval: DefaultIntervalFor(name)}
			}
		}
	}

	return state, nil
}

// LoadAnimationStateFrom loads saved animation data from file.
func LoadAnimationStateFrom(path string) (*AnimationState, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(path) //nolint:gosec // ok
	if err != nil {
		return nil, fmt.Errorf("reading Highway animation data: %w", err)
	}

	var state AnimationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("decoding Highway animation data: %w", err)
	}

	return &state, nil
}

var DefaultSpinnerIntervals = map[string]int{
	"amour":             100,
	"ascii-landscape":   100,
	"barcode":           100,
	"bounce":            100,
	"braille":           80,
	"braille-pulse":     180,
	"braille-wave":      100,
	"breathe":           100,
	"breathing-circles": 150,
	"cascade":           60,
	"checkerboard":      250,
	"classic-waveform":  100,
	"columns":           60,
	"diagswipe":         60,
	"dna":               80,
	"dot":               100,
	"ellipsis":          333,
	"fairlight":         100,
	"fillsweep":         100,
	"globe":             250,
	"gradient-flow":     90,
	"hamburger":         333,
	"heart-throb":       100,
	"helix":             80,
	"infantry":          100,
	"jamboree":          100,
	"jump":              100,
	"matrix-rain":       70,
	"meter":             143,
	"monkey":            333,
	"morse":             100,
	"moon":              125,
	"musical":           100,
	"network-graph":     100,
	"orbit":             100,
	"particle-drift":    80,
	"points":            143,
	"pulse":             125,
	"pulsing-rings":     120,
	"rain":              100,
	"scan":              70,
	"scanline":          120,
	"snake":             80,
	"sparkle":           150,
	"spinner":           100,
	"starlight":         100,
	"trinkets":          100,
	"wave":              100,
	"waverows":          90,
}

// DefaultIntervalFor returns the default interval in ms for a spinner type.
func DefaultIntervalFor(name string) int {
	if v, ok := DefaultSpinnerIntervals[name]; ok {
		return v
	}
	return 100
}
