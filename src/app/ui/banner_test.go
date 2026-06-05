package ui

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/app/bedrock"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

var _ = Describe("resolveBannerConfig", func() {
	When("a steps override is provided", func() {
		It("propagates the override to the resolved config", func() {
			raw := bedrock.BannerSubConfig{
				Position: contract.PositionBottom,
				Tick:     100,
				Steps:    48,
			}

			cfg := resolveBannerConfig(raw, contract.Palette{})

			Expect(cfg.Position).To(Equal(contract.PositionBottom))
			Expect(cfg.Tick).To(Equal(100))
			Expect(cfg.StepsOverride).To(Equal(48))
		})
	})

	When("the steps override is omitted", func() {
		It("leaves StepsOverride as zero so the gradient's own steps are used", func() {
			raw := bedrock.BannerSubConfig{
				Position: contract.PositionTop,
			}

			cfg := resolveBannerConfig(raw, contract.Palette{})

			Expect(cfg.StepsOverride).To(Equal(0))
		})

		It("treats a negative value the same as zero", func() {
			raw := bedrock.BannerSubConfig{Steps: -5}

			cfg := resolveBannerConfig(raw, contract.Palette{})

			Expect(cfg.StepsOverride).To(Equal(0))
		})
	})

	When("the banner is disabled", func() {
		It("still preserves the steps override", func() {
			raw := bedrock.BannerSubConfig{
				Disable: true,
				Steps:   24,
			}

			cfg := resolveBannerConfig(raw, contract.Palette{})

			Expect(cfg.Disable).To(BeTrue())
			Expect(cfg.StepsOverride).To(Equal(24))
		})
	})
})

// bannerGradientPalette returns a Palette with a banner-control
// component bound to a small named gradient. The gradient has 4 steps
// by default so the tests can distinguish "no override" (4 steps) from
// "override applied" (e.g. 48 steps).
func bannerGradientPalette() contract.Palette {
	return contract.Palette{
		Highlights: contract.HighlightsConfig{
			Gradients: map[string]contract.GradientDef{
				"banner-gradient": {
					Steps: 4,
					Hi:    &contract.SemanticColour{ANSI16: "cyan", TrueColor: "#00E5FF"},
					Lo:    &contract.SemanticColour{ANSI16: "magenta", TrueColor: "#B388FF"},
				},
			},
			Components: map[string]string{
				contract.GradientComponentBanner: "banner-gradient",
			},
		},
	}
}

// newBannerPresenter constructs a highwayPresenter with the supplied
// banner config. The palette is wired up so the banner-control
// component resolves to a real gradient.
func newBannerPresenter(bannerCfg BannerConfig) *highwayPresenter {
	theme, err := contract.NewTheme(bannerGradientPalette(), &bytes.Buffer{})
	Expect(err).NotTo(HaveOccurred())

	return &highwayPresenter{
		cfg: HighwayConfig{
			Banner: bannerCfg,
		},
		theme: theme,
	}
}

var _ = Describe("buildBannerInfo - steps override", func() {
	It("uses the gradient's own steps when no override is configured", func() {
		h := newBannerPresenter(BannerConfig{
			Position: contract.PositionTop,
		})

		info := h.buildBannerInfo()

		Expect(info.Gradient).NotTo(BeNil())
		Expect(info.Gradient.Steps).To(Equal(4))
	})

	It("applies the override when the user sets steps in the UI config", func() {
		h := newBannerPresenter(BannerConfig{
			Position:      contract.PositionTop,
			StepsOverride: 48,
		})

		info := h.buildBannerInfo()

		Expect(info.Gradient).NotTo(BeNil())
		Expect(info.Gradient.Steps).To(Equal(48))
	})

	It("uses the override rather than the gradient's Hi/Lo endpoints", func() {
		h := newBannerPresenter(BannerConfig{
			Position:      contract.PositionTop,
			StepsOverride: 16,
		})

		info := h.buildBannerInfo()

		// Steps change but the colours are still inherited from
		// the shared gradient.
		Expect(info.Gradient.Steps).To(Equal(16))
		Expect(info.Gradient.Hi).NotTo(BeNil())
		Expect(info.Gradient.Lo).NotTo(BeNil())
	})

	It("leaves the gradient state at the override value", func() {
		h := newBannerPresenter(BannerConfig{
			Position:      contract.PositionTop,
			StepsOverride: 32,
		})

		info := h.buildBannerInfo()

		Expect(info.State).NotTo(BeNil())
		Expect(info.State.TotalSteps).To(Equal(32))
	})
})
