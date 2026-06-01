package core

import (
	"io/fs"

	"github.com/snivilised/jaywalk/src/agenor/enums"
)

type (
	// TraverseFilter filter that can be applied to file system entries. When specified,
	// the callback will only be invoked for file system nodes that pass the filter.
	TraverseFilter interface {
		// Description describes filter
		Description() string

		// Validate ensures the filter definition is valid, returns
		// error when invalid
		Validate() error

		// Source, filter definition (comes from filter definition Pattern)
		Source() string

		// IsMatch does this node match the filter
		IsMatch(node *Node) bool

		// IsApplicable is this filter applicable to this node's scope
		IsApplicable(node *Node) bool

		// Scope, what items this filter applies to
		Scope() enums.FilterScope
	}

	// FilterDef defines a filter that can be applied to file system entries.
	FilterDef struct {
		// Type specifies the type of filter (mandatory)
		Type enums.FilterType

		// Description describes filter (optional)
		Description string

		// Pattern filter definition (mandatory)
		Pattern string

		// Scope which file system entries this filter applies to (defaults
		// to ScopeAllEn)
		Scope enums.FilterScope

		// Negate, reverses the applicability of the filter (Defaults to false)
		Negate bool

		// IfNotApplicable, when the filter does not apply to a directory entry,
		// this value determines whether the callback is invoked for this entry
		// or not (defaults to true).
		IfNotApplicable enums.TriStateBool

		// Poly allows for the definition of a PolyFilter which contains separate
		// filters that target files and directories separately. If present, then
		// all other fields are redundant, since the filter definitions inside
		// Poly should be referred to instead.
		Poly *PolyFilterDef
	}

	// PolyFilterDef defines a filter that can be applied to a directory's collection of entries
	// when subscription is set to ScopeDirectory or ScopeAllEntries
	PolyFilterDef struct {
		File      FilterDef
		Directory FilterDef
	}

	// ChildTraverseFilter filter that can be applied to a directory's collection of entries
	// when subscription is set to ScopeDirectory or ScopeAllEntries
	ChildTraverseFilter interface {
		// Description describes filter
		Description() string

		// Validate ensures the filter definition is valid, returns
		// error when invalid
		Validate() error

		// Source, filter definition (comes from filter definition Pattern)
		Source() string

		// Matching returns the collection of files contained within this
		// item's directory that matches this filter.
		Matching(children []fs.DirEntry) []fs.DirEntry
	}

	// ChildFilterDef defines a filter that can be applied to a directory's collection of entries
	// when subscription is set to ScopeDirectory or ScopeAllEntries
	ChildFilterDef struct {
		// Type specifies the type of filter (mandatory)
		Type enums.FilterType

		// Description describes filter (optional)
		Description string

		// Pattern filter definition (mandatory)
		Pattern string

		// Negate, reverses the applicability of the filter (Defaults to false)
		Negate bool
	}

	// SampleFilterDef defines a filter that can be applied to a directory's collection of entries
	// when subscription is set to ScopeDirectory or ScopeAllEntries
	SampleFilterDef struct {
		// Type specifies the type of filter (mandatory)
		Type enums.FilterType

		// Description describes filter (optional)
		Description string

		// Pattern filter definition (mandatory except if using Custom)
		Pattern string

		// Scope which file system entries this filter applies to;
		// for sampling, only ScopeFile and ScopeDirectory are valid.
		Scope enums.FilterScope

		// Negate, reverses the applicability of the filter (Defaults to false)
		Negate bool

		// Poly allows for the definition of a PolyFilter which contains separate
		// filters that target files and directories separately. If present, then
		// all other fields are redundant, since the filter definitions inside
		// Poly should be referred to instead.
		Poly *PolyFilterDef

		// Custom client defined sampling filter
		//
		Custom SampleTraverseFilter
	}

	// SampleTraverseFilter filter that can be applied to a directory's collection of entries
	// when subscription is set to ScopeDirectory or ScopeAllEntries
	SampleTraverseFilter interface {
		// Description describes filter
		Description() string

		// Validate ensures the filter definition is valid, panics when invalid
		Validate() error

		// Matching returns the collection of files contained within this
		// item's directory that matches this filter.
		Matching(children []fs.DirEntry) []fs.DirEntry
	}
)

// BenignNodeFilterDef is a filter definition that matches all nodes and can
// be used as a default filter when no filtering is desired. It uses a regex
// pattern that matches any string (".") and applies to the entire tree (ScopeTree).
var BenignNodeFilterDef = FilterDef{
	Type:        enums.FilterTypeRegex,
	Description: "benign allow all",
	Pattern:     ".",
	Scope:       enums.ScopeTree,
}

// IsBenign reports whether the filter is the BenignNodeFilterDef (the
// "allow everything" default used as a placeholder when the user has
// not supplied a filter of the relevant kind). The display layer uses
// this to distinguish a user-specified filter from a placeholder that
// happens to be active because the user only set the other one.
func (f FilterDef) IsBenign() bool {
	return f == BenignNodeFilterDef
}
