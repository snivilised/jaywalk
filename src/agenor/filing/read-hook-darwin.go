//go:build darwin

package filing

import (
	"io/fs"
	"strings"

	lo "github.com/snivilised/jaywalk/src/third/lo"
)

// isUnixMetadataFile returns true for common metadata files that may appear
// on macOS when sharing with Windows or other platforms.
func isUnixMetadataFile(name string) bool {
	switch name {
	case "Thumbs.db", "ehthumbs.db", "desktop.ini", ".DS_Store":
		return true
	default:
		return false
	}
}

// isHiddenUnixEntry returns true if the entry is a hidden file on Unix-like systems.
// On macOS, hidden files start with ".".
// We also exclude known cross-platform metadata files.
func isHiddenUnixEntry(entry fs.DirEntry) bool {
	name := entry.Name()

	// Exclude known metadata files (can appear on macOS when sharing with Windows)
	if isUnixMetadataFile(name) {
		return true
	}

	// Standard Unix hidden files start with "."
	return strings.HasPrefix(name, ".")
}

// DefaultReadEntriesHook reads the contents of a directory.
// On macOS, it excludes:
//   - Files starting with "." (standard Unix hidden files)
//   - Known cross-platform metadata files: Thumbs.db, ehthumbs.db, desktop.ini, .DS_Store
//
// The resulting slice is left un-sorted.
func DefaultReadEntriesHook(sys fs.ReadDirFS, dirname string) ([]fs.DirEntry, error) {
	contents, err := fs.ReadDir(sys, dirname)
	if err != nil {
		return nil, err
	}

	return lo.Filter(contents, func(item fs.DirEntry, _ int) bool {
		return !isHiddenUnixEntry(item)
	}), nil
}
