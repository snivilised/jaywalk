package command

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cubiest/jibberjabber"
	"github.com/snivilised/jaywalk/src/agenor/pref"
	"github.com/snivilised/jaywalk/src/locale"
	"github.com/snivilised/jaywalk/src/third/lo"
	"github.com/snivilised/li18ngo"
	"github.com/snivilised/mamba/assist"
	macfg "github.com/snivilised/mamba/assist/cfg"
	si18n "github.com/snivilised/mamba/locale"
	"github.com/snivilised/mamba/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/language"

	"github.com/snivilised/jaywalk/src/app/bedrock"
	jac "github.com/snivilised/jaywalk/src/app/controller"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/app/shell"
	"github.com/snivilised/jaywalk/src/app/ui"
	"github.com/snivilised/jaywalk/src/prism/flow"
)

// ---------------------------------------------------------------------------
// Locale detection
// ---------------------------------------------------------------------------

// LocaleDetector abstracts the detection of the user's preferred
// language as a BCP 47 language tag.
type LocaleDetector interface {
	Scan() language.Tag
}

// Jabber is a LocaleDetector implemented using jibberjabber.
type Jabber struct{}

// Scan returns the detected language tag.
func (j *Jabber) Scan() language.Tag {
	lang, _ := jibberjabber.DetectIETF()
	return language.MustParse(lang)
}

// ---------------------------------------------------------------------------
// ConfigureOptions
// ---------------------------------------------------------------------------

// AppConfigInfo describes the configuration file that should be loaded.
type AppConfigInfo struct {
	Name       string
	ConfigType string
	ConfigPath string
	Viper      macfg.ViperConfig
}

// ConfigureAppOptions groups options that influence how Bootstrap
// initialises localisation and configuration.
type ConfigureAppOptions struct {
	Detector   LocaleDetector
	ConfigInfo AppConfigInfo
	GetForest  pref.BuildForest
}

// ConfigureAppOptionFn is a functional option used to modify
// ConfigureOptions before Bootstrap performs its setup.
type ConfigureAppOptionFn func(*ConfigureAppOptions)

// ---------------------------------------------------------------------------
// Per-leaf param-set state
// ---------------------------------------------------------------------------

// navState holds the nav-level param-sets that every navigation leaf command
// (walk, sprint, query) registers independently on its own flag set.
// Each leaf owns its own navState so that flags appear only on the command
// that declares them and nowhere else.
type navState struct {
	navPs       *assist.ParamSet[NavParameterSet]
	previewFam  *assist.ParamSet[store.PreviewParameterSet]
	cascadeFam  *assist.ParamSet[store.CascadeParameterSet]
	samplingFam *assist.ParamSet[store.SamplingParameterSet]
	polyFam     *assist.ParamSet[store.PolyFilterParameterSet]
}

// walkState holds all param-sets owned exclusively by the walk command.
type walkState struct {
	navState
	execPs *assist.ParamSet[ExecParameterSet]
}

// sprintState holds all param-sets owned exclusively by the sprint command.
// sprint is the only command that owns the worker-pool family.
type sprintState struct {
	navState
	execPs        *assist.ParamSet[ExecParameterSet]
	workerPoolFam *assist.ParamSet[store.WorkerPoolParameterSet]
}

// queryState holds all param-sets owned exclusively by the query command.
// query intentionally omits execPs: it is a read-only traversal that
// cannot be resumed.
type queryState struct {
	navState
}

// ---------------------------------------------------------------------------
// Bootstrap
// ---------------------------------------------------------------------------

