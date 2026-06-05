package legend_test

import (
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/legend"
	lab "github.com/snivilised/jaywalk/test/laboratory"
)

// ---------------------------------------------------------------------------
// Test fixtures
// ---------------------------------------------------------------------------

const (
	// filesGlobPattern is the --files-glob value used in the
	// legend specs. Lifted into a constant for the same reason
	// as contract.Static.Emoji.Padlock.
	filesGlobPattern = "*.go"
)

func testTheme() contract.Theme {
	t, err := contract.NewTheme(contract.SystemPalette(), io.Discard)
	if err != nil {
		panic(err)
	}
	return t
}

func makeStyles() legend.Styles {
	th := testTheme()
	// Caller responsibility: strip the column width from the label
	// style (see Styles docstring). The highway view does the same.
	return legend.Styles{
		LabelStyle:  th.SummaryLabelStyle.Width(0),
		ValueStyle:  th.SummaryValueStyle,
		BorderStyle: th.BorderStyle,
	}
}

func modelWith(cascade, filesGlob, dirsGlob string,
	numFiles, numFolders uint, sampleLast bool) legend.Info {
	return legend.Info{
		Position: contract.PositionBottom,
		Header: contract.HeaderInfo{
			CascadeDisplay: cascade,
			FilesGlob:      filesGlob,
			DirsGlob:       dirsGlob,
			NumFiles:       numFiles,
			NumFolders:     numFolders,
			SampleLast:     sampleLast,
		},
	}
}

// ---------------------------------------------------------------------------
// NewModel
// ---------------------------------------------------------------------------

var _ = Describe("NewModel", func() {
	It("captures info from WithInfo", func() {
		info := modelWith(contract.Static.Emoji.Padlock, "", "", 0, 0, false)
		m := legend.NewModel(legend.WithInfo(info))
		out := m.View()
		Expect(out).To(ContainSubstring(contract.Static.Emoji.Padlock))
	})

	It("captures styles from WithStyles", func() {
		info := modelWith(contract.Static.Emoji.Padlock, "", "", 0, 0, false)
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(80),
			legend.WithStyles(makeStyles()),
		)
		out := m.View()
		Expect(out).NotTo(BeEmpty())
	})

	It("uses WithWidth when set", func() {
		// Use a wide header so width differences are observable
		// through wrap behaviour.
		info := modelWith(contract.Static.Emoji.Padlock, filesGlobPattern, "src/*", 100, 50, true)
		m1 := legend.NewModel(legend.WithInfo(info), legend.WithWidth(200))
		m2 := legend.NewModel(legend.WithInfo(info), legend.WithWidth(30))
		// Narrower width should produce more lines (wrap) than wider.
		Expect(strings.Count(m1.View(), "\n")).To(BeNumerically("<", strings.Count(m2.View(), "\n")))
	})

	It("with no options yields a no-op View", func() {
		m := legend.NewModel()
		Expect(m.View()).To(BeEmpty())
	})
})

// ---------------------------------------------------------------------------
// View - position gating
// ---------------------------------------------------------------------------

var _ = Describe("View position gating", func() {
	styles := makeStyles()
	active := modelWith(contract.Static.Emoji.Padlock, "", "", 0, 0, false)

	It("returns '' when Position is empty", func() {
		info := active
		info.Position = ""
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(80),
			legend.WithStyles(styles),
		)
		Expect(m.View()).To(BeEmpty())
	})

	It("returns '' when Position is unrecognised", func() {
		info := active
		info.Position = "sideways"
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(80),
			legend.WithStyles(styles),
		)
		Expect(m.View()).To(BeEmpty())
	})

	It("emits output for PositionTop", func() {
		info := active
		info.Position = contract.PositionTop
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(80),
			legend.WithStyles(styles),
		)
		Expect(m.View()).NotTo(BeEmpty())
	})

	It("emits output for PositionBottom", func() {
		info := active
		info.Position = contract.PositionBottom
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(80),
			legend.WithStyles(styles),
		)
		Expect(m.View()).NotTo(BeEmpty())
	})
})

// ---------------------------------------------------------------------------
// View - no-op when no flag is active
// ---------------------------------------------------------------------------

var _ = Describe("View when no flag is active", func() {
	styles := makeStyles()

	It("returns '' when no flag is set", func() {
		info := modelWith("", "", "", 0, 0, false)
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(80),
			legend.WithStyles(styles),
		)
		Expect(m.View()).To(BeEmpty())
	})
})

// ---------------------------------------------------------------------------
// View - composition order (cascade / filter / sampler)
// ---------------------------------------------------------------------------

