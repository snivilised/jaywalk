package ui_test

import (
	"io"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/app/bedrock"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/app/ui"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

var _ = Describe("Registry", func() {

	// ------------------------------------------------------------------
	// New - takes *bedrock.FullViewConfig returned by LoadConfig
	// ------------------------------------------------------------------

	Describe("New", func() {
		DescribeTable("returns a Presenter for known modes",
			func(mode string) {
				palette := contract.SystemPalette()

				cfg, err := bedrock.LoadConfig(nil, palette)
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

				_, err := ui.New("nonexistent-mode", palette, &bedrock.FullViewConfig{})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("nonexistent-mode"))
			})
		})

		Context("when the palette contains an invalid ansi16 name", func() {
			It("returns an error propagated from prism", func() {
				palette := contract.SystemPalette()
				palette.Directory = contract.SemanticColour{ANSI16: "notacolour"}

				cfg, err := bedrock.LoadConfig(nil, palette)
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

				cfg, err := bedrock.LoadConfig(nil, palette)
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
	// LoadConfig - returns *bedrock.FullViewConfig
	// ------------------------------------------------------------------

	Describe("LoadConfig", func() {
		Context("with nil source", func() {
			It("returns a FullViewConfig with zero-value sections", func() {
				palette := contract.SystemPalette()
				cfg, err := bedrock.LoadConfig(nil, palette)
				Expect(err).To(BeNil())
				Expect(cfg).NotTo(BeNil())
			})
		})
	})
})
