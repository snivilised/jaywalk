package prism_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism"
)

var _ = Describe("Palette", func() {

	// ------------------------------------------------------------------
	// ResolveANSI16
	// ------------------------------------------------------------------

	Describe("ResolveANSI16", func() {
		// In lipgloss v2, Color is a function returning color.Color
		// (image/color.Color). We verify the returned value is non-nil
		// and equal to calling lipgloss.Color with the expected number.
		DescribeTable("valid colour names resolve to the correct ANSI number",
			func(input, expectedNumber string) {
				c, err := prism.ResolveANSI16(input)

				Expect(err).To(BeNil())
				Expect(c).NotTo(BeNil())
				Expect(c).To(Equal(lipgloss.Color(expectedNumber)))
			},
			Entry("black", "black", "0"),
			Entry("red", "red", "1"),
			Entry("green", "green", "2"),
			Entry("yellow", "yellow", "3"),
			Entry("blue", "blue", "4"),
			Entry("magenta", "magenta", "5"),
			Entry("cyan", "cyan", "6"),
			Entry("white", "white", "7"),
			Entry("bright-black", "bright-black", "8"),
			Entry("bright-red", "bright-red", "9"),
			Entry("bright-green", "bright-green", "10"),
			Entry("bright-yellow", "bright-yellow", "11"),
			Entry("bright-blue", "bright-blue", "12"),
			Entry("bright-magenta", "bright-magenta", "13"),
			Entry("bright-cyan", "bright-cyan", "14"),
			Entry("bright-white", "bright-white", "15"),
		)

		DescribeTable("valid raw number strings are passed through",
			func(input string) {
				c, err := prism.ResolveANSI16(input)

				Expect(err).To(BeNil())
				Expect(c).NotTo(BeNil())
				Expect(c).To(Equal(lipgloss.Color(input)))
			},
			Entry("0", "0"),
			Entry("7", "7"),
			Entry("8", "8"),
			Entry("15", "15"),
		)

		Context("when the input is empty", func() {
			It("returns nil without error", func() {
				c, err := prism.ResolveANSI16("")

				Expect(err).To(BeNil())
				Expect(c).To(BeNil())
			})
		})

		DescribeTable("unrecognised values return an error",
			func(input string) {
				c, err := prism.ResolveANSI16(input)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(input))
				Expect(c).To(BeNil())
			},
			Entry("unknown name", "turquoise"),
			Entry("CSS name not supported", "cornflowerblue"),
			Entry("hex not accepted at ansi16 tier", "#FF0000"),
			Entry("out-of-range number", "16"),
			Entry("negative number", "-1"),
		)
	})

	// ------------------------------------------------------------------
	// SemanticColour.Resolve
	// ------------------------------------------------------------------

	Describe("SemanticColour.Resolve", func() {
		Context("when all tiers are populated with valid values", func() {
			It("returns three non-nil color.Color values", func() {
				sc := prism.SemanticColour{
					ANSI16:    "cyan",
					ANSI256:   "116",
					TrueColor: "#89DCEB",
				}

				ansi, ansi256, trueCol, err := sc.Resolve()

				Expect(err).To(BeNil())
				Expect(ansi).To(Equal(lipgloss.Color("6")))
				Expect(ansi256).To(Equal(lipgloss.Color("116")))
				Expect(trueCol).To(Equal(lipgloss.Color("#89DCEB")))
			})
		})

		Context("when only the ansi16 tier is set", func() {
			It("returns a non-nil ansi colour and nil upper tiers", func() {
				sc := prism.SemanticColour{ANSI16: "red"}

				ansi, ansi256, trueCol, err := sc.Resolve()

				Expect(err).To(BeNil())
				Expect(ansi).To(Equal(lipgloss.Color("1")))
				Expect(ansi256).To(BeNil())
				Expect(trueCol).To(BeNil())
			})
		})

		Context("when all tiers are empty", func() {
			It("returns three nil color.Color values without error", func() {
				sc := prism.SemanticColour{}

				ansi, ansi256, trueCol, err := sc.Resolve()

				// nil satisfies color.Color - it is the zero value for the
				// interface, indicating no colour is set for that tier.
				Expect(err).To(BeNil())
				Expect(ansi).To(BeNil())
				Expect(ansi256).To(BeNil())
				Expect(trueCol).To(BeNil())
			})
		})

		Context("when the ansi16 value is unrecognised", func() {
			It("returns an error and nil colours", func() {
				sc := prism.SemanticColour{
					ANSI16:    "turquoise",
					ANSI256:   "116",
					TrueColor: "#89DCEB",
				}

				ansi, _, _, err := sc.Resolve()

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("turquoise"))
				Expect(ansi).To(BeNil())
			})
		})
	})

	// ------------------------------------------------------------------
	// SystemPalette
	// ------------------------------------------------------------------

	Describe("SystemPalette", func() {
		It("returns a palette where all ANSI16 fields resolve without error", func() {
			palette := prism.SystemPalette()

			fields := []prism.SemanticColour{
				palette.Directory,
				palette.File,
				palette.Root,
				palette.Action,
				palette.Pipeline,
				palette.Skipped,
				palette.Error,
				palette.Muted,
				palette.Progress,
				palette.BoxBorder,
				palette.SummaryLabel,
				palette.SummaryValue,
				palette.Worker,
				palette.WorkerIdle,
				palette.LaneHeader,
				palette.Header,
				palette.Frame,
				palette.Border,
				palette.BarFilled,
				palette.BarEmpty,
			}

			for _, sc := range fields {
				_, _, _, err := sc.Resolve()
				Expect(err).To(BeNil(),
					"expected all SystemPalette entries to resolve without error",
				)
			}
		})

		It("returns a palette with no TrueColor or ANSI256 values set", func() {
			palette := prism.SystemPalette()

			fields := []prism.SemanticColour{
				palette.Directory,
				palette.File,
				palette.Root,
				palette.Action,
				palette.Pipeline,
				palette.Skipped,
				palette.Error,
				palette.Muted,
				palette.Progress,
				palette.BoxBorder,
				palette.SummaryLabel,
				palette.SummaryValue,
				palette.Worker,
				palette.WorkerIdle,
				palette.LaneHeader,
				palette.Header,
				palette.Frame,
				palette.Border,
				palette.BarFilled,
				palette.BarEmpty,
			}

			for _, sc := range fields {
				Expect(sc.TrueColor).To(BeEmpty(),
					"SystemPalette should not set TrueColor values",
				)
				Expect(sc.ANSI256).To(BeEmpty(),
					"SystemPalette should not set ANSI256 values",
				)
			}
		})

		It("returns a palette where all ANSI16 names are non-empty", func() {
			palette := prism.SystemPalette()

			// Verify every entry has an ANSI16 value - the system palette
			// must always be fully populated so it works without config.
			fields := map[string]string{
				"Directory":    palette.Directory.ANSI16,
				"File":         palette.File.ANSI16,
				"Root":         palette.Root.ANSI16,
				"Action":       palette.Action.ANSI16,
				"Pipeline":     palette.Pipeline.ANSI16,
				"Skipped":      palette.Skipped.ANSI16,
				"Error":        palette.Error.ANSI16,
				"Muted":        palette.Muted.ANSI16,
				"Progress":     palette.Progress.ANSI16,
				"BoxBorder":    palette.BoxBorder.ANSI16,
				"SummaryLabel": palette.SummaryLabel.ANSI16,
				"SummaryValue": palette.SummaryValue.ANSI16,
				"Worker":       palette.Worker.ANSI16,
				"WorkerIdle":   palette.WorkerIdle.ANSI16,
				"LaneHeader":   palette.LaneHeader.ANSI16,
				"Header":       palette.Header.ANSI16,
				"Frame":        palette.Frame.ANSI16,
				"Border":       palette.Border.ANSI16,
				"BarFilled":    palette.BarFilled.ANSI16,
				"BarEmpty":     palette.BarEmpty.ANSI16,
			}

			for name, ansi16 := range fields {
				Expect(ansi16).NotTo(BeEmpty(),
					"SystemPalette.%s.ANSI16 must not be empty", name,
				)
			}
		})
	})

	// ------------------------------------------------------------------
	// Ensure color.Color interface is satisfied
	// ------------------------------------------------------------------

	Describe("lipgloss.Color return type", func() {
		It("satisfies the color.Color interface", func() {
			// lipgloss.Color() returns color.Color - verify this compiles
			// and the value is usable as a standard library colour.
			var c = lipgloss.Color("6")
			Expect(c).NotTo(BeNil())

			// color.Color requires RGBA() - verify it is callable.
			r, g, b, a := c.RGBA()
			// ANSI colours have implementation-defined RGBA values but
			// must not panic.
			_ = r
			_ = g
			_ = b
			_ = a
		})
	})
})
