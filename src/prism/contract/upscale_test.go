package contract_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

var _ = Describe("UpscalePalette", func() {
	DescribeTable("SemanticColour rules",
		func(input, expected contract.SemanticColour) {
			result := contract.UpscalePalette(contract.Palette{Directory: input})
			Expect(result.Directory).To(Equal(expected))
		},
		Entry("true-color only: derives ansi256 and ansi16",
			contract.SemanticColour{TrueColor: "#CC0000"},
			contract.SemanticColour{TrueColor: "#CC0000", ANSI256: "1", ANSI16: "red"}),
		Entry("true-color + ansi256: derives ansi16 only",
			contract.SemanticColour{TrueColor: "#CC0000", ANSI256: "196"},
			contract.SemanticColour{TrueColor: "#CC0000", ANSI256: "196", ANSI16: "red"}),
		Entry("true-color + ansi16: derives ansi256 only",
			contract.SemanticColour{TrueColor: "#CC0000", ANSI16: "red"},
			contract.SemanticColour{TrueColor: "#CC0000", ANSI256: "1", ANSI16: "red"}),
		Entry("ansi16 only: promotes to hex, derives ansi256",
			contract.SemanticColour{ANSI16: "cyan"},
			contract.SemanticColour{TrueColor: "#06989A", ANSI256: "30", ANSI16: "cyan"}),
		Entry("ansi256 only: promotes to hex, derives ansi16",
			contract.SemanticColour{ANSI256: "196"},
			contract.SemanticColour{TrueColor: "#FF0000", ANSI256: "196", ANSI16: "bright-red"}),
		Entry("ansi16 + ansi256: promotes true-color from ansi16",
			contract.SemanticColour{ANSI16: "red", ANSI256: "196"},
			contract.SemanticColour{TrueColor: "#CC0000", ANSI256: "196", ANSI16: "red"}),
		Entry("all three present: no change",
			contract.SemanticColour{TrueColor: "#CC0000", ANSI256: "196", ANSI16: "red"},
			contract.SemanticColour{TrueColor: "#CC0000", ANSI256: "196", ANSI16: "red"}),
		Entry("none present: no change",
			contract.SemanticColour{},
			contract.SemanticColour{}),
	)

	Context("gradient endpoints", func() {
		It("upscales hi and lo on every gradient definition", func() {
			p := contract.Palette{
				Directory: contract.SemanticColour{ANSI16: "cyan"},
				Highlights: contract.HighlightsConfig{
					Gradients: map[string]contract.GradientDef{
						"test-gradient": {
							Steps: 8,
							Hi:    &contract.SemanticColour{TrueColor: "#00E5FF"},
							Lo:    &contract.SemanticColour{ANSI16: "magenta"},
						},
					},
				},
			}

			result := contract.UpscalePalette(p)

			gd := result.Highlights.Gradients["test-gradient"]
			Expect(gd.Hi).NotTo(BeNil())
			Expect(gd.Hi.TrueColor).To(Equal("#00E5FF"))
			Expect(gd.Hi.ANSI256).NotTo(BeEmpty())
			Expect(gd.Hi.ANSI16).NotTo(BeEmpty())

			Expect(gd.Lo).NotTo(BeNil())
			Expect(gd.Lo.TrueColor).NotTo(BeEmpty())
			Expect(gd.Lo.ANSI256).NotTo(BeEmpty())
			Expect(gd.Lo.ANSI16).To(Equal("magenta"))
		})
	})

	Context("immutability", func() {
		It("does not mutate the input palette", func() {
			original := contract.Palette{
				Directory: contract.SemanticColour{ANSI16: "cyan"},
				Error:     contract.SemanticColour{TrueColor: "#CC0000"},
			}
			originalCopy := original

			_ = contract.UpscalePalette(original)

			Expect(original.Directory).To(Equal(originalCopy.Directory))
			Expect(original.Error).To(Equal(originalCopy.Error))
		})
	})
})
