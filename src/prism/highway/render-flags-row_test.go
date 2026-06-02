package highway

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	lab "github.com/snivilised/jaywalk/test/laboratory"
)

const (
	// filesGlobPattern is the --files-glob value used in the
	// flags-row specs. Lifted into a constant for the same reason
	// as contract.Static.Emoji.Padlock.
	filesGlobPattern = "*.go"
)

var _ = Describe("wrapFlagsRow", func() {
	It("returns nil for an empty entry slice", func() {
		Expect(wrapFlagsRow(nil, " | ", 80)).To(BeNil())
		Expect(wrapFlagsRow([]string{}, " | ", 80)).To(BeNil())
	})

	It("places a single entry on one line", func() {
		got := wrapFlagsRow([]string{"alpha"}, " | ", 80)
		Expect(got).To(Equal([]string{"alpha"}))
	})

	It("joins entries on a single line when they fit", func() {
		got := wrapFlagsRow([]string{"a", "b", "c"}, " | ", 80)
		Expect(got).To(Equal([]string{"a | b | c"}))
	})

	It("wraps to a new line when the next entry would overflow", func() {
		got := wrapFlagsRow([]string{"aaaaa", "bbbbb", "ccccc"}, " | ", 14)
		// aaaaa (5) + " | " (3) + bbbbb (5) = 13, fits within 14
		// adding ccccc would be 13 + 3 + 5 = 21, so it wraps to a new line
		Expect(got).To(Equal([]string{"aaaaa | bbbbb", "ccccc"}))
	})

	It("each entry becomes its own line when no two fit together", func() {
		got := wrapFlagsRow([]string{"aaaaa", "bbbbb", "ccccc"}, " | ", 11)
		// 5 + 3 + 5 = 13, exceeds 11, so no two fit together
		Expect(got).To(Equal([]string{"aaaaa", "bbbbb", "ccccc"}))
	})

	It("keeps a single overlong entry on its own line", func() {
		got := wrapFlagsRow([]string{"way-too-long"}, " | ", 5)
		Expect(got).To(Equal([]string{"way-too-long"}))
	})

	It("preserves the configured separator verbatim", func() {
		got := wrapFlagsRow([]string{"a", "b"}, "##", 80)
		Expect(got[0]).To(ContainSubstring("##"))
		Expect(got[0]).NotTo(ContainSubstring("|"))
	})
})

var _ = Describe("composeFlagsRowEntries", func() {
	// helper that returns a model with the given string fields set
	modelWith := func(cascade, filesGlob, dirsGlob string,
		numFiles, numFolders uint, sampleLast bool) Model {
		return Model{
			header: contract.HeaderInfo{
				CascadeDisplay: cascade,
				FilesGlob:      filesGlob,
				DirsGlob:       dirsGlob,
				NumFiles:       numFiles,
				NumFolders:     numFolders,
				SampleLast:     sampleLast,
			},
		}
	}

	It("returns an empty slice when no flags are set", func() {
		m := modelWith("", "", "", 0, 0, false)
		entries := m.composeFlagsRowEntries(testTheme().HeaderStyle, testTheme().SummaryValueStyle)
		Expect(entries).To(BeEmpty())
	})

	It("emits a single entry for the cascade widget", func() {
		m := modelWith(contract.Static.Emoji.Padlock, "", "", 0, 0, false)
		entries := m.composeFlagsRowEntries(testTheme().HeaderStyle, testTheme().SummaryValueStyle)
		Expect(entries).To(HaveLen(1))
		Expect(entries[0]).To(ContainSubstring(contract.Static.Emoji.Padlock))
	})

	It("emits one entry per active filter", func() {
		m := modelWith("", filesGlobPattern, "src/*", 0, 0, false)
		entries := m.composeFlagsRowEntries(testTheme().HeaderStyle, testTheme().SummaryValueStyle)
		Expect(entries).To(HaveLen(1)) // filter widget bundles all filters into one entry
		Expect(entries[0]).To(ContainSubstring("files glob"))
		Expect(entries[0]).To(ContainSubstring("dirs glob"))
	})

	It("emits a sampler entry when sampleLast is set", func() {
		m := modelWith("", "", "", 0, 0, true)
		entries := m.composeFlagsRowEntries(testTheme().HeaderStyle, testTheme().SummaryValueStyle)
		Expect(entries).To(HaveLen(1))
		Expect(entries[0]).To(ContainSubstring(contract.Static.Emoji.Snail))
	})

	It("emits all three widget entries in spec order", func() {
		m := modelWith(contract.Static.Emoji.Padlock, filesGlobPattern, "", 10, 5, true)
		entries := m.composeFlagsRowEntries(testTheme().HeaderStyle, testTheme().SummaryValueStyle)
		Expect(entries).To(HaveLen(3))
		// First entry is the cascade
		Expect(entries[0]).To(ContainSubstring(contract.Static.Emoji.Padlock))
		// Second is the filter
		Expect(entries[1]).To(ContainSubstring("files glob"))
		// Third is the sampler
		Expect(entries[2]).To(ContainSubstring(contract.Static.Emoji.Snail))
		Expect(entries[2]).To(ContainSubstring("#files"))
		Expect(entries[2]).To(ContainSubstring("#dirs"))
	})
})

