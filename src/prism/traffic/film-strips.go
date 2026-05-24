package traffic

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

const filmStripWindow = 7

var (
	waveStrip       = []rune("⋅.˳˳.⋅ॱ˙˙ॱ⋅.˳˳.⋅ॱ˙˙ॱᐧ.˳˳.⋅")
	fairlightStrip  = []rune("꧁∙·▫ₒₒ▫ᵒᴼᵒ▫ₒₒ▫꧁ fairlight ꧂▫ₒₒ▫ᵒᴼᵒ▫ₒₒ▫·∙꧂")
	amourStrip      = []rune("♥¸.•**◦༄◦°˚°◦.¸¸彡彡")
	jamboreeStrip   = []rune("✯¸.•´*¨`*•✿ ✿•*`¨*`•.¸✯")
	musicalStrip    = []rune("♪°•°∞°•°♪°•°∞°•°♪°•°∞°•°♪°•°∞°•°♪")
	trinketsStrip   = []rune("⭑*•̩̩͙⊱••••✩••••̩̩͙⊰•*⭑")
	morseStrip      = []rune("━◦○◦━◦○◦━◦○◦━◦○◦━◦○◦━◦○◦━")
	starlightStrip  = []rune("───✱*.｡:｡✱*.:｡✧*.｡✰*.:｡✧*.｡:｡*.｡✱ ───")
	infantryStrip   = []rune("°l||l°l||l°l||l°l||l°l||l°l||l°l||l°l||l°l||l°l||l°")
	heartThrobStrip = []rune("︵‿︵‿୨♡୧‿︵‿︵")
	barcodeStrip    = []rune(" █║▌│ █│║▌ ║││█║▌ │║║█║ │║║█║")
)

const (
	CategoryFilmStrip = "film-strip-set"
	CategoryBraille   = "braille-set"
	CategoryCharm     = "charm-set"
	CategorySkins     = "skin-set"
	CategoryBasics    = "basics-set"
)

var SpinnerCategories = map[string][]string{
	CategoryFilmStrip: {
		SpinnerTypeWave,
		SpinnerTypeFairlight,
		SpinnerTypeAmour,
		SpinnerTypeJamboree,
		SpinnerTypeMusical,
		SpinnerTypeTrinkets,
		SpinnerTypeMorse,
		SpinnerTypeStarlight,
		SpinnerTypeInfantry,
		SpinnerTypeHeartThrob,
		SpinnerTypeBarcode,
	},
	CategoryBraille: {
		SpinnerTypeBraille,
		SpinnerTypeBrailleWave,
		SpinnerTypeDNA,
		SpinnerTypeScan,
		SpinnerTypeRain,
		SpinnerTypeScanLine,
		SpinnerTypeBraillePulse,
		SpinnerTypeSnake,
		SpinnerTypeSparkle,
		SpinnerTypeCascade,
		SpinnerTypeColumns,
		SpinnerTypeOrbit,
		SpinnerTypeBreathe,
		SpinnerTypeWaveRows,
		SpinnerTypeCheckerboard,
		SpinnerTypeHelix,
		SpinnerTypeFillSweep,
		SpinnerTypeDiagSwipe,
	},
	CategoryCharm: {
		SpinnerTypeDefault,
		SpinnerTypeDot,
		SpinnerTypeJump,
		SpinnerTypePulse,
		SpinnerTypePoints,
		SpinnerTypeGlobe,
		SpinnerTypeMoon,
		SpinnerTypeMonkey,
		SpinnerTypeMeter,
		SpinnerTypeHamburger,
		SpinnerTypeEllipsis,
	},
	CategorySkins: {
		SpinnerTypeClassicWaveform,
		SpinnerTypeParticleDrift,
		SpinnerTypePulsingRings,
		SpinnerTypeASCIILandscape,
		SpinnerTypeMatrixRain,
		SpinnerTypeGradientFlow,
		SpinnerTypeBreathingCircles,
		SpinnerTypeNetworkGraph,
	},
	CategoryBasics: {
		SpinnerTypeDefault,
		SpinnerTypeBounce,
	},
}

func ExpandNames(names []string) []string {
	var expanded []string
	for _, name := range names {
		if members, ok := SpinnerCategories[name]; ok {
			expanded = append(expanded, members...)
		} else {
			expanded = append(expanded, name)
		}
	}
	return expanded
}

func newFilmStripFrame(strip []rune) func(tick int) string {
	return func(tick int) string {
		start := tick % len(strip)
		var seg []rune
		for i, w := 0, 0; i < len(strip) && w < filmStripWindow; i++ {
			r := strip[(start+i)%len(strip)]
			rw := runewidth.RuneWidth(r)
			if w+rw > filmStripWindow {
				break
			}
			seg = append(seg, r)
			w += rw
		}
		segStr := string(seg)
		if pad := filmStripWindow - runewidth.StringWidth(segStr); pad > 0 {
			segStr += strings.Repeat(" ", pad)
		}
		return "┃" + segStr + "┃"
	}
}

func bounceFrame(tick int) string {
	const width = 9
	total := 2 * (width - 2)
	pos := tick % total
	if pos >= width-2 {
		pos = total - pos - 1
	}
	filled := strings.Repeat("█", pos+1)
	empty := strings.Repeat("░", width-2-(pos+1))
	return "┃" + filled + empty + "┃"
}
