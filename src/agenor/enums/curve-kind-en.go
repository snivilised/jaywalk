package enums

import "fmt"

//go:generate stringer -type=CurveKind -linecomment -trimprefix=CurveKind -output curve-kind-en-auto.go

// CurveKind controls the shape of the interpolation path between the
// Hi and Lo gradient endpoints. The zero value is treated as CurveKindLinear.
type CurveKind uint

const (
	// CurveKindLinear interpolates each channel at a constant rate.
	// This is the default and matches pre-issue behaviour exactly.
	CurveKindLinear CurveKind = iota // linear

	// CurveKindSine applies a sine-wave shape to the interpolation.
	CurveKindSine // sine

	// CurveKindQuadraticIn accelerates toward the Lo endpoint.
	CurveKindQuadraticIn // quadratic-in

	// CurveKindQuadraticOut decelerates toward the Lo endpoint.
	CurveKindQuadraticOut // quadratic-out

	// CurveKindCubic applies a cubic (S-curve) shape.
	CurveKindCubic // cubic
)

// UnmarshalText implements encoding.TextUnmarshaler so that CurveKind
// can be decoded from YAML/mapstructure string values.
func (k *CurveKind) UnmarshalText(data []byte) error {
	switch string(data) {
	case "linear":
		*k = CurveKindLinear
	case "sine":
		*k = CurveKindSine
	case "quadratic-in":
		*k = CurveKindQuadraticIn
	case "quadratic-out":
		*k = CurveKindQuadraticOut
	case "cubic":
		*k = CurveKindCubic
	default:
		return fmt.Errorf("unknown curve kind %q: valid values are linear, sine, quadratic-in, quadratic-out, cubic", string(data))
	}
	return nil
}