var _ = Describe("renderFlagsRow", func() {
	It("is a no-op when no flag is active", func() {
		m := baseModel(1)
		m.FlagsRowPosition = FlagsRowPositionTop
		var b strings.Builder
		m.renderFlagsRow(&b)
		Expect(b.String()).To(BeEmpty())
	})

	It("emits a separator border when at least one flag is active", func() {
		m := baseModel(1)
		m.FlagsRowPosition = FlagsRowPositionBottom
		m.header.CascadeDisplay = contract.Static.Emoji.Padlock
		var b strings.Builder
		m.renderFlagsRow(&b)
		out := b.String()
		Expect(out).To(ContainSubstring(contract.Static.Emoji.Padlock))
		Expect(out).To(ContainSubstring("├"))
		Expect(out).To(ContainSubstring("┤"))
	})

	It("composes multiple widgets into a single line with the pipe separator", func() {
		m := baseModel(1)
		m.FlagsRowPosition = FlagsRowPositionBottom
		m.header.CascadeDisplay = contract.Static.Emoji.Padlock
		m.header.FilesGlob = filesGlobPattern
		var b strings.Builder
		m.renderFlagsRow(&b)
		out := b.String()
		Expect(out).To(ContainSubstring(contract.Static.Emoji.Padlock))
		Expect(out).To(ContainSubstring("|"))
		Expect(out).To(ContainSubstring("files glob"))
	})

	It("places the colon immediately after the label (no column padding)", func() {
		m := baseModel(1)
		m.FlagsRowPosition = FlagsRowPositionBottom
		m.header.FilesGlob = filesGlobPattern
		var b strings.Builder
		m.renderFlagsRow(&b)
		out := b.String()
		// The label "files glob" must be followed immediately by ":" with
		// no intermediate padding. SummaryLabelStyle has Width(16) for
		// the closing summary's column alignment, but the flags row
		// renders inline so renderFlagsRow must strip that width.
		// We strip ANSI escape codes before substring matching, because
		// the styled label injects reset codes that interrupt a literal
		// "files glob:" match.
		Expect(lab.StripANSI(out)).To(ContainSubstring("files glob:"))
		Expect(lab.StripANSI(out)).NotTo(ContainSubstring("files glob      :"))
	})

	It("places the pipe separator between entries with surrounding spaces", func() {
		m := baseModel(1)
		m.FlagsRowPosition = FlagsRowPositionBottom
		m.header.CascadeDisplay = contract.Static.Emoji.Padlock
		m.header.FilesGlob = filesGlobPattern
		m.header.NumFiles = 10
		var b strings.Builder
		m.renderFlagsRow(&b)
		out := b.String()
		// Three entries (cascade, filter, sampler) should produce two
		// " | " separators on the same line.
		Expect(strings.Count(out, " | ")).To(BeNumerically(">=", 2))
	})

	It("produces the spec example format with all secondary flags set", func() {
		m := baseModel(1)
		m.width = 200 // wide enough to fit everything on one line
		m.FlagsRowPosition = FlagsRowPositionBottom
		// Note: glob/regex are mutually exclusive per family (enforced
		// at flag binding), but the spec example shows both for format
		// illustration. We pick the realistic case: one per family.
		m.header.CascadeDisplay = "depth:3"
		m.header.FilesGlob = "*.go"
		m.header.DirsRegex = `^\.`
		m.header.NumFiles = 100
		m.header.NumFolders = 10
		m.header.SampleLast = true
		var b strings.Builder
		m.renderFlagsRow(&b)
		out := lab.StripANSI(b.String())
		// spec example shape: `depth: 3 | files glob: *.go | ... | dirs regex: ^\. | 🐌 | #files: 100 | #dirs: 10`
		Expect(out).To(ContainSubstring("depth:3"))
		Expect(out).To(ContainSubstring("files glob: *.go"))
		Expect(out).To(ContainSubstring("dirs regex: ^\\."))
		Expect(out).To(ContainSubstring(contract.Static.Emoji.Snail))
		Expect(out).To(ContainSubstring("#files: 100"))
		Expect(out).To(ContainSubstring("#dirs: 10"))
		// Every flag should be joined with " | " - cascade + filter +
		// sampler emits 3 entries joined by 2 separators, plus the
		// filter widget internally uses " | " too (1) and the sampler
		// widget does too (2). Total: 2 + 1 + 2 = 5.
		Expect(strings.Count(out, " | ")).To(Equal(5))
	})

	// regression: when only --files is set, the user reported
	// "files glob: *|.go, dirs regex: ." (with stale benign default
	// bleeding through). After the fix, only the user's actual
	// filter should appear.
	It("does not surface a benign default for un-set dir filter when only --files is set", func() {
		m := baseModel(1)
		m.width = 200
		m.FlagsRowPosition = FlagsRowPositionBottom
		m.header.FilesGlob = "*|.go"
		var b strings.Builder
		m.renderFlagsRow(&b)
		out := lab.StripANSI(b.String())
		Expect(out).To(ContainSubstring("files glob: *|.go"))
		Expect(out).NotTo(ContainSubstring("dirs regex"))
		Expect(out).NotTo(ContainSubstring("dirs glob"))
		Expect(out).NotTo(ContainSubstring("files regex"))
	})

	It("wraps onto a second line when the content exceeds the width", func() {
		m := baseModel(1)
		m.width = 30 // narrow
		m.FlagsRowPosition = FlagsRowPositionBottom
		m.header.CascadeDisplay = contract.Static.Emoji.Padlock
		m.header.FilesGlob = filesGlobPattern
		m.header.DirsGlob = "src/*"
		m.header.NumFiles = 100
		m.header.NumFolders = 50
		m.header.SampleLast = true
		var b strings.Builder
		m.renderFlagsRow(&b)
		out := b.String()
		// With 4 chars of borders, available width is 26; total unwrapped content
		// is ~60 chars, so we expect at least 2 wrapped content lines (plus the
		// closing separator border makes 3 newline-terminated lines).
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		Expect(len(lines)).To(BeNumerically(">=", 2))
	})
})

