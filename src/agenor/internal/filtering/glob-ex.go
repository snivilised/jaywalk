package filtering

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/locale"
	"github.com/snivilised/jaywalk/src/third/lo"
)

const (
	excludeExtSymbol = '!'
	wildcard         = "*"

	// GlobExPattern matches a file pattern specification, which consists of
	// the base and extension parts
	GlobExPattern = `^(?P<base>.*?)\.(?P<neg>!)?(?P<ext>[^.]+)$`
)

func createGlobExFilter(def *core.FilterDef,
	ifNotApplicable bool,
) (filter core.TraverseFilter, err error) {
	var (
		segments, patterns []string
		gs                 *globSpec
	)

	if segments, patterns, err = splitGlobExPattern(def.Pattern); err != nil {
		return nil, err
	}

	// we ignore the scope on the filter because use of this type
	// of filter implies the correct scope, ie file scope.
	def.Scope = enums.ScopeFile

	directoryBase, directoryExclusion := splitGlob(segments[0])

	gs, err = newSpec(directoryBase, directoryExclusion, patterns)
	filter = &GlobEx{
		Base: Base{
			name:            def.Description,
			scope:           def.Scope,
			pattern:         def.Pattern,
			negate:          def.Negate,
			ifNotApplicable: ifNotApplicable,
		},
		spec: gs,
	}

	return filter, err
}

// GlobEx is a filter that matches files based on a glob pattern. An extended
// glob pattern allows for multiple patterns to be defined that is applied to
// the filename itself: a base glob that must match the filename, and a set of
// suffixes that must also match the filename.
type GlobEx struct {
	Base
	spec *globSpec
}

// IsApplicable determines if the filter is applicable to the given node.
// It returns true if the filter matches the node, false otherwise.
func (f *GlobEx) IsApplicable(node *core.Node) bool {
	if f.Base.IsApplicable(node) {
		return f.isBaseMatch(node)
	}

	return false
}

func (f *GlobEx) isBaseMatch(node *core.Node) bool {
	name := strings.ToLower(node.Extension.Name)

	if result, _ := filepath.Match(
		f.spec.directoryGlob, name,
	); !result {
		return false
	}

	if excluded, _ := filepath.Match(
		f.spec.directoryExclusion, name,
	); excluded {
		return false
	}

	return true
}

// IsMatch does this node match the filter; It's important to note that
// IfNotApplicable on the filter definition should be set correctly to avoid
// an unexpected result in certain situations. By default, IfNotApplicable = true,
// but this behaviour could cause confusion if not well understood. If for example
// using a filter pattern like "c*|*.*", (which means any file whose parent
// directory starts with "c"), files whose parent directory doesn't match
// this condition still pass through. This is because, when this filter is
// not applicable to non conforming files, returning IfNotApplicable's default
// value of true, will permit the non conforming file to be invoked. To resolve
// this situation, the client should also set the filter's IfNotApplicable to false.
func (f *GlobEx) IsMatch(node *core.Node) bool {
	if f.IsApplicable(node) {
		return f.invert(f.spec.IsMatch(node.Extension.Name))
	}

	return f.ifNotApplicable
}

// patternSpec represents a file pattern specification, which consists of
// the base and extension parts
type (
	patternSpec struct {
		base       string
		ext        string
		grossExt   string
		matcher    string
		excludeExt bool
	}

	globSpec struct {
		specs              []*patternSpec // */<excl>|*.o*.flac,f*.jpg
		directoryGlob      string         // *
		directoryExclusion string         // <excl>
	}
)

func parse(pattern string, re *regexp.Regexp) (spec *patternSpec, err error) {
	subMatches := re.FindStringSubmatch(pattern)

	if subMatches == nil {
		return nil, locale.NewInvalidExtendedGlobPatternError(pattern)
	}

	result := make(map[string]string)

	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = subMatches[i]
		}
	}

	ext := result["ext"]
	base := result["base"]
	grossExt := "." + ext

	return &patternSpec{
		base:       base,
		ext:        ext,
		grossExt:   grossExt,
		excludeExt: result["neg"] == string(excludeExtSymbol),
		matcher: lo.Ternary(base == "",
			wildcard+grossExt,
			base+grossExt,
		),
	}, nil
}

func newSpec(directoryBase, directoryExclusion string, patterns []string) (*globSpec, error) {
	re := regexp.MustCompile(GlobExPattern)

	specs, err := lo.MapE(patterns, func(pattern string, _ int) (*patternSpec, error) {
		pattern = strings.ToLower(strings.TrimSpace(pattern))

		return parse(pattern, re)
	})
	if err != nil {
		return nil, err
	}

	return &globSpec{
		specs:              specs,
		directoryGlob:      directoryBase,
		directoryExclusion: directoryExclusion,
	}, nil
}

func (s *patternSpec) excluded(name string) bool {
	if !s.excludeExt {
		return false
	}

	m, _ := filepath.Match(s.matcher, name)

	return m
}

func (s *patternSpec) match(name string) bool {
	m, _ := filepath.Match(s.matcher, name)
	return m
}

// IsMatch determines if the glob spec matches the given name.
func (s globSpec) IsMatch(name string) bool {
	for _, spec := range s.specs {
		if spec.excluded(name) {
			return false
		}

		if spec.excludeExt || spec.match(name) {
			return true
		}
	}

	return false
}
