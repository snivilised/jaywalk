package controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/agenor"
	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/agenor/filing"
	"github.com/snivilised/jaywalk/src/agenor/pref"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

// extractHeaderInfo is exercised indirectly via Coordinator.coordinate
// at runtime. These specs invoke it directly to verify the benign
// default filter placeholder does not surface in the flags row as a
// user-supplied filter (the bug that caused "dirs regex: ." to appear
// even when the user had not specified --dirs-regex).
var _ = Describe("extractHeaderInfo", func() {
	buildRequest := func(settings ...pref.Option) *Request {
		return &Request{
			Settings: settings,
		}
	}

	It("ignores the benign default in the file slot when only --dirs-regex is set", func() {
		req := buildRequest(pref.WithFilter(&pref.FilterOptions{
			Node: &core.FilterDef{
				Type: enums.FilterTypePoly,
				Poly: &core.PolyFilterDef{
					File: core.BenignNodeFilterDef, // placeholder
					Directory: core.FilterDef{
						Type:        enums.FilterTypeRegex,
						Pattern:     "src/.*",
						Description: "dirs-regex:src/.*",
					},
				},
			},
		}))

		got := extractHeaderInfo(req)
		Expect(got.FilesGlob).To(BeEmpty())
		Expect(got.FilesRegex).To(BeEmpty())
		Expect(got.DirsRegex).To(Equal("src/.*"))
		Expect(got.DirsGlob).To(BeEmpty())
	})

	It("ignores the benign default in the directory slot when only --files is set", func() {
		req := buildRequest(pref.WithFilter(&pref.FilterOptions{
			Node: &core.FilterDef{
				Type: enums.FilterTypePoly,
				Poly: &core.PolyFilterDef{
					File: core.FilterDef{
						Type:        enums.FilterTypeGlobEx,
						Pattern:     "*|.go",
						Description: "files-glob:*|.go",
					},
					Directory: core.BenignNodeFilterDef, // placeholder
				},
			},
		}))

		got := extractHeaderInfo(req)
		Expect(got.FilesGlob).To(Equal("*|.go"))
		Expect(got.FilesRegex).To(BeEmpty())
		Expect(got.DirsRegex).To(BeEmpty(),
			"the benign '.' regex must not surface as a user flag")
		Expect(got.DirsGlob).To(BeEmpty())
	})

	It("surfaces both filters when both are user-supplied", func() {
		req := buildRequest(pref.WithFilter(&pref.FilterOptions{
			Node: &core.FilterDef{
				Type: enums.FilterTypePoly,
				Poly: &core.PolyFilterDef{
					File: core.FilterDef{
						Type:        enums.FilterTypeGlobEx,
						Pattern:     "*|.go",
						Description: "files-glob:*|.go",
					},
					Directory: core.FilterDef{
						Type:        enums.FilterTypeRegex,
						Pattern:     "^\\.",
						Description: "dirs-regex:^\\.",
					},
				},
			},
		}))

		got := extractHeaderInfo(req)
		Expect(got.FilesGlob).To(Equal("*|.go"))
		Expect(got.DirsRegex).To(Equal("^\\."))
	})

	It("surfaces the cascade display when --no-recurse is set", func() {
		req := buildRequest(agenor.WithNoRecurse())
		got := extractHeaderInfo(req)
		Expect(got.CascadeDisplay).To(Equal(contract.Static.Emoji.Padlock))
	})

	// regression: any option that mutates Options.Hooks (eg
	// WithHookReadDirectory, introduced for --include-hidden) must not
	// panic when replayed against the scratch Options built inside
	// extractHeaderInfo. Previously a bare &pref.Options{} was used,
	// so Hooks.ReadDirectory was a nil interface and .Tap(hook)
	// dereferenced a nil itab, segfaulting.
	It("does not panic when a Hook option (eg WithHookReadDirectory) is in settings", func() {
		req := buildRequest(
			agenor.WithHookReadDirectory(filing.ReadEntriesAll),
		)

		var got contract.HeaderInfo
		Expect(func() { got = extractHeaderInfo(req) }).NotTo(Panic())
		_ = got
	})
})
