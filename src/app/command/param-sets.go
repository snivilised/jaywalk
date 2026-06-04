package command

import (
	"github.com/snivilised/mamba/store"
)

// ---------------------------------------------------------------------------
// Subscription flag values - what the user types on the command line
// ---------------------------------------------------------------------------

const (
	// SubscribeFlagFiles subscribes to file nodes only.
	SubscribeFlagFiles = "files"

	// SubscribeFlagDirs subscribes to directory nodes only.
	SubscribeFlagDirs = "dirs"

	// SubscribeFlagAll subscribes to all nodes (files and directories).
	SubscribeFlagAll = "all"

	// SubscribeFlagDefault is the default subscription if not specified.
	SubscribeFlagDefault = SubscribeFlagAll
)

// ---------------------------------------------------------------------------
// Root parameter set
// ---------------------------------------------------------------------------

// RootParameterSet holds flags defined on the root command that are
// inherited by all sub-commands via PersistentFlags.
type RootParameterSet struct {
	store.ParameterSetWithOverrides

	// Language sets the IETF BCP 47 language tag for i18n output.
	Language string

	// TUI selects the display mode. Corresponds to --tui(-t) <mode>.
	// Defaults to "linear". Future values: "porthole", "lanes".
	TUI string

	// Theme selects the colour theme. Corresponds to --theme <name>.
	// Defaults to "system" which uses ANSI-16 colours set by the
	// user's terminal theme. Any other value is resolved to a YAML
	// file in the themes directory (~/.config/jay/themes/<name>.yaml
	// or $JAY_THEMES_DIR/<name>.yaml).
	Theme string
}

// ---------------------------------------------------------------------------
// Param-set and family registration name constants
// ---------------------------------------------------------------------------

// Each navigation leaf command (walk, sprint, query) registers its own
// independent copy of the nav param-set and families. The constants are
// prefixed by command name to avoid collisions in the CobraContainer registry.

const (
	RootPsName = "root"

	// walk
	WalkNavPsName       = "walk-nav"
	WalkExecPsName      = "walk-exec"
	WalkPreviewFamName  = "walk-preview"
	WalkCascadeFamName  = "walk-cascade"
	WalkSamplingFamName = "walk-sampling"
	WalkPolyFamName     = "walk-poly"

	// sprint
	SprintNavPsName         = "sprint-nav"
	SprintExecPsName        = "sprint-exec"
	SprintPreviewFamName    = "sprint-preview"
	SprintCascadeFamName    = "sprint-cascade"
	SprintSamplingFamName   = "sprint-sampling"
	SprintPolyFamName       = "sprint-poly"
	SprintWorkerPoolFamName = "sprint-worker-pool"

	// query
	QueryNavPsName       = "query-nav"
	QueryPreviewFamName  = "query-preview"
	QueryCascadeFamName  = "query-cascade"
	QuerySamplingFamName = "query-sampling"
	QueryPolyFamName     = "query-poly"
)

// ---------------------------------------------------------------------------
// Nav parameter set
// ---------------------------------------------------------------------------

// NavParameterSet holds the nav-level flags registered directly on each
// navigation leaf command (walk, sprint, query). Each leaf owns its own
// independent instance; there is no shared ghost parent.
type NavParameterSet struct {
	store.ParameterSetWithOverrides

	// Subscribe controls which node types are visited.
	// Valid values: "files", "dirs", "all". Maps to --subscribe(-s).
	Subscribe string

	// Action names the config-defined action to invoke for each node.
	// Maps to --action(-a).
	Action string

	// Pipeline names the config-defined pipeline to execute.
	// Maps to --pipeline(-p).
	Pipeline string

	// IncludeHidden switches off the platform's hidden-file filter
	// for the directory read hook, so dotfiles and known metadata
	// files are visited. Maps to --include-hidden(-Z).
	IncludeHidden bool
}

// ---------------------------------------------------------------------------
// Exec parameter set
// ---------------------------------------------------------------------------

// ExecParameterSet holds the exec-level flags registered directly on
// execution commands (walk, sprint). Query intentionally omits this
// param-set since it is a read-only traversal that cannot be resumed.
type ExecParameterSet struct {
	store.ParameterSetWithOverrides

	// Resume defines the resume strategy for interrupted traversals.
	// Maps to --resume(-r).
	// Valid values: "spawn", "fast". Empty means prime (no resume).
	Resume string
}
