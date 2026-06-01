package filter

import (
	"charm.land/lipgloss/v2"
)

// Styles defines the styles used to render the filter info widget.
type Styles struct {
	// LabelStyle is applied to the label part of each rendered entry
	// (e.g. "files glob" in "files glob: *.go").
	LabelStyle lipgloss.Style

	// ValueStyle is applied to the value part of each rendered entry
	// (e.g. "*.go" in "files glob: *.go").
	ValueStyle lipgloss.Style
}

// Render renders the filter information widget as a list of labelled
// values for any active filters. Active filters are those whose
// corresponding pattern is non-empty. Returns an empty string when no
// filters are active.
//
// Labels are human-readable and use spaces (not the CLI dash form):
//
//	files glob: <pattern>
//	files regex: <pattern>
//	dirs glob: <pattern>
//	dirs regex: <pattern>
//
// Per the spec (issue 568), consecutive entries are separated by
// " | " (single space, pipe, single space) - the same separator used
// between distinct widgets in the flags row. This produces the
// expected output:
//
//	files glob: *.go | dirs regex: src/.*
func Render(filesGlob, filesRegex, dirsGlob, dirsRegex string,
	fileTypeMode, dirTypeMode string, styles Styles) string {
	if filesGlob == "" && filesRegex == "" &&
		dirsGlob == "" && dirsRegex == "" {
		return ""
	}

	var parts []string

	if filesGlob != "" {
		parts = append(parts, labelValue(styles, "files glob", filesGlob))
	} else if filesRegex != "" {
		parts = append(parts, labelValue(styles, "files regex", filesRegex))
	}

	if dirsGlob != "" {
		parts = append(parts, labelValue(styles, "dirs glob", dirsGlob))
	} else if dirsRegex != "" {
		parts = append(parts, labelValue(styles, "dirs regex", dirsRegex))
	}

	if len(parts) == 0 {
		return ""
	}

	return joinParts(parts)
}

// labelValue renders "<label>: <value>" with the label and value each
// receiving their respective style.
func labelValue(styles Styles, label, value string) string {
	return styles.LabelStyle.Render(label) + ": " + styles.ValueStyle.Render(value)
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
