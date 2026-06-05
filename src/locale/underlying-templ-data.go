package locale

import (
	lingo "github.com/snivilised/li18ngo/locale"
	"github.com/snivilised/li18ngo/locale/enums"
)

//go:generate lingo

// underliers is the single source of truth for all i18n messages in this
// package. Edit this map and run go generate to regenerate the auto files.
var _ = lingo.Underliers{
	// -------------------------------------------------------------------------
	// flags: Cobra messages
	// -------------------------------------------------------------------------

	"tui-flag-description": {
		MessageID:   "tui-flag-description",
		Seed:        "TuiFlagDesc",
		TypeName:    enums.UnderlyingTypeStaticCobra,
		Description: "Cobra flag description for TUI flag",
		Story:       "Cobra flag description for TUI flag",
		Other:       "tui denotes what view to use (default: linear)",
		File:        "flags",
	},

	"theme-flag-description": {
		MessageID:   "theme-flag-description",
		Seed:        "ThemeFlagDesc",
		TypeName:    enums.UnderlyingTypeDynamicCobra,
		Description: "Cobra flag description for theme flag",
		Story:       "Cobra flag description for theme flag",
		Other:       "theme denotes colour theme selection (default: system), loaded from '{{.Path}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Path",
				GoType: "string",
				Tale:   "The path to the colour themes file",
			},
		},
		File: "flags",
	},

	"subscribe-flag-description": {
		MessageID: "subscribe-flag-description",
		Seed:      "SubscribeFlagDesc",
		TypeName:  enums.UnderlyingTypeStaticCobra,
		Description: "Cobra flag description for subscribe flag; values are static " +
			"so do not translate them",
		Story: "Cobra flag description for subscribe flag",
		Other: "subscribe denotes the node types to visit: 'files', 'dirs' or 'all' (default)",
		File:  "flags",
	},

	"action-flag-description": {
		MessageID:   "action-flag-description",
		Seed:        "ActionFlagDesc",
		TypeName:    enums.UnderlyingTypeStaticCobra,
		Description: "Cobra flag description for action flag",
		Story:       "Cobra flag description for action flag",
		Other: "action denotes the executable or script function to invoke for " +
			"each matching node, as defined in config",
		File: "flags",
	},

	"pipeline-flag-description": {
		MessageID:   "pipeline-flag-description",
		Seed:        "PipelineFlagDesc",
		TypeName:    enums.UnderlyingTypeStaticCobra,
		Description: "Cobra flag description for pipeline flag",
		Story:       "Cobra flag description for pipeline flag",
		Other: "pipeline denotes defines a sequence of actions to invoke for " +
			"each matching node, as defined in config",
		File: "flags",
	},

	"resume-flag-description": {
		MessageID: "resume-flag-description",
		Seed:      "ResumeFlagDesc",
		TypeName:  enums.UnderlyingTypeStaticCobra,
		Description: "Cobra flag description for resume flag; values are static" +
			"so do not translate them",
		Story: "Cobra flag description for resume flag",
		Other: "resume denotes the strategy to use to continue a previously interrupted " +
			"navigation: 'spawn' or 'ff'",
		File: "flags",
	},

	"dry-run-flag-description": {
		MessageID:   "dry-run-flag-description",
		Seed:        "DryRunFlagDesc",
		TypeName:    enums.UnderlyingTypeStaticCobra,
		Description: "Cobra flag description for dry-run flag",
		Story:       "Cobra flag description for dry-run flag",
		Other:       "dry-run shows what would be executed without actually running the commands",
		File:        "flags",
	},

	"include-hidden-flag-description": {
		MessageID:   "include-hidden-flag-description",
		Seed:        "IncludeHiddenFlagDesc",
		TypeName:    enums.UnderlyingTypeStaticCobra,
		Description: "Cobra flag description for include-hidden flag",
		Story:       "Cobra flag description for include-hidden flag",
		Other:       "include-hidden includes hidden files (dot files) and known metadata files in the traversal",
		File:        "flags",
	},

	// -------------------------------------------------------------------------
	// root-cmd: Cobra messages
	// -------------------------------------------------------------------------

	"root-command-short-description": {
		MessageID:   "root-command-short-description",
		Seed:        "RootCmdShortDesc",
		TypeName:    enums.UnderlyingTypeStaticCobra,
		Description: "Navigates a directory tree",
		Story: "RootCmdShortDesc is the short description shown in" +
			" cobra help output for the root command.",
		Other: "Navigates a directory tree",
		File:  "root-cmd",
	},

	"root-command-long-description": {
		MessageID:   "root-command-long-description",
		Seed:        "RootCmdLongDesc",
		TypeName:    enums.UnderlyingTypeStaticCobra,
		Description: "Navigates a directory tree and shows matching nodes",
		Story: "RootCmdLongDesc is the long description shown in" +
			" cobra help output for the root command.",
		Other: `Navigates a directory tree and shows matching nodes. The root command
does not take any action for each visited node. It is simply a visualiser for
traversal.`,
		File: "root-cmd",
	},

	"root-command-config-file-usage": {
		MessageID:   "root-command-config-file-usage",
		Seed:        "RootCmdConfigFileUsage",
		TypeName:    enums.UnderlyingTypeDynamicCobra,
		Description: "root command config flag usage",
		Story: "RootCmdConfigFileUsage is the usage string for the" +
			" config file flag on the root command.",
		Other: "Config file (default is $HOME/{{.ConfigFileName}}.yml)",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "ConfigFileName",
				GoType: "string",
				Tale:   "is the base name of the config file without extension",
			},
		},
		File: "root-cmd",
	},

	"root-command-language-usage": {
		MessageID:   "root-command-language-usage",
		Seed:        "RootCmdLangUsage",
		TypeName:    enums.UnderlyingTypeStaticCobra,
		Description: "root command lang usage",
		Story: "RootCmdLangUsage is the usage string for the" +
			" language flag on the root command.",
		Other: "'lang' defines the language according to IETF BCP 47",
		File:  "root-cmd",
	},

	// -------------------------------------------------------------------------
	// walk-cmd: Cobra messages
	// -------------------------------------------------------------------------

	"walk-command-short-description": {
		MessageID:   "walk-command-short-description",
		Seed:        "WalkCmdShortDesc",
		TypeName:    enums.UnderlyingTypeStaticCobra,
		Description: "Navigates a directory tree on a single cpu core",
		Story: "RootCmdShortDesc is the short description shown in" +
			" cobra help output for the root command.",
		Other: "Executes actions and pipelines for matching nodes on a single cpu core",
		File:  "walk-cmd",
	},

	"walk-command-long-description": {
		MessageID: "walk-command-long-description",
		Seed:      "WalkCmdLongDesc",
		TypeName:  enums.UnderlyingTypeStaticCobra,
		Description: "Navigates a directory tree on a single cpu core and executes " +
			"actions and pipelines for matching nodes",
		Story: "WalkCmdLongDesc is the long description shown in" +
			" cobra help output for the walk command.",
		Other: `Navigates a directory tree and executes actions and pipelines for 
matching nodes. All actions are executed sequentially on a
single cpu core. This is the simplest and most compatible
execution mode, but may be slower for large traversals, depending
on the configured actions and pipelines. Actions support invocation
of any external command found on PATH, shell built-in function or
script.
Use --action or --pipeline to name a config-defined operation.
`,
		File: "walk-cmd",
	},

	// -------------------------------------------------------------------------
	// sprint-cmd: Cobra messages
	// -------------------------------------------------------------------------

	"sprint-command-short-description": {
		MessageID:   "sprint-command-short-description",
		Seed:        "SprintCmdShortDesc",
		TypeName:    enums.UnderlyingTypeStaticCobra,
		Description: "Navigates a directory tree on a single cpu core",
		Story: "RootCmdShortDesc is the short description shown in" +
			" cobra help output for the root command.",
		Other: "Executes actions and pipelines for matching nodes on a single cpu core",
		File:  "sprint-cmd",
	},

	"sprint-command-long-description": {
		MessageID: "sprint-command-long-description",
		Seed:      "SprintCmdLongDesc",
		TypeName:  enums.UnderlyingTypeStaticCobra,
		Description: "Navigates a directory tree on multiple single cpu cores and executes " +
			"actions and pipelines for matching nodes",
		Story: "SprintCmdLongDesc is the long description shown in" +
			" cobra help output for the sprint command.",
		Other: `Navigates a directory tree and executes actions and pipelines for 
matching nodes. All actions are executed concurrently on multiple
cpu cores. All available cpu cores can be utilised by specifying
the --cpu flag, or a specific set of cores can be targeted using
the --now flag. The sprint command is designed to be run when bulk
processing of a directory tree is required where the configured
action is heavily IO bound. Actions and pipelines defined for walk
run in the same way as they do for sprint.
Use --action or --pipeline to name a config-defined operation.
`,
		File: "sprint-cmd",
	},

	// -------------------------------------------------------------------------
	// query-cmd: Cobra messages
	// -------------------------------------------------------------------------

	"query-command-short-description": {
		MessageID:   "query-command-short-description",
		Seed:        "QueryCmdShortDesc",
		TypeName:    enums.UnderlyingTypeStaticCobra,
		Description: "query acts as a non destructive directory tree inquiry",
		Story: "QueryCmdShortDesc is the short description shown in" +
			" cobra help output for the query command.",
		Other: "query navigates the directory tree via a walk traversal",
		File:  "query-cmd",
	},

	"query-command-long-description": {
		MessageID:   "query-command-long-description",
		Seed:        "QueryCmdLongDesc",
		TypeName:    enums.UnderlyingTypeStaticCobra,
		Description: "query acts as a non destructive directory tree inquiry",
		Story: "QueryCmdLongDesc is the long description shown in" +
			" cobra help output for the nav command.",
		Other: "query command walks the directory tree showing nodes that " +
			" satisfy filtering criteria when present. Can be used as a kind of dry " +
			"run traversal, except no actions or pipelines are invoked as in the case " +
			"for walk and sprint commands.",
		File: "query-cmd",
	},

	// -------------------------------------------------------------------------
	// ghost-cmds: Cobra messages
	// -------------------------------------------------------------------------

	"command-is-a-ghost": {
		MessageID:   "command-is-a-ghost",
		Seed:        "CommandIsAGhost",
		TypeName:    enums.UnderlyingTypeDynamicCobra,
		Description: "command is a ghost",
		Story:       "CommandIsAGhost is use for the long and short text of a ghost command",
		Other:       "{{.Command}} is a ghost",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Command",
				GoType: "string",
				Tale:   "is the name of the ghost command",
			},
		},
		File: "ghost-cmds",
	},

	// -------------------------------------------------------------------------
	// ghost-cmds: General messages
	// -------------------------------------------------------------------------

	"command-is-not-user-invocable-prompt": {
		MessageID:   "command-is-not-user-invocable-prompt",
		Seed:        "CommandIsNotUserInvocablePrompt",
		TypeName:    enums.UnderlyingTypeDynamicGeneral,
		Description: "command is not a user invocable",
		Story: "CommandIsNotUserInvocablePrompt is the message shown during startup" +
			" if the user accidentally invokes this ghost sub command, before exiting.",
		Other: "{{.Command}} is hidden and not designed to be directly invoked",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Command",
				GoType: "string",
				Tale:   "is the name of the ghost command",
			},
		},
		File: "ghost-cmds",
	},

	// -------------------------------------------------------------------------
	// shell:Error messages
	// -------------------------------------------------------------------------

	"psm-set-no-powershell-exe-found.static-error": {
		MessageID:   "psm-set-no-powershell-exe-found.static-error",
		Seed:        "PSMSetNoPowerShellExeFound",
		TypeName:    enums.UnderlyingTypeStaticErrorWrapperMsg,
		Description: "PSModulePath is set but no PowerShell executable found",
		Story:       "PSModulePath is set but no PowerShell executable found",
		Other:       "PSModulePath is set but no PowerShell executable found: '{{.Wrapped}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Wrapped",
				GoType: "error",
				Tale:   "The underlying error representing the failure to find a PowerShell executable",
			},
		},
		File: "shell",
	},

	"cmd-not-found-in-env.dynamic-error": {
		MessageID:   "cmd-not-found-in-env.dynamic-error",
		Seed:        "CmdNotFoundInEnv",
		TypeName:    enums.UnderlyingTypeDynamicError,
		Description: "Command not found in environment error",
		Story: "Error returned when a command specified in the config is not " +
			"found in the current environment",
		Other: "'{{.Command}}' Not found in current environment",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Command",
				GoType: "string",
				Tale:   "The command that was not found",
			},
		},
		File: "shell",
	},

	"neither-pwsh-or-powershell-exe-found.static-error": {
		MessageID:   "neither-pwsh-or-powershell-exe-found.static-error",
		Seed:        "NeitherPwshOrPowerShellExeFound",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "neither pwsh.exe nor powershell.exe found on PATH",
		Story:       "neither pwsh.exe nor powershell.exe found on PATH",
		Other:       "Neither pwsh.exe nor powershell.exe found on PATH",
		File:        "shell",
	},

	"cmd-not-found-as-path-binary-or-builtin.dynamic-error": {
		MessageID: "cmd-not-found-as-path-binary-or-builtin.dynamic-error",
		Seed:      "CmdNotFoundAsPathBinaryOrBuiltin",
		TypeName:  enums.UnderlyingTypeDynamicError,
		Description: "Command not found in environment as a PATH binary" +
			" or cmd.exe builtin error",
		Story: "Error returned when a command specified in the config is not " +
			"found in the current environment as PATH binary or cmd.exe builtin",
		Other: "'{{.Command}}' Not found in current environment",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Command",
				GoType: "string",
				Tale:   "The command that was not found",
			},
		},
		File: "shell",
	},

	// -------------------------------------------------------------------------
	// ui: Error messages
	// -------------------------------------------------------------------------

	"porthole-view-not-supported-by-sprint.static-error": {
		MessageID:   "porthole-view-not-supported-by-sprint.static-error",
		Seed:        "PortholeViewNotSupportedBySprint",
		TypeName:    enums.UnderlyingTypeStaticErrorWrapperMsg,
		Description: "Porthole view not supported by sprint command error",
		Story:       "Error returned when porthole view is requested via the sprint command",
		Other:       "{{.Wrapped}}",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Wrapped",
				GoType: "error",
				Tale:   "The underlying error representing the failure to use porthole with sprint",
			},
		},
		File: "ui",
	},

	// -------------------------------------------------------------------------
	// config: General messages
	// -------------------------------------------------------------------------

	"using-config-file": {
		MessageID:   "using-config-file",
		Seed:        "UsingConfigFile",
		TypeName:    enums.UnderlyingTypeDynamicGeneral,
		Description: "Message to indicate which config is being used",
		Story: "UsingConfigFile is printed on startup to indicate" +
			" which configuration file has been loaded.",
		Other: "Using config file: '{{.ConfigFileName}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "ConfigFileName",
				GoType: "string",
				Tale:   "is the name of the config file being used",
			},
		},
		File: "config",
	},

	// -------------------------------------------------------------------------
	// config: Error messages
	// -------------------------------------------------------------------------

	// TODO: rename bedrock errors; shouldn't use bedrock in the name
	"bedrock-load-viper-setup.static-error": {
		MessageID: "bedrock-load-viper-setup.static-error",
		Seed:      "BedrockLoadViperSetup",
		TypeName:  enums.UnderlyingTypeStaticErrorWrapper,
		Description: "Error returned when viper setup fails" +
			" during bedrock.Load",
		Story: "BedrockLoadViperSetup indicates that viper could not" +
			" be configured during the bedrock.Load call.",
		Other: "bedrock.Load: viper setup",
		File:  "config",
	},

	"bedrock-load-reading-config.static-error": {
		MessageID: "bedrock-load-reading-config.static-error",
		Seed:      "BedrockLoadReadingConfig",
		TypeName:  enums.UnderlyingTypeStaticErrorWrapper,
		Description: "Error returned when reading the config file" +
			" fails during bedrock.Load",
		Story: "BedrockLoadReadingConfig indicates that the" +
			" configuration file could not be read during the" +
			" bedrock.Load call.",
		Other: "bedrock.Load: reading config",
		File:  "config",
	},

	"bedrock-load-decoding.static-error": {
		MessageID: "bedrock-load-decoding.static-error",
		Seed:      "BedrockLoadDecoding",
		TypeName:  enums.UnderlyingTypeStaticErrorWrapper,
		Description: "Error returned when config decoding fails" +
			" during bedrock.Load",
		Story: "BedrockLoadDecoding indicates that the configuration" +
			" could not be decoded into the target struct during the" +
			" bedrock.Load call.",
		Other: "bedrock.Load: decoding",
		File:  "config",
	},

	"bedrock-load-validation.static-error": {
		MessageID: "bedrock-load-validation.static-error",
		Seed:      "BedrockLoadValidation",
		TypeName:  enums.UnderlyingTypeStaticErrorWrapper,
		Description: "Error returned when config validation fails" +
			" during bedrock.Load",
		Story: "BedrockLoadValidation indicates that the decoded" +
			" configuration failed validation during the bedrock.Load call.",
		Other: "bedrock.Load: validation",
		File:  "config",
	},

	"unsupported-format.dynamic-error": {
		MessageID: "unsupported-format.dynamic-error",
		Seed:      "UnsupportedFormat",
		TypeName:  enums.UnderlyingTypeDynamicError,
		Description: "Error returned when an unregistered config" +
			" format is requested",
		Story: "UnsupportedFormat indicates that the configuration" +
			" format requested by the caller has not been registered" +
			" with the format registry.",
		Other: "Unsupported format '{{.Format}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Format",
				GoType: "string",
				Tale:   "is the unsupported format",
			},
		},
		File: "config",
	},

	"creating-decoder-for.dynamic-error": {
		MessageID: "creating-decoder-for.dynamic-error",
		Seed:      "CreatingDecoderFor",
		TypeName:  enums.UnderlyingTypeDynamicErrorWrapper,
		Description: "Error returned when a mapstructure decoder" +
			" cannot be created for a config section",
		Story: "CreatingDecoderFor indicates that a mapstructure" +
			" decoder could not be constructed for the named" +
			" configuration section.",
		Other: "Creating decoder for '{{.Key}}': '{{.Wrapped}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Wrapped",
				GoType: "error",
				Tale:   "is the underlying error",
			},
			{
				Note:   "Key",
				GoType: "string",
				Tale:   "is the configuration section name",
			},
		},
		File: "config",
	},

	"decoding-section.dynamic-error": {
		MessageID: "decoding-section.dynamic-error",
		Seed:      "DecodingSection",
		TypeName:  enums.UnderlyingTypeDynamicErrorWrapper,
		Description: "Error returned when mapstructure decoding of" +
			" a config section fails",
		Story: "DecodingSection indicates that mapstructure failed" +
			" to decode the named configuration section into its" +
			" target struct.",
		Other: "Decoding section '{{.Key}}': '{{.Wrapped}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Wrapped",
				GoType: "error",
				Tale:   "is the underlying error",
			},
			{
				Note:   "Key",
				GoType: "string",
				Tale:   "is the configuration section name",
			},
		},
		File: "config",
	},

	"flags-section-unexpected-type.dynamic-error": {
		MessageID: "flags-section-unexpected-type.dynamic-error",
		Seed:      "FlagsSectionUnexpectedType",
		TypeName:  enums.UnderlyingTypeDynamicError,
		Description: "Error returned when the flags config section" +
			" has an unexpected type",
		Story: "FlagsSectionUnexpectedType indicates that the flags" +
			" section of the configuration file was decoded into an" +
			" unexpected Go type.",
		Other: "Flags section has unexpected type '{{.TypeName}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "TypeName",
				GoType: "string",
				Tale:   "is the unexpected type",
			},
		},
		File: "config",
	},

	"action-not-found.dynamic-error": {
		MessageID:   "action-not-found.dynamic-error",
		Seed:        "ActionNotFound",
		TypeName:    enums.UnderlyingTypeDynamicError,
		Description: "Action not found in config error",
		Story: "ActionNotFound indicates that an action with the specified" +
			" name was not found in the traversal configuration.",
		Other: "Action not found: '{{.Action}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Action",
				GoType: "string",
				Tale:   "is the action name that was not found",
			},
		},
		File: "config",
	},

	"action-has-empty-cmd.dynamic-error": {
		MessageID:   "action-has-empty-cmd.dynamic-error",
		Seed:        "ActionHasEmptyCmd",
		TypeName:    enums.UnderlyingTypeDynamicError,
		Description: "Action has empty cmd string error",
		Story: "ActionHasEmptyCmd indicates that an action with the specified" +
			" name has an empty command string.",
		Other: "Action has empty cmd string: '{{.Action}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Action",
				GoType: "string",
				Tale:   "is the action name that has an empty cmd string",
			},
		},
		File: "config",
	},

	"pipeline-not-found.dynamic-error": {
		MessageID:   "pipeline-not-found.dynamic-error",
		Seed:        "PipelineNotFound",
		TypeName:    enums.UnderlyingTypeDynamicError,
		Description: "Pipeline not found in config error",
		Story: "PipelineNotFound indicates that a pipeline with the specified" +
			" name was not found in the traversal configuration.",
		Other: "Pipeline not found: '{{.Pipeline}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Pipeline",
				GoType: "string",
				Tale:   "is the pipeline name that was not found",
			},
		},
		File: "config",
	},

	// -------------------------------------------------------------------------
	// filter: Error messages
	// -------------------------------------------------------------------------

	"filter-is-nil.static-error": {
		MessageID:   "filter-is-nil.static-error",
		Seed:        "FilterIsNil",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "filter is nil error",
		Story: "FilterIsNil indicates that the caller passed a nil" +
			" filter reference where a concrete filter implementation" +
			" was required.",
		Other: "Filter is nil",
		File:  "filter",
	},

	"filter-missing-type.static-error": {
		MessageID:   "filter-missing-type.static-error",
		Seed:        "FilterMissingType",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "filter missing type",
		Story: "FilterMissingType indicates that the filter definition" +
			" is missing a required type field.",
		Other: "Filter missing type",
		File:  "filter",
	},

	"custom-filter-not-supported-for-children.static-error": {
		MessageID:   "custom-filter-not-supported-for-children.static-error",
		Seed:        "FilterCustomNotSupported",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "custom filter not supported for children",
		Story: "FilterCustomNotSupported indicates that custom filters" +
			" cannot be applied to child nodes in this context.",
		Other: "Custom filter not supported for children",
		File:  "filter",
	},

	"glob-ex-filter-not-supported-for-children.static-error": {
		MessageID:   "glob-ex-filter-not-supported-for-children.static-error",
		Seed:        "FilterChildGlobExNotSupported",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "glob-ex filter not supported for children",
		Story: "FilterChildGlobExNotSupported indicates that glob-ex" +
			" filters cannot be applied to child nodes in this context.",
		Other: "Glob-ex filter not supported for children",
		File:  "filter",
	},

	"filter-is-undefined.static-error": {
		MessageID:   "filter-is-undefined.static-error",
		Seed:        "FilterUndefined",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "filter is undefined error",
		Story: "FilterUndefined indicates that the filter referenced" +
			" in the traversal options has not been defined.",
		Other: "Filter is undefined",
		File:  "filter",
	},

	"invalid-extended-glob-pattern.dynamic-error": {
		MessageID:   "invalid-extended-glob-pattern.dynamic-error",
		Seed:        "InvalidExtendedGlobPattern",
		TypeName:    enums.UnderlyingTypeDynamicError,
		Description: "invalid extended glob pattern",
		Story:       "InvalidExtendedGlobPattern indicates that the extended glob pattern is invalid.",
		Other:       "Invalid extended glob pattern: '{{.Pattern}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Pattern",
				GoType: "string",
				Tale:   "Is the invalid filter pattern",
			},
		},
		File: "filter",
	},

	// -------------------------------------------------------------------------
	// filter: Error messages
	// -------------------------------------------------------------------------

	"missing-custom-filter-definition.static-error": {
		MessageID:   "missing-custom-filter-definition.static-error",
		Seed:        "MissingCustomFilterDefinition",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "config error missing-custom-filter-definition",
		Story: "MissingCustomFilterDefinition indicates that the" +
			" traversal configuration references a custom filter but" +
			" no definition for it was found.",
		Other: "Missing custom filter definition (config error)",
		File:  "filter",
	},

	"invalid-glob-ex-filter-missing-separator.dynamic-error": {
		MessageID: "invalid-glob-ex-filter-missing-separator" +
			".dynamic-error",
		Seed:     "InvalidExtGlobFilterMissingSeparator",
		TypeName: enums.UnderlyingTypeDynamicError,
		Description: "invalid glob ex filter definition;" +
			" pattern is missing separator",
		Story: "InvalidExtGlobFilterMissingSeparator indicates that" +
			" an extended glob filter definition is invalid because" +
			" the pattern is missing the required separator character.",
		Other: "Extended glob pattern missing separator, pattern: '{{.Pattern}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Pattern",
				GoType: "string",
				Tale:   "is the invalid filter pattern",
			},
		},
		File: "filter",
	},

	"invalid-extended-glob-filter-missing-separator.sentinel-error": {
		MessageID: "invalid-extended-glob-filter-missing-separator" +
			".sentinel-error",
		Seed:     "CoreInvalidExtGlobFilterMissingSeparator",
		TypeName: enums.UnderlyingTypeSentinelError,
		Description: "invalid glob ex filter definition;" +
			" pattern is missing separator",
		Story: "CoreInvalidExtGlobFilterMissingSeparator is the" +
			" sentinel core error for an invalid extended glob filter" +
			" definition. Wrap this error using" +
			" NewInvalidExtGlobFilterMissingSeparatorError.",
		Other: "Invalid glob ex filter definition;" +
			" pattern is missing separator",
		File: "filter",
	},

	"poly-filter-is-invalid.static-error": {
		MessageID:   "poly-filter-is-invalid.static-error",
		Seed:        "PolyFilterIsInvalid",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "poly filter definition is invalid error",
		Story: "PolyFilterIsInvalid indicates that a poly filter" +
			" definition fails validation.",
		Other: "poly filter definition is invalid",
		File:  "filter",
	},

	"failed-to-get-navigator-driver.static-error": {
		MessageID:   "failed-to-get-navigator-driver.static-error",
		Seed:        "InternalFailedToGetNavigatorDriver",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "failed to get navigator driver",
		Story: "InternalFailedToGetNavigatorDriver indicates an" +
			" internal failure when resolving the navigator driver." +
			" This is not expected during normal operation.",
		Other: "Failed to get navigator driver",
	},

	// TODO: Need to to check where this is invoked from, may not be required.
	"invalid-incase-filter-definition.dynamic-error": {
		MessageID: "invalid-incase-filter-definition.dynamic-error",
		Seed:      "InvalidInCaseFilterDef",
		TypeName:  enums.UnderlyingTypeDynamicErrorWrapper,
		Description: "invalid incase filter definition; pattern is" +
			" missing separator wrapper error",
		Story: "InvalidInCaseFilterDef indicates that a case-insensitive" +
			" filter definition is invalid because the pattern is missing" +
			" the required separator.",
		Other: "'{{.Wrapped}}', pattern: '{{.Pattern}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Wrapped",
				GoType: "error",
				Tale:   "is the underlying error",
			},
			{
				Note:   "Pattern",
				GoType: "string",
				Tale:   "is the invalid filter pattern",
			},
		},
		File: "filter",
	},

	// TODO: Need to to check where this is invoked from, may not be required.
	"invalid-incase-filter-definition.sentinel-error": {
		MessageID: "invalid-incase-filter-definition.sentinel-error",
		Seed:      "CoreInvalidInCaseFilterDef",
		TypeName:  enums.UnderlyingTypeSentinelError,
		Description: "invalid incase filter definition; pattern is" +
			" missing separator core error",
		Story: "CoreInvalidInCaseFilterDef is the sentinel core error" +
			" for an invalid case-insensitive filter definition. Wrap" +
			" this error using NewInvalidInCaseFilterDefError.",
		Other: "Invalid incase filter definition;" +
			" pattern is missing separator",
		File: "filter",
	},

	// -------------------------------------------------------------------------
	// sample: Error messages
	// -------------------------------------------------------------------------

	"invalid-file-sampling-spec-missing-files.static-error": {
		MessageID: "invalid-file-sampling-spec-missing-files.static-error",
		Seed:      "InvalidFileSamplingSpecMissingFiles",
		TypeName:  enums.UnderlyingTypeStaticError,
		Description: "invalid file sampling specification," +
			" missing no of files",
		Story: "InvalidFileSamplingSpecMissingFiles indicates that" +
			" the file sampling specification is invalid because the" +
			" required number-of-files field is absent.",
		Other: "Invalid file sampling specification," +
			" missing no of files",
		File: "sample",
	},

	"invalid-file-sampling-spec-missing-directories.static-error": {
		MessageID: "invalid-file-sampling-spec-missing-directories.static-error",
		Seed:      "InvalidSamplingSpecMissingDirectories",
		TypeName:  enums.UnderlyingTypeStaticError,
		Description: "invalid file sampling specification," +
			" missing no of directories",
		Story: "InvalidSamplingSpecMissingDirectories indicates that" +
			" the file sampling specification is invalid because the" +
			" required number-of-directories field is absent.",
		Other: "Invalid file sampling specification," +
			" missing no of directories",
		File: "sample",
	},

	// -------------------------------------------------------------------------
	// General messages
	// -------------------------------------------------------------------------

	"node-visited": {
		MessageID:   "node-visited",
		Seed:        "NodeVisited",
		TypeName:    enums.UnderlyingTypeDynamicGeneral,
		Description: "Printed for each node visited during traversal",
		Story: "NodeVisited is printed for each filesystem node" +
			" encountered during traversal.",
		Other: "Node Visited -> '{{.Path}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Path",
				GoType: "string",
				Tale:   "is the path of the node being visited",
			},
		},
	},

	"action-failed": {
		MessageID: "action-failed",
		Seed:      "ActionFailed",
		TypeName:  enums.UnderlyingTypeDynamicGeneral,
		Description: "Printed when an action fails on a node" +
			" during traversal",
		Story: "ActionFailed is printed when a named action returns" +
			" an error while processing a node.",
		Other: "[!] Action '{{.Name}}' failed on '{{.Path}}': '{{.Err}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Name",
				GoType: "string",
				Tale:   "is the name of the action that failed",
			},
			{
				Note:   "Path",
				GoType: "string",
				Tale:   "is the path of the node on which the action failed",
			},
			{
				Note:   "Err",
				GoType: "string",
				Tale:   "is the error message",
			},
		},
	},

	"action-visited": {
		MessageID: "action-visited",
		Seed:      "ActionVisited",
		TypeName:  enums.UnderlyingTypeDynamicGeneral,
		Description: "Printed for each node successfully processed" +
			" by an action",
		Story: "ActionVisited is printed for each node successfully" +
			" processed by a named action.",
		Other: "[+] Actioned for '{{.Name}}' at '{{.Path}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Name",
				GoType: "string",
				Tale:   "is the name of the action that succeeded",
			},
			{
				Note:   "Path",
				GoType: "string",
				Tale:   "is the path of the node successfully processed",
			},
		},
	},

	"pipeline-failed": {
		MessageID: "pipeline-failed",
		Seed:      "PipelineFailed",
		TypeName:  enums.UnderlyingTypeDynamicGeneral,
		Description: "Printed when a pipeline fails on a node" +
			" during traversal",
		Story: "PipelineFailed is printed when a named pipeline" +
			" returns an error while processing a node.",
		Other: "[!] Pipeline '{{.Name}}' failed on '{{.Path}}': '{{.Err}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Name",
				GoType: "string",
				Tale:   "is the name of the pipeline that failed",
			},
			{
				Note:   "Path",
				GoType: "string",
				Tale:   "is the path of the node on which the pipeline failed",
			},
			{
				Note:   "Err",
				GoType: "string",
				Tale:   "is the error message",
			},
		},
	},

	"pipeline-visited": {
		MessageID: "pipeline-visited",
		Seed:      "PipelineVisited",
		TypeName:  enums.UnderlyingTypeDynamicGeneral,
		Description: "Printed for each node successfully processed" +
			" by a pipeline",
		Story: "PipelineVisited is printed for each node successfully" +
			" processed by a named pipeline.",
		Other: "[+] Pipeline succeeded for '{{.Name}}' at '{{.Path}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Name",
				GoType: "string",
				Tale:   "is the name of the pipeline that succeeded",
			},
			{
				Note:   "Path",
				GoType: "string",
				Tale:   "is the path of the node successfully processed",
			},
		},
	},

	"traversal-failed": {
		MessageID:   "traversal-failed",
		Seed:        "TraversalFailed",
		TypeName:    enums.UnderlyingTypeDynamicGeneral,
		Description: "Printed when the traversal itself fails",
		Story: "TraversalFailed is printed when the traversal" +
			" operation itself encounters a fatal error.",
		Other: "[!] Traversal failed: '{{.Err}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Err",
				GoType: "string",
				Tale:   "Is the error message",
			},
		},
	},

	"action-skipped": {
		MessageID:   "action-skipped",
		Seed:        "ActionSkipped",
		TypeName:    enums.UnderlyingTypeDynamicGeneral,
		Description: "Printed when an action is skipped",
		Story:       "ActionSkipped is printed when an action is skipped",
		Other:       "[!] Action '{{.Name}}' skipped at '{{.Path}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Name",
				GoType: "string",
				Tale:   "Name of the action that was skipped",
			},
			{
				Note:   "Path",
				GoType: "string",
				Tale:   "Full path of the node that action that was skipped for",
			},
		},
	},

	"placeholder-breach": {
		MessageID:   "placeholder-breach",
		Seed:        "PlaceholderBreach",
		TypeName:    enums.UnderlyingTypeDynamicGeneral,
		Description: "Printed when placeholder breach occurs during expansion",
		Story:       "Printed when a placeholder breach occurs during expansion.",
		Other:       "[!] Placeholder: '{{.Placeholder}}' resolved to '{{.ResolvedPath}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Placeholder",
				GoType: "string",
				Tale:   "The placeholder that breached above the traversal root",
			},
			{
				Note:   "ResolvedPath",
				GoType: "string",
				Tale:   "The resolved path of the placeholder that breached above the traversal root",
			},
		},
	},

	"traversal-complete": {
		MessageID:   "traversal-complete",
		Seed:        "TraversalComplete",
		TypeName:    enums.UnderlyingTypeDynamicGeneral,
		Description: "Printed on successful completion of a traversal",
		Story: "TraversalComplete is printed when a traversal finishes" +
			" successfully, summarising the nodes visited and time elapsed.",
		Other: "[+] Traversal complete successfully: {{.Files}} files, {{.Dirs}} dirs" +
			" visited in {{.Elapsed}}",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Files",
				GoType: "uint",
				Tale:   "is the total number of files visited",
			},
			{
				Note:   "Dirs",
				GoType: "uint",
				Tale:   "is the total number of directories visited",
			},
			{
				Note:   "Elapsed",
				GoType: "string",
				Tale:   "is the total time elapsed during the traversal",
			},
		},
	},

	// -------------------------------------------------------------------------
	// Error messages
	// -------------------------------------------------------------------------

	"pipeline-preflight-failure.dynamic-error": {
		MessageID:   "pipeline-preflight-failure.dynamic-error",
		Seed:        "PipelinePreflightFailure",
		TypeName:    enums.UnderlyingTypeDynamicErrorWrapper,
		Description: "Error occurred during pipeline preflight checks",
		Story:       "Error occurred during pipeline preflight checks",
		Other:       "'{{.Wrapped}}' Preflight check failed for pipeline: '{{.Pipeline}}' ",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Pipeline",
				GoType: "string",
				Tale:   "The name of the pipeline for which the preflight check failed",
			},
			{
				Note:   "Wrapped",
				GoType: "error",
				Tale:   "The underlying error that caused the preflight check to fail",
			},
		},
	},

	"failed-to-create-worker-pool.static-error": {
		MessageID:   "failed-to-create-worker-pool.static-error",
		Seed:        "WorkerPoolCreationFailed",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "failed to create worker pool",
		Story: "WorkerPoolCreationFailed indicates that the worker" +
			" pool could not be initialised.",
		Other: "Failed to create worker pool",
	},

	"usage-missing-tree-path.static-error": {
		MessageID:   "usage-missing-tree-path.static-error",
		Seed:        "UsageMissingTreePath",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "usage missing tree path",
		Story: "UsageMissingTreePath indicates that the command was" +
			" invoked without the required tree path argument.",
		Other: "Usage missing tree path",
	},

	"usage-missing-restore-path.static-error": {
		MessageID:   "usage-missing-restore-path.static-error",
		Seed:        "UsageMissingRestorePath",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "usage missing restore path",
		Story: "UsageMissingRestorePath indicates that the command was" +
			" invoked without the required restore path argument.",
		Other: "Usage missing restore path",
	},

	"usage-missing-subscription.static-error": {
		MessageID:   "usage-missing-subscription.static-error",
		Seed:        "UsageMissingSubscription",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "usage missing subscription",
		Story: "UsageMissingSubscription indicates that the command" +
			" was invoked without specifying a subscription type.",
		Other: "Usage missing subscription",
	},

	"usage-missing-handler.static-error": {
		MessageID:   "usage-missing-handler.static-error",
		Seed:        "UsageMissingHandler",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "usage missing handler",
		Story: "UsageMissingHandler indicates that the command was" +
			" invoked without registering a required handler.",
		Other: "Usage missing handler",
	},

	"id-generator-func-cant-be-nil.static-error": {
		MessageID:   "id-generator-func-cant-be-nil.static-error",
		Seed:        "IDGeneratorFuncCantBeNil",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "id generator func is nil, should be defined",
		Story: "IDGeneratorFuncCantBeNil indicates that a nil function" +
			" was supplied where an ID generator func is required.",
		Other: "ID generator func can't be nil",
	},

	"un-equal-conversion.sentinel-error": {
		MessageID:   "un-equal-conversion.sentinel-error",
		Seed:        "UnEqualJSONConversion",
		TypeName:    enums.UnderlyingTypeSentinelError,
		Description: "JSON options conversion error",
		Story: "UnEqualJSONConversion indicates that a round-trip" +
			" JSON conversion produced a result that does not equal" +
			" the original value.",
		Other: "Unequal JSON conversion",
	},

	"invalid-path.dynamic-error": {
		MessageID:   "invalid-path.dynamic-error",
		Seed:        "InvalidPath",
		TypeName:    enums.UnderlyingTypeDynamicError,
		Description: "invalid path (dynamic error)",
		Story: "InvalidPath indicates that a path supplied by the" +
			" caller fails validation.",
		Other: "Invalid Path: '{{.Path}}'",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Path",
				GoType: "string",
				Tale:   "is the invalid path",
			},
		},
	},

	"traverse-fs-mismatch.static-error": {
		MessageID:   "traverse-fs-mismatch.static-error",
		Seed:        "TraverseFsMismatch",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "traverse fs mismatch error",
		Story: "TraverseFsMismatch indicates that the filesystem" +
			" passed to the traversal does not match the filesystem" +
			" recorded at the point the session was saved.",
		Other: "traverse-fs file system mismatch",
	},

	"resume-fs-mismatch.static-error": {
		MessageID:   "resume-fs-mismatch.static-error",
		Seed:        "ResumeFsMismatch",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "resume fs mismatch error",
		Story: "ResumeFsMismatch indicates that the filesystem passed" +
			" to a resume operation does not match the filesystem" +
			" recorded at the point the session was saved.",
		Other: "resume-fs file system mismatch",
	},

	"resume-fs-mismatch.sentinel-error": {
		MessageID:   "resume-fs-mismatch.sentinel-error",
		Seed:        "CoreResumeFsMismatch",
		TypeName:    enums.UnderlyingTypeSentinelError,
		Description: "core resume file system mismatch error",
		Story: "CoreResumeFsMismatch is the sentinel core error for" +
			" a filesystem mismatch detected during traversal or resume." +
			" Wrap using NewTraverseFsMismatchError or" +
			" NewResumeFsMismatchError.",
		Other: "Resume file system mismatch",
	},

	"panic-occurred.sentinel-error": {
		MessageID:   "panic-occurred.sentinel-error",
		Seed:        "CorePanicOccurred",
		TypeName:    enums.UnderlyingTypeSentinelError,
		Description: "core error",
		Story: "CorePanicOccurred is the sentinel core error indicating" +
			" that a panic was intercepted during traversal.",
		Other: "Panic occurred",
	},

	"invalid-subscription.static-error": {
		MessageID:   "invalid-subscription.static-error",
		Seed:        "InvalidSubscription",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "invalid subscription type",
		Story: "InvalidSubscription indicates that the subscription" +
			" type supplied by the caller is not one of the accepted values.",
		Other: "Invalid subscription type," +
			" must be one of: [files, dirs, all]",
	},

	"invalid-resume-strategy.static-error": {
		MessageID:   "invalid-resume-strategy.static-error",
		Seed:        "InvalidResumeStrategy",
		TypeName:    enums.UnderlyingTypeStaticError,
		Description: "invalid resume strategy type",
		Story: "InvalidResumeStrategy indicates that the resume strategy" +
			" supplied by the caller is not one of the accepted values.",
		Other: "Invalid resume strategy, must be one of: [spawn, fast]",
	},

	"traversal-saved.dynamic-error": {
		MessageID: "traversal-saved.dynamic-error",
		Seed:      "TraversalSaved",
		TypeName:  enums.UnderlyingTypeDynamicErrorWrapper,
		Description: "Error returned when a panic occurs during traversal" +
			" and save of the traversal state succeeds",
		Story: "Traversal state saved as a result of a panic occurring during traversal",
		Other: "'{{.Wrapped}}' Traversal state saved successfully to: '{{.SavedTo}}' ",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "SavedTo",
				GoType: "string",
				Tale:   "The path the traversal state was saved to",
			},
			{
				Note:   "Wrapped",
				GoType: "error",
				Tale:   "The underlying error representing the panic that occurred",
			},
		},
	},

	"traversal-not-saved.dynamic-error": {
		MessageID: "traversal-not-saved.dynamic-error",
		Seed:      "TraversalNotSaved",
		TypeName:  enums.UnderlyingTypeDynamicErrorWrapper,
		Description: "Error returned when a panic occurs during traversal" +
			" and save of the traversal state fails",
		Story: "Traversal state not saved as a result of a panic occurring during traversal",
		Other: "'{{.Wrapped}}' Failed to save traversal state to: '{{.SavedTo}}' ",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "SavedTo",
				GoType: "string",
				Tale:   "The path the traversal state was attempted to be saved to",
			},
			{
				Note:   "Wrapped",
				GoType: "error",
				Tale: "The underlying error representing the panic that occurred and the" +
					" failure to save",
			},
		},
	},

	"invalid-resume-value.dynamic-error": {
		MessageID:   "invalid-resume-value.dynamic-error",
		Seed:        "InvalidResumeValue",
		TypeName:    enums.UnderlyingTypeDynamicError,
		Description: "Invalid resume value format error",
		Story: "InvalidResumeValue indicates that the resume value supplied by the caller" +
			" is not in a valid value.",
		Other: "Invalid resume value, actual: '{{.Actual}}', must be: {{.Values}}",
		Fields: []lingo.UnderlyingField{
			{
				Note:   "Actual",
				GoType: "string",
				Tale:   "The resume value provided",
			},
			{
				Note:   "Values",
				GoType: "string",
				Tale:   "The valid values for resume, composed together, probably as CSV",
			},
		},
	},

	// words

	"prohibitive.word": {
		MessageID:   "prohibitive.word",
		Seed:        "Prohibitive",
		TypeName:    enums.UnderlyingTypeStaticGeneral,
		Description: "Word: prohibitive",
		Story:       "Word: prohibitive",
		Other:       "prohibitive",
	},

	"permissive.word": {
		MessageID:   "permissive.word",
		Seed:        "Permissive",
		TypeName:    enums.UnderlyingTypeStaticGeneral,
		Description: "Word: permissive",
		Story:       "Word: permissive",
		Other:       "permissive",
	},
}
