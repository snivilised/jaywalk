package command_test

import (
	"errors"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/agenor/test/hanno"
	"github.com/snivilised/jaywalk/src/app/command"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/locale"
	"github.com/snivilised/jaywalk/src/prism/movies"
	"github.com/snivilised/li18ngo"
	nef "github.com/snivilised/nefilim"
	"github.com/spf13/cobra"
	"golang.org/x/text/language"
)

const (
	configName = "jay-test"
	configPath = "../../test/data/configuration"
)

// DetectorStub is a test double that satisfies the LocaleDetector
// interface and always returns British English.
type DetectorStub struct{}

func (j *DetectorStub) Scan() language.Tag {
	return language.BritishEnglish
}

func applyTestConfig(co *command.ConfigureAppOptions) {
	co.Detector = &DetectorStub{}
	co.ConfigInfo.Name = configName
	co.ConfigInfo.ConfigPath = configPath
}

func buildRoot() *cobra.Command {
	return (&command.Bootstrap{}).Root(applyTestConfig)
}

var _ = Describe("Bootstrap", Ordered, func() {
	var (
		repo     string
		l10nPath string
	)

	BeforeAll(func() {
		Expect(li18ngo.Register()).To(Succeed())
	})

	BeforeAll(func() {
		repo = hanno.Repo("")
		l10nPath = hanno.Combine(repo, "test/data/l10n")
		fS := nef.NewUniversalABS()
		Expect(fS.DirectoryExists(l10nPath)).To(BeTrue())
	})

	Context("given: root defined", func() {
		It("🧪 should: build command tree without error", func() {
			Expect(buildRoot()).NotTo(BeNil())
		})
	})

	Context("given: command tree built", func() {

		// ---------------------------------------------------------------
		// Discoverability: walk, sprint, query must be visible direct
		// children of root so they appear in 'jay --help'.
		// ---------------------------------------------------------------

		Context("discoverability", func() {
			It("🧪 should: register walk as a direct child of root", func() {
				walkCmd, _, err := buildRoot().Find([]string{"walk"})
				Expect(err).To(BeNil())
				Expect(walkCmd).NotTo(BeNil())
				Expect(walkCmd.Name()).To(Equal("walk"))
			})

			It("🧪 should: register sprint as a direct child of root", func() {
				sprintCmd, _, err := buildRoot().Find([]string{"sprint"})
				Expect(err).To(BeNil())
				Expect(sprintCmd).NotTo(BeNil())
				Expect(sprintCmd.Name()).To(Equal("sprint"))
			})

			It("🧪 should: register query as a direct child of root", func() {
				queryCmd, _, err := buildRoot().Find([]string{"query"})
				Expect(err).To(BeNil())
				Expect(queryCmd).NotTo(BeNil())
				Expect(queryCmd.Name()).To(Equal("query"))
			})

			It("🧪 should: not register any hidden ghost commands", func() {
				root := buildRoot()
				for _, sub := range root.Commands() {
					Expect(sub.Hidden).To(BeFalse(),
						"command %q should not be hidden", sub.Name())
				}
			})
		})

		// ---------------------------------------------------------------
		// Nav flags: --subscribe, --action, --pipeline must appear as
		// local flags on walk, sprint, and query.
		// ---------------------------------------------------------------

		Context("nav flags", func() {
			It("🧪 should: expose --subscribe on walk", func() {
				walkCmd, _, err := buildRoot().Find([]string{"walk"})
				Expect(err).To(BeNil())
				Expect(walkCmd.Flags().Lookup("subscribe")).NotTo(BeNil())
			})

			It("🧪 should: expose --subscribe on sprint", func() {
				sprintCmd, _, err := buildRoot().Find([]string{"sprint"})
				Expect(err).To(BeNil())
				Expect(sprintCmd.Flags().Lookup("subscribe")).NotTo(BeNil())
			})

			It("🧪 should: expose --subscribe on query", func() {
				queryCmd, _, err := buildRoot().Find([]string{"query"})
				Expect(err).To(BeNil())
				Expect(queryCmd.Flags().Lookup("subscribe")).NotTo(BeNil())
			})

			It("🧪 should: expose --action on walk", func() {
				walkCmd, _, err := buildRoot().Find([]string{"walk"})
				Expect(err).To(BeNil())
				Expect(walkCmd.Flags().Lookup("action")).NotTo(BeNil())
			})

			It("🧪 should: expose --action on sprint", func() {
				sprintCmd, _, err := buildRoot().Find([]string{"sprint"})
				Expect(err).To(BeNil())
				Expect(sprintCmd.Flags().Lookup("action")).NotTo(BeNil())
			})

			It("🧪 should: expose --action on query", func() {
				queryCmd, _, err := buildRoot().Find([]string{"query"})
				Expect(err).To(BeNil())
				Expect(queryCmd.Flags().Lookup("action")).NotTo(BeNil())
			})

			It("🧪 should: expose --pipeline on walk", func() {
				walkCmd, _, err := buildRoot().Find([]string{"walk"})
				Expect(err).To(BeNil())
				Expect(walkCmd.Flags().Lookup("pipeline")).NotTo(BeNil())
			})

			It("🧪 should: expose --pipeline on sprint", func() {
				sprintCmd, _, err := buildRoot().Find([]string{"sprint"})
				Expect(err).To(BeNil())
				Expect(sprintCmd.Flags().Lookup("pipeline")).NotTo(BeNil())
			})

			It("🧪 should: expose --pipeline on query", func() {
				queryCmd, _, err := buildRoot().Find([]string{"query"})
				Expect(err).To(BeNil())
				Expect(queryCmd.Flags().Lookup("pipeline")).NotTo(BeNil())
			})
		})

		// ---------------------------------------------------------------
		// Exec flags: --resume must appear on walk and sprint only.
		// query must not expose it in any form.
		// ---------------------------------------------------------------

		Context("exec flags", func() {
			It("🧪 should: expose --resume on walk", func() {
				walkCmd, _, err := buildRoot().Find([]string{"walk"})
				Expect(err).To(BeNil())
				Expect(walkCmd.Flags().Lookup("resume")).NotTo(BeNil())
			})

			It("🧪 should: expose --resume on sprint", func() {
				sprintCmd, _, err := buildRoot().Find([]string{"sprint"})
				Expect(err).To(BeNil())
				Expect(sprintCmd.Flags().Lookup("resume")).NotTo(BeNil())
			})

			It("🧪 should: not expose --resume on query", func() {
				queryCmd, _, err := buildRoot().Find([]string{"query"})
				Expect(err).To(BeNil())
				Expect(queryCmd.Flags().Lookup("resume")).To(BeNil())
				Expect(queryCmd.InheritedFlags().Lookup("resume")).To(BeNil())
			})
		})

		// ---------------------------------------------------------------
		// Root persistent flags: --tui and --theme must be inherited by
		// all navigation commands. Nav and exec flags must NOT leak onto
		// root's persistent flag set.
		// ---------------------------------------------------------------

		Context("root persistent flags", func() {
			It("🧪 should: expose --theme on walk via inheritance", func() {
				walkCmd, _, err := buildRoot().Find([]string{"walk"})
				Expect(err).To(BeNil())
				Expect(walkCmd.InheritedFlags().Lookup("theme")).NotTo(BeNil())
			})

			It("🧪 should: expose --theme on sprint via inheritance", func() {
				sprintCmd, _, err := buildRoot().Find([]string{"sprint"})
				Expect(err).To(BeNil())
				Expect(sprintCmd.InheritedFlags().Lookup("theme")).NotTo(BeNil())
			})

			It("🧪 should: expose --theme on query via inheritance", func() {
				queryCmd, _, err := buildRoot().Find([]string{"query"})
				Expect(err).To(BeNil())
				Expect(queryCmd.InheritedFlags().Lookup("theme")).NotTo(BeNil())
			})

			It("🧪 should: not register --subscribe as a root persistent flag", func() {
				Expect(buildRoot().PersistentFlags().Lookup("subscribe")).To(BeNil())
			})

			It("🧪 should: not register --action as a root persistent flag", func() {
				Expect(buildRoot().PersistentFlags().Lookup("action")).To(BeNil())
			})

			It("🧪 should: not register --pipeline as a root persistent flag", func() {
				Expect(buildRoot().PersistentFlags().Lookup("pipeline")).To(BeNil())
			})

			It("🧪 should: not register --resume as a root persistent flag", func() {
				Expect(buildRoot().PersistentFlags().Lookup("resume")).To(BeNil())
			})
		})

		// ---------------------------------------------------------------
		// Worker-pool flags: --cpu and --now are sprint-exclusive.
		// They must not appear on walk or query in any form.
		// ---------------------------------------------------------------

		Context("worker-pool flags - sprint exclusive", func() {
			It("🧪 should: expose --cpu as a local flag on sprint", func() {
				sprintCmd, _, err := buildRoot().Find([]string{"sprint"})
				Expect(err).To(BeNil())
				Expect(sprintCmd.Flags().Lookup("cpu")).NotTo(BeNil())
			})

			It("🧪 should: expose --now as a local flag on sprint", func() {
				sprintCmd, _, err := buildRoot().Find([]string{"sprint"})
				Expect(err).To(BeNil())
				Expect(sprintCmd.Flags().Lookup("now")).NotTo(BeNil())
			})

			It("🧪 should: not expose --cpu on walk", func() {
				walkCmd, _, err := buildRoot().Find([]string{"walk"})
				Expect(err).To(BeNil())
				Expect(walkCmd.Flags().Lookup("cpu")).To(BeNil())
				Expect(walkCmd.InheritedFlags().Lookup("cpu")).To(BeNil())
			})

			It("🧪 should: not expose --now on walk", func() {
				walkCmd, _, err := buildRoot().Find([]string{"walk"})
				Expect(err).To(BeNil())
				Expect(walkCmd.Flags().Lookup("now")).To(BeNil())
				Expect(walkCmd.InheritedFlags().Lookup("now")).To(BeNil())
			})

			It("🧪 should: not expose --cpu on query", func() {
				queryCmd, _, err := buildRoot().Find([]string{"query"})
				Expect(err).To(BeNil())
				Expect(queryCmd.Flags().Lookup("cpu")).To(BeNil())
				Expect(queryCmd.InheritedFlags().Lookup("cpu")).To(BeNil())
			})

			It("🧪 should: not expose --now on query", func() {
				queryCmd, _, err := buildRoot().Find([]string{"query"})
				Expect(err).To(BeNil())
				Expect(queryCmd.Flags().Lookup("now")).To(BeNil())
				Expect(queryCmd.InheritedFlags().Lookup("now")).To(BeNil())
			})
		})

		// ---------------------------------------------------------------
		// Flag isolation: nav and exec flags must not appear on utility
		// commands (verify, theme) that are direct children of root.
		// ---------------------------------------------------------------

		XContext("flag isolation - utility commands", Label("pending"), func() {
			It("🧪 should: not expose --subscribe on verify", func() {
				verifyCmd, _, err := buildRoot().Find([]string{"verify"})
				Expect(err).To(BeNil())
				Expect(verifyCmd.Flags().Lookup("subscribe")).To(BeNil())
				Expect(verifyCmd.InheritedFlags().Lookup("subscribe")).To(BeNil())
			})

			It("🧪 should: not expose --resume on verify", func() {
				verifyCmd, _, err := buildRoot().Find([]string{"verify"})
				Expect(err).To(BeNil())
				Expect(verifyCmd.Flags().Lookup("resume")).To(BeNil())
				Expect(verifyCmd.InheritedFlags().Lookup("resume")).To(BeNil())
			})

			It("🧪 should: not expose --subscribe on theme", func() {
				themeCmd, _, err := buildRoot().Find([]string{"theme"})
				Expect(err).To(BeNil())
				Expect(themeCmd.Flags().Lookup("subscribe")).To(BeNil())
				Expect(themeCmd.InheritedFlags().Lookup("subscribe")).To(BeNil())
			})

			It("🧪 should: not expose --resume on theme", func() {
				themeCmd, _, err := buildRoot().Find([]string{"theme"})
				Expect(err).To(BeNil())
				Expect(themeCmd.Flags().Lookup("resume")).To(BeNil())
				Expect(themeCmd.InheritedFlags().Lookup("resume")).To(BeNil())
			})
		})

		// ---------------------------------------------------------------
		// Lazy view bootstrap (issue 583): PersistentPreRunE must
		// construct the selected view on-demand without any pre-registration
		// step. The concrete presenter type reflects the --tui mode.
		// ---------------------------------------------------------------

		Context("lazy view bootstrap - issue 583", Label("issue-583"), func() {
			// presenterType returns the reflect-qualified name of the
			// concrete presenter so we can distinguish linear from
			// highway without exposing the unexported ui types.
			presenterType := func(p report.Presenter) string {
				if p == nil {
					return "<nil>"
				}
				return reflect.TypeOf(p).String()
			}

			Context("given: --tui linear on walk", func() {
				It("🧪 should: build the linear presenter on-demand", func() {
					bootstrap := command.Bootstrap{}
					tester := hanno.CommandTester{
						// walk requires 1 arg + --action/--pipeline. We
						// supply a dummy directory and a noop action so
						// PersistentPreRunE runs and b.UI is populated
						// before any traversal error.
						Args: []string{
							"--tui", "linear",
							"walk", ".",
							"--action", "noop",
							"--theme", "system",
						},
						Root: bootstrap.Root(applyTestConfig),
					}
					_, _ = tester.Execute()

					Expect(bootstrap.UI).NotTo(BeNil(),
						"PersistentPreRunE should have populated b.UI")
					Expect(presenterType(bootstrap.UI)).To(Equal("*ui.linearPresenter"))
				})
			})

			Context("given: --tui highway on walk", func() {
				It("🧪 should: build the highway presenter on-demand", func() {
					bootstrap := command.Bootstrap{}
					tester := hanno.CommandTester{
						Args: []string{
							"--tui", "highway",
							"walk", ".",
							"--action", "noop",
							"--theme", "system",
						},
						Root: bootstrap.Root(applyTestConfig),
					}
					_, _ = tester.Execute()

					Expect(bootstrap.UI).NotTo(BeNil())
					Expect(presenterType(bootstrap.UI)).To(Equal("*ui.highwayPresenter"))
				})
			})

			Context("given: default mode (no --tui) on walk", func() {
				It("🧪 should: fall through to the linear presenter", func() {
					bootstrap := command.Bootstrap{}
					tester := hanno.CommandTester{
						Args: []string{
							"walk", ".",
							"--action", "noop",
							"--theme", "system",
						},
						Root: bootstrap.Root(applyTestConfig),
					}
					_, _ = tester.Execute()

					Expect(bootstrap.UI).NotTo(BeNil())
					Expect(presenterType(bootstrap.UI)).To(Equal("*ui.linearPresenter"))
				})
			})

			Context("given: default mode on sprint", func() {
				It("🧪 should: override the default to highway", func() {
					bootstrap := command.Bootstrap{}
					tester := hanno.CommandTester{
						Args: []string{
							"sprint", ".",
							"--action", "noop",
							"--theme", "system",
						},
						Root: bootstrap.Root(applyTestConfig),
					}
					_, _ = tester.Execute()

					Expect(bootstrap.UI).NotTo(BeNil())
					Expect(presenterType(bootstrap.UI)).To(Equal("*ui.highwayPresenter"))
				})
			})

			Context("given: highway presenter construction", func() {
				It("🧪 should: invoke movies.RegisterAll() before building", func() {
					// newHighwayPresenter calls movies.RegisterAll() as
					// its first action. We verify the registration
					// occurred indirectly: a known spinner name must be
					// resolvable. We trigger the presenter via the
					// command tester, then query the movies package
					// directly.
					bootstrap := command.Bootstrap{}
					tester := hanno.CommandTester{
						Args: []string{
							"--tui", "highway",
							"walk", ".",
							"--action", "noop",
							"--theme", "system",
						},
						Root: bootstrap.Root(applyTestConfig),
					}
					_, _ = tester.Execute()

					Expect(bootstrap.UI).NotTo(BeNil())
					def, ok := movies.Lookup(movies.SpinnerTypeDefault)
					Expect(ok).To(BeTrue(),
						"movies.RegisterAll() should have populated the default spinner")
					Expect(def.Frames).NotTo(BeNil())
					// And a frame can be produced:
					frame := def.Frames(0)
					Expect(frame).NotTo(BeEmpty())
				})
			})

			// ---------------------------------------------------------------
			// View config (issue 583 follow-up): the
			// presenter must be constructable via the FullViewConfig
			// returned by ui.LoadConfig. The highway view must work
			// even when no jay.ui.yml is present (zero-value config
			// is still a valid FullViewConfig).
			// ---------------------------------------------------------------

			Context("view config", func() {
				It("🧪 should: build highway presenter via zero-value FullViewConfig when jay.ui.yml is absent", func() {
					bootstrap := command.Bootstrap{}
					tester := hanno.CommandTester{
						Args: []string{
							"--tui", "highway",
							"walk", ".",
							"--action", "noop",
							"--theme", "system",
						},
						Root: bootstrap.Root(applyTestConfig),
					}
					_, _ = tester.Execute()

					Expect(bootstrap.UI).NotTo(BeNil())
					Expect(presenterType(bootstrap.UI)).To(Equal("*ui.highwayPresenter"),
						"highway mode should produce highwayPresenter even without jay.ui.yml")
				})

				It("🧪 should: build linear presenter without consuming any view config", func() {
					bootstrap := command.Bootstrap{}
					tester := hanno.CommandTester{
						Args: []string{
							"--tui", "linear",
							"walk", ".",
							"--action", "noop",
							"--theme", "system",
						},
						Root: bootstrap.Root(applyTestConfig),
					}
					_, _ = tester.Execute()

					Expect(bootstrap.UI).NotTo(BeNil())
					Expect(presenterType(bootstrap.UI)).To(Equal("*ui.linearPresenter"))
				})
			})
		})

		// ---------------------------------------------------------------
		// Porthole view rejection: porthole is incompatible with sprint;
		// the guard in PersistentPreRunE must return an error before
		// constructing any presenter.
		// ---------------------------------------------------------------

		Context("given: --tui porthole on sprint", func() {
			It("🧪 should: return an error because porthole is incompatible with sprint", func() {
				bootstrap := command.Bootstrap{}
				tester := hanno.CommandTester{
					Args: []string{
						"--tui", "porthole",
						"sprint", ".",
						"--action", "noop",
						"--theme", "system",
					},
					Root: bootstrap.Root(applyTestConfig),
				}
				_, err := tester.Execute()

				Expect(err).NotTo(BeNil(),
					"sprint with --tui porthole must fail")

				var target *locale.PortholeViewNotSupportedBySprintError
				Expect(errors.As(err, &target)).To(BeTrue(),
					"error should be PortholeViewNotSupportedBySprintError, got: %v", err)

				Expect(bootstrap.UI).To(BeNil(),
					"UI must remain nil when porthole guard rejects sprint")
			})
		})
	})
})
