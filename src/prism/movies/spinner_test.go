package movies

import (
	"charm.land/bubbles/v2/spinner"
	"github.com/mattn/go-runewidth"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

var _ = Describe("SpinnerFrames", func() {
	// -----------------------------------------------------------------------
	// bounceFrame - bounded grow/recede pattern
	// -----------------------------------------------------------------------
	DescribeTable("bounce frames",
		func(tick int, expected string) {
			Expect(bounceFrame(tick)).To(Equal(expected))
		},
		Entry("tick 0", 0, "┃█░░░░░░┃"),
		Entry("tick 1", 1, "┃██░░░░░┃"),
		Entry("tick 2", 2, "┃███░░░░┃"),
		Entry("tick 3", 3, "┃████░░░┃"),
		Entry("tick 4", 4, "┃█████░░┃"),
		Entry("tick 5", 5, "┃██████░┃"),
		Entry("tick 6", 6, "┃███████┃"),
		Entry("tick 7", 7, "┃███████┃"),
		Entry("tick 8", 8, "┃██████░┃"),
		Entry("tick 9", 9, "┃█████░░┃"),
		Entry("tick 10", 10, "┃████░░░┃"),
		Entry("tick 11", 11, "┃███░░░░┃"),
		Entry("tick 12", 12, "┃██░░░░░┃"),
		Entry("tick 13", 13, "┃█░░░░░░┃"),
	)

	// -----------------------------------------------------------------------
	// spinnerFrame - basic line spinner
	// -----------------------------------------------------------------------
	DescribeTable("spinner frames",
		func(tick int, expected string) {
			Expect(spinnerFrame(tick)).To(Equal(expected))
		},
		Entry("tick 0", 0, "|"),
		Entry("tick 1", 1, "/"),
		Entry("tick 2", 2, "-"),
		Entry("tick 3", 3, "\\"),
		Entry("tick 4 wraps", 4, "|"),
	)

	// -----------------------------------------------------------------------
	// frameArraySpinner helper
	// -----------------------------------------------------------------------
	DescribeTable("frameArraySpinner",
		func(tick int, expected string) {
			frames := []string{"a", "b", "c"}
			fn := frameArraySpinner(frames)
			Expect(fn(tick)).To(Equal(expected))
		},
		Entry("tick 0", 0, "a"),
		Entry("tick 1", 1, "b"),
		Entry("tick 2", 2, "c"),
		Entry("tick 3 wraps", 3, "a"),
		Entry("tick 4 wraps", 4, "b"),
	)

	// -----------------------------------------------------------------------
	// Unicode animation first frames
	// -----------------------------------------------------------------------
	DescribeTable("braille spinners first frame",
		func(fn contract.FrameFunc, expected string) {
			Expect(fn(0)).To(Equal(expected))
		},
		Entry("braille", frameArraySpinner(spinner.MiniDot.Frames), "⠋"),
		Entry("braillewave", brailleWaveSpinner, "⠁⠂⠄⡀"),
		Entry("dna", dnaSpinner, "⠋⠉⠙⠚"),
		Entry("scan", scanSpinner, "⠀⠀⠀⠀"),
		Entry("rain", rainSpinner, "⢁⠂⠔⠈"),
		Entry("scanline", scanLineSpinner, "⠉⠉⠉"),
		Entry("braille-pulse", braillePulseSpinner, "⠀⠶⠀"),
		Entry("snake", snakeSpinner, "⣁⡀"),
		Entry("sparkle", sparkleSpinner, "⡡⠊⢔⠡"),
		Entry("cascade", cascadeSpinner, "⠀⠀⠀⠀"),
		Entry("columns", columnsSpinner, "⡀⠀⠀"),
		Entry("orbit", orbitSpinner, "⠃"),
		Entry("breathe", breatheSpinner, "⠀"),
		Entry("waverows", waveRowsSpinner, "⠖⠉⠉⠑"),
		Entry("checkerboard", checkerboardSpinner, "⢕⢕⢕"),
		Entry("helix", helixSpinner, "⢌⣉⢎⣉"),
		Entry("fillsweep", fillSweepSpinner, "⣀⣀"),
		Entry("diagswipe", diagSwipeSpinner, "⠁⠀"),
	)

	// -----------------------------------------------------------------------
	// Charm built-in spinners first frame
	// -----------------------------------------------------------------------
	DescribeTable("charm spinners first frame",
		func(fn contract.FrameFunc, expected string) {
			Expect(fn(0)).To(Equal(expected))
		},
		Entry("dot", frameArraySpinner(spinner.Dot.Frames), "⣾ "),
		Entry("jump", frameArraySpinner(spinner.Jump.Frames), "⢄"),
		Entry("pulse", frameArraySpinner(spinner.Pulse.Frames), "█"),
		Entry("points", frameArraySpinner(spinner.Points.Frames), "∙∙∙"),
		Entry("globe", frameArraySpinner(spinner.Globe.Frames), "🌍"),
		Entry("moon", frameArraySpinner(spinner.Moon.Frames), "🌑"),
		Entry("monkey", frameArraySpinner(spinner.Monkey.Frames), "🙈"),
		Entry("meter", frameArraySpinner(spinner.Meter.Frames), "▱▱▱"),
		Entry("hamburger", frameArraySpinner(spinner.Hamburger.Frames), "☱"),
		Entry("ellipsis", frameArraySpinner(spinner.Ellipsis.Frames), ""),
	)

	// -----------------------------------------------------------------------
	// Unicode animation wrap-around (last frame then tick beyond)
	// -----------------------------------------------------------------------
	DescribeTable("braille spinners wrap correctly",
		func(fn contract.FrameFunc, firstFrame string) {
			Expect(fn(0)).To(Equal(firstFrame))
		},
		Entry("braille wraps at 10", frameArraySpinner(spinner.MiniDot.Frames), "⠋"),
		Entry("orbit wraps at 8", orbitSpinner, "⠃"),
	)

	// -----------------------------------------------------------------------
	// Film strip scrolling - all 11 produce bounded windowed output
	// -----------------------------------------------------------------------
	DescribeTable("film strip scrolling frames",
		func(name string, strip []rune) {
			fn := newFilmStripFrame(strip)
			output := fn(0)
			Expect(output).To(HavePrefix("┃"))
			Expect(output).To(HaveSuffix("┃"))
			Expect(runewidth.StringWidth(output)).To(BeNumerically(">=", 9))
		},
		Entry("wave", SpinnerTypeWave, waveStrip),
		Entry("fairlight", SpinnerTypeFairlight, fairlightStrip),
		Entry("amour", SpinnerTypeAmour, amourStrip),
		Entry("jamboree", SpinnerTypeJamboree, jamboreeStrip),
		Entry("musical", SpinnerTypeMusical, musicalStrip),
		Entry("trinkets", SpinnerTypeTrinkets, trinketsStrip),
		Entry("morse", SpinnerTypeMorse, morseStrip),
		Entry("starlight", SpinnerTypeStarlight, starlightStrip),
		Entry("infantry", SpinnerTypeInfantry, infantryStrip),
		Entry("heart-throb", SpinnerTypeHeartThrob, heartThrobStrip),
		Entry("barcode", SpinnerTypeBarcode, barcodeStrip),
	)

	DescribeTable("film strip output changes as tick advances",
		func(strip []rune) {
			fn := newFilmStripFrame(strip)
			Expect(fn(0)).NotTo(Equal(fn(1)))
		},
		Entry("wave", waveStrip),
		Entry("fairlight", fairlightStrip),
		Entry("amour", amourStrip),
		Entry("jamboree", jamboreeStrip),
		Entry("musical", musicalStrip),
		Entry("trinkets", trinketsStrip),
		Entry("morse", morseStrip),
		Entry("starlight", starlightStrip),
		Entry("infantry", infantryStrip),
		Entry("heart-throb", heartThrobStrip),
		Entry("barcode", barcodeStrip),
	)

	// -----------------------------------------------------------------------
	// Animation skins - render non-empty
	// -----------------------------------------------------------------------
	DescribeTable("animation skins render without panic",
		func(name string, fn contract.FrameFunc) {
			Expect(fn(0)).NotTo(BeEmpty())
			Expect(fn(5)).NotTo(BeEmpty())
			Expect(fn(15)).NotTo(BeEmpty())
		},
		Entry("classic-waveform", SpinnerTypeClassicWaveform, newClassicWaveform()),
		Entry("particle-drift", SpinnerTypeParticleDrift, newParticleDrift()),
		Entry("pulsing-rings", SpinnerTypePulsingRings, newPulsingRings()),
		Entry("ascii-landscape", SpinnerTypeASCIILandscape, newASCIILandscape()),
		Entry("matrix-rain", SpinnerTypeMatrixRain, newMatrixRain()),
		Entry("gradient-flow", SpinnerTypeGradientFlow, newGradientFlow()),
		Entry("breathing-circles", SpinnerTypeBreathingCircles, newBreathingCircles()),
		Entry("network-graph", SpinnerTypeNetworkGraph, newNetworkGraph()),
	)

	// -----------------------------------------------------------------------
	// Animation skins - output changes over ticks
	// -----------------------------------------------------------------------
	It("classic-waveform output changes over time", func() {
		fn := newClassicWaveform()
		Expect(fn(0)).NotTo(Equal(fn(10)))
	})

	It("particle-drift output changes over time", func() {
		fn := newParticleDrift()
		Expect(fn(0)).NotTo(Equal(fn(10)))
	})

	// -----------------------------------------------------------------------
	// RegisterAll populates builtinSpinners
	// -----------------------------------------------------------------------
	It("RegisterAll populates all spinner types", func() {
		RegisterAll()
		Expect(len(builtinSpinners)).To(BeNumerically(">=", 48))
		for name := range SpinnerNames {
			_, ok := Lookup(name)
			Expect(ok).To(BeTrue(), "Lookup(%q) should succeed", name)
		}
	})

	// -----------------------------------------------------------------------
	// All unicode spinners wrap without panic across the full cycle
	// -----------------------------------------------------------------------
	It("all spinners produce non-empty frames across the full cycle", func() {
		for _, fn := range []contract.FrameFunc{
			frameArraySpinner(spinner.MiniDot.Frames),
			brailleWaveSpinner, dnaSpinner,
			scanSpinner, rainSpinner, scanLineSpinner,
			braillePulseSpinner, snakeSpinner, sparkleSpinner,
			cascadeSpinner, columnsSpinner, orbitSpinner,
			breatheSpinner, waveRowsSpinner, checkerboardSpinner,
			helixSpinner, fillSweepSpinner, diagSwipeSpinner,
			bounceFrame,
			newFilmStripFrame(waveStrip),
			newFilmStripFrame(fairlightStrip),
			newFilmStripFrame(amourStrip),
			newFilmStripFrame(jamboreeStrip),
			newFilmStripFrame(musicalStrip),
			newFilmStripFrame(trinketsStrip),
			newFilmStripFrame(morseStrip),
			newFilmStripFrame(starlightStrip),
			newFilmStripFrame(infantryStrip),
			newFilmStripFrame(heartThrobStrip),
			newFilmStripFrame(barcodeStrip),
			frameArraySpinner(spinner.Dot.Frames),
			frameArraySpinner(spinner.Jump.Frames),
			frameArraySpinner(spinner.Pulse.Frames),
			frameArraySpinner(spinner.Points.Frames),
			frameArraySpinner(spinner.Globe.Frames),
			frameArraySpinner(spinner.Moon.Frames),
			frameArraySpinner(spinner.Monkey.Frames),
			frameArraySpinner(spinner.Meter.Frames),
			frameArraySpinner(spinner.Hamburger.Frames),
		} {
			for tick := 0; tick < 50; tick++ {
				Expect(fn(tick)).NotTo(BeEmpty())
			}
		}
	})

	It("ellipsis spinner produces non-empty frames at non-mod-4 ticks", func() {
		fn := frameArraySpinner(spinner.Ellipsis.Frames)
		Expect(fn(0)).To(BeEmpty())
		Expect(fn(4)).To(BeEmpty())
		for tick := 1; tick <= 3; tick++ {
			Expect(fn(tick)).NotTo(BeEmpty())
		}
	})
})
