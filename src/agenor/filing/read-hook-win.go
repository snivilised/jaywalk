//go:build windows

package filing

import (
	"io/fs"
	"strings"
	"syscall"

	lo "github.com/snivilised/jaywalk/src/third/lo"
)

// isWindowsMetadataFile returns true for common Windows/macOS metadata files
// that users typically don't want to interact with directly.
func isWindowsMetadataFile(name string) bool {
	switch name {
	case "Thumbs.db", "ehthumbs.db", "desktop.ini", ".DS_Store":
		return true
	default:
		return false
	}
}

// isHiddenWindowsEntry returns true if the entry is a hidden/system file
// or a known metadata file that should be excluded on Windows.
func isHiddenWindowsEntry(entry fs.DirEntry) (bool, error) {
	name := entry.Name()

	// Always exclude known metadata files
	if isWindowsMetadataFile(name) {
		return true, nil
	}

	// Check Windows file attributes (Hidden and System)
	info, err := entry.Info()
	if err != nil {
		return false, err
	}

	if sys, ok := info.Sys().(*syscall.Win32FileAttributeData); ok {
		hidden := sys.FileAttributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0
		system := sys.FileAttributes&syscall.FILE_ATTRIBUTE_SYSTEM != 0
		if hidden || system {
			return true, nil
		}
	}

	// Fallback: also treat dotfiles as hidden
	return strings.HasPrefix(name, "."), nil
}

// DefaultReadEntriesHook reads the contents of a directory.
// On Windows, it excludes:
//   - Hidden files (FILE_ATTRIBUTE_HIDDEN)
//   - System files (FILE_ATTRIBUTE_SYSTEM)
//   - Known metadata files: Thumbs.db, ehthumbs.db, desktop.ini, .DS_Store
//
// The resulting slice is left un-sorted.
func DefaultReadEntriesHook(sys fs.ReadDirFS, dirname string) ([]fs.DirEntry, error) {
	contents, err := fs.ReadDir(sys, dirname)
	if err != nil {
		return nil, err
	}

	return lo.Filter(contents, func(item fs.DirEntry, _ int) bool {
		hidden, err := isHiddenWindowsEntry(item)
		if err != nil {
			// On error, treat as not hidden to avoid losing entries
			return true
		}
		return !hidden
	}), nil
}