var _ = Describe("Model.FlagsRowPosition defaulting", func() {
	It("defaults to bottom when the OvertureMsg position is empty", func() {
		m := baseModel(1)
		m.FlagsRowPosition = ""
		updated, _ := update(m, OvertureMsg{FlagsRowPosition: ""})
		Expect(updated.FlagsRowPosition).To(Equal(FlagsRowPositionBottom))
	})

	It("defaults to bottom when the OvertureMsg position is unrecognised", func() {
		m := baseModel(1)
		updated, _ := update(m, OvertureMsg{FlagsRowPosition: "sideways"})
		Expect(updated.FlagsRowPosition).To(Equal(FlagsRowPositionBottom))
	})

	It("preserves a top position from the OvertureMsg", func() {
		m := baseModel(1)
		updated, _ := update(m, OvertureMsg{FlagsRowPosition: FlagsRowPositionTop})
		Expect(updated.FlagsRowPosition).To(Equal(FlagsRowPositionTop))
	})

	It("preserves an explicit bottom position from the OvertureMsg", func() {
		m := baseModel(1)
		updated, _ := update(m, OvertureMsg{FlagsRowPosition: FlagsRowPositionBottom})
		Expect(updated.FlagsRowPosition).To(Equal(FlagsRowPositionBottom))
	})
})
