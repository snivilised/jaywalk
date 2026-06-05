package movies

import (
	"sync"

	"charm.land/bubbles/v2/spinner"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

// SpinnerDef holds a spinner's frame generator.
// Interval is NOT stored here - per-lane animation speed is controlled by
// the config override mechanism (see HighwayConfig.Overrides →
// Lane.IntervalMs → initLaneSkip in the highway model).
type SpinnerDef struct {
	Frames contract.FrameFunc
}

var (
	builtinSpinners map[string]SpinnerDef
	registerOnce    sync.Once
	SpinnerNames    map[string]struct{}
)

func RegisterAll() {
	registerOnce.Do(func() {
		builtinSpinners = make(map[string]SpinnerDef)
		SpinnerNames = make(map[string]struct{})

		add := func(name string, def SpinnerDef) {
			builtinSpinners[name] = def
			SpinnerNames[name] = struct{}{}
		}

		add(SpinnerTypeBounce, SpinnerDef{Frames: bounceFrame})
		add(SpinnerTypeDefault, SpinnerDef{Frames: spinnerFrame})

		add(SpinnerTypeBraille, SpinnerDef{Frames: frameArraySpinner(spinner.MiniDot.Frames)})
		add(SpinnerTypeBrailleWave, SpinnerDef{Frames: brailleWaveSpinner})
		add(SpinnerTypeDNA, SpinnerDef{Frames: dnaSpinner})
		add(SpinnerTypeScan, SpinnerDef{Frames: scanSpinner})
		add(SpinnerTypeRain, SpinnerDef{Frames: rainSpinner})
		add(SpinnerTypeScanLine, SpinnerDef{Frames: scanLineSpinner})
		add(SpinnerTypeBraillePulse, SpinnerDef{Frames: braillePulseSpinner})
		add(SpinnerTypeSnake, SpinnerDef{Frames: snakeSpinner})
		add(SpinnerTypeSparkle, SpinnerDef{Frames: sparkleSpinner})
		add(SpinnerTypeCascade, SpinnerDef{Frames: cascadeSpinner})
		add(SpinnerTypeColumns, SpinnerDef{Frames: columnsSpinner})
		add(SpinnerTypeOrbit, SpinnerDef{Frames: orbitSpinner})
		add(SpinnerTypeBreathe, SpinnerDef{Frames: breatheSpinner})
		add(SpinnerTypeWaveRows, SpinnerDef{Frames: waveRowsSpinner})
		add(SpinnerTypeCheckerboard, SpinnerDef{Frames: checkerboardSpinner})
		add(SpinnerTypeHelix, SpinnerDef{Frames: helixSpinner})
		add(SpinnerTypeFillSweep, SpinnerDef{Frames: fillSweepSpinner})
		add(SpinnerTypeDiagSwipe, SpinnerDef{Frames: diagSwipeSpinner})

		add(SpinnerTypeClassicWaveform, SpinnerDef{Frames: newClassicWaveform()})
		add(SpinnerTypeParticleDrift, SpinnerDef{Frames: newParticleDrift()})
		add(SpinnerTypePulsingRings, SpinnerDef{Frames: newPulsingRings()})
		add(SpinnerTypeASCIILandscape, SpinnerDef{Frames: newASCIILandscape()})
		add(SpinnerTypeMatrixRain, SpinnerDef{Frames: newMatrixRain()})
		add(SpinnerTypeGradientFlow, SpinnerDef{Frames: newGradientFlow()})
		add(SpinnerTypeBreathingCircles, SpinnerDef{Frames: newBreathingCircles()})
		add(SpinnerTypeNetworkGraph, SpinnerDef{Frames: newNetworkGraph()})

		add(SpinnerTypeWave, SpinnerDef{Frames: newFilmStripFrame(waveStrip)})
		add(SpinnerTypeFairlight, SpinnerDef{Frames: newFilmStripFrame(fairlightStrip)})
		add(SpinnerTypeAmour, SpinnerDef{Frames: newFilmStripFrame(amourStrip)})
		add(SpinnerTypeJamboree, SpinnerDef{Frames: newFilmStripFrame(jamboreeStrip)})
		add(SpinnerTypeMusical, SpinnerDef{Frames: newFilmStripFrame(musicalStrip)})
		add(SpinnerTypeTrinkets, SpinnerDef{Frames: newFilmStripFrame(trinketsStrip)})
		add(SpinnerTypeMorse, SpinnerDef{Frames: newFilmStripFrame(morseStrip)})
		add(SpinnerTypeStarlight, SpinnerDef{Frames: newFilmStripFrame(starlightStrip)})
		add(SpinnerTypeInfantry, SpinnerDef{Frames: newFilmStripFrame(infantryStrip)})
		add(SpinnerTypeHeartThrob, SpinnerDef{Frames: newFilmStripFrame(heartThrobStrip)})
		add(SpinnerTypeBarcode, SpinnerDef{Frames: newFilmStripFrame(barcodeStrip)})

		add(SpinnerTypeDot, SpinnerDef{Frames: frameArraySpinner(spinner.Dot.Frames)})
		add(SpinnerTypeJump, SpinnerDef{Frames: frameArraySpinner(spinner.Jump.Frames)})
		add(SpinnerTypePulse, SpinnerDef{Frames: frameArraySpinner(spinner.Pulse.Frames)})
		add(SpinnerTypePoints, SpinnerDef{Frames: frameArraySpinner(spinner.Points.Frames)})
		add(SpinnerTypeGlobe, SpinnerDef{Frames: frameArraySpinner(spinner.Globe.Frames)})
		add(SpinnerTypeMoon, SpinnerDef{Frames: frameArraySpinner(spinner.Moon.Frames)})
		add(SpinnerTypeMonkey, SpinnerDef{Frames: frameArraySpinner(spinner.Monkey.Frames)})
		add(SpinnerTypeMeter, SpinnerDef{Frames: frameArraySpinner(spinner.Meter.Frames)})
		add(SpinnerTypeHamburger, SpinnerDef{Frames: frameArraySpinner(spinner.Hamburger.Frames)})
		add(SpinnerTypeEllipsis, SpinnerDef{Frames: frameArraySpinner(spinner.Ellipsis.Frames)})
	})
}

func spinnerFrame(tick int) string {
	return spinner.Line.Frames[tick%len(spinner.Line.Frames)]
}

func frameArraySpinner(frames []string) contract.FrameFunc {
	return func(tick int) string {
		return frames[tick%len(frames)]
	}
}

func Lookup(name string) (SpinnerDef, bool) {
	def, ok := builtinSpinners[name]
	return def, ok
}
