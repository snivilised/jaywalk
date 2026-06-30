package contract

// ansi16Canonical maps ANSI16 colour names to their canonical
// true-color hex equivalents (xterm defaults).
var ansi16Canonical = map[string]string{
	"black":          "#000000",
	"red":            "#CC0000",
	"green":          "#4E9A06",
	"yellow":         "#C4A000",
	"blue":           "#3465A4",
	"magenta":        "#75507B",
	"cyan":           "#06989A",
	"white":          "#D3D7CF",
	"bright-black":   "#555753",
	"bright-red":     "#EF2929",
	"bright-green":   "#8AE234",
	"bright-yellow":  "#FCE94F",
	"bright-blue":    "#729FCF",
	"bright-magenta": "#AD7FA8",
	"bright-cyan":    "#34E2E2",
	"bright-white":   "#EEEEEC",
}

// Ansi16ToHex returns the canonical hex string for the given ANSI16
// colour name, and true if found. Returns "", false for unknown names.
func Ansi16ToHex(name string) (string, bool) {
	hex, ok := ansi16Canonical[name]
	return hex, ok
}
