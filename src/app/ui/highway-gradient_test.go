//go:build !race
// +build !race

package ui_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism"
)

var _ = Describe("Highway Gradient API", func() {
	var (
		palette prism.Palette
		w       *bytes.Buffer
		err     error
		theme   prism.Theme
	)

	Describe("GradientFor component-based lookup", func() {
		Context("with properly configured highlights", func() {
			It("retrieves gradient using component name constant", func() {
				highlights := prism.HighlightsConfig{
					Gradients: map[string]prism.GradientDef{
						"aurora-borealis": {
							Steps: 8,
							Hi:    &prism.SemanticColour{ANSI16: "cyan", TrueColor: "#00E5FF"},
							Lo:    &prism.SemanticColour{ANSI16: "magenta", TrueColor: "#B388FF"},
						},
					},
					Components: map[string]string{
						prism.GradientComponentActivity: "aurora-borealis",
					},
				}

				palette.Highlights = highlights
				w = &bytes.Buffer{}

				theme, err = prism.NewTheme(palette, w)
				Expect(err).To(BeNil())
				Expect(theme).NotTo(BeNil())

				gGradient, has := theme.GradientFor(prism.GradientComponentActivity)
				Expect(has).To(BeTrue())
				Expect(gGradient.Steps).To(Equal(8))
				Expect(gGradient.Hi).NotTo(BeNil())
				Expect(gGradient.Lo).NotTo(BeNil())

				// Verify colours are within expected ranges (non-zero for Hi endpoint)
				r, gColor, bVal, _ := gGradient.Hi.RGBA()
				Expect(r).NotTo(Equal(0))
				Expect(gColor).NotTo(Equal(0))
				Expect(bVal).NotTo(Equal(0))
			})

			It("returns false when component is not configured", func() {
				highlights := prism.HighlightsConfig{
					Components: map[string]string{}, // empty
				}

				palette.Highlights = highlights
				w.Reset()

				theme, err = prism.NewTheme(palette, w)
				Expect(err).To(BeNil())

				gGrad, has := theme.GradientFor(prism.GradientComponentActivity)
				Expect(has).To(BeFalse())
				Expect(gGrad.Steps).To(Equal(0))
				Expect(gGrad.Hi).To(BeNil())
			})
		})

		Context("handles missing gradient definition", func() {
			It("fails gracefully when component maps to non-existent gradient", func() {
				highlights := prism.HighlightsConfig{
					Gradients: map[string]prism.GradientDef{
						"existing-gradient": {
							Hi: &prism.SemanticColour{ANSI16: "red"},
							Lo: &prism.SemanticColour{ANSI16: "blue"},
						},
					},
					Components: map[string]string{
						prism.GradientComponentActivity: "nonexistent", // doesn't exist
					},
				}

				palette.Highlights = highlights
				w.Reset()

				theme, err = prism.NewTheme(palette, w)
				Expect(err).To(BeNil())

				_, has := theme.GradientFor(prism.GradientComponentActivity)
				// Component exists but gradient doesn't → lookup fails gracefully
				Expect(has).To(BeFalse())
			})
		})

		Context("handles missing component mapping", func() {
			It("fails gracefully when no component entry exists", func() {
				highlights := prism.HighlightsConfig{
					Gradients: map[string]prism.GradientDef{
						"some-gradient": {
							Hi: &prism.SemanticColour{ANSI16: "cyan"},
							Lo: &prism.SemanticColour{ANSI16: "magenta"},
						},
					},
					// No Components mapping at all or no activity-control entry
				}

				palette.Highlights = highlights
				w.Reset()

				theme, err = prism.NewTheme(palette, w)
				Expect(err).To(BeNil())

				_, has := theme.GradientFor(prism.GradientComponentActivity)
				Expect(has).To(BeFalse())
			})
		})

		Context("works with partial gradient endpoints", func() {
			It("loads gradient when only Hi specified (derives Lo)", func() {
				highlights := prism.HighlightsConfig{
					Gradients: map[string]prism.GradientDef{
						"hi-only": {
							Hi:    &prism.SemanticColour{ANSI16: "cyan", TrueColor: "#00FFFF"},
							Steps: 8,
						},
					},
					Components: map[string]string{
						prism.GradientComponentActivity: "hi-only",
					},
				}

				palette.Highlights = highlights
				w.Reset()

				theme, err = prism.NewTheme(palette, w)
				Expect(err).To(BeNil())

				gGrad, has := theme.GradientFor(prism.GradientComponentActivity)
				Expect(has).To(BeTrue())
				Expect(gGrad.Steps).To(Equal(8))
				Expect(gGrad.Hi).NotTo(BeNil())
				// Lo is nil since only Hi was provided; renderer will derive it
			})

			It("loads gradient when only Lo specified (derives Hi)", func() {
				highlights := prism.HighlightsConfig{
					Gradients: map[string]prism.GradientDef{
						"lo-only": {
							Lo:    &prism.SemanticColour{ANSI16: "magenta", TrueColor: "#FF1744"},
							Steps: 6,
						},
					},
					Components: map[string]string{
						prism.GradientComponentActivity: "lo-only",
					},
				}

				palette.Highlights = highlights
				w.Reset()

				theme, err = prism.NewTheme(palette, w)
				Expect(err).To(BeNil())

				gGrad, has := theme.GradientFor(prism.GradientComponentActivity)
				Expect(has).To(BeTrue())
				Expect(gGrad.Steps).To(Equal(6))
				Expect(gGrad.Lo).NotTo(BeNil())
				// Hi is nil since only Lo was provided; renderer will derive it
			})

			It("uses default step count when steps=0", func() {
				highlights := prism.HighlightsConfig{
					Gradients: map[string]prism.GradientDef{
						"zero-steps": {
							Steps: 0, // triggers default
							Hi:    &prism.SemanticColour{ANSI16: "cyan"},
							Lo:    &prism.SemanticColour{ANSI16: "magenta"},
						},
					},
					Components: map[string]string{
						prism.GradientComponentActivity: "zero-steps",
					},
				}

				palette.Highlights = highlights
				w.Reset()

				theme, err = prism.NewTheme(palette, w)
				Expect(err).To(BeNil())

				gGrad, has := theme.GradientFor(prism.GradientComponentActivity)
				Expect(has).To(BeTrue())
				// Should default to 8 steps when not specified
				Expect(gGrad.Steps).To(Equal(8))
			})
		})

		Context("handles multiple gradient definitions", func() {
			It("looks up each independent gradient correctly", func() {
				highlights := prism.HighlightsConfig{
					Gradients: map[string]prism.GradientDef{
						"gradient-alpha": {
							Hi: &prism.SemanticColour{ANSI16: "red", TrueColor: "#FF4D6D"},
							Lo: &prism.SemanticColour{ANSI16: "green", TrueColor: "#B9FBC0"},
						},
						"gradient-beta": {
							Hi: &prism.SemanticColour{ANSI16: "blue", TrueColor: "#3D5AFE"},
							Lo: &prism.SemanticColour{ANSI16: "yellow", TrueColor: "#FFEA00"},
						},
					},
					Components: map[string]string{
						prism.GradientComponentActivity: "gradient-alpha", // highway uses alpha only
					},
				}

				palette.Highlights = highlights
				w.Reset()

				theme, err = prism.NewTheme(palette, w)
				Expect(err).To(BeNil())

				gAlpha, hasAlpha := theme.GradientFor(prism.GradientComponentActivity)
				Expect(hasAlpha).To(BeTrue())
				Expect(gAlpha.Steps).To(Equal(8)) // default steps from GradientDef when not specified

				// Verify beta is separate and independently loaded
				_, hasBeta := theme.HighlightGradients["gradient-beta"]
				Expect(hasBeta).To(BeTrue())
			})
		})
	})

	Describe("Theme gradient map structure", func() {
		It("populates HighlightGradients map when gradients defined", func() {
			highlights := prism.HighlightsConfig{
				Gradients: map[string]prism.GradientDef{
					"test-gradient": {
						Steps: 4,
						Hi:    &prism.SemanticColour{ANSI16: "cyan"},
						Lo:    &prism.SemanticColour{ANSI16: "magenta"},
					},
				},
			}

			palette.Highlights = highlights
			w.Reset()

			theme, err = prism.NewTheme(palette, w)
			Expect(err).To(BeNil())

			Expect(theme.HighlightGradients).NotTo(BeNil())
			Expect(len(theme.HighlightGradients)).To(Equal(1))
			Expect(theme.HighlightGradients["test-gradient"]).NotTo(BeNil())
		})

		It("populates HighlightsComponents map when components defined", func() {
			highlights := prism.HighlightsConfig{
				Components: map[string]string{
					prism.GradientComponentActivity: "some-gradient",
				},
			}

			palette.Highlights = highlights
			w.Reset()

			theme, err = prism.NewTheme(palette, w)
			Expect(err).To(BeNil())

			Expect(theme.HighlightsComponents).NotTo(BeNil())
			Expect(len(theme.HighlightsComponents)).To(Equal(1))
			val, ok := theme.HighlightsComponents[prism.GradientComponentActivity]
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal("some-gradient"))
		})
	})

	Describe("Component-based vs direct gradient name lookup", func() {
		It("fails when passing gradient name directly to GradientFor()", func() {
			highlights := prism.HighlightsConfig{
				Gradients: map[string]prism.GradientDef{
					"aurora-borealis": {
						Hi: &prism.SemanticColour{ANSI16: "cyan"},
						Lo: &prism.SemanticColour{ANSI16: "magenta"},
					},
				},
				Components: map[string]string{
					prism.GradientComponentActivity: "aurora-borealis",
				},
			}

			palette.Highlights = highlights
			w.Reset()

			theme, err = prism.NewTheme(palette, w)
			Expect(err).To(BeNil())

			gGrad, has := theme.GradientFor("aurora-borealis")
			_ = gGrad
			Expect(has).To(BeFalse()) // because HighlightsComponents["aurora-borealis"] doesn't exist!

			// Correct way: use component-based lookup
			gGrad, has = theme.GradientFor(prism.GradientComponentActivity)
			Expect(has).To(BeTrue())
			Expect(gGrad).NotTo(BeNil())
		})
	})

	Describe("sendMotif gradient resolution", func() {
		It("resolves gradient via component name (sendMotif pattern), not gradient name", func() {
			highlights := prism.HighlightsConfig{
				Gradients: map[string]prism.GradientDef{
					"aurora-borealis": {
						Steps: 8,
						Hi:    &prism.SemanticColour{ANSI16: "cyan"},
						Lo:    &prism.SemanticColour{ANSI16: "magenta"},
					},
				},
				Components: map[string]string{
					prism.GradientComponentActivity: "aurora-borealis",
				},
			}

			palette.Highlights = highlights
			w.Reset()

			theme, err = prism.NewTheme(palette, w)
			Expect(err).To(BeNil())

			// Correct: resolve using component name (what sendMotif must do)
			grad, has := theme.GradientFor(prism.GradientComponentActivity)
			Expect(has).To(BeTrue())
			Expect(grad.Steps).To(Equal(8))
			Expect(grad.Hi).NotTo(BeNil())
			Expect(grad.Lo).NotTo(BeNil())

			// Anti-pattern: passing gradient name directly to GradientFor fails
			// because GradientFor expects a component name, not a gradient name.
			// sendMotif must NOT pre-resolve the component → gradient name manually.
			_, hasGradName := theme.GradientFor("aurora-borealis")
			Expect(hasGradName).To(BeFalse())
		})
	})

	Describe("Empty highlights config", func() {
		It("creates valid Theme with empty gradient maps", func() {
			palette = prism.SystemPalette() // has no highlights
			w.Reset()

			theme, err = prism.NewTheme(palette, w)
			Expect(err).To(BeNil())

			Expect(theme.HighlightGradients).NotTo(BeNil())
			Expect(len(theme.HighlightGradients)).To(Equal(0))

			Expect(theme.HighlightsComponents).NotTo(BeNil())
			Expect(len(theme.HighlightsComponents)).To(Equal(0))

			gGrad, has := theme.GradientFor(prism.GradientComponentActivity)
			Expect(has).To(BeFalse())
			Expect(gGrad.Steps).To(Equal(0))
		})
	})

	Describe("Periscope component gradient", func() {
		It("retrieves periscope gradient using component name constant", func() {
			highlights := prism.HighlightsConfig{
				Gradients: map[string]prism.GradientDef{
					"aurora-borealis": {
						Steps: 8,
						Hi:    &prism.SemanticColour{ANSI16: "cyan", TrueColor: "#00E5FF"},
						Lo:    &prism.SemanticColour{ANSI16: "magenta", TrueColor: "#B388FF"},
					},
				},
				Components: map[string]string{
					prism.GradientComponentActivity:  "aurora-borealis",
					prism.GradientComponentPeriscope: "aurora-borealis",
				},
			}

			palette.Highlights = highlights
			w = &bytes.Buffer{}

			theme, err = prism.NewTheme(palette, w)
			Expect(err).To(BeNil())

			pgGrad, has := theme.GradientFor(prism.GradientComponentPeriscope)
			Expect(has).To(BeTrue())
			Expect(pgGrad.Steps).To(Equal(8))
			Expect(pgGrad.Hi).NotTo(BeNil())
			Expect(pgGrad.Lo).NotTo(BeNil())
		})

		It("returns false when periscope component is not configured", func() {
			highlights := prism.HighlightsConfig{
				Gradients: map[string]prism.GradientDef{
					"aurora-borealis": {
						Steps: 8,
						Hi:    &prism.SemanticColour{ANSI16: "cyan"},
						Lo:    &prism.SemanticColour{ANSI16: "magenta"},
					},
				},
				Components: map[string]string{
					prism.GradientComponentActivity: "aurora-borealis",
				},
			}

			palette.Highlights = highlights
			w.Reset()

			theme, err = prism.NewTheme(palette, w)
			Expect(err).To(BeNil())

			_, has := theme.GradientFor(prism.GradientComponentPeriscope)
			Expect(has).To(BeFalse())
		})

		It("independent from activity-control gradient", func() {
			highlights := prism.HighlightsConfig{
				Gradients: map[string]prism.GradientDef{
					"aurora-borealis": {
						Steps: 8,
						Hi:    &prism.SemanticColour{ANSI16: "cyan", TrueColor: "#00E5FF"},
						Lo:    &prism.SemanticColour{ANSI16: "magenta", TrueColor: "#B388FF"},
					},
					"ember-glow": {
						Hi: &prism.SemanticColour{ANSI16: "red", TrueColor: "#FF6F00"},
						Lo: &prism.SemanticColour{ANSI16: "yellow", TrueColor: "#FFD54F"},
					},
				},
				Components: map[string]string{
					prism.GradientComponentActivity:  "aurora-borealis",
					prism.GradientComponentPeriscope: "ember-glow",
				},
			}

			palette.Highlights = highlights
			w.Reset()

			theme, err = prism.NewTheme(palette, w)
			Expect(err).To(BeNil())

			aGrad, hasActivity := theme.GradientFor(prism.GradientComponentActivity)
			Expect(hasActivity).To(BeTrue())
			Expect(aGrad.Hi).NotTo(BeNil())

			pGrad, hasPeriscope := theme.GradientFor(prism.GradientComponentPeriscope)
			Expect(hasPeriscope).To(BeTrue())
			Expect(pGrad.Hi).NotTo(BeNil())

			Expect(aGrad.Hi).NotTo(Equal(pGrad.Hi))
		})
	})
})
