package pref

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/third/lo"
)

// DefaultQueryStatusHook query the status of the path on the file system
// provided.
func DefaultQueryStatusHook(qsys fs.StatFS, path string) (fs.FileInfo, error) {
	return qsys.Stat(path)
}

// CaseSensitiveSortHook hook function for case sensitive directory traversal. A
// directory of "a" will be visited after a sibling directory "B".
func CaseSensitiveSortHook(entries []fs.DirEntry, _ ...any) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
}

// DefaultCaseInSensitiveSortHook hook function for case insensitive directory traversal. A
// directory of "a" will be visited before a sibling directory "B".
func DefaultCaseInSensitiveSortHook(entries []fs.DirEntry, _ ...any) {
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})
}

// tail extracts the end part of a string, starting from the offset
func tail(input string, offset int) string {
	asRunes := []rune(input)

	if offset >= len(asRunes) {
		return ""
	}

	return string(asRunes[offset:])
}

// difference returns the difference between a child path and a parent path
// Designed to be used with paths created from the file system rather than
// custom created or user provided input. For this reason, if there is no
// relationship between the parent and child paths provided then a panic
// may occur.
func difference(parent, child string) string {
	return tail(child, len(parent))
}

// RootItemSubPathHook returns the sub path for a root item. The sub path is the
// difference between the tree path and the node path. This is because the root
// item is the only item that has a path that is not a sub path of its parent.
func RootItemSubPathHook(info *core.SubPathInfo) string {
	return difference(info.Tree, info.Node.Path)
}

// DefaultSubPathHook returns the sub path for a node. The sub path is the
// difference between the tree path and the node path. If the node is a top
// level item then the sub path is either an empty string or a separator
// depending on the value of KeepTrailingSep.
func DefaultSubPathHook(info *core.SubPathInfo) string {
	if info.Node.Extension.Scope == enums.ScopeTop {
		return lo.Ternary(info.KeepTrailingSep, string(filepath.Separator), "")
	}

	return difference(info.Tree, info.Node.Extension.Parent)
}

// DefaultFaultHandler is the default handler for navigation faults. It simply
// returns the error contained in the fault.
func DefaultFaultHandler(fault *NavigationFault) error {
	return fault.Err
}

// DefaultPanicHandler is the default handler for panics. It simply saves the
// panic data using the provided recovery mechanism and returns the resulting
// path and error.
func DefaultPanicHandler(recovery Recovery, data RescueData) (string, error) {
	return recovery.Save(data)
}

// DefaultSkipHandler is the default handler for skipping traversal. It simply
// returns a nil error and indicates that no traversal should be skipped.
func DefaultSkipHandler(*core.Node,
	core.DirectoryContents, error,
) (enums.SkipTraversal, error) {
	return enums.SkipNoneTraversal, nil
}
