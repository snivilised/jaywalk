package contract

// Colour is a simple RGBA colour type for in-package gradient interpolation.
// This mirrors image/color.RGBA but without alpha support (alpha always 255).
type Colour struct {
	R, G, B uint8
}
