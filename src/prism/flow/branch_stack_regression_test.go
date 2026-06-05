package flow

import (
	"testing"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// TestUpdateBranchStack_Regression is a standalone test that verifies
// updateBranchStack and buildBranchPrefix work correctly together across
// a realistic multi-level tree traversal. This is a regression test for
// the bug where depth-1 nodes were not being recorded in the branch
// stack, causing nested children to lose their vertical continuation
// bars.
func TestUpdateBranchStack_Regression(t *testing.T) {
	// Build a tree:
	//   src (depth 0, root)
	//   ├── app (depth 1, NOT last)         <- joint at depth 1
	//   │   ├── controller (depth 2, NOT last)
	//   │   │   └── dispatch.go (depth 3, last)
	//   │   └── command (depth 2, last)
	//   │       └── bootstrap.go (depth 3, last)
	//   └── ui (depth 1, last)               <- last at depth 1
	//       └── porthole.go (depth 2, last)
	//
	// branchStack progresses:
	//   src        -> stack becomes nil  (depth 0 short-circuit)
	//   app        -> stack becomes [true]   (depth 1, joint)
	//   controller -> stack becomes [true, true]  (depth 2, joint)
	//   dispatch   -> stack becomes [true, true, false]  (depth 3, last)
	//   command    -> stack becomes [true, false]  (depth 2, last -> trim)
	//   bootstrap  -> stack becomes [true, false, false]
	//   ui         -> stack becomes [false]  (depth 1, last -> trim)
	//   porthole   -> stack becomes [false, false]

	icons := contract.TreeIcons{
		contract.TreeIconBranchVertical: "│",
		contract.TreeIconBranchIndent:   "  ",
		contract.TreeIconBranchJoint:    "├── ",
		contract.TreeIconBranchLast:     "└── ",
	}

	type call struct {
		name   string
		depth  uint
		isLast bool
		want   string // expected branch prefix
		// expected stack after this call
		wantStackLen    int
		wantStackValues []bool
	}

	cases := []call{
		{name: "src", depth: 0, isLast: true, want: "",
			wantStackLen: 0},
		{name: "app", depth: 1, isLast: false, want: "├── ",
			wantStackLen: 1, wantStackValues: []bool{true}},
		{name: "controller", depth: 2, isLast: false, want: "│  ├── ",
			wantStackLen: 2, wantStackValues: []bool{true, true}},
		{name: "dispatch.go", depth: 3, isLast: true, want: "│  │  └── ",
			wantStackLen: 3, wantStackValues: []bool{true, true, false}},
		{name: "command", depth: 2, isLast: true, want: "│  └── ",
			wantStackLen: 2, wantStackValues: []bool{true, false}},
		{name: "bootstrap.go", depth: 3, isLast: true, want: "│     └── ",
			wantStackLen: 3, wantStackValues: []bool{true, false, false}},
		{name: "ui", depth: 1, isLast: true, want: "└── ",
			wantStackLen: 1, wantStackValues: []bool{false}},
		{name: "porthole.go", depth: 2, isLast: true, want: "   └── ",
			wantStackLen: 2, wantStackValues: []bool{false, false}},
	}

	stack := []bool(nil)
	for _, c := range cases {
		got := buildBranchPrefix(c.depth, c.isLast, stack, icons)
		if got != c.want {
			t.Errorf("buildBranchPrefix(%s, depth=%d, isLast=%v) = %q, want %q",
				c.name, c.depth, c.isLast, got, c.want)
		}
		stack = updateBranchStack(c.depth, c.isLast, stack)
		if len(stack) != c.wantStackLen {
			t.Errorf("after %s: stack len = %d, want %d (stack=%v)",
				c.name, len(stack), c.wantStackLen, stack)
		}
		if c.wantStackValues != nil {
			for i, v := range c.wantStackValues {
				if i < len(stack) && stack[i] != v {
					t.Errorf("after %s: stack[%d] = %v, want %v (full stack=%v)",
						c.name, i, stack[i], v, stack)
				}
			}
		}
	}
}
