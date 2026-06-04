package filing

import (
	"io/fs"
)

// ReadEntriesAll returns the full directory contents without filtering.
func ReadEntriesAll(sys fs.ReadDirFS, dirname string) ([]fs.DirEntry, error) {
	return fs.ReadDir(sys, dirname)
}
