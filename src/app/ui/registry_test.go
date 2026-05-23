package ui_test

import (
	"io"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/app/ui"
	"github.com/snivilised/jaywalk/src/prism"
)

var _ = Describe("Registry", func() {

	// ------------------------------------------------------------------
	// New
	// ------------------------------------------------------------------

	Describe("New", func() {
		DescribeTable("returns a Presenter for known modes",
			func(mode string) {
				palette := prism.SystemPalette()

				presenter, err := ui.New(mode, palette, ui.HighwayConfig{})

				Expect(err).To(BeNil())
				Expect(presenter).NotTo(BeNil())
			},
			Entry("explicit linear mode", ui.ModeLinear),
			Entry("empty string defaults to linear", ""),
		)

		Context("when the mode is not registered", func() {
			It("returns an error containing the unknown mode name", func() {
				palette := prism.SystemPalette()

				_, err := ui.New("nonexistent-mode", palette, ui.HighwayConfig{})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("nonexistent-mode"))
			})
		})

		Context("when the palette contains an invalid ansi16 name", func() {
			It("returns an error propagated from prism", func() {
				palette := prism.SystemPalette()
				palette.Directory = prism.SemanticColour{ANSI16: "notacolour"}

				_, err := ui.New(ui.ModeLinear, palette, ui.HighwayConfig{})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("notacolour"))
			})
		})

		Context("when the palette contains custom tree icons", func() {
			It("renders the custom icons via the selected view", func() {
				palette := prism.SystemPalette()
				palette.TreeIcons = map[string]string{
					prism.TreeIconRoot:           "*",
					prism.TreeIconDirectory:      "D",
					prism.TreeIconFile:           "F",
					prism.TreeIconElapsed:        "E",
					prism.TreeIconSkipped:        "S",
					prism.TreeIconBranchVertical: "|",
					prism.TreeIconBranchJoint:    "+-- ",
					prism.TreeIconBranchLast:     "L-- ",
					prism.TreeIconBranchIndent:   "  ",
				}

				origStdout := os.Stdout
				r, w, err := os.Pipe()
				Expect(err).To(BeNil())
				os.Stdout = w

				presenter, err := ui.New(ui.ModeLinear, palette, ui.HighwayConfig{})
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
})
