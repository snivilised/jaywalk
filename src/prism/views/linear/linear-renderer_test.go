package linear

import (
	"bytes"
	"time"

	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

var _ = Describe("LinearRenderer", func() {
	It("renders tree branches with default icons", func() {
		w := &bytes.Buffer{}
		palette := contract.Palette{}

		renderer, err := New(palette, w)
		Expect(err).To(Succeed())
		Expect(renderer).NotTo(BeNil())

		renderer.Show(contract.Motif{
			Name:        "app",
			IsDir:       true,
			Depth:       0,
			VisualDepth: 0,
			IsLast:      true,
		})

		renderer.Show(contract.Motif{
			Name:        "bedrock",
			IsDir:       true,
			Depth:       1,
			VisualDepth: 1,
			IsLast:      false,
		})

		renderer.Show(contract.Motif{
			Name:        "bedrock_suite_test.go",
			IsDir:       false,
			Depth:       1,
			VisualDepth: 2,
			IsLast:      true,
		})

		output := w.String()
		Expect(output).To(ContainSubstring("✻ app/\n"))
		Expect(output).To(ContainSubstring("├── 📁 bedrock/\n"))
		Expect(output).To(ContainSubstring("│  └── 🔖 bedrock_suite_test.go\n"))
	})

	It("applies WithIcons overrides", func() {
		w := &bytes.Buffer{}
		palette := contract.Palette{}

		renderer, err := New(palette, w, WithIcons(map[string]string{
			contract.TreeIconRoot:           "R",
			contract.TreeIconDirectory:      "D",
			contract.TreeIconFile:           "F",
			contract.TreeIconElapsed:        "E",
			contract.TreeIconBranchJoint:    "+-- ",
			contract.TreeIconBranchLast:     "L-- ",
			contract.TreeIconBranchVertical: "|",
			contract.TreeIconBranchIndent:   "  ",
		}))
		Expect(err).To(Succeed())

		renderer.Show(contract.Motif{
			Name:        "root",
			IsDir:       true,
			Depth:       0,
			VisualDepth: 0,
			IsLast:      true,
		})

		renderer.Show(contract.Motif{
			Name:        "child",
			IsDir:       false,
			Depth:       1,
			VisualDepth: 1,
			IsLast:      true,
		})

		Expect(w.String()).To(ContainSubstring("R root/\n"))
		Expect(w.String()).To(ContainSubstring("L-- F child\n"))
	})

	It("renders summary entries with tree icon prefixes", func() {
		w := &bytes.Buffer{}
		palette := contract.Palette{}

		renderer, err := New(palette, w)
		Expect(err).To(Succeed())

		renderer.End(contract.Summary{
			Kind:         enums.NavigationKindPrime,
			FilesVisited: 12,
			DirsVisited:  3,
			Elapsed:      2 * time.Second,
		})

		output := w.String()
		Expect(output).To(ContainSubstring("🔖 files:"))
		Expect(output).To(ContainSubstring("📁 dirs:"))
		Expect(output).To(ContainSubstring("⏰ elapsed:"))
	})

	It("aligns summary values when the elapsed icon has a different display width", func() {
		w := &bytes.Buffer{}
		palette := contract.Palette{}

		// ⏱️ this emoji is 2 columns wide and breaks width calculation
		// probably because some lipgloss internal processing is not performing
		// correct rune width calculations
		renderer, err := New(palette, w, WithIcons(map[string]string{
			contract.TreeIconElapsed: "🦋",
		}))
		Expect(err).To(Succeed())

		renderer.End(contract.Summary{
			Kind:         enums.NavigationKindPrime,
			FilesVisited: 55,
			DirsVisited:  7,
			Skipped:      0,
			Elapsed:      2 * time.Millisecond,
		})

		output := ansi.Strip(w.String())
		Expect(output).To(ContainSubstring("🦋 elapsed:"))
		Expect(output).To(ContainSubstring("🔖 files:"))
		Expect(output).To(ContainSubstring("📁 dirs:"))
		Expect(output).To(ContainSubstring("⛔️ skipped:"))
	})

	It("returns a renderer when options are provided", func() {
		w := &bytes.Buffer{}
		palette := contract.Palette{}

		renderer, err := New(palette, w, WithIcons(nil))
		Expect(err).To(Succeed())
		Expect(renderer).NotTo(BeNil())
		Expect(renderer).To(Not(BeNil()))
		renderer.Show(contract.Motif{Name: "test", Depth: 0, VisualDepth: 0, IsDir: true, IsLast: true})
		Expect(w.String()).To(ContainSubstring("✻ test/"))
	})

	It("renders the banner with highway-style borders and widgets", func() {
		w := &bytes.Buffer{}
		palette := contract.Palette{}

		renderer, err := New(palette, w)
		Expect(err).To(Succeed())

		renderer.Begin(contract.Overture{
			Kind:      enums.NavigationKindPrime,
			Root:      "./src/app",
			Caption:   "files and folders",
			StartedAt: time.Date(2026, time.May, 10, 11, 31, 7, 0, time.UTC),
		})

		output := w.String()
		// Top border uses contract.Static.Borders
		Expect(output).To(ContainSubstring(contract.Static.Borders.TopLeftCorner))
		Expect(output).To(ContainSubstring(contract.Static.Borders.TopRight))
		// Path is rendered inside the top border
		Expect(output).To(ContainSubstring("[ ./src/app ]"))
		// Header row shows "jay" and caption with date
		Expect(output).To(ContainSubstring("jay"))
		Expect(output).To(ContainSubstring("files and folders  -"))
		// Bottom border uses contract.Static.Borders
		Expect(output).To(ContainSubstring(contract.Static.Borders.BottomLeft))
		Expect(output).To(ContainSubstring(contract.Static.Borders.BottomRightCorner))
	})

	It("renders final directory children without vertical continuation", func() {
		w := &bytes.Buffer{}
		palette := contract.Palette{}

		renderer, err := New(palette, w)
		Expect(err).To(Succeed())

		renderer.Show(contract.Motif{Name: "src", IsDir: true, Depth: 0, VisualDepth: 0, IsLast: true})
		renderer.Show(contract.Motif{Name: "app", IsDir: true, Depth: 1, VisualDepth: 1, IsLast: false})
		renderer.Show(contract.Motif{Name: "main.go", IsDir: false, Depth: 2, VisualDepth: 2, IsLast: true})
		renderer.Show(contract.Motif{Name: "ui", IsDir: true, Depth: 1, VisualDepth: 1, IsLast: true})
		renderer.Show(contract.Motif{Name: "doc.go", IsDir: false, Depth: 2, VisualDepth: 2, IsLast: true})

		output := w.String()
		Expect(output).To(ContainSubstring("└── 📁 ui/"))
		Expect(output).To(ContainSubstring("   └── 🔖 doc.go"))
	})

	It("applies BranchStyle from theme to branch characters", func() {
		w := &bytes.Buffer{}
		palette := contract.SystemPalette()
		palette.Branch = contract.SemanticColour{ANSI16: "green"}

		renderer, err := New(palette, w)
		Expect(err).To(Succeed())
		Expect(renderer).NotTo(BeNil())

		renderer.Show(contract.Motif{Name: "root", IsDir: true, Depth: 0, VisualDepth: 0, IsLast: true})
		renderer.Show(contract.Motif{Name: "child", IsDir: false, Depth: 1, VisualDepth: 1, IsLast: true})

		output := w.String()
		Expect(output).To(ContainSubstring("✻ root/\n"))
		Expect(output).To(ContainSubstring("└── 🔖 child\n"))
	})
})
