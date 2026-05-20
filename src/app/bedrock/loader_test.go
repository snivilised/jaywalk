package bedrock_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/app/bedrock"
	"github.com/snivilised/jaywalk/src/locale"
	"github.com/snivilised/li18ngo"
)

var _ = Describe("Load", Ordered, func() {
	BeforeAll(func() {
		Expect(li18ngo.Use(
			func(o *li18ngo.UseOptions) {
				o.From.Sources = li18ngo.TranslationFiles{
					locale.SourceID: li18ngo.TranslationSource{Name: "agenor"},
				}
			},
		)).To(Succeed())
	})

	// -----------------------------------------------------------------------
	// Happy path
	// -----------------------------------------------------------------------
	Context("given a complete, valid YAML config", func() {
		var (
			config *bedrock.Config
			err    error
		)

		BeforeEach(func() {
			config, err = bedrock.Load(bedrock.LoadOptions{
				ViperInstance: viperFromYAML(fullYAML),
			})
		})

		It("succeeds without error", func() {
			Expect(err).To(BeNil())
			Expect(config).NotTo(BeNil())
		})

		// ── Mapped: Interaction ──────────────────────────────────────────
		Describe("mapped interaction section", func() {
			It("decodes per-item-delay as a Duration", func() {
				Expect(config.Mapped.Interaction.TUI.PerItemDelay).To(Equal(1 * time.Second))
			})
		})

		// ── Mapped: Advanced ────────────────────────────────────────────
		Describe("mapped advanced section", func() {
			It("decodes abort-on-error", func() {
				Expect(config.Mapped.Advanced.AbortOnError).To(BeFalse())
			})
			It("decodes overwrite-on-collision", func() {
				Expect(config.Mapped.Advanced.OverwriteOnCollision).To(BeFalse())
			})
			It("decodes extensions.suffixes-csv", func() {
				Expect(config.Mapped.Advanced.Extensions.SuffixesCSV).To(Equal("jpg,jpeg,png"))
			})
			It("decodes extensions.transforms-csv", func() {
				Expect(config.Mapped.Advanced.Extensions.TransformsCSV).To(Equal("lower"))
			})
			It("decodes extensions.map", func() {
				Expect(config.Mapped.Advanced.Extensions.Map).To(HaveKeyWithValue("jpeg", "jpg"))
			})
		})

		// ── Mapped: Logging ──────────────────────────────────────────────
		Describe("mapped logging section", func() {
			It("decodes max-size-in-mb", func() {
				Expect(config.Mapped.Logging.MaxSizeInMB).To(Equal(10))
			})
			It("decodes max-backups", func() {
				Expect(config.Mapped.Logging.MaxBackups).To(Equal(3))
			})
			It("decodes max-age-in-days", func() {
				Expect(config.Mapped.Logging.MaxAgeInDays).To(Equal(30))
			})
			It("decodes level", func() {
				Expect(config.Mapped.Logging.Level).To(Equal("info"))
			})
			It("decodes time-format", func() {
				Expect(config.Mapped.Logging.TimeFormat).To(Equal("2006-01-02 15:04:05"))
			})
		})

		// ── Raw: Actions ─────────────────────────────────────────────────
		Describe("raw actions section", func() {
			It("contains all four actions", func() {
				Expect(config.Raw.Actions).To(HaveLen(4))
			})
			It("preserves komp-2 cmd template", func() {
				Expect(config.Raw.Actions["komp-2"].Cmd).To(
					ContainSubstring("ffmpeg"))
			})
			It("preserves komp-2 when expression", func() {
				Expect(config.Raw.Actions["komp-2"].When).To(
					Equal("isVideo && isLarge && !isHidden"))
			})
		})

		// ── Raw: Pipelines ───────────────────────────────────────────────
		Describe("raw pipelines section", func() {
			It("contains two pipelines", func() {
				Expect(config.Raw.Pipelines).To(HaveLen(2))
			})
			It("video-workflow has three steps", func() {
				Expect(config.Raw.Pipelines["video-workflow"].Steps).To(HaveLen(3))
			})
			It("quick-transcode steps are correct", func() {
				Expect(config.Raw.Pipelines["quick-transcode"].Steps).To(
					ConsistOf("komp-18", "upload-s3"))
			})
		})

		// ── Raw: Flags ───────────────────────────────────────────────────
		Describe("raw flags section", func() {
			It("contains short override for walk.foo", func() {
				Expect(config.Raw.Flags.Short["walk"]["foo"]).To(Equal("F"))
			})
			It("contains short override for sprint.bar", func() {
				Expect(config.Raw.Flags.Short["sprint"]["bar"]).To(Equal("Z"))
			})
			It("contains invoke.any.files default", func() {
				val, ok := config.Raw.Flags.Invoke["any"]["files"]
				Expect(ok).To(BeTrue())
				Expect(val).To(BeNumerically("==", 2))
			})
			It("contains component.sampler.folders default", func() {
				val, ok := config.Raw.Flags.Component["sampler"]["folders"]
				Expect(ok).To(BeTrue())
				Expect(val).To(BeNumerically("==", 1))
			})
		})
	})

	// -----------------------------------------------------------------------
	// Minimal config
	// -----------------------------------------------------------------------
	Context("given a minimal valid YAML config", func() {
		It("loads successfully with zero-value mapped sections", func() {
			config, err := bedrock.Load(bedrock.LoadOptions{
				ViperInstance: viperFromYAML(minimalYAML),
			})
			Expect(err).To(BeNil())
			Expect(config).NotTo(BeNil())
			Expect(config.Mapped.Advanced.AbortOnError).To(BeFalse())
			Expect(config.Raw.Actions).To(BeEmpty())
		})
	})
})
