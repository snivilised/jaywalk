package kernel_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/agenor"
	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/agenor/internal/enclave"
	"github.com/snivilised/jaywalk/src/agenor/pref"
	"github.com/snivilised/jaywalk/src/agenor/test/hanno"
	"github.com/snivilised/jaywalk/src/agenor/tfs"
	"github.com/snivilised/jaywalk/src/internal/services"
	"github.com/snivilised/jaywalk/src/locale"
	"github.com/snivilised/jaywalk/src/third/lo"
	lab "github.com/snivilised/jaywalk/test/laboratory"
	"github.com/snivilised/li18ngo"
	"github.com/snivilised/nefilim/test/luna"
	"github.com/snivilised/pants"
)

const (
	ResumeAtTeenageColor = "RETRO-WAVE/College/Teenage Color"
)

func PrimeBuilder(post *lab.AsyncPostage) *agenor.Builders {
	return agenor.Prime(
		&pref.Using{
			Tree:         post.Path,
			Subscription: post.Entry.Subscription,
			Head: pref.Head{
				Handler: post.Entry.Callback,
				GetForest: func(_ string) *core.Forest {
					return &core.Forest{
						T: post.FS,
						R: tfs.New(),
					}
				},
			},
		},
		Settings(post)...,
	)
}

func ResumeBuilder(post *lab.AsyncPostage) *agenor.Builders {
	return agenor.Resume(
		&pref.Relic{
			From:     post.Path,
			Strategy: enums.ResumeStrategyFastward,
			Head: pref.Head{
				Handler: post.Entry.Callback,
				GetForest: func(_ string) *core.Forest {
					return &core.Forest{
						T: post.FS,
						R: tfs.New(),
					}
				},
			},
		},
		Settings(post)...,
	)
}

func Settings(post *lab.AsyncPostage) []pref.Option {
	return []pref.Option{
		agenor.WithOnBegin(lab.Begin("🛡️")),
		agenor.WithOnEnd(lab.End("🏁")),
		agenor.IfOptionF(post.Entry.NoWorkers > 0, func() pref.Option {
			return agenor.WithNoW(post.Entry.NoWorkers)
		}),
		agenor.IfOptionF(post.Entry.CPU, func() pref.Option {
			return agenor.WithCPU()
		}),
		agenor.IfOptionF(post.Entry.Consume, func() pref.Option {
			return agenor.WithOutput(&pref.OutputOptions{
				CheckCloseInterval: time.Second / 10,
				TimeoutOnSend:      time.Second * 3,
				On:                 post.On,
			})
		}),
	}
}

func AsyncNoWCallback(magnitude string, now uint) core.Client {
	return func(servant core.Servant) error {
		node := servant.Node()
		name := node.Extension.Name
		GinkgoWriter.Printf("---> 🌀 ASYNC//%v/%v-CALLBACK(NoW=%v): '%v'\n",
			name, magnitude, node.Path,
			lo.Ternary(now == 0, "CPU", fmt.Sprintf("%v", now)),
		)

		return nil
	}
}

func consumeOk(_ context.Context,
	outs core.OutputStream,
	wg pants.WaitGroup,
) {
	defer wg.Done()

	// We don't need to use a timeout on the observe channel
	// because the navigator invokes Conclude, which results in
	// the observe channel being closed, terminating the range.
	// This aspect is specific to this example and to test
	// cancellation, we'll have to use a select.
	//
	for output := range outs {
		GinkgoWriter.Printf("🍒 payload: '%v', id: '%v', seq: '%v' (e: '%v')\n",
			output.Payload, output.ID, output.SequenceNo, output.Error,
		)
	}

	GinkgoWriter.Println("===> 🍒 finished consuming output channel.")
}

