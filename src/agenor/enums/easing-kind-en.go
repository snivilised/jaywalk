package enums

import "fmt"

//go:generate stringer -type=EasingKind -linecomment -trimprefix=EasingKind -output easing-kind-en-auto.go

// EasingKind controls the distribution of steps along the curve.
// The zero value is treated as EasingKindUniform.
type EasingKind uint

const (
	// EasingKindUniform distributes steps evenly. Default; matches
	// pre-issue behaviour exactly.
	EasingKindUniform EasingKind = iota // uniform

	// EasingKindEaseIn clusters steps toward the start of the gradient.
	EasingKindEaseIn // ease-in

	// EasingKindEaseOut clusters steps toward the end of the gradient.
	EasingKindEaseOut // ease-out

	// EasingKindEaseInOut clusters steps toward both ends.
	EasingKindEaseInOut // ease-in-out
)

// UnmarshalText implements encoding.TextUnmarshaler so that EasingKind
// can be decoded from YAML/mapstructure string values.
func (k *EasingKind) UnmarshalText(data []byte) error {
	switch string(data) {
	case "uniform":
		*k = EasingKindUniform
	case "ease-in":
		*k = EasingKindEaseIn
	case "ease-out":
		*k = EasingKindEaseOut
	case "ease-in-out":
		*k = EasingKindEaseInOut
	default:
		return fmt.Errorf("unknown easing kind %q: valid values are uniform, ease-in, ease-out, ease-in-out", string(data))
	}
	return nil
}
