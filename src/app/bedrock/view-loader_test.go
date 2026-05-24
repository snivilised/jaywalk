package bedrock_test

import (
	"testing/fstest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/app/bedrock"
	"github.com/snivilised/jaywalk/src/locale"
	"github.com/snivilised/li18ngo"
	"github.com/snivilised/nefilim/test/luna"
)

var _ = Describe("ViewConfigLoader", Ordered, func() {
	BeforeAll(func() {
		Expect(li18ngo.Use(
			func(o *li18ngo.UseOptions) {
				o.From.Sources = li18ngo.TranslationFiles{
					locale.SourceID: li18ngo.TranslationSource{Name: "agenor"},
				}
			},
		)).To(Succeed())
	})

	const configHome = "root/config"

	Describe("Load", func() {
		It("loads highway config section successfully", func() {
			fS := luna.NewMemFS()
			fS.MapFS[configHome+"/jay.ui.yaml"] = &fstest.MapFile{
			Data: []byte(`
ui:
  highway:
    emoji-pool: '😎 😄 👽 🤖 🦊 🐯 👻 🧙 🦄 🦁'
    separator: ' '
    animation:
      spinners:
        enabled: ['wave', 'braille']
        override:
          wave:
            interval: 200
`),
		}

		loader := bedrock.NewViewConfigLoaderWithFS(configHome, fS)
		var cfg bedrock.HighwayConfig
		err := loader.Load("highway", &cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Pool).To(Equal("😎 😄 👽 🤖 🦊 🐯 👻 🧙 🦄 🦁"))
		Expect(cfg.Separator).To(Equal(" "))
		Expect(cfg.AnimationData.Spinners.Enabled).To(ConsistOf("wave", "braille"))
		Expect(cfg.AnimationData.Spinners.Override).NotTo(BeNil())
		Expect(cfg.AnimationData.Spinners.Override["wave"]).NotTo(BeNil())
		Expect(cfg.AnimationData.Spinners.Override["wave"].Interval).To(Equal(200))
		})

		It("returns nil when config file does not exist", func() {
			fS := luna.NewMemFS()
			loader := bedrock.NewViewConfigLoaderWithFS(configHome, fS)
			var cfg bedrock.HighwayConfig
			err := loader.Load("highway", &cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Pool).To(BeEmpty())
		})

		It("returns nil when view section is absent", func() {
			fS := luna.NewMemFS()
			fS.MapFS[configHome+"/jay.ui.yaml"] = &fstest.MapFile{
				Data: []byte(`
ui:
  other-view:
    setting: value
`),
			}

			loader := bedrock.NewViewConfigLoaderWithFS(configHome, fS)
			var cfg bedrock.HighwayConfig
			err := loader.Load("highway", &cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Pool).To(BeEmpty())
		})

		It("loads empty animation types gracefully", func() {
			fS := luna.NewMemFS()
			fS.MapFS[configHome+"/jay.ui.yaml"] = &fstest.MapFile{
				Data: []byte(`
ui:
  highway:
    emoji-pool: '😎 😄 👽 🤖 🦊 🐯 👻 🧙 🦄 🦁'
`),
			}

			loader := bedrock.NewViewConfigLoaderWithFS(configHome, fS)
			var cfg bedrock.HighwayConfig
			err := loader.Load("highway", &cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(cfg.AnimationData.Spinners.Enabled)).To(Equal(0))
		})

		It("loads from .yml extension as fallback", func() {
			fS := luna.NewMemFS()
			fS.MapFS[configHome+"/jay.ui.yml"] = &fstest.MapFile{
				Data: []byte(`
ui:
  highway:
    emoji-pool: '😎 😄 👽 🤖 🦊 🐯 👻 🧙 🦄 🦁'
`),
			}

			loader := bedrock.NewViewConfigLoaderWithFS(configHome, fS)
			var cfg bedrock.HighwayConfig
			err := loader.Load("highway", &cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Pool).To(Equal("😎 😄 👽 🤖 🦊 🐯 👻 🧙 🦄 🦁"))
		})

		It("prefers .yaml over .yml", func() {
			fS := luna.NewMemFS()
			fS.MapFS[configHome+"/jay.ui.yaml"] = &fstest.MapFile{
				Data: []byte(`
ui:
  highway:
    emoji-pool: 'from-yaml'
`),
			}
			fS.MapFS[configHome+"/jay.ui.yml"] = &fstest.MapFile{
				Data: []byte(`
ui:
  highway:
    emoji-pool: 'from-yml'
`),
			}

			loader := bedrock.NewViewConfigLoaderWithFS(configHome, fS)
			var cfg bedrock.HighwayConfig
			err := loader.Load("highway", &cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Pool).To(Equal("from-yaml"))
		})

		It("returns error on malformed YAML", func() {
			fS := luna.NewMemFS()
			fS.MapFS[configHome+"/jay.ui.yaml"] = &fstest.MapFile{
				Data: []byte("garbage: [unclosed"),
			}

			loader := bedrock.NewViewConfigLoaderWithFS(configHome, fS)
			var cfg bedrock.HighwayConfig
			err := loader.Load("highway", &cfg)
			Expect(err).To(HaveOccurred())
		})
	})
})