var _ = Describe("View composition", func() {
	styles := makeStyles()
	width := 200 // wide enough to keep all entries on one line

	It("emits a single entry for the cascade widget", func() {
		info := modelWith(contract.Static.Emoji.Padlock, "", "", 0, 0, false)
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(width),
			legend.WithStyles(styles),
		)
		out := lab.StripANSI(m.View())
		Expect(out).To(ContainSubstring(contract.Static.Emoji.Padlock))
	})

	It("emits one entry per active filter (bundled by filter widget)", func() {
		info := modelWith("", filesGlobPattern, "src/*", 0, 0, false)
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(width),
			legend.WithStyles(styles),
		)
		out := lab.StripANSI(m.View())
		Expect(out).To(ContainSubstring("files glob"))
		Expect(out).To(ContainSubstring("dirs glob"))
	})

	It("emits a sampler entry when SampleLast is set", func() {
		info := modelWith("", "", "", 0, 0, true)
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(width),
			legend.WithStyles(styles),
		)
		out := lab.StripANSI(m.View())
		Expect(out).To(ContainSubstring(contract.Static.Emoji.Snail))
	})

	It("emits all three entries in spec order", func() {
		info := modelWith(contract.Static.Emoji.Padlock, filesGlobPattern, "", 10, 5, true)
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(width),
			legend.WithStyles(styles),
		)
		out := lab.StripANSI(m.View())
		// Locate each marker and assert its byte index.
		idxCascade := strings.Index(out, contract.Static.Emoji.Padlock)
		idxFilter := strings.Index(out, "files glob")
		idxSampler := strings.Index(out, contract.Static.Emoji.Snail)
		Expect(idxCascade).To(BeNumerically(">=", 0))
		Expect(idxFilter).To(BeNumerically(">", idxCascade))
		Expect(idxSampler).To(BeNumerically(">", idxFilter))
		Expect(out).To(ContainSubstring("#files"))
		Expect(out).To(ContainSubstring("#dirs"))
	})
})

// ---------------------------------------------------------------------------
// View - column padding regression
// ---------------------------------------------------------------------------

var _ = Describe("View label/column regression", func() {
	styles := makeStyles()
	width := 200

	It("places the colon immediately after the label (no column padding)", func() {
		// The label "files glob" must be followed immediately by ":"
		// with no intermediate padding. The closing summary uses
		// Width(16) to align labels in a column, but the flags row
		// renders inline so the caller must strip that width.
		// We strip ANSI escape codes before substring matching,
		// because the styled label injects reset codes that
		// interrupt a literal "files glob:" match.
		info := modelWith("", filesGlobPattern, "", 0, 0, false)
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(width),
			legend.WithStyles(styles),
		)
		out := lab.StripANSI(m.View())
		Expect(out).To(ContainSubstring("files glob:"))
		Expect(out).NotTo(ContainSubstring("files glob      :"))
	})
})

// ---------------------------------------------------------------------------
// View - separator and pipe placement
// ---------------------------------------------------------------------------

var _ = Describe("View separator and pipes", func() {
	styles := makeStyles()
	width := 200

	It("does NOT emit any surrounding border (the view controls layout)", func() {
		// The legend widget is layout-agnostic: it renders the flag
		// entries only. Surrounding ├─────┤ borders are the parent
		// view's responsibility, so the legend must not include them
		// in its own output.
		info := modelWith(contract.Static.Emoji.Padlock, "", "", 0, 0, false)
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(width),
			legend.WithStyles(styles),
		)
		out := m.View()
		Expect(out).NotTo(ContainSubstring("├"))
		Expect(out).NotTo(ContainSubstring("┤"))
	})

	It("composes multiple widgets into a single line with the pipe separator", func() {
		info := modelWith(contract.Static.Emoji.Padlock, filesGlobPattern, "", 0, 0, false)
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(width),
			legend.WithStyles(styles),
		)
		out := m.View()
		Expect(out).To(ContainSubstring(contract.Static.Emoji.Padlock))
		Expect(out).To(ContainSubstring("|"))
		Expect(out).To(ContainSubstring("files glob"))
	})

	It("places the pipe separator between entries with surrounding spaces", func() {
		info := modelWith(contract.Static.Emoji.Padlock, filesGlobPattern, "", 10, 0, false)
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(width),
			legend.WithStyles(styles),
		)
		out := m.View()
		// Three entries (cascade, filter, sampler) should produce
		// at least 2 " | " separators on the same line, plus the
		// filter widget's internal " | " (1) and the sampler
		// widget's internal separators. We just assert the
		// minimum for the inter-widget case.
		Expect(strings.Count(out, " | ")).To(BeNumerically(">=", 2))
	})

	It("produces the spec example format with all secondary flags set", func() {
		info := legend.Info{
			Position: contract.PositionBottom,
			Header: contract.HeaderInfo{
				CascadeDisplay: "depth:3",
				FilesGlob:      "*.go",
				DirsRegex:      `^\.`,
				NumFiles:       100,
				NumFolders:     10,
				SampleLast:     true,
			},
		}
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(width),
			legend.WithStyles(styles),
		)
		out := lab.StripANSI(m.View())
		// spec example shape: `depth: 3 | files glob: *.go | ... | dirs regex: ^\. | 🐌 | #files: 100 | #dirs: 10`
		Expect(out).To(ContainSubstring("depth:3"))
		Expect(out).To(ContainSubstring("files glob: *.go"))
		Expect(out).To(ContainSubstring("dirs regex: ^\\."))
		Expect(out).To(ContainSubstring(contract.Static.Emoji.Snail))
		Expect(out).To(ContainSubstring("#files: 100"))
		Expect(out).To(ContainSubstring("#dirs: 10"))
		// cascade + filter + sampler emits 3 entries joined by 2
		// separators, plus the filter widget internally uses " | "
		// (1) and the sampler widget does too (2). Total: 2 + 1 + 2 = 5.
		Expect(strings.Count(out, " | ")).To(Equal(5))
	})
})