var _ = Describe("Navigator", Ordered, func() {
	var (
		from string
		fS   *luna.MemFS
	)

	BeforeAll(func() {
		Expect(li18ngo.Register(
			func(o *li18ngo.UseOptions) {
				o.From.Sources = li18ngo.TranslationFiles{
					locale.SourceID: li18ngo.TranslationSource{Name: "agenor"},
				}
			},
		)).To(Succeed())

		fS = hanno.Nuxx(verbose, lab.Static.RetroWave)
		from = lab.GetJSONPath()
	})

	BeforeEach(func() {
		services.Reset()
	})

	DescribeTable("sprint",
		func(specCtx SpecContext, entry *lab.AsyncOkTE) {
			lab.WithTestContext(specCtx, func(ctx context.Context, _ context.CancelFunc) {
				var (
					wg sync.WaitGroup
					on core.OutputFunc
				)

				if entry.Consume {
					on = func(outs core.OutputStream) {
						wg.Add(1)

						go consumeOk(ctx, outs, &wg)
					}
				}

				result, err := agenor.Sprint(&wg).Configure(enclave.Loader(func(active *core.ActiveState) {
					GinkgoWriter.Printf("===> 🐚 restoring state: resume at=%v, subscription=%v\n",
						entry.Resume.At, entry.Subscription,
					)

					active.Tree = lab.Static.RetroWave
					active.Depth = 2
					active.TraverseDescription.IsRelative = true
					active.ResumeDescription.IsRelative = false
					active.Subscription = entry.Subscription
					active.CurrentPath = entry.Resume.At
				})).Extent(
					entry.Builder(&lab.AsyncPostage{
						Entry: entry,
						Path:  entry.Path(),
						FS:    fS,
						On:    on,
					}),
				).Navigate(ctx)

				wg.Wait()

				Expect(err).To(Succeed())
				Expect(result).NotTo(BeNil())
			})
		},
		func(entry *lab.AsyncOkTE) string {
			return fmt.Sprintf("🧪 ===> given: '%v', should: '%v'",
				entry.Given, entry.Should,
			)
		},

		Entry(nil, &lab.AsyncOkTE{
			AsyncTE: lab.AsyncTE{
				Given:        "Primary Session WithCPUPool",
				Should:       "sprint with context",
				Callback:     AsyncNoWCallback("PRIME", 0),
				Builder:      PrimeBuilder,
				Path:         func() string { return lab.Static.RetroWave },
				Subscription: enums.SubscribeUniversal,
				CPU:          true,
			},
		}, SpecTimeout(time.Second*2)),

		Entry(nil, &lab.AsyncOkTE{
			AsyncTE: lab.AsyncTE{
				Given:        "Primary Session NoW=3",
				Should:       "sprint with context",
				Callback:     AsyncNoWCallback("PRIME", 3),
				Builder:      PrimeBuilder,
				Path:         func() string { return lab.Static.RetroWave },
				Subscription: enums.SubscribeUniversal,
				NoWorkers:    3,
			},
		}, SpecTimeout(time.Second*2)),

		Entry(nil, &lab.AsyncOkTE{
			AsyncTE: lab.AsyncTE{
				Given:        "Resume Session NoW=3",
				Should:       "sprint with context",
				Callback:     AsyncNoWCallback("RESUME", 3),
				Builder:      ResumeBuilder,
				Path:         func() string { return from },
				Subscription: enums.SubscribeUniversal,
				NoWorkers:    3,
				Resume: lab.AsyncResumeTE{
					At:       ResumeAtTeenageColor,
					Strategy: enums.ResumeStrategyFastward,
				},
			},
		}, SpecTimeout(time.Second*2)),

		Entry(nil, &lab.AsyncOkTE{
			AsyncTE: lab.AsyncTE{
				Given:        "Primary Session With Output",
				Should:       "consume output",
				Callback:     AsyncNoWCallback("PRIME", 3),
				Builder:      PrimeBuilder,
				Path:         func() string { return lab.Static.RetroWave },
				Subscription: enums.SubscribeUniversal,
				NoWorkers:    3,
			},
			Consume: true,
		}, SpecTimeout(time.Second*2)),

		Entry(nil, &lab.AsyncOkTE{
			AsyncTE: lab.AsyncTE{
				Given:    "Resume Session With Output",
				Should:   "consume output",
				Callback: AsyncNoWCallback("RESUME", 3),
				Builder:  ResumeBuilder,
				Resume: lab.AsyncResumeTE{
					At:       lab.Static.TeenageColor,
					Strategy: enums.ResumeStrategyFastward,
				},
				Path:         lab.GetJSONPath,
				Subscription: enums.SubscribeUniversal,
				NoWorkers:    3,
			},
			Consume: true,
		}, SpecTimeout(time.Second*2)),
	)
})

var _ = Describe("Navigator cancellation", func() {
	var fS *luna.MemFS

	BeforeEach(func() {
		Expect(li18ngo.Register(
			func(o *li18ngo.UseOptions) {
				o.From.Sources = li18ngo.TranslationFiles{
					locale.SourceID: li18ngo.TranslationSource{Name: "agenor"},
				}
			},
		)).To(Succeed())
		services.Reset()
		fS = hanno.Nuxx(verbose, lab.Static.RetroWave)
	})

	It("should return saved error when context is cancelled mid-sprint", SpecTimeout(time.Second*2), func(specCtx SpecContext) {
		lab.WithTestContext(specCtx, func(ctx context.Context, cancel context.CancelFunc) {
			var wg sync.WaitGroup

			navCtx, navCancel := context.WithCancel(ctx)
			defer navCancel()

			cancelled := false
			callback := func(servant core.Servant) error {
				if !cancelled {
					cancelled = true
					navCancel()
				}
				return nil
			}

			result, err := agenor.Sprint(&wg).Configure(
				enclave.Loader(func(active *core.ActiveState) {
					active.Tree = lab.Static.RetroWave
					active.Subscription = enums.SubscribeUniversal
				}),
			).Extent(
				agenor.Prime(
					&pref.Using{
						Tree:         lab.Static.RetroWave,
						Subscription: enums.SubscribeUniversal,
						Head: pref.Head{
							Handler: callback,
							GetForest: func(_ string) *core.Forest {
								return &core.Forest{
									T: fS,
									R: tfs.New(),
								}
							},
						},
					},
					agenor.WithNoW(3),
				),
			).Navigate(navCtx)

			wg.Wait()

			Expect(err).To(HaveOccurred())
			var savedErr *locale.TraversalSavedError
			Expect(errors.As(err, &savedErr)).To(BeTrue())
			Expect(result).NotTo(BeNil())
		})
	})
})
