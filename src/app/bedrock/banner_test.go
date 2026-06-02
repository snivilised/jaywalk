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

var _ = Describe("BannerSubConfig loading", Ordered, func() {
	BeforeAll(func() {
		Expect(li18ngo.Use(
			func(o *li18ngo.UseOptions) {
				o.From.Sources = li18ngo.TranslationFiles{
					locale.SourceID: li18ngo.TranslationSource{Name: "agenor"},
				}
			},
		)).To(Succeed())
	})

	const configHome = "root/banner"

	newLoader := func(yaml string) *bedrock.ViewConfigLoader {
		fS := luna.NewMemFS()
		fS.MapFS[configHome+"/jay.ui.yaml"] = &fstest.MapFile{
			Data: []byte(yaml),
		}
		return bedrock.NewViewConfigLoaderWithFS(configHome, fS)
	}

	Describe("steps override", func() {
		It("defaults Steps to zero when the section is absent", func() {
			loader := newLoader(`
ui:
  highway: {}
`)
			var cfg bedrock.HighwayConfig
			Expect(loader.Load("highway", &cfg)).To(Succeed())
			Expect(cfg.Banner.Steps).To(Equal(0))
		})

		It("parses the steps override from jay.ui.yml", func() {
			loader := newLoader(`
ui:
  highway:
    banner:
      position: bottom
      tick: 100
      steps: 48
`)
			var cfg bedrock.HighwayConfig
			Expect(loader.Load("highway", &cfg)).To(Succeed())
			Expect(cfg.Banner.Position).To(Equal("bottom"))
			Expect(cfg.Banner.Tick).To(Equal(100))
			Expect(cfg.Banner.Steps).To(Equal(48))
		})

		It("keeps the other banner settings intact when steps is set", func() {
			loader := newLoader(`
ui:
  highway:
    banner:
      disable: false
      position: top
      tick: 250
      justify: center
      steps: 32
`)
			var cfg bedrock.HighwayConfig
			Expect(loader.Load("highway", &cfg)).To(Succeed())
			Expect(cfg.Banner.Disable).To(BeFalse())
			Expect(cfg.Banner.Position).To(Equal("top"))
			Expect(cfg.Banner.Tick).To(Equal(250))
			Expect(cfg.Banner.Justify).To(Equal("center"))
			Expect(cfg.Banner.Steps).To(Equal(32))
		})
	})
})