// Bootstrap is a pure composition root. Its sole responsibility is
// wiring: it constructs the cobra command tree, registers param-sets,
// creates the Coordinator, and connects everything together.
// It contains no business logic and no traversal decisions.
//
// Command hierarchy:
//
//	root               (persistent flags: --tui, --theme)
//	  ├── walk         (flags: nav + families + --resume)
//	  ├── sprint       (flags: nav + families + --resume + worker-pool)
//	  ├── query        (flags: nav + families)
//	  ├── verify       (flags: tbd)
//	  └── theme        (flags: tbd)
//
// walk, sprint, and query are direct children of root. Each registers its
// own copy of the nav flags and families on its local flag set so that
// flags appear only on the commands that own them. There are no ghost
// intermediary commands. See doc.go for the full flag inventory.
//
// Code smell checklist (should remain clean):
//   - No direct calls to agenor from Bootstrap
//   - No UI rendering logic in Bootstrap
//   - No flag interpretation beyond resolving --tui and --theme
type Bootstrap struct {
	container *assist.CobraContainer
	options   ConfigureAppOptions

	// AppConfig is populated after configure() reads viper.
	AppConfig *bedrock.Config

	// UI is resolved from --tui and --theme in PersistentPreRunE and
	// passed into requests. Bootstrap does not use it directly.
	UI report.Presenter

	// fileManager provides XDG-compliant path resolution for all
	// file system locations used by jay (config, state, cache, themes).
	fileManager *bedrock.FileManager

	// logger is the structured logger used by the application.
	// Created once in Root() and passed to the Coordinator.
	logger *slog.Logger

	// themeLoader resolves named themes from the themes directory.
	// Constructed once in Root() and reused in PersistentPreRunE.
	themeLoader *bedrock.ThemeLoader

	// viewConfigLoader loads per-view configuration from jay.ui.yml
	// on-demand when a view is selected (e.g. --tui highway).
	viewConfigLoader *bedrock.ViewConfigLoader

	// coord is the single Coordinator instance wired at startup and
	// shared by all command handlers.
	coord *jac.Coordinator

	// root param-set
	rootPs *assist.ParamSet[RootParameterSet]

	// per-leaf param-set state; each leaf owns its flags independently.
	walk   walkState
	sprint sprintState
	query  queryState
}

// ---------------------------------------------------------------------------
// Root
// ---------------------------------------------------------------------------

func (b *Bootstrap) prepare() {
	b.fileManager = bedrock.NewFileManager()

	b.options = ConfigureAppOptions{
		Detector: &Jabber{},
		ConfigInfo: AppConfigInfo{
			Name:       bedrock.AppName,
			ConfigType: "", // no forced type; viper discovers .yaml, .yml, etc.
			ConfigPath: b.fileManager.ConfigHome(),
			Viper:      &macfg.GlobalViperConfig{},
		},
	}
}

// Root builds the command tree and returns the root command, ready to
// be executed. This is the composition root: all wiring happens here
// and only here.
func (b *Bootstrap) Root(options ...ConfigureAppOptionFn) *cobra.Command {
	b.prepare()

	for _, fo := range options {
		fo(&b.options)
	}

	b.configure()

	env, err := shell.Detect()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	b.themeLoader = bedrock.NewThemeLoaderWithDir(b.fileManager.ThemesDir())
	b.viewConfigLoader = bedrock.NewViewConfigLoader(b.fileManager.ConfigHome())
	b.logger = b.createLogger()
	b.coord = jac.New(b.AppConfig,
		jac.WithLocate(env.Locate),
		jac.WithExec(env.Execute),
		jac.WithShell(env.Command),
		jac.WithForest(b.options.GetForest),
		jac.WithAdminPath(b.fileManager.AdminPath()),
		jac.WithLogger(b.logger),
	)

	b.container = assist.NewCobraContainer(
		&cobra.Command{
			Use:     bedrock.AppName,
			Short:   li18ngo.Text(locale.RootCmdShortDescTemplData{}),
			Long:    li18ngo.Text(locale.RootCmdLongDescTemplData{}),
			Version: fmt.Sprintf("'%v'", Version),

			PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
				mode := b.rootPs.Native.TUI
				if mode == ui.ModeDefault || mode == "" {
					if cmd.Name() == "sprint" {
						mode = ui.ModeHighway
					}
				}

				// Register only the selected view factory; avoid registering
				// all known views when only one will be used.
				switch mode {
				case ui.ModeLinear:
					flow.Register()

				case ui.ModeHighway:
					var cfg bedrock.HighwayConfig
					if err := b.viewConfigLoader.Load("highway", &cfg); err != nil {
						return err
					}
					highwayPath := b.viewConfigLoader.ResolvePath()
					if highwayPath == "" {
						highwayPath = filepath.Join(b.fileManager.ConfigHome(), "jay.ui.yml")
					}
					fmt.Printf("[DEBUG] Highway view configuration loaded from %s\n",
						highwayPath)
					fmt.Printf("[DEBUG]  Emoji pool: %q\n", cfg.Pool)
					if len(cfg.AnimationData.EnabledTypes) > 0 {
						fmt.Printf("[DEBUG]  Animation types enabled: %v\n",
							cfg.AnimationData.EnabledTypes)
					} else {
						fmt.Println("[DEBUG]  (No animation types specified, using defaults)")
					}
					// TODO: register highway view factory when view is implemented
				}

				palette, err := b.themeLoader.Load(b.rootPs.Native.Theme)
				if err != nil {
					return err
				}

				mgr, err := ui.New(mode, palette)
				if err != nil {
					return err
				}

				b.UI = mgr

				return nil
			},

			Run: func(_ *cobra.Command, _ []string) {
				fmt.Println("=== jay ===")
			},
		},
	)

	// Registration order: root first, then its direct children.
	b.buildRootCommand(b.container)
	b.buildWalkCommand(b.container)
	b.buildSprintCommand(b.container)
	b.buildQueryCommand(b.container)

	return b.container.Root()
}

