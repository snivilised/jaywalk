package bedrock_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/app/bedrock"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/nefilim/test/luna"
)

var _ = Describe("ThemeWriter", func() {
	const themesDir = "themes"

	var (
		fS     *luna.MemFS
		writer *bedrock.ThemeWriter
	)

	BeforeEach(func() {
		fS = luna.NewMemFS()
		writer = bedrock.NewThemeWriterWithFS(themesDir, fS)
	})

	Describe("Write", func() {
		It("creates the file under themesDir with .yaml extension", func() {
			palette := contract.SystemPalette()
			err := writer.Write("my-theme", palette)
			Expect(err).NotTo(HaveOccurred())

			_, err = fS.Stat(themesDir + "/my-theme.yaml")
			Expect(err).NotTo(HaveOccurred())
		})

		It("wraps the palette under the top-level palette: key", func() {
			palette := contract.SystemPalette()
			palette.Directory = contract.SemanticColour{
				ANSI16:    "cyan",
				ANSI256:   "116",
				TrueColor: "#89DCEB",
			}

			err := writer.Write("check-envelope", palette)
			Expect(err).NotTo(HaveOccurred())

			data, err := fS.ReadFile(themesDir + "/check-envelope.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("palette:"))
			Expect(string(data)).To(ContainSubstring("directory:"))
			Expect(string(data)).To(ContainSubstring("ansi16: cyan"))
		})

		Describe("round-trip", func() {
			It("written file is loadable by ThemeLoader and produces an equal palette", func() {
				palette := contract.Palette{
					Directory: contract.SemanticColour{
						ANSI16:    "cyan",
						ANSI256:   "116",
						TrueColor: "#89DCEB",
					},
					Error: contract.SemanticColour{
						ANSI16:    "red",
						ANSI256:   "211",
						TrueColor: "#F38BA8",
					},
					TreeIcons: contract.TreeIcons{
						contract.TreeIconRoot:      "✻",
						contract.TreeIconDirectory: "📁",
						contract.TreeIconFile:      "🔖",
					},
					Highlights: contract.HighlightsConfig{
						Gradients: map[string]contract.GradientDef{
							"aurora-borealis": {
								Steps: 8,
								Hi: &contract.SemanticColour{
									ANSI16:    "cyan",
									TrueColor: "#00E5FF",
								},
								Lo: &contract.SemanticColour{
									ANSI16:    "magenta",
									TrueColor: "#B388FF",
								},
								Animate: boolPtr(true),
							},
						},
						Components: map[string]string{
							"banner-control": "aurora-borealis",
						},
					},
				}

				err := writer.Write("round-trip", palette)
				Expect(err).NotTo(HaveOccurred())

				loader := bedrock.NewThemeLoaderWithFS(themesDir, fS)
				loaded, err := loader.Load("round-trip")
				Expect(err).NotTo(HaveOccurred())

				Expect(loaded.Directory.ANSI16).To(Equal(palette.Directory.ANSI16))
				Expect(loaded.Directory.ANSI256).To(Equal(palette.Directory.ANSI256))
				Expect(loaded.Directory.TrueColor).To(Equal(palette.Directory.TrueColor))
				Expect(loaded.Error.ANSI16).To(Equal(palette.Error.ANSI16))
				Expect(loaded.Error.ANSI256).To(Equal(palette.Error.ANSI256))
				Expect(loaded.Error.TrueColor).To(Equal(palette.Error.TrueColor))
				Expect(loaded.TreeIcons[contract.TreeIconRoot]).To(Equal("✻"))
				Expect(loaded.TreeIcons[contract.TreeIconDirectory]).To(Equal("📁"))
				Expect(loaded.TreeIcons[contract.TreeIconFile]).To(Equal("🔖"))

				gd := loaded.Highlights.Gradients["aurora-borealis"]
				Expect(gd.Steps).To(Equal(8))
				Expect(gd.Hi).NotTo(BeNil())
				Expect(gd.Hi.ANSI16).To(Equal("cyan"))
				Expect(gd.Hi.TrueColor).To(Equal("#00E5FF"))
				Expect(gd.Lo).NotTo(BeNil())
				Expect(gd.Lo.ANSI16).To(Equal("magenta"))
				Expect(gd.Lo.TrueColor).To(Equal("#B388FF"))
				Expect(gd.Animate).NotTo(BeNil())
				Expect(*gd.Animate).To(BeTrue())
				Expect(loaded.Highlights.Components["banner-control"]).To(Equal("aurora-borealis"))
			})
		})

		Describe("safe write", func() {
			It("leaves no .tmp file on success", func() {
				palette := contract.SystemPalette()
				err := writer.Write("no-tmp", palette)
				Expect(err).NotTo(HaveOccurred())

				entries, err := fS.ReadDir(themesDir)
				Expect(err).NotTo(HaveOccurred())
				for _, entry := range entries {
					Expect(entry.Name()).NotTo(HavePrefix(".tmp-"))
				}
			})

			It("overwrites an existing file without error", func() {
				palette := contract.SystemPalette()
				err := writer.Write("overwrite", palette)
				Expect(err).NotTo(HaveOccurred())

				palette.Directory = contract.SemanticColour{ANSI16: "red"}
				err = writer.Write("overwrite", palette)
				Expect(err).NotTo(HaveOccurred())

				loader := bedrock.NewThemeLoaderWithFS(themesDir, fS)
				loaded, err := loader.Load("overwrite")
				Expect(err).NotTo(HaveOccurred())
				Expect(loaded.Directory.ANSI16).To(Equal("red"))
			})
		})

		It("round-trips all semantic colour roles", func() {
			palette := contract.Palette{
				Directory:    contract.SemanticColour{TrueColor: "#111111"},
				File:         contract.SemanticColour{TrueColor: "#222222"},
				Root:         contract.SemanticColour{TrueColor: "#333333"},
				Branch:       contract.SemanticColour{TrueColor: "#444444"},
				Action:       contract.SemanticColour{TrueColor: "#555555"},
				Pipeline:     contract.SemanticColour{TrueColor: "#666666"},
				LandingStrip: contract.SemanticColour{TrueColor: "#777777"},
				Skipped:      contract.SemanticColour{TrueColor: "#888888"},
				Error:        contract.SemanticColour{TrueColor: "#999999"},
				Muted:        contract.SemanticColour{TrueColor: "#AAAAAA"},
				Progress:     contract.SemanticColour{TrueColor: "#BBBBBB"},
				BoxBorder:    contract.SemanticColour{TrueColor: "#CCCCCC"},
				SummaryLabel: contract.SemanticColour{TrueColor: "#DDDDDD"},
				SummaryValue: contract.SemanticColour{TrueColor: "#EEEEEE"},
				Worker:       contract.SemanticColour{TrueColor: "#FF0000"},
				WorkerIdle:   contract.SemanticColour{TrueColor: "#00FF00"},
				LaneHeader:   contract.SemanticColour{TrueColor: "#0000FF"},
				Header:       contract.SemanticColour{TrueColor: "#FFFF00"},
				Frame:        contract.SemanticColour{TrueColor: "#FF00FF"},
				Border:       contract.SemanticColour{TrueColor: "#00FFFF"},
				BarFilled:    contract.SemanticColour{TrueColor: "#AA0000"},
				BarEmpty:     contract.SemanticColour{TrueColor: "#00AA00"},
			}

			err := writer.Write("all-roles", palette)
			Expect(err).NotTo(HaveOccurred())

			loader := bedrock.NewThemeLoaderWithFS(themesDir, fS)
			loaded, err := loader.Load("all-roles")
			Expect(err).NotTo(HaveOccurred())

			Expect(loaded.Directory.TrueColor).To(Equal("#111111"))
			Expect(loaded.File.TrueColor).To(Equal("#222222"))
			Expect(loaded.Root.TrueColor).To(Equal("#333333"))
			Expect(loaded.Branch.TrueColor).To(Equal("#444444"))
			Expect(loaded.Action.TrueColor).To(Equal("#555555"))
			Expect(loaded.Pipeline.TrueColor).To(Equal("#666666"))
			Expect(loaded.LandingStrip.TrueColor).To(Equal("#777777"))
			Expect(loaded.Skipped.TrueColor).To(Equal("#888888"))
			Expect(loaded.Error.TrueColor).To(Equal("#999999"))
			Expect(loaded.Muted.TrueColor).To(Equal("#AAAAAA"))
			Expect(loaded.Progress.TrueColor).To(Equal("#BBBBBB"))
			Expect(loaded.BoxBorder.TrueColor).To(Equal("#CCCCCC"))
			Expect(loaded.SummaryLabel.TrueColor).To(Equal("#DDDDDD"))
			Expect(loaded.SummaryValue.TrueColor).To(Equal("#EEEEEE"))
			Expect(loaded.Worker.TrueColor).To(Equal("#FF0000"))
			Expect(loaded.WorkerIdle.TrueColor).To(Equal("#00FF00"))
			Expect(loaded.LaneHeader.TrueColor).To(Equal("#0000FF"))
			Expect(loaded.Header.TrueColor).To(Equal("#FFFF00"))
			Expect(loaded.Frame.TrueColor).To(Equal("#FF00FF"))
			Expect(loaded.Border.TrueColor).To(Equal("#00FFFF"))
			Expect(loaded.BarFilled.TrueColor).To(Equal("#AA0000"))
			Expect(loaded.BarEmpty.TrueColor).To(Equal("#00AA00"))
		})
	})
})

func boolPtr(b bool) *bool {
	return &b
}
