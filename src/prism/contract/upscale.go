package contract

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// UpscalePalette returns a new Palette in which every SemanticColour
// has all three tier fields populated. Missing fields are derived from
// whichever fields are present using the following rules:
//
//   - true-color present, ansi256 absent: derive via NearestAnsi256
//   - true-color present, ansi16 absent:  derive via NearestAnsi16Name
//   - ansi16 present, true-color absent:  promote via Ansi16ToHex,
//     then derive ansi256
//   - ansi256 present, true-color absent: promote via xterm256 table,
//     then derive ansi16
//   - all three present: no action
//   - none present: no action (role intentionally empty)
//
// The input palette is not mutated. A fully populated copy is returned.
func UpscalePalette(p Palette) Palette {
	out := p

	out.Directory = upscaleColour(p.Directory)
	out.File = upscaleColour(p.File)
	out.Root = upscaleColour(p.Root)
	out.Branch = upscaleColour(p.Branch)
	out.TreeIcons = p.TreeIcons
	out.Action = upscaleColour(p.Action)
	out.Pipeline = upscaleColour(p.Pipeline)
	out.LandingStrip = upscaleColour(p.LandingStrip)
	out.Skipped = upscaleColour(p.Skipped)
	out.Error = upscaleColour(p.Error)
	out.Muted = upscaleColour(p.Muted)
	out.Progress = upscaleColour(p.Progress)
	out.BoxBorder = upscaleColour(p.BoxBorder)
	out.SummaryLabel = upscaleColour(p.SummaryLabel)
	out.SummaryValue = upscaleColour(p.SummaryValue)
	out.Worker = upscaleColour(p.Worker)
	out.WorkerIdle = upscaleColour(p.WorkerIdle)
	out.LaneHeader = upscaleColour(p.LaneHeader)
	out.Header = upscaleColour(p.Header)
	out.Frame = upscaleColour(p.Frame)
	out.Border = upscaleColour(p.Border)
	out.BarFilled = upscaleColour(p.BarFilled)
	out.BarEmpty = upscaleColour(p.BarEmpty)

	out.Highlights.Gradients = make(map[string]GradientDef, len(p.Highlights.Gradients))
	for name, gd := range p.Highlights.Gradients {
		gdCopy := gd
		if gd.Hi != nil {
			hiCopy := upscaleColour(*gd.Hi)
			gdCopy.Hi = &hiCopy
		}
		if gd.Lo != nil {
			loCopy := upscaleColour(*gd.Lo)
			gdCopy.Lo = &loCopy
		}
		out.Highlights.Gradients[name] = gdCopy
	}

	out.Highlights.Components = p.Highlights.Components

	return out
}

// upscaleColour applies the upscaling rules to a single SemanticColour.
func upscaleColour(sc SemanticColour) SemanticColour {
	hasTC := sc.TrueColor != ""
	has256 := sc.ANSI256 != ""
	has16 := sc.ANSI16 != ""

	switch {
	case hasTC && !has256 && !has16:
		// Derive both from true-color.
		tc, ok := parseHex(sc.TrueColor)
		if !ok {
			return sc
		}
		return SemanticColour{
			TrueColor: sc.TrueColor,
			ANSI256:   fmt.Sprintf("%d", NearestAnsi256(tc)),
			ANSI16:    NearestAnsi16Name(tc),
		}

	case hasTC && has256 && !has16:
		// Derive ANSI16 from true-color.
		tc, ok := parseHex(sc.TrueColor)
		if !ok {
			return sc
		}
		return SemanticColour{
			TrueColor: sc.TrueColor,
			ANSI256:   sc.ANSI256,
			ANSI16:    NearestAnsi16Name(tc),
		}

	case hasTC && !has256 && has16:
		// Derive ANSI256 from true-color.
		tc, ok := parseHex(sc.TrueColor)
		if !ok {
			return sc
		}
		return SemanticColour{
			TrueColor: sc.TrueColor,
			ANSI256:   fmt.Sprintf("%d", NearestAnsi256(tc)),
			ANSI16:    sc.ANSI16,
		}

	case !hasTC && has16 && !has256:
		// Promote ANSI16 to true-color, then derive ANSI256.
		hex, ok := Ansi16ToHex(sc.ANSI16)
		if !ok {
			return sc
		}
		tc, _ := parseHex(hex)
		return SemanticColour{
			TrueColor: hex,
			ANSI256:   fmt.Sprintf("%d", NearestAnsi256(tc)),
			ANSI16:    sc.ANSI16,
		}

	case !hasTC && !has16 && has256:
		// Promote ANSI256 to true-color, then derive ANSI16.
		idx, err := strconv.Atoi(sc.ANSI256)
		if err != nil || idx < 0 || idx > 255 {
			return sc
		}
		entry := xterm256Palette[idx]
		hex := fmt.Sprintf("#%02X%02X%02X", entry[0], entry[1], entry[2])
		tc := color.RGBA{R: entry[0], G: entry[1], B: entry[2], A: 255}
		return SemanticColour{
			TrueColor: hex,
			ANSI256:   sc.ANSI256,
			ANSI16:    NearestAnsi16Name(tc),
		}

	case !hasTC && has16 && has256:
		// Promote ANSI16 to true-color.
		hex, ok := Ansi16ToHex(sc.ANSI16)
		if !ok {
			return sc
		}
		return SemanticColour{
			TrueColor: hex,
			ANSI256:   sc.ANSI256,
			ANSI16:    sc.ANSI16,
		}

	case hasTC && has256 && has16:
		// All three present: no action.
		return sc

	default:
		// None present: no action.
		return sc
	}
}

// parseHex converts a "#RRGGBB" hex string to a color.RGBA.
func parseHex(hex string) (color.RGBA, bool) {
	hex = strings.TrimSpace(hex)
	if !strings.HasPrefix(hex, "#") || len(hex) != 7 {
		return color.RGBA{}, false
	}

	r, err := strconv.ParseUint(hex[1:3], 16, 8)
	if err != nil {
		return color.RGBA{}, false
	}
	g, err := strconv.ParseUint(hex[3:5], 16, 8)
	if err != nil {
		return color.RGBA{}, false
	}
	b, err := strconv.ParseUint(hex[5:7], 16, 8)
	if err != nil {
		return color.RGBA{}, false
	}

	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, true
}