// ---------------------------------------------------------------------------
// configure
// ---------------------------------------------------------------------------

func (b *Bootstrap) configure() {
	vc := b.options.ConfigInfo.Viper
	ci := b.options.ConfigInfo

	vc.SetConfigName(ci.Name)
	if ci.ConfigType != "" {
		vc.SetConfigType(ci.ConfigType)
	}
	vc.AddConfigPath(ci.ConfigPath) // XDG config home (primary)
	vc.AddConfigPath(".")           // current dir
	if def := b.fileManager.ResolvePath("$HOME/.config/jay"); def != ci.ConfigPath {
		vc.AddConfigPath(def) // legacy location (fallback)
	}
	vc.AddConfigPath("/etc/jay") // system-wide
	vc.AutomaticEnv()

	err := vc.ReadInConfig()

	handleLangSetting()

	if err != nil {
		msg := li18ngo.Text(locale.UsingConfigFileTemplData{
			ConfigFileName: b.options.ConfigInfo.Name,
		})
		fmt.Fprintln(os.Stderr, msg)
	}

	b.AppConfig, err = bedrock.Load(bedrock.LoadOptions{
		ViperInstance: viper.GetViper(),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handleLangSetting() {
	tag := lo.TernaryF(viper.InConfig("lang"),
		func() language.Tag {
			lang := viper.GetString("lang")
			parsedTag, err := language.Parse(lang)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			return parsedTag
		},
		func() language.Tag {
			return li18ngo.DefaultLanguage
		},
	)

	err := li18ngo.Register(func(uo *li18ngo.UseOptions) {
		uo.Tag = tag
		uo.From = li18ngo.LoadFrom{
			Sources: li18ngo.TranslationFiles{
				SourceID:            li18ngo.TranslationSource{Name: bedrock.AppName},
				si18n.MambaSourceID: li18ngo.TranslationSource{Name: "mamba"},
			},
		}
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// createLogger
// ---------------------------------------------------------------------------

func (b *Bootstrap) createLogger() *slog.Logger {
	logPath := b.fileManager.LogPath()
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0o750); err != nil { // originally 0o755
		fmt.Fprintln(os.Stderr, "warning: could not create log directory:", err)
	}

	//nolint:gosec // logPath comes from user's own config, not untrusted input
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600) // originally 0o644
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning: could not open log file:", err)
		return slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	level := slog.LevelInfo
	if b.AppConfig != nil {
		level = parseLogLevel(b.AppConfig.Mapped.Logging.Level)
	}

	return slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: level,
	}))
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// ---------------------------------------------------------------------------
// buildRootCommand
// ---------------------------------------------------------------------------

// buildRootCommand registers the truly global persistent flags on root:
// --tui and --theme. These are the only flags that propagate to all
// sub-commands including utility commands (verify, theme).
func (b *Bootstrap) buildRootCommand(container *assist.CobraContainer) {
	root := container.Root()

	b.rootPs = assist.NewParamSet[RootParameterSet](root)

	b.rootPs.BindString(
		assist.NewFlagInfoOnFlagSet(
			li18ngo.Text(locale.TuiFlagDescTemplData{}),
			"t",
			ui.ModeDefault,
			root.PersistentFlags(),
		),
		&b.rootPs.Native.TUI,
	)

	b.rootPs.BindString(
		assist.NewFlagInfoOnFlagSet(
			li18ngo.Text(locale.NewThemeFlagDescTemplData(b.themeLoader.ThemesDir())),
			"",
			bedrock.ThemeSystemName,
			root.PersistentFlags(),
		),
		&b.rootPs.Native.Theme,
	)

	container.MustRegisterParamSet(RootPsName, b.rootPs)
}

// ---------------------------------------------------------------------------
// bindNavFlags
// ---------------------------------------------------------------------------

// bindNavFlags registers the nav param-set and all four nav flag families
// onto the supplied cobra command using its local flag set. It is called
// once per navigation leaf (walk, sprint, query) so that each command owns
// its flags independently with no cross-contamination.
func (b *Bootstrap) bindNavFlags(cmd *cobra.Command, ns *navState) {
	fs := cmd.Flags()

	ns.navPs = assist.NewParamSet[NavParameterSet](cmd)

	ns.navPs.BindString(
		assist.NewFlagInfoOnFlagSet(
			li18ngo.Text(locale.SubscribeFlagDescTemplData{}),
			"s",
			SubscribeFlagDefault,
			fs,
		),
		&ns.navPs.Native.Subscribe,
	)

	ns.navPs.BindString(
		assist.NewFlagInfoOnFlagSet(
			li18ngo.Text(locale.ActionFlagDescTemplData{}),
			"a",
			"",
			fs,
		),
		&ns.navPs.Native.Action,
	)

	ns.navPs.BindString(
		assist.NewFlagInfoOnFlagSet(
			li18ngo.Text(locale.PipelineFlagDescTemplData{}),
			"p",
			"",
			fs,
		),
		&ns.navPs.Native.Pipeline,
	)

	// family: preview [--dry-run]
	ns.previewFam = assist.NewParamSet[store.PreviewParameterSet](cmd)
	ns.previewFam.Native.BindAll(ns.previewFam, fs)

	// family: cascade [--depth, --no-recurse]
	ns.cascadeFam = assist.NewParamSet[store.CascadeParameterSet](cmd)
	ns.cascadeFam.Native.BindAll(ns.cascadeFam, fs)

	// family: sampling [--sample, --num-files, --num-folders, --last]
	ns.samplingFam = assist.NewParamSet[store.SamplingParameterSet](cmd)
	ns.samplingFam.Native.BindAll(ns.samplingFam, fs)

	// family: poly-filter [--files-glob, --file-regex, --folders-glob, --folders-regex]
	ns.polyFam = assist.NewParamSet[store.PolyFilterParameterSet](cmd)
	ns.polyFam.Native.BindAll(ns.polyFam, fs)
}

// ---------------------------------------------------------------------------
// bindExecFlags
// ---------------------------------------------------------------------------

// bindExecFlags registers --resume onto the supplied command's local flag
// set and populates the provided ParamSet pointer. Called only by walk and
// sprint; query intentionally omits this since it cannot be resumed.
func (b *Bootstrap) bindExecFlags(cmd *cobra.Command, ep **assist.ParamSet[ExecParameterSet]) {
	// TODO: WTF **??
	*ep = assist.NewParamSet[ExecParameterSet](cmd)

	(*ep).BindString(
		assist.NewFlagInfoOnFlagSet(
			li18ngo.Text(locale.ResumeFlagDescTemplData{}),
			"r",
			"",
			cmd.Flags(),
		),
		&(*ep).Native.Resume,
	)
}

// ---------------------------------------------------------------------------
// navFamilies
// ---------------------------------------------------------------------------

// navFamilies builds a NavFamilies bundle from a navState pointer.
// Each runner passes its own leaf's navState so the correct param-sets
// are used regardless of which command is executing.
func navFamilies(ns *navState) NavFamilies {
	return NavFamilies{
		Preview:  ns.previewFam,
		Cascade:  ns.cascadeFam,
		Sampling: ns.samplingFam,
		PolyFam:  ns.polyFam,
	}
}
