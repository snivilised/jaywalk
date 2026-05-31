package bedrock

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/snivilised/jaywalk/src/prism/movies"
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
		for _, name := range defaultSpinnerList {
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

// DefaultIntervalFor returns the default interval in ms for a spinner type.
func DefaultIntervalFor(name string) int {
	switch name {
	case "spinner":
		return 100
	case "braille":
		return 80
	case "braillewave":
		return 100
	case "dna":
		return 80
	case "scan":
		return 70
	case "rain":
		return 100
	case "scanline":
		return 120
	case "braille-pulse":
		return 180
	case "snake":
		return 80
	case "sparkle":
		return 150
	case "cascade":
		return 60
	case "columns":
		return 60
	case "orbit":
		return 100
	case "breathe":
		return 100
	case "waverows":
		return 90
	case "checkerboard":
		return 250
	case "helix":
		return 80
	case "fillsweep":
		return 100
	case "diagswipe":
		return 60
	case "classic-waveform":
		return 100
	case "particle-drift":
		return 80
	case "pulsing-rings":
		return 120
	case "ascii-landscape":
		return 100
	case "matrix-rain":
		return 70
	case "gradient-flow":
		return 90
	case "breathing-circles":
		return 150
	case "network-graph":
		return 100
	case "bounce":
		return 100
	case "wave":
		return 100
	case "fairlight":
		return 100
	case "amour":
		return 100
	case "jamboree":
		return 100
	case "musical":
		return 100
	case "trinkets":
		return 100
	case "morse":
		return 100
	case "starlight":
		return 100
	case "infantry":
		return 100
	case "heart-throb":
		return 100
	case "barcode":
		return 100
	case "dot":
		return 100
	case "jump":
		return 100
	case "pulse":
		return 125
	case "points":
		return 143
	case "globe":
		return 250
	case "moon":
		return 125
	case "monkey":
		return 333
	case "meter":
		return 143
	case "hamburger":
		return 333
	case "ellipsis":
		return 333
	default:
		return 100
	}
}

var defaultSpinnerList = []string{
	"wave",
	"fairlight",
	"amour",
	"jamboree",
	"musical",
	"trinkets",
	"morse",
	"starlight",
	"infantry",
	"heart-throb",
	"barcode",
	"bounce",
	"spinner",
	"braille",
	"braillewave",
	"dna",
	"scan",
	"rain",
	"scanline",
	"braille-pulse",
	"snake",
	"sparkle",
	"cascade",
	"columns",
	"orbit",
	"breathe",
	"waverows",
	"checkerboard",
	"helix",
	"fillsweep",
	"diagswipe",
	"classic-waveform",
	"particle-drift",
	"pulsing-rings",
	"ascii-landscape",
	"matrix-rain",
	"gradient-flow",
	"breathing-circles",
	"network-graph",
	"dot",
	"jump",
	"pulse",
	"points",
	"globe",
	"moon",
	"monkey",
	"meter",
	"hamburger",
	"ellipsis",
}
