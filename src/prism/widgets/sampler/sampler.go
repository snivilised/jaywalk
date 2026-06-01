package sampler

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

// Styles defines the styles used to render the sampler info widget.
type Styles struct {
	// LabelStyle is applied to the label part of each rendered entry
	// (e.g. "#files" in "#files: 10") and to the 🐌 indicator which
	// has no associated value.
	LabelStyle lipgloss.Style

	// ValueStyle is applied to the value part of each rendered entry
	// (e.g. "10" in "#files: 10").
	ValueStyle lipgloss.Style
}

// Render renders the sampler info widget as a list of labelled values
// for any active sampler flags. Returns an empty string when sampling
// is not active (no fields set).
//
// Item order is fixed and matches the spec:
//
//	🐌  (when SampleLast is true)
//	#files: <n>   (when NumFiles > 0)
//	#dirs: <n>    (when NumFolders > 0)
//
// The 🐌 emoji is a boolean indicator for the --last flag and always
// precedes the count items when present. Numeric values of zero are
// treated as unset so the caller does not need to track whether the
// user supplied a zero count.
//
// Per the spec (issue 568), consecutive entries are separated by
// " | " (single space, pipe, single space) - the same separator used
// between distinct widgets in the flags row.
func Render(numFiles, numFolders uint, sampleLast bool, styles Styles) string {
	if numFiles == 0 && numFolders == 0 && !sampleLast {
		return ""
	}

	var parts []string

	if sampleLast {
		parts = append(parts, styles.LabelStyle.Render("🐌"))
	}
	if numFiles > 0 {
		parts = append(parts, labelValue(styles, "#files", numFiles))
	}
	if numFolders > 0 {
		parts = append(parts, labelValue(styles, "#dirs", numFolders))
	}

	if len(parts) == 0 {
		return ""
	}

	return joinParts(parts)
}

// labelValue renders "<label>: <n>" with the label and value each
// receiving their respective style.
func labelValue(styles Styles, label string, value uint) string {
	return styles.LabelStyle.Render(label) + ": " +
		styles.ValueStyle.Render(fmt.Sprintf("%d", value))
}

// joinParts joins the rendered parts with " | " (single space, pipe,
// single space). The separator is unstyled so that it does not get
// coloured. This matches the spec's field separator.
func joinParts(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += " | "
		}
		out += p
	}
	return out
}
