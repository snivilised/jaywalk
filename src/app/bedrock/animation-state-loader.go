package bedrock

import (
	"encoding/json"
	"fmt"
	"os"
)

// AnimationState holds loaded animation data from config.
// This is loaded on-demand when Highway view is requested.
type AnimationState struct {
	FilmStrip   *FilmStripData   `json:"film-strip,omitempty"`
	SpaceFilled *SpaceFilledData `json:"space-filled,omitempty"`
	Spinner     *SpinnerData     `json:"spinner,omitempty"`
}

// FilmStripData holds loaded film strip animation data from config.
type FilmStripData struct {
	Speed     int `json:"speed"`     // ms per frame
	Amplitude int `json:"amplitude"` // intensity
}

// SpaceFilledData holds loaded space-filled bar animation data from config.
type SpaceFilledData struct {
	GradientSteps int `json:"gradient-steps"`
}

// SpinnerData holds loaded spinner animation data from config.
type SpinnerData struct {
	RotationSpeed float64 `json:"rotation-speed"`
}

// LoadAnimationData decodes animation data from the Highway configuration.
// This is called ONLY when Highway view is requested.
func LoadAnimationData(config *HighwayConfig) (*AnimationState, error) {
	state := &AnimationState{}

	if len(config.AnimationData.EnabledTypes) == 0 {
		state.FilmStrip = &FilmStripData{Speed: 100, Amplitude: 5}
		state.SpaceFilled = &SpaceFilledData{GradientSteps: 8}
		state.Spinner = &SpinnerData{RotationSpeed: 0.1}
	} else {
		for _, typeName := range config.AnimationData.EnabledTypes {
			switch typeName {
			case "film-strip":
				if config.AnimationData.FilmStrip != nil {
					state.FilmStrip = &FilmStripData{
						Speed:     config.AnimationData.FilmStrip.Speed,
						Amplitude: config.AnimationData.FilmStrip.Amplitude,
					}
				} else {
					state.FilmStrip = &FilmStripData{Speed: 100, Amplitude: 5}
				}
			case "space-filled":
				if config.AnimationData.SpaceFilled != nil {
					state.SpaceFilled = &SpaceFilledData{
						GradientSteps: config.AnimationData.SpaceFilled.GradientSteps,
					}
				} else {
					state.SpaceFilled = &SpaceFilledData{GradientSteps: 8}
				}
			case "spinner":
				if config.AnimationData.Spinner != nil {
					state.Spinner = &SpinnerData{
						RotationSpeed: config.AnimationData.Spinner.RotationSpeed,
					}
				} else {
					state.Spinner = &SpinnerData{RotationSpeed: 0.1}
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