// ---------------------------------------------------------------------------
// View - wrap behaviour
// ---------------------------------------------------------------------------

var _ = Describe("View wrap behaviour", func() {
	styles := makeStyles()

	It("wraps onto a second line when the content exceeds the width", func() {
		info := modelWith(contract.Static.Emoji.Padlock, filesGlobPattern, "src/*", 100, 50, true)
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(30), // narrow
			legend.WithStyles(styles),
		)
		out := m.View()
		// With 4 chars of borders, available width is 26; total
		// unwrapped content is ~60 chars, so we expect at least
		// 2 wrapped content lines (plus the closing separator
		// makes 3 newline-terminated lines).
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		Expect(len(lines)).To(BeNumerically(">=", 2))
	})
})

// ---------------------------------------------------------------------------
// Height - informs parent viewports how many rows the legend will occupy
// ---------------------------------------------------------------------------

var _ = Describe("Height", func() {
	styles := makeStyles()
	width := 200

	It("returns 0 when position is empty (no flags rendered)", func() {
		m := legend.NewModel(
			legend.WithInfo(legend.Info{Position: ""}),
			legend.WithWidth(width),
			legend.WithStyles(styles),
		)
		Expect(m.Height()).To(Equal(0))
	})

	It("returns 0 when position is set but no flag is active", func() {
		m := legend.NewModel(
			legend.WithInfo(legend.Info{
				Position: contract.PositionBottom,
				Header:   contract.HeaderInfo{},
			}),
			legend.WithWidth(width),
			legend.WithStyles(styles),
		)
		Expect(m.Height()).To(Equal(0))
	})

	It("returns the number of entry lines (no surrounding borders) when at least one flag is active", func() {
		// The legend is layout-agnostic: it returns the count of
		// entry lines it will render. Surrounding borders are the
		// view's responsibility and are NOT included in this count.
		info := modelWith(contract.Static.Emoji.Padlock, "*.go", "", 0, 0, false)
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(width),
			legend.WithStyles(styles),
		)
		h := m.Height()
		Expect(h).To(BeNumerically(">=", 1)) // >= 1 entry
	})
})

// ---------------------------------------------------------------------------
// View - benign default regression
// ---------------------------------------------------------------------------

var _ = Describe("View benign default regression", func() {
	styles := makeStyles()
	width := 200

	// regression: when only --files is set, the user reported
	// "files glob: *|.go, dirs regex: ." (with stale benign default
	// bleeding through). After the fix, only the user's actual
	// filter should appear.
	It("does not surface a benign default for un-set dir filter when only --files is set", func() {
		info := modelWith("", "*|.go", "", 0, 0, false)
		m := legend.NewModel(
			legend.WithInfo(info),
			legend.WithWidth(width),
			legend.WithStyles(styles),
		)
		out := lab.StripANSI(m.View())
		Expect(out).To(ContainSubstring("files glob: *|.go"))
		Expect(out).NotTo(ContainSubstring("dirs regex"))
		Expect(out).NotTo(ContainSubstring("dirs glob"))
		Expect(out).NotTo(ContainSubstring("files regex"))
	})
})
