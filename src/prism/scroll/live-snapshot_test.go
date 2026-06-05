package scroll_test

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/flow"
	"github.com/snivilised/jaywalk/src/prism/scroll"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
)

// Live-snapshot specs render the porthole view at several terminal
// sizes using a realistic multi-level tree (./src/app -> bedrock ->
// *.go + data/) and dump the View() output via GinkgoWriter for
// visual inspection when debugging porthole rendering issues.
//
// These specs are intentionally non-failing: their primary purpose
// is to produce readable output. The Expect assertions are kept
// minimal so the snapshot stays useful as a regression target.

var _ = Describe("Live snapshot", func() {
	var (
		palette contract.Palette
		theme   contract.Theme
		grad    contract.ResolvedGradient
	)

	BeforeEach(func() {
		palette = contract.SystemPalette()
		var err error
		theme, err = contract.NewTheme(palette, &strings.Builder{})
		Expect(err).To(BeNil())

		if g, ok := theme.GradientFor(contract.GradientComponentBanner); ok {
			grad = g
		}
	})

	// buildModel constructs a Model with the given terminal size and
	// the OvertureMsg populated with a sample header/banner.
	buildModel := func(width, height int) scroll.Model {
		model := scroll.NewModel("/test", 0, theme, false)
		model, _ = update(model, tea.WindowSizeMsg{
			Width: width, Height: height,
		})
		model, _ = update(model, scroll.OvertureMsg{
			Root:      "/Users/x/project",
			Caption:   "Snapshot",
			StartedAt: core.Now(),
			Header: contract.HeaderInfo{
				FilesGlob: "*|.go",
				DirsRegex: ".",
			},
			Banner: banner.Info{
				Disable:  true,
				Position: contract.PositionTop,
				Width:    width,
				Gradient: &grad,
			},
			FlagsRowPosition: contract.PositionBottom,
		})
		return model
	}

	// visualDepth mirrors Node.VisualDepth(): files are one level
	// deeper than their containing directory.
	visualDepth := func(depth core.TraversalDepth, isDir bool) uint {
		vd := uint(depth)
		if !isDir {
			vd++
		}
		return vd
	}

	// dump writes the rendered View() content to GinkgoWriter framed
	// by banner lines so the snapshot is easy to spot in test output.
	dump := func(label string, w, h int, content string) {
		GinkgoWriter.Printf("\n====== %s (terminal %dx%d) ======\n", label, w, h)
		for i, line := range strings.Split(content, "\n") {
			GinkgoWriter.Printf("%2d: %s\n", i, line)
		}
		GinkgoWriter.Println("====== END ======")
	}

	// Mirrors the real ./src/app structure observed in:
	//   jay walk ./src/app --action boo --theme starship -f '*|.go' --tui porthole
	//
	// The Navigator reports structural depth. Files are visually
	// one level deeper than their containing directory. RenderLine
	// is invoked with (depth=Extension.Depth, visualDepth=VisualDepth()).
	type item struct {
		name   string
		isDir  bool
		depth  core.TraversalDepth
		isLast bool
	}
	fileItems := []item{
		// app is the root, structural depth 0
		{name: "app", isDir: true, depth: 0, isLast: true},
		// bedrock is a directory inside app, structural depth 1
		{name: "bedrock", isDir: true, depth: 1, isLast: true},
		// Files inside bedrock: structural depth 1, visual depth 2
		{name: "animation-state-loader.go", isDir: false, depth: 1},
		{name: "file-manager.go", isDir: false, depth: 1},
		// bedrock/data is a directory inside bedrock, structural depth 2
		{name: "data", isDir: true, depth: 2, isLast: true},
	}

	// actionItems mirror the structural items above but with an
	// action name and command output so the right-justified landing
	// strip is visible in the snapshot.
	type actionItem struct {
		name       string
		isDir      bool
		depth      core.TraversalDepth
		isLast     bool
		actionName string
		commandOut string
	}
	actionItems := []actionItem{
		{name: "app", isDir: true, depth: 0, isLast: true,
			actionName: "boo", commandOut: "sleep 0.1s"},
		{name: "bedrock", isDir: true, depth: 1, isLast: true,
			actionName: "boo", commandOut: "sleep 3.00s"},
		{name: "animation-state-loader.go", isDir: false, depth: 1,
			actionName: "boo", commandOut: "sleep 1.0s"},
		{name: "file-manager.go", isDir: false, depth: 1,
			actionName: "boo", commandOut: "sleep 0.5s"},
		{name: "data", isDir: true, depth: 2, isLast: true,
			actionName: "boo", commandOut: "sleep 0.001s"},
	}

	// buildAndFeed rebuilds the model and feeds the supplied items
	// through flow.RenderLine, sending each rendered line as a
	// ContentLineMsg. The model is closed with CompleteMsg before
	// the view is rendered.
	type feeder struct {
		name       string
		actionName string
		commandOut string
	}
	buildAndFeed := func(width, height int, feeders []feeder) string {
		model := buildModel(width, height)
		stack := []bool{}
		for _, it := range feeders {
			// Locate the matching item in fileItems to get the
			// structural depth (fileItems is the source of truth
			// for isDir/depth/isLast).
			var fi item
			for _, candidate := range fileItems {
				if candidate.name == it.name {
					fi = candidate
					break
				}
			}
			result := flow.RenderLine(
				it.name, it.name, fi.isDir, uint(fi.depth),
				it.actionName, "", it.commandOut, "", false, nil,
				fi.isLast, false, false, visualDepth(fi.depth, fi.isDir),
				stack,
				uint(width-3), // bodyWidth for right-justification
				theme,
				"", // activityFrame
			)
			stack = result.BranchStack
			model, _ = update(model, scroll.ContentLineMsg{Line: result.Line})
		}
		model, _ = update(model, scroll.CompleteMsg{
			Files: 2, Dirs: 3, Elapsed: time.Second,
		})
		return model.View().Content
	}

	// toFeeders projects fileItems to feeders (no action) so the
	// same buildAndFeed helper serves both the structural and the
	// action cases.
	toFeeders := func() []feeder {
		out := make([]feeder, len(fileItems))
		for i, it := range fileItems {
			out[i] = feeder{name: it.name}
		}
		return out
	}

	// toActionFeeders projects actionItems to feeders so the
	// action name and command output are passed through to
	// flow.RenderLine.
	toActionFeeders := func() []feeder {
		out := make([]feeder, len(actionItems))
		for i, it := range actionItems {
			out[i] = feeder{
				name:       it.name,
				actionName: it.actionName,
				commandOut: it.commandOut,
			}
		}
		return out
	}

	Describe("structural tree", func() {
		It("renders the tree at 80x24", func() {
			content := buildAndFeed(80, 24, toFeeders())
			dump("80x24", 80, 24, content)

			Expect(content).To(ContainSubstring("app/"))
			Expect(content).To(ContainSubstring("bedrock/"))
			Expect(content).To(ContainSubstring("data/"))
			Expect(content).To(ContainSubstring("animation-state-loader.go"))
			Expect(content).To(ContainSubstring("file-manager.go"))
			Expect(content).To(ContainSubstring("press space to exit"))
		})

		It("renders the tree at 120x30", func() {
			content := buildAndFeed(120, 30, toFeeders())
			dump("120x30", 120, 30, content)

			Expect(content).To(ContainSubstring("app/"))
			Expect(content).To(ContainSubstring("bedrock/"))
			Expect(content).To(ContainSubstring("press space to exit"))
		})
	})

	Describe("tree with actions (right-justified strips)", func() {
		It("renders the action tree at 80x24", func() {
			content := buildAndFeed(80, 24, toActionFeeders())
			dump("actions 80x24", 80, 24, content)

			// Each action item should appear with its command
			// output somewhere on the same line. Right-
			// justification is visual but substring presence
			// confirms the strip survived truncation in the
			// viewport.
			Expect(content).To(ContainSubstring("app/"))
			Expect(content).To(ContainSubstring("sleep 0.1s"))
			Expect(content).To(ContainSubstring("sleep 3.00s"))
			Expect(content).To(ContainSubstring("sleep 1.0s"))
			Expect(content).To(ContainSubstring("sleep 0.5s"))
			Expect(content).To(ContainSubstring("sleep 0.001s"))
			Expect(content).To(ContainSubstring("press space to exit"))
		})

		It("renders the action tree at 120x30", func() {
			content := buildAndFeed(120, 30, toActionFeeders())
			dump("actions 120x30", 120, 30, content)

			Expect(content).To(ContainSubstring("app/"))
			Expect(content).To(ContainSubstring("sleep 0.1s"))
			Expect(content).To(ContainSubstring("sleep 3.00s"))
			Expect(content).To(ContainSubstring("press space to exit"))
		})
	})
})
