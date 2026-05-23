package bedrock

import (
	"encoding/json"
	"fmt"
	"os"
)

// AnimationState holds loaded animation data from config.
// This is loaded on-demand when Highway view is requested.
type AnimationState struct {
	FilmStrip   *SpinnerItemConfig `json:"film-strip,omitempty"`
	SpaceFilled *SpinnerItemConfig `json:"space-filled,omitempty"`
	Spinner     *SpinnerItemConfig `json:"spinner,omitempty"`
}

// LoadAnimationData decodes animation data from the Highway configuration.
// This is called ONLY when Highway view is requested.
func LoadAnimationData(config *HighwayConfig) (*AnimationState, error) {
	state := &AnimationState{}
	spinners := &config.AnimationData.Spinners

	if len(spinners.Enabled) == 0 {
		state.FilmStrip = &SpinnerItemConfig{Interval: 100}
		state.SpaceFilled = &SpinnerItemConfig{Interval: 80}
		state.Spinner = &SpinnerItemConfig{Interval: 100}
	} else {
		for _, typeName := range spinners.Enabled {
			switch typeName {
			case "film-strip":
				if spinners.FilmStrip != nil {
					state.FilmStrip = &SpinnerItemConfig{
						Interval: spinners.FilmStrip.Interval,
					}
				} else {
					state.FilmStrip = &SpinnerItemConfig{Interval: 100}
				}
			case "space-filled":
				if spinners.SpaceFilled != nil {
					state.SpaceFilled = &SpinnerItemConfig{
						Interval: spinners.SpaceFilled.Interval,
					}
				} else {
					state.SpaceFilled = &SpinnerItemConfig{Interval: 80}
				}
			case "spinner":
				if spinners.Spinner != nil {
					state.Spinner = &SpinnerItemConfig{
						Interval: spinners.Spinner.Interval,
					}
				} else {
					state.Spinner = &SpinnerItemConfig{Interval: 100}
				}
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
