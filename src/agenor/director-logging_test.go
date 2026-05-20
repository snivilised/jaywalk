package agenor_test

import (
	"bytes"
	"context"
	"log/slog"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/agenor"
	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/agenor/pref"
	"github.com/snivilised/jaywalk/src/agenor/test/hanno"
	"github.com/snivilised/jaywalk/src/agenor/tfs"
	"github.com/snivilised/jaywalk/src/internal/services"
	"github.com/snivilised/jaywalk/src/locale"
	lab "github.com/snivilised/jaywalk/test/laboratory"
	"github.com/snivilised/li18ngo"
	"github.com/snivilised/nefilim/test/luna"
)

var _ = Describe("Navigation logging", Ordered, func() {
	var (
		fS   *luna.MemFS
		tree string
	)

	BeforeAll(func() {
		Expect(li18ngo.Register(
			func(o *li18ngo.UseOptions) {
				o.From.Sources = li18ngo.TranslationFiles{
					locale.SourceID: li18ngo.TranslationSource{Name: "agenor"},
				}
			},
		)).To(Succeed())

		fS = hanno.Nuxx(false, lab.Static.RetroWave)
		tree = lab.Static.RetroWave
	})

	BeforeEach(func() {
		services.Reset()
	})

	It("🧪 should: log navigation started with root tree path", func(specCtx SpecContext) {
		lab.WithTestContext(specCtx, func(ctx context.Context, _ context.CancelFunc) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, nil))

			_, err := agenor.Walk().Configure().Extent(agenor.Prime(
				&pref.Using{
					Subscription: enums.SubscribeFiles,
					Head: pref.Head{
						Handler: noOpHandler,
						GetForest: func(_ string) *core.Forest {
							return &core.Forest{
								T: fS,
								R: tfs.New(),
							}
						},
					},
					Tree: tree,
				},
				agenor.WithLogger(logger),
				agenor.WithFaultHandler(agenor.Accepter(lab.IgnoreFault)),
			)).Navigate(ctx)

			Expect(err).To(Succeed())
			output := buf.String()
			Expect(output).To(ContainSubstring("navigation started"))
			Expect(output).To(ContainSubstring(tree))
		})
	})

	It("🧪 should: log navigation completed with file and dir counts", func(specCtx SpecContext) {
		lab.WithTestContext(specCtx, func(ctx context.Context, _ context.CancelFunc) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, nil))

			result, err := agenor.Walk().Configure().Extent(agenor.Prime(
				&pref.Using{
					Subscription: enums.SubscribeUniversal,
					Head: pref.Head{
						Handler: noOpHandler,
						GetForest: func(_ string) *core.Forest {
							return &core.Forest{
								T: fS,
								R: tfs.New(),
							}
						},
					},
					Tree: tree,
				},
				agenor.WithLogger(logger),
				agenor.WithFaultHandler(agenor.Accepter(lab.IgnoreFault)),
			)).Navigate(ctx)

			Expect(err).To(Succeed())
			output := buf.String()
			Expect(output).To(ContainSubstring("navigation completed"))
			Expect(output).To(ContainSubstring("files"))
			Expect(output).To(ContainSubstring("directories"))
			Expect(result).NotTo(BeNil())
		})
	})

	Context("error path", func() {
		It("🧪 should: log error, not completion, when navigation fails", func(specCtx SpecContext) {
			lab.WithTestContext(specCtx, func(ctx context.Context, _ context.CancelFunc) {
				var buf bytes.Buffer
				logger := slog.New(slog.NewTextHandler(&buf, nil))

				_, err := agenor.Walk().Configure().Extent(agenor.Prime(
					&pref.Using{
						Subscription: enums.SubscribeFiles,
						Head: pref.Head{
							Handler: noOpHandler,
						},
						Tree: "non-existent-path",
					},
					agenor.WithLogger(logger),
				)).Navigate(ctx)

				Expect(err).NotTo(Succeed())
				output := buf.String()
				Expect(output).To(ContainSubstring("ERROR"))
			})
		})
	})
})
