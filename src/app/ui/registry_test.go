package ui_test

import (
	"io"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/app/ui"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

var _ = Describe("Registry", func() {

	// ------------------------------------------------------------------
	// New - takes the polymorphic ViewConfig returned by LoadConfig
	// ------------------------------------------------------------------

	Describe("New", func() {
		DescribeTable("returns a Presenter for known modes",
			func(mode string) {
				palette := contract.SystemPalette()

				cfg, err := ui.LoadConfig(mode, nil, palette)
				Expect(err).To(BeNil())

				presenter, err := ui.New(mode, palette, cfg)

				Expect(err).To(BeNil())
				Expect(presenter).NotTo(BeNil())
			},
			Entry("explicit linear mode", ui.ModeLinear),
			Entry("empty string defaults to linear", ""),
		)

		Context("when the mode is not registered", func() {
			It("returns an error containing the unknown mode name", func() {
				palette := contract.SystemPalette()

				_, err := ui.New("nonexistent-mode", palette, ui.LinearConfig{})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("nonexistent-mode"))
			})
		})

		Context("when highway is requested with the wrong config type", func() {
			It("returns a type-mismatch error", func() {
				palette := contract.SystemPalette()

				_, err := ui.New(ui.ModeHighway, palette, ui.LinearConfig{})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("HighwayConfig"))
			})
		})

		Context("when the palette contains an invalid ansi16 name", func() {
			It("returns an error propagated from prism", func() {
				palette := contract.SystemPalette()
				palette.Directory = contract.SemanticColour{ANSI16: "notacolour"}

				cfg, err := ui.LoadConfig(ui.ModeLinear, nil, palette)
				Expect(err).To(BeNil())

				_, err = ui.New(ui.ModeLinear, palette, cfg)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("notacolour"))
			})
		})

		Context("when the palette contains custom tree icons", func() {
			It("renders the custom icons via the selected view", func() {
				palette := contract.SystemPalette()
				palette.TreeIcons = map[string]string{
					contract.TreeIconRoot:           "*",
					contract.TreeIconDirectory:      "D",
					contract.TreeIconFile:           "F",
					contract.TreeIconElapsed:        "E",
					contract.TreeIconSkipped:        "S",
					contract.TreeIconBranchVertical: "|",
					contract.TreeIconBranchJoint:    "+-- ",
					contract.TreeIconBranchLast:     "L-- ",
					contract.TreeIconBranchIndent:   "  ",
				}

				origStdout := os.Stdout
				r, w, err := os.Pipe()
				Expect(err).To(BeNil())
				os.Stdout = w

				cfg, err := ui.LoadConfig(ui.ModeLinear, nil, palette)
				Expect(err).To(BeNil())

				presenter, err := ui.New(ui.ModeLinear, palette, cfg)
				Expect(err).To(BeNil())
				Expect(presenter).NotTo(BeNil())

				node := &core.Node{
					Path: "./test/file.txt",
					Extension: core.Extension{
						Depth: 1,
						Name:  "file.txt",
					},
				}

				presenter.OnNodeEvent(&report.NeutralEvent{
					DisplayEvent: report.DisplayEvent{
						Node:   node,
						IsLast: true,
					},
				})

				Expect(w.Close()).To(Succeed())
				os.Stdout = origStdout

				output, err := io.ReadAll(r)
				Expect(err).To(BeNil())
				Expect(string(output)).To(ContainSubstring("L-- F file.txt"))
			})
		})
	})

	// ------------------------------------------------------------------
	// LoadConfig - returns the polymorphic ViewConfig for the mode
	// ------------------------------------------------------------------

	Describe("LoadConfig", func() {
		Context("given: linear mode", func() {
			It("returns a LinearConfig without touching the source", func() {
				palette := contract.SystemPalette()
				cfg, err := ui.LoadConfig(ui.ModeLinear, nil, palette)
				Expect(err).To(BeNil())
				_, ok := cfg.(ui.LinearConfig)
				Expect(ok).To(BeTrue())
			})
		})

		Context("given: highway mode", func() {
			It("returns a HighwayConfig", func() {
				palette := contract.SystemPalette()
				cfg, err := ui.LoadConfig(ui.ModeHighway, nil, palette)
				Expect(err).To(BeNil())
				_, ok := cfg.(ui.HighwayConfig)
				Expect(ok).To(BeTrue())
			})
		})

		Context("given: empty mode", func() {
			It("defaults to linear", func() {
				palette := contract.SystemPalette()
				cfg, err := ui.LoadConfig("", nil, palette)
				Expect(err).To(BeNil())
				_, ok := cfg.(ui.LinearConfig)
				Expect(ok).To(BeTrue())
			})
		})

		Context("given: unknown mode", func() {
			It("returns an error naming the bad mode", func() {
				palette := contract.SystemPalette()
				_, err := ui.LoadConfig("orbiter", nil, palette)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("orbiter"))
			})
		})
	})

	// ------------------------------------------------------------------
	// Sealed ViewConfig - the marker method prevents external
	// implementations.
	// ------------------------------------------------------------------

	Describe("ViewConfig marker", func() {
		It("linear and highway configs both satisfy ViewConfig", func() {
			var _ ui.ViewConfig = ui.LinearConfig{}
			var _ ui.ViewConfig = ui.HighwayConfig{}
		})
	})
})
