package contract_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"charm.land/lipgloss/v2"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

var _ = Describe("Theme", func() {

	// ------------------------------------------------------------------
	// NewTheme with SystemPalette
	// ------------------------------------------------------------------

	Describe("NewTheme with SystemPalette", func() {
		It("constructs a Theme without error", func() {
			w := &bytes.Buffer{}
			palette := contract.SystemPalette()

			theme, err := contract.NewTheme(palette, w)

			Expect(err).To(BeNil())
			Expect(theme).NotTo(Equal(contract.Theme{}))
		})

		It("includes all style fields in the Theme", func() {
			w := &bytes.Buffer{}
			palette := contract.SystemPalette()

			theme, err := contract.NewTheme(palette, w)

			Expect(err).To(BeNil())
			Expect(theme.DirStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.FileStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.ActionStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.PipelineStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.LandingStripStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.SkippedStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.BoxStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.SummaryLabelStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.SummaryValueStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.ErrorStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.MutedStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.ProgressStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.WorkerStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.WorkerIdleStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.LaneHeaderStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.HeaderStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.FrameStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.BorderStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.BarFilledStyle).NotTo(Equal(lipgloss.NewStyle()))
			Expect(theme.BarEmptyStyle).NotTo(Equal(lipgloss.NewStyle()))
		})

		It("sets bold on appropriate styles", func() {
			w := &bytes.Buffer{}
			palette := contract.SystemPalette()

			theme, err := contract.NewTheme(palette, w)

			Expect(err).To(BeNil())

			dirStyleBold := theme.DirStyle.Bold(false)
			Expect(theme.DirStyle).NotTo(Equal(dirStyleBold))

			rootStyleBold := theme.RootStyle.Bold(false)
			Expect(theme.RootStyle).NotTo(Equal(rootStyleBold))

			actionStyleBold := theme.ActionStyle.Bold(false)
			Expect(theme.ActionStyle).NotTo(Equal(actionStyleBold))

			pipelineStyleBold := theme.PipelineStyle.Bold(false)
			Expect(theme.PipelineStyle).NotTo(Equal(pipelineStyleBold))

			summaryLabelStyleBold := theme.SummaryLabelStyle.Bold(false)
			Expect(theme.SummaryLabelStyle).NotTo(Equal(summaryLabelStyleBold))

			errorStyleBold := theme.ErrorStyle.Bold(false)
			Expect(theme.ErrorStyle).NotTo(Equal(errorStyleBold))

			workerStyleBold := theme.WorkerStyle.Bold(false)
			Expect(theme.WorkerStyle).NotTo(Equal(workerStyleBold))

			laneHeaderStyleBold := theme.LaneHeaderStyle.Bold(false)
			Expect(theme.LaneHeaderStyle).NotTo(Equal(laneHeaderStyleBold))

			headerStyleBold := theme.HeaderStyle.Bold(false)
			Expect(theme.HeaderStyle).NotTo(Equal(headerStyleBold))
		})

		It("sets faint on appropriate styles", func() {
			w := &bytes.Buffer{}
			palette := contract.SystemPalette()

			theme, err := contract.NewTheme(palette, w)

			Expect(err).To(BeNil())

			skippedStyleFaint := theme.SkippedStyle.Faint(false)
			Expect(theme.SkippedStyle).NotTo(Equal(skippedStyleFaint))

			mutedStyleFaint := theme.MutedStyle.Faint(false)
			Expect(theme.MutedStyle).NotTo(Equal(mutedStyleFaint))

			workerIdleStyleFaint := theme.WorkerIdleStyle.Faint(false)
			Expect(theme.WorkerIdleStyle).NotTo(Equal(workerIdleStyleFaint))
		})

		It("sets up BoxStyle with rounded border", func() {
			w := &bytes.Buffer{}
			palette := contract.SystemPalette()

			theme, err := contract.NewTheme(palette, w)

			Expect(err).To(BeNil())

			// Check that BoxStyle has a border by creating a copy without it
			noBorderStyle := theme.BoxStyle.Border(lipgloss.NormalBorder())
			Expect(theme.BoxStyle).NotTo(Equal(noBorderStyle))
		})

		It("sets up SummaryLabelStyle with width constraint", func() {
			w := &bytes.Buffer{}
			palette := contract.SystemPalette()

			theme, err := contract.NewTheme(palette, w)

			Expect(err).To(BeNil())

			// Check that SummaryLabelStyle has a non-zero width by creating a copy without it
			noWidthStyle := theme.SummaryLabelStyle.Width(0)
			Expect(theme.SummaryLabelStyle).NotTo(Equal(noWidthStyle))
		})

		It("sets up SummaryValueStyle", func() {
			w := &bytes.Buffer{}
			palette := contract.SystemPalette()

			theme, err := contract.NewTheme(palette, w)

			Expect(err).To(BeNil())
			Expect(theme.SummaryValueStyle).NotTo(Equal(lipgloss.NewStyle()))
		})

		It("includes TreeIcons with default values", func() {
			w := &bytes.Buffer{}
			palette := contract.SystemPalette()

			theme, err := contract.NewTheme(palette, w)

			Expect(err).To(BeNil())
			Expect(theme.TreeIcons).NotTo(BeNil())
			Expect(len(theme.TreeIcons)).To(BeNumerically(">", 0))
		})

		It("includes empty HighlightGradients map", func() {
			w := &bytes.Buffer{}
			palette := contract.SystemPalette()

			theme, err := contract.NewTheme(palette, w)

			Expect(err).To(BeNil())
			Expect(theme.HighlightGradients).NotTo(BeNil())
			Expect(len(theme.HighlightGradients)).To(Equal(0))
		})

		It("includes empty HighlightsComponents map", func() {
			w := &bytes.Buffer{}
			palette := contract.SystemPalette()

			theme, err := contract.NewTheme(palette, w)

			Expect(err).To(BeNil())
			Expect(theme.HighlightsComponents).NotTo(BeNil())
			Expect(len(theme.HighlightsComponents)).To(Equal(0))
		})
	})

	// ------------------------------------------------------------------
	// NewTheme with custom palette values
	// ------------------------------------------------------------------

	Describe("NewTheme with custom palette", func() {
		Context("accepts a fully populated TrueColor palette", func() {
			w := &bytes.Buffer{}
			palette := contract.SystemPalette()
			palette.Directory = contract.SemanticColour{TrueColor: "#00FF00"}
			palette.File = contract.SemanticColour{TrueColor: "#FFFF00"}

			theme, err := contract.NewTheme(palette, w)

			Expect(err).To(BeNil())

			// Check that Foreground is set by creating a copy without it
			dirNoForeground := theme.DirStyle.Foreground(nil)
			Expect(theme.DirStyle).NotTo(Equal(dirNoForeground))

			fileNoForeground := theme.FileStyle.Foreground(nil)
			Expect(theme.FileStyle).NotTo(Equal(fileNoForeground))
		})

		It("accepts a mixed-tier palette", func() {
			w := &bytes.Buffer{}
			palette := contract.SystemPalette()
			palette.Directory = contract.SemanticColour{ANSI16: "cyan", TrueColor: "#00FF00"}
			palette.File = contract.SemanticColour{ANSI16: "white", ANSI256: "255", TrueColor: "#FFFF00"}

			_, err := contract.NewTheme(palette, w)

			Expect(err).To(BeNil())
		})
	})

	// ------------------------------------------------------------------
	// Gradient lookup tests (NEW - Highway gradient fix)
	// ------------------------------------------------------------------

	Describe("GradientFor", func() {
		Context("lookup failures", func() {
			It("returns empty with unknown component name", func() {
				palette := contract.SystemPalette()
				w := &bytes.Buffer{}

				theme, err := contract.NewTheme(palette, w)
				Expect(err).To(BeNil())

				grad, has := theme.GradientFor("some-random-component")
				Expect(has).To(BeFalse())
				Expect(grad).To(Equal(contract.ResolvedGradient{}))
			})
		})

		Context("with highlights configured in palette", func() {
			It("loads gradients when defined in palette", func() {
				palette := contract.SystemPalette()

				// Configure a gradient and its component mapping
				highlights := contract.HighlightsConfig{
					Gradients: map[string]contract.GradientDef{
						"aurora-borealis": {
							Steps: 8,
							Hi:    &contract.SemanticColour{ANSI16: "cyan", TrueColor: "#00E5FF"},
							Lo:    &contract.SemanticColour{ANSI16: "magenta", TrueColor: "#B388FF"},
						},
					},
					Components: map[string]string{
						contract.GradientComponentActivity: "aurora-borealis",
					},
				}

				palette.Highlights = highlights
				w := &bytes.Buffer{}

				theme, err := contract.NewTheme(palette, w)
				Expect(err).To(BeNil())

				// Verify gradient was loaded
				Expect(theme.HighlightGradients).NotTo(BeNil())
				Expect(len(theme.HighlightGradients)).To(Equal(1))
				Expect(theme.HighlightGradients["aurora-borealis"]).NotTo(BeNil())

				// Verify components map
				Expect(theme.HighlightsComponents).NotTo(BeNil())
				Expect(len(theme.HighlightsComponents)).To(Equal(1))
				gradientName, ok := theme.HighlightsComponents[contract.GradientComponentActivity]
				Expect(ok).To(BeTrue())
				Expect(gradientName).To(Equal("aurora-borealis"))
			})

			It("retrieves gradient via GradientFor successfully", func() {
				palette := contract.SystemPalette()

				highlights := contract.HighlightsConfig{
					Gradients: map[string]contract.GradientDef{
						"ember-glow": {
							Steps: 4,
							Hi:    &contract.SemanticColour{ANSI16: "red", TrueColor: "#FF5733"},
							Lo:    &contract.SemanticColour{ANSI16: "yellow", TrueColor: "#FFA500"},
						},
					},
					Components: map[string]string{
						contract.GradientComponentActivity: "ember-glow",
					},
				}

				palette.Highlights = highlights
				w := &bytes.Buffer{}

				theme, err := contract.NewTheme(palette, w)
				Expect(err).To(BeNil())

				grad, has := theme.GradientFor(contract.GradientComponentActivity)
				Expect(has).To(BeTrue())
				Expect(grad.Steps).To(Equal(4))
				// Verify Hi and Lo are non-nil (they're resolved by NewTheme)
				Expect(grad.Hi).NotTo(BeNil())
				Expect(grad.Lo).NotTo(BeNil())

				// Verify colours are reasonable (not zero values)
				r, g, b, _ := grad.Hi.RGBA()
				Expect(r).NotTo(Equal(0))
				Expect(g).NotTo(Equal(0))
				Expect(b).NotTo(Equal(0))
			})

			It("handles missing gradient definition gracefully", func() {
				palette := contract.SystemPalette()

				highlights := contract.HighlightsConfig{
					Gradients: map[string]contract.GradientDef{
						"other-gradient": {
							Hi: &contract.SemanticColour{ANSI16: "red"},
							Lo: &contract.SemanticColour{ANSI16: "blue"},
						},
					},
					Components: map[string]string{
						contract.GradientComponentActivity: "unknown-gradient", // gradient doesn't exist
					},
				}

				palette.Highlights = highlights
				w := &bytes.Buffer{}

				theme, err := contract.NewTheme(palette, w)
				Expect(err).To(BeNil())

				grad, has := theme.GradientFor(contract.GradientComponentActivity)
				// Component exists but gradient doesn't → lookup fails gracefully
				Expect(has).To(BeFalse())
				Expect(grad.Steps).To(Equal(0))
				Expect(grad.Hi).To(BeNil())
				Expect(grad.Lo).To(BeNil())
			})

			It("handles missing component mapping gracefully", func() {
				palette := contract.SystemPalette()

				highlights := contract.HighlightsConfig{
					Gradients: map[string]contract.GradientDef{
						"existing-gradient": {
							Hi: &contract.SemanticColour{ANSI16: "red"},
							Lo: &contract.SemanticColour{ANSI16: "blue"},
						},
					},
					// No component mapping for activity-control
				}

				palette.Highlights = highlights
				w := &bytes.Buffer{}

				theme, err := contract.NewTheme(palette, w)
				Expect(err).To(BeNil())

				grad, has := theme.GradientFor(contract.GradientComponentActivity)
				// Component doesn't exist → lookup fails gracefully
				Expect(has).To(BeFalse())
				Expect(grad.Steps).To(Equal(0))
			})

			It("works with only Hi endpoint (derives Lo)", func() {
				palette := contract.SystemPalette()

				highlights := contract.HighlightsConfig{
					Gradients: map[string]contract.GradientDef{
						"hi-only": {
							Hi:    &contract.SemanticColour{ANSI16: "cyan"},
							Steps: 8,
						},
					},
					Components: map[string]string{
						contract.GradientComponentActivity: "hi-only",
					},
				}

				palette.Highlights = highlights
				w := &bytes.Buffer{}

				theme, err := contract.NewTheme(palette, w)
				Expect(err).To(BeNil())

				grad, has := theme.GradientFor(contract.GradientComponentActivity)
				Expect(has).To(BeTrue())
				Expect(grad.Steps).To(Equal(8))
				Expect(grad.Hi).NotTo(BeNil())
				// Lo is derived from Hi; should exist and be dimmed (not zero values)
				Expect(grad.Lo).NotTo(BeNil())
				rLo, gLo, bLo, _ := grad.Lo.RGBA()
				Expect(rLo).NotTo(Equal(0))
				Expect(gLo).NotTo(Equal(0))
				Expect(bLo).NotTo(Equal(0))
			})

			It("works with only Lo endpoint (derives Hi)", func() {
				palette := contract.SystemPalette()

				highlights := contract.HighlightsConfig{
					Gradients: map[string]contract.GradientDef{
						"lo-only": {
							Lo:    &contract.SemanticColour{ANSI16: "magenta"},
							Steps: 8,
						},
					},
					Components: map[string]string{
						contract.GradientComponentActivity: "lo-only",
					},
				}

				palette.Highlights = highlights
				w := &bytes.Buffer{}

				theme, err := contract.NewTheme(palette, w)
				Expect(err).To(BeNil())

				grad, has := theme.GradientFor(contract.GradientComponentActivity)
				Expect(has).To(BeTrue())
				Expect(grad.Steps).To(Equal(8))
				Expect(grad.Lo).NotTo(BeNil())
				// Hi is derived from Lo; should exist and be brightened (not zero values)
				Expect(grad.Hi).NotTo(BeNil())
				rHi, gHi, bHi, _ := grad.Hi.RGBA()
				Expect(rHi).NotTo(Equal(0))
				Expect(gHi).NotTo(Equal(0))
				Expect(bHi).NotTo(Equal(0))
			})
		})

		Context("handles zero steps", func() {
			It("uses default step count when steps=0", func() {
				palette := contract.SystemPalette()

				highlights := contract.HighlightsConfig{
					Gradients: map[string]contract.GradientDef{
						"zero-steps": {
							Steps: 0, // Should use default
							Hi:    &contract.SemanticColour{ANSI16: "cyan"},
							Lo:    &contract.SemanticColour{ANSI16: "magenta"},
						},
					},
					Components: map[string]string{
						contract.GradientComponentActivity: "zero-steps",
					},
				}

				palette.Highlights = highlights
				w := &bytes.Buffer{}

				theme, err := contract.NewTheme(palette, w)
				Expect(err).To(BeNil())

				grad, has := theme.GradientFor(contract.GradientComponentActivity)
				Expect(has).To(BeTrue())
				// Should use default step count (8) when 0 is specified
				Expect(grad.Steps).To(Equal(8))
			})
		})
	})

	Describe("Gradient lookup roundtrip", func() {
		It("roundtrips component -> gradient -> resolved gradient correctly", func() {
			palette := contract.SystemPalette()

			highlights := contract.HighlightsConfig{
				Gradients: map[string]contract.GradientDef{
					"test-gradient-1": {
						Steps: 6,
						Hi:    &contract.SemanticColour{ANSI16: "blue"},
						Lo:    &contract.SemanticColour{ANSI16: "yellow"},
					},
				},
				Components: map[string]string{
					contract.GradientComponentActivity: "test-gradient-1",
				},
			}

			palette.Highlights = highlights
			w := &bytes.Buffer{}

			theme, err := contract.NewTheme(palette, w)
			Expect(err).To(BeNil())

			// Roundtrip test: component name -> gradient name -> resolved gradient
			componentName := contract.GradientComponentActivity
			actualGradientName, exists := theme.HighlightsComponents[componentName]
			Expect(exists).To(BeTrue())
			Expect(actualGradientName).To(Equal("test-gradient-1"))

			resolved, has := theme.GradientFor(componentName)
			Expect(has).To(BeTrue())
			Expect(resolved.Steps).To(Equal(6))
			Expect(resolved.Hi).NotTo(BeNil())
			Expect(resolved.Lo).NotTo(BeNil())
		})

		It("works with multiple gradients but correct component lookup", func() {
			palette := contract.SystemPalette()

			highlights := contract.HighlightsConfig{
				Gradients: map[string]contract.GradientDef{
					"gradient-alpha": {
						Steps: 8,
						Hi:    &contract.SemanticColour{ANSI16: "red"},
						Lo:    &contract.SemanticColour{ANSI16: "green"},
					},
					"gradient-beta": {
						Steps: 4,
						Hi:    &contract.SemanticColour{ANSI16: "blue"},
						Lo:    &contract.SemanticColour{ANSI16: "yellow"},
					},
				},
				Components: map[string]string{
					contract.GradientComponentActivity: "gradient-alpha", // highway uses alpha only
				},
			}

			palette.Highlights = highlights
			w := &bytes.Buffer{}

			theme, err := contract.NewTheme(palette, w)
			Expect(err).To(BeNil())

			// Component should resolve to "gradient-alpha", not "gradient-beta"
			resolved, has := theme.GradientFor(contract.GradientComponentActivity)
			Expect(has).To(BeTrue())
			Expect(resolved.Steps).To(Equal(8)) // gradient-alpha's steps
			Expect(resolved.Hi).NotTo(BeNil())
			Expect(resolved.Lo).NotTo(BeNil())

			// Verify beta is a separate gradient
			beta, has := theme.HighlightGradients["gradient-beta"]
			Expect(has).To(BeTrue())
			Expect(beta.Steps).To(Equal(4))
		})
	})
})
