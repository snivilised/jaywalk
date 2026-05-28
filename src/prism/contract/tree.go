package contract

// TreeIcons maps named tree rendering tokens to their icon or glyph
// values. These are configured via the theme file and may be extended
// in future without changing the prism public API.
type TreeIcons map[string]string

const (
	TreeIconRoot           = "root-icon"
	TreeIconDirectory      = "directory-icon"
	TreeIconFile           = "file-icon"
	TreeIconElapsed        = "elapsed-icon"
	TreeIconSkipped        = "skipped-icon"
	TreeIconError          = "error-icon"
	TreeIconBranchVertical = "branch-vertical"
	TreeIconBranchJoint    = "branch-joint"
	TreeIconBranchLast     = "branch-last"
	TreeIconBranchIndent   = "branch-indent"
)
