package contract

// HeaderInfo groups the supplementary flag values that need to be
// extracted from the resolved traversal options for display in the
// highway view's header and flags row. The fields are organised into
// three loose clusters (cascade, filter, sampler) but are kept flat
// on the struct to keep call-site field access simple.
//
// HeaderInfo is defined in the prism/contract package because it
// crosses the prism/app boundary: the controller (app) populates it
// from traversal options, then the highway model (prism) reads it
// to render the flags row. Placing it in a contract keeps both
// sides decoupled from each other.
type HeaderInfo struct {
	// CascadeDisplay holds the cascade widget value: "🔒" when
	// --no-recurse is set, "depth:<n>" when --depth is set, or empty
	// when neither is active.
	CascadeDisplay string

	// FilesGlob holds the --files-glob pattern (empty when unset).
	FilesGlob string
	// FilesRegex holds the --files-regex pattern (empty when unset).
	FilesRegex string
	// DirsGlob holds the --dirs-glob pattern (empty when unset).
	DirsGlob string
	// DirsRegex holds the --dirs-regex pattern (empty when unset).
	DirsRegex string
	// FileTypeMode is "regex" or "glob" indicating which of FilesRegex/
	// FilesGlob the renderer should prefer. Defaults to "glob".
	FileTypeMode string
	// DirTypeMode is "regex" or "glob" indicating which of DirsRegex/
	// DirsGlob the renderer should prefer. Defaults to "glob".
	DirTypeMode string

	// NumFiles is the --num-files value (0 when unset).
	NumFiles uint
	// NumFolders is the --num-folders value (0 means unset).
	NumFolders uint
	// SampleLast is true when --last was specified.
	SampleLast bool
}
