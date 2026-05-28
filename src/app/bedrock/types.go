package bedrock

import (
	"time"
)

// ---------------------------------------------------------------------------
// Mapped sections - decoded via mapstructure
// ---------------------------------------------------------------------------

// TUIConfig holds settings for the terminal user interface.
type TUIConfig struct {
	PerItemDelay time.Duration `mapstructure:"per-item-delay"`
}

// InteractionConfig groups all user-interaction knobs.
type InteractionConfig struct {
	TUI        TUIConfig `mapstructure:"tui"`
	DateFormat string    `mapstructure:"date-format"`
}

// ExtensionsConfig controls how file extensions are normalised.
type ExtensionsConfig struct {
	SuffixesCSV   string            `mapstructure:"suffixes-csv"`
	TransformsCSV string            `mapstructure:"transforms-csv"`
	Map           map[string]string `mapstructure:"map"`
}

// ExecConfig controls command execution settings.
type ExecConfig struct {
	Truncate int `mapstructure:"truncate"`
}

// OutputConfig controls output processing settings.
type OutputConfig struct {
	Exec ExecConfig `mapstructure:"exec"`
}

// AdvancedConfig holds low-level behavioural switches.
type AdvancedConfig struct {
	AbortOnError         bool             `mapstructure:"abort-on-error"`
	OverwriteOnCollision bool             `mapstructure:"overwrite-on-collision"`
	Extensions           ExtensionsConfig `mapstructure:"extensions"`
	Output               OutputConfig     `mapstructure:"output"`
}

// LoggingConfig controls the jay log file.
// The log path is managed by FileManager and JAY_STATE_DIR, not via config.
type LoggingConfig struct {
	MaxSizeInMB  int    `mapstructure:"max-size-in-mb"`
	MaxBackups   int    `mapstructure:"max-backups"`
	MaxAgeInDays int    `mapstructure:"max-age-in-days"`
	Level        string `mapstructure:"level"`
	TimeFormat   string `mapstructure:"time-format"`
}

// MappedConfig bundles all sections that are decoded into concrete types.
type MappedConfig struct {
	Interaction InteractionConfig `mapstructure:"interaction"`
	Advanced    AdvancedConfig    `mapstructure:"advanced"`
	Logging     LoggingConfig     `mapstructure:"logging"`
	Highway     HighwayConfig     `mapstructure:"highway"`
}

// ---------------------------------------------------------------------------
// Raw sections - consumer-driven, arbitrary user content
// ---------------------------------------------------------------------------

// RawAction is one entry from the actions block.  The cmd and when strings
// are kept verbatim; jay's action-runner is responsible for interpreting them.
type RawAction struct {
	Cmd     string `mapstructure:"cmd"`
	When    string `mapstructure:"when"`
	Capture string `mapstructure:"capture"`
}

// RawPipeline is one entry from the pipelines block.
type RawPipeline struct {
	Steps []string `mapstructure:"steps"`
}

// FlagShortOverride captures per-command short-flag re-mappings.
//
//	flags.short.overrides.cmds.<cmd>.<flag> = <letter>
type FlagShortOverride map[string]map[string]string

// FlagInvokeDefaults captures command-level flag defaults.
//
//	flags.invoke.cmds.<cmd>.<flag> = <value>
type FlagInvokeDefaults map[string]map[string]any

// FlagComponentDefaults captures component-level flag defaults.
//
//	flags.component.<component>.<flag> = <value>
type FlagComponentDefaults map[string]map[string]any

// FlagsConfig aggregates all flags sub-sections.
type FlagsConfig struct {
	Short     FlagShortOverride     `mapstructure:"short"`
	Invoke    FlagInvokeDefaults    `mapstructure:"invoke"`
	Component FlagComponentDefaults `mapstructure:"component"`
}

// RawConfig holds all unstructured sections verbatim.
type RawConfig struct {
	Actions   map[string]RawAction   `mapstructure:"actions"`
	Pipelines map[string]RawPipeline `mapstructure:"pipelines"`
	Flags     FlagsConfig            `mapstructure:"flags"`
}

// ---------------------------------------------------------------------------
// Top-level unified config
// ---------------------------------------------------------------------------

// Config is the root configuration object handed to callers after a
// successful Load + Validate cycle.
type Config struct {
	Mapped MappedConfig
	Raw    RawConfig
}

// ---------------------------------------------------------------------------
// View Configuration — loaded on-demand from jay.ui.yml
// ---------------------------------------------------------------------------

// HighwayConfig holds the emoji pool and animation data for Highway view.
// This configuration is loaded from jay.ui.yml when Highway view is requested.
type HighwayConfig struct {
	// WorkerPool is a space-separated list of emoji runes for worker/lane decoration.
	// Recommended: at least 10 emojis for variety across lanes/workers.
	WorkerPool string `mapstructure:"worker-emoji-pool"`

	// JobPool is a space-separated list of emoji runes for job decoration.
	// Each new job arrival gets the next emoji from this pool.
	// Recommended: at least 6 emojis for variety.
	JobPool string `mapstructure:"job-emoji-pool"`

	// Separator between emoji and content info (default: " ").
	Separator string `mapstructure:"separator"`

	// AnimationData holds animation type configurations loaded from config.
	// These are loaded on-demand when Highway view is activated.
	AnimationData HighwayAnimationConfig `mapstructure:"animation,omitempty"`

	// AnimationGradient specifies which gradient to apply to highway animation frames.
	// Must match a key in palette.highlights.gradients (e.g., "aurora-borealis").
	// When empty, no gradient is applied and default styling is used.
	AnimationGradient string `mapstructure:"animation-gradient,omitempty"`
}

// HighwayAnimationConfig holds animation data configuration for Highway view.
type HighwayAnimationConfig struct {
	// Spinners holds per-spinner configuration.
	Spinners SpinnerAnimationConfig `mapstructure:"spinners,omitempty"`
}

// SpinnerAnimationConfig groups all spinner-type configurations.
type SpinnerAnimationConfig struct {
	// Enabled lists which spinner types should be loaded on demand.
	// All registered spinner names are valid. Only spinners in this list
	// will be loaded when Highway view starts.
	Enabled []string `mapstructure:"enabled"`

	// Override holds per-spinner config overrides keyed by spinner name.
	// Only entries listed in Enabled are consulted. The most useful override
	// is "interval" (milliseconds), which controls the animation speed of
	// that spinner in the highway view. Larger = slower (e.g. 5000 advances
	// the frame every 5 seconds). Value 0 or omitted = full speed (advances
	// every global tick, ~50ms).
	// Example:
	//   spinners:
	//     enabled: [wave]
	//     override:
	//       wave:
	//         interval: 5000   # wave cycles once every 5 seconds
	Override map[string]*SpinnerItemConfig `mapstructure:"override,omitempty"`
}

// SpinnerItemConfig holds per-spinner settings.
type SpinnerItemConfig struct {
	// Interval is the time between frames in milliseconds.
	Interval int `mapstructure:"interval"`
}
