package linear

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

var _ = Describe("RenderLine regression", func() {
	DescribeTable("files as children (visual depth)",
		func(name string, isDir bool, visual uint, isLast bool, wantPrefix string) {
			icons := contract.TreeIcons{
				contract.TreeIconBranchVertical: "│",
				contract.TreeIconBranchIndent:   "  ",
				contract.TreeIconBranchJoint:    "├── ",
				contract.TreeIconBranchLast:     "└── ",
			}
			stack := []bool{}
			// Process all prior entries to build up the stack,
			// then check this entry's prefix.
			type entry struct {
				name   string
				isDir  bool
				visual uint
				isLast bool
			}
			all := []entry{
				{name: "app", isDir: true, visual: 0, isLast: true},
				{name: "bedrock", isDir: true, visual: 1, isLast: true},
				{name: "animation-state-loader.go", isDir: false, visual: 2, isLast: false},
				{name: "file-manager.go", isDir: false, visual: 2, isLast: false},
				{name: "data", isDir: true, visual: 2, isLast: true},
			}
			for _, e := range all {
				stack = updateBranchPrefix(e.visual, e.isLast, stack, icons)
				if e.name == name {
					break
				}
			}
			got := buildBranchPrefix(visual, isLast, stack, icons)
			Expect(got).To(Equal(wantPrefix))
		},
		Entry("app (root)", "app", true, uint(0), true, ""),
		Entry("bedrock (dir, last)", "bedrock", true, uint(1), true, "└── "),
		Entry("animation-state-loader.go (file, not last)", "animation-state-loader.go", false, uint(2), false, "   ├── "),
		Entry("file-manager.go (file, not last)", "file-manager.go", false, uint(2), false, "   ├── "),
		Entry("data (dir, last)", "data", true, uint(2), true, "   └── "),
	)

	It("sanity checks prefix starts with ancestor column", func() {
		icons := contract.TreeIcons{
			contract.TreeIconBranchVertical: "│",
			contract.TreeIconBranchIndent:   "  ",
			contract.TreeIconBranchJoint:    "├── ",
			contract.TreeIconBranchLast:     "└── ",
		}
		got := buildBranchPrefix(2, false, []bool{false, true}, icons)
		Expect(got).To(SatisfyAny(
			HavePrefix("  "),
			HavePrefix("│"),
		))
	})
})

// updateBranchPrefix builds the branch prefix and updates the stack in
// one step, mirroring how RenderLine uses them together.
func updateBranchPrefix(depth uint, isLast bool, stack []bool, icons contract.TreeIcons) []bool {
	_ = buildBranchPrefix(depth, isLast, stack, icons)
	return updateBranchStack(depth, isLast, stack)
}
