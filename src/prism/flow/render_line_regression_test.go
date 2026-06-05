package flow

import (
	"strings"
	"testing"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// TestRenderLine_FilesAsChildren is a regression test for the bug where
// the porthole presenter passed structural depth to RenderLine instead of
// visual depth. In the real-world ./src/app structure:
//
//	app/  (depth 0, root)
//	└── bedrock/  (depth 1, dir)
//	    ├── *.go files  (structural depth 1, VISUAL depth 2)
//	    └── data/  (depth 2, dir)
//
// Files in bedrock/ should render as children (visual depth 2), not
// siblings of bedrock (visual depth 1).
func TestRenderLine_FilesAsChildren(t *testing.T) {
	icons := contract.TreeIcons{
		contract.TreeIconBranchVertical: "│",
		contract.TreeIconBranchIndent:   "  ",
		contract.TreeIconBranchJoint:    "├── ",
		contract.TreeIconBranchLast:     "└── ",
	}

	type call struct {
		name   string
		isDir  bool
		visual uint
		isLast bool
		// exact expected prefix
		wantPrefix string
	}
	calls := []call{
		{name: "app", isDir: true, visual: 0, isLast: true,
			wantPrefix: ""},
		{name: "bedrock", isDir: true, visual: 1, isLast: true,
			wantPrefix: "└── "},
		// Files inside bedrock: VISUAL depth 2 (one deeper than structural)
		{name: "animation-state-loader.go", isDir: false, visual: 2,
			wantPrefix: "   ├── "},
		{name: "file-manager.go", isDir: false, visual: 2,
			wantPrefix: "   ├── "},
		// bedrock/data at visual depth 2, last
		{name: "data", isDir: true, visual: 2, isLast: true,
			wantPrefix: "   └── "},
	}

	stack := []bool{}
	for _, c := range calls {
		got := buildBranchPrefix(c.visual, c.isLast, stack, icons)
		if got != c.wantPrefix {
			t.Errorf("buildBranchPrefix(%q, visual=%d, isLast=%v, stack=%v)\n  got:  %q\n  want: %q",
				c.name, c.visual, c.isLast, stack, got, c.wantPrefix)
		}
		stack = updateBranchStack(c.visual, c.isLast, stack)
	}

	// Sanity check: the strings contains prefix as a prefix
	got := buildBranchPrefix(2, false, []bool{false, true}, icons)
	if !strings.HasPrefix(got, "  ") && !strings.HasPrefix(got, "│") {
		t.Errorf("expected prefix to start with ancestor column, got %q", got)
	}
}
