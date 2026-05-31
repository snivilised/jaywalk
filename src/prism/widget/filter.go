package widget

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// FilterStyles defines the styles used to render the filter info widget.
type FilterStyles struct {
	// InfoStyle is applied to the filter information text.
	InfoStyle lipgloss.Style
}

// Filter renders the filter information widget.
// It displays active filters (filesGlob, filesRegex, dirsGlob, dirsRegex) with precedence rules.
// Returns an empty string if no filters are active.
// The output format is " └─ [ filter1:value1, filter2:value2 ]".
func Filter(filesGlob, filesRegex, dirsGlob, dirsRegex string,
	fileTypeMode, dirTypeMode string, styles FilterStyles) string {
	if filesGlob == "" && filesRegex == "" &&
		dirsGlob == "" && dirsRegex == "" {
		return ""
	}

	var parts []string

	if filesGlob != "" {
		parts = append(parts, fmt.Sprintf("files-glob:%s", filesGlob))
	} else if filesRegex != "" {
		parts = append(parts, fmt.Sprintf("files-regex:%s", filesRegex))
	}

	if dirsGlob != "" {
		parts = append(parts, fmt.Sprintf("dirs-glob:%s", dirsGlob))
	} else if dirsRegex != "" {
		parts = append(parts, fmt.Sprintf("dirs-regex:%s", dirsRegex))
	}

	if len(parts) == 0 {
		return ""
	}

	return styles.InfoStyle.Render(" └─ [ " + strings.Join(parts, ", ") + " ]")
}
