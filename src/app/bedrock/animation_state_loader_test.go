package bedrock_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/app/bedrock"
	"github.com/snivilised/jaywalk/src/locale"
	"github.com/snivilised/li18ngo"
)

var _ = Describe("LoadAnimationData", Ordered, func() {
	BeforeAll(func() {
		Expect(li18ngo.Use(
			func(o *li18ngo.UseOptions) {
				o.From.Sources = li18ngo.TranslationFiles{
					locale.SourceID: li18ngo.TranslationSource{Name: "agenor"},
				}
			},
		)).To(Succeed())
	})

	It("loads all defaults when enabled is empty", func() {
		cfg := &bedrock.HighwayConfig{}
		state, err := bedrock.LoadAnimationData(cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.Data).NotTo(BeEmpty())
		Expect(len(state.Data)).To(BeNumerically(">=", 48))
	})

	It("loads specified spinners from config", func() {
		cfg := &bedrock.HighwayConfig{}
		cfg.AnimationData.Spinners.Enabled = []string{"braille", "wave"}
		state, err := bedrock.LoadAnimationData(cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.Data).To(HaveLen(2))
		Expect(state.Data["braille"]).NotTo(BeNil())
		Expect(state.Data["wave"]).NotTo(BeNil())
	})

	It("expands category to member spinners", func() {
		cfg := &bedrock.HighwayConfig{}
		cfg.AnimationData.Spinners.Enabled = []string{"film-strip-set"}
		state, err := bedrock.LoadAnimationData(cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(state.Data)).To(BeNumerically(">=", 11))
		Expect(state.Data["wave"]).NotTo(BeNil())
		Expect(state.Data["fairlight"]).NotTo(BeNil())
		Expect(state.Data["barcode"]).NotTo(BeNil())
	})

	It("applies override interval to individual spinner", func() {
		cfg := &bedrock.HighwayConfig{}
		cfg.AnimationData.Spinners.Enabled = []string{"wave"}
		cfg.AnimationData.Spinners.Override = map[string]*bedrock.SpinnerItemConfig{
			"wave": {Interval: 250},
		}
		state, err := bedrock.LoadAnimationData(cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.Data["wave"].Interval).To(Equal(250))
	})

	It("uses default interval when no override exists", func() {
		cfg := &bedrock.HighwayConfig{}
		cfg.AnimationData.Spinners.Enabled = []string{"braille"}
		state, err := bedrock.LoadAnimationData(cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.Data["braille"].Interval).To(Equal(80))
	})
})
