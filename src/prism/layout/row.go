package layout

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

// SegType describes how a segment's width is determined.
type SegType int

const (
	// ContentSized means width = lipgloss.Width(content).
	ContentSized SegType = iota

	// FixedWidth means width = the explicit Width field.
	FixedWidth
)

// Segment is one piece in a row layout.
// Content should be pre-styled by the caller.
// FixedWidth segments are padded with trailing spaces to reach Width.
type Segment struct {
	Content  string
	Type     SegType
	Width    int // used when Type == FixedWidth
	GapAfter int
}

type flexSeg struct {
	content  string
	truncate bool
	gapAfter int
}

type target byte

const (
	targetLeft target = iota
	targetFlex
	targetRight
)

// Row partitions a fixed width among left-aligned segments, an optional
// flexible segment in the middle, optional left segments after the flex,
// and an optional right-aligned group at the end.
//
//	┌──────────────────────────────────────────────────────┐
//	│ left … flex … left-after-flex … filler … right …  │
//	└──────────────────────────────────────────────────────┘
//
// Layout algorithm:
//  1. Sum measured widths of all left segments (before and after flex)
//     plus all right segments.
//  2. If flex exists: flexWidth = width - sum(all fixed) - flex gap
//  3. Render in order: left-before-flex + flex + left-after-flex +
//     filler + right + caps
type Row struct {
	width    int
	left     []Segment
	flex     *flexSeg
	right    []Segment
	flexIdx  int // number of left segments added BEFORE Flex() was called
	leftCap  string
	rightCap string
	pivot    target
}

// NewRow creates a row with the given total inner width (excluding borders).
func NewRow(width int) *Row {
	if width < 1 {
		width = 1
	}
	return &Row{width: width, flexIdx: -1}
}

// Caps sets the left and right border strings. These are rendered
// before the first segment and after the last segment respectively.
func (r *Row) Caps(left, right string) *Row {
	r.leftCap = left
	r.rightCap = right
	return r
}

// Content adds a content-sized segment to the left zone.
// If Flex has been called, this segment is placed after the flex content
// but still left-aligned.
func (r *Row) Content(s string) *Row {
	r.left = append(r.left, Segment{Content: s, Type: ContentSized})
	r.pivot = targetLeft
	return r
}

// Fixed adds a fixed-width segment to the left zone.
func (r *Row) Fixed(w int, s string) *Row {
	r.left = append(r.left, Segment{Content: s, Type: FixedWidth, Width: w})
	r.pivot = targetLeft
	return r
}

// RightContent adds a content-sized segment to the right zone.
// Right-zone segments are right-aligned as a group (pushed to the
// right edge of the row).
func (r *Row) RightContent(s string) *Row {
	r.right = append(r.right, Segment{Content: s, Type: ContentSized})
	r.pivot = targetRight
	return r
}

// RightFixed adds a fixed-width segment to the right zone.
func (r *Row) RightFixed(w int, s string) *Row {
	r.right = append(r.right, Segment{Content: s, Type: FixedWidth, Width: w})
	r.pivot = targetRight
	return r
}

// Flex enables the flexible middle segment. Call this after any left
// segments that should appear before the flex content and before any
// right segments. Segments added via Content/Fixed after Flex() are
// placed after the flex content but remain left-aligned.
// If truncate is true, the flex content is truncated with "…" when
// it exceeds its allocated width.
func (r *Row) Flex(truncate bool) *Row {
	r.flex = &flexSeg{truncate: truncate}
	r.flexIdx = len(r.left)
	r.pivot = targetFlex
	return r
}

// Gap sets the gap-after on the most recently added segment
// (or the flex segment if the last call was Flex).
func (r *Row) Gap(n int) *Row {
	if n <= 0 {
		return r
	}
	switch r.pivot {
	case targetLeft:
		if len(r.left) > 0 {
			r.left[len(r.left)-1].GapAfter += n
		}
	case targetFlex:
		if r.flex != nil {
			r.flex.gapAfter += n
		}
	case targetRight:
		if len(r.right) > 0 {
			r.right[len(r.right)-1].GapAfter += n
		}
	}
	return r
}

// summedWidth returns the total rendered width of a segment list,
// including each segment's content width plus its gap-after.
// FixedWidth segments contribute Width; ContentSized contributes lipgloss.Width.
func summedWidth(segments []Segment) int {
	total := 0
	for _, s := range segments {
		w := s.Width
		if s.Type == ContentSized {
			w = lipgloss.Width(s.Content)
		}
		total += w + s.GapAfter
	}
	return total
}

// computeFlex returns the width allocated to the flex segment.
// Returns 0 if there is no flex segment.
func (r *Row) computeFlex() int {
	if r.flex == nil {
		return 0
	}
	leftTotal := summedWidth(r.left)
	rightTotal := summedWidth(r.right)

	fw := r.width - leftTotal - rightTotal - r.flex.gapAfter
	if fw < 1 {
		fw = 1
	}
	return fw
}

// FlexWidth returns the width allocated to the flexible segment.
// Call this before SetFlexContent so the caller can render the
// flex content with the correct width constraint.
// Returns 0 if no flex segment was added.
func (r *Row) FlexWidth() int {
	return r.computeFlex()
}

// SetFlexContent sets the content for the flexible segment.
// Call this after FlexWidth so the content can be rendered
// at the correct width.
func (r *Row) SetFlexContent(s string) *Row {
	if r.flex != nil {
		r.flex.content = s
	}
	return r
}

// renderSegments writes each segment's content to the builder,
// followed by its gap-after spacer. FixedWidth segments are padded
// with trailing spaces to reach their declared Width.
func renderSegments(b *strings.Builder, segments []Segment) {
	for _, s := range segments {
		b.WriteString(s.Content)
		if s.Type == FixedWidth {
			cw := lipgloss.Width(s.Content)
			if pad := s.Width - cw; pad > 0 {
				b.WriteString(strings.Repeat(" ", pad))
			}
		}
		if s.GapAfter > 0 {
			b.WriteString(strings.Repeat(" ", s.GapAfter))
		}
	}
}

// RenderTo writes the fully assembled row to the builder.
func (r *Row) RenderTo(b *strings.Builder) {
	// Split left segments into "before flex" and "after flex" groups
	var leftBefore, leftAfter []Segment
	if r.flexIdx >= 0 && r.flexIdx < len(r.left) {
		leftBefore = r.left[:r.flexIdx]
		leftAfter = r.left[r.flexIdx:]
	} else {
		leftBefore = r.left
	}

	leftBeforeTotal := summedWidth(leftBefore)
	leftAfterTotal := summedWidth(leftAfter)
	rightTotal := summedWidth(r.right)

	flexWidth := 0
	flexGap := 0
	if r.flex != nil {
		flexWidth = r.computeFlex()
		flexGap = r.flex.gapAfter
	}

	// Left cap
	b.WriteString(r.leftCap)

	// Left-before-flex
	renderSegments(b, leftBefore)

	// Flex zone
	flexContentWidth := 0
	if r.flex != nil {
		content := r.flex.content
		contentWidth := lipgloss.Width(content)
		if contentWidth > flexWidth && r.flex.truncate {
			runes := []rune(content)
			keep := flexWidth - 1
			if keep < 0 {
				keep = 0
			}
			width := 0
			truncated := make([]rune, 0, keep)
			for _, ru := range runes { //nolint:staticcheck // runes required due to tricky nature of multi-byte emojis
				charWidth := lipgloss.Width(string(ru))
				if width+charWidth > keep {
					break
				}
				width += charWidth
				truncated = append(truncated, ru)
			}
			content = string(truncated) + contract.Ellipses
		}
		b.WriteString(content)
		flexContentWidth = lipgloss.Width(content)

		if flexGap > 0 {
			b.WriteString(strings.Repeat(" ", flexGap))
		}
	}

	// Left-after-flex
	renderSegments(b, leftAfter)

	// Filler between left zone and right zone
	allLeft := leftBeforeTotal + flexContentWidth + flexGap + leftAfterTotal
	filler := r.width - allLeft - rightTotal
	if filler > 0 {
		b.WriteString(strings.Repeat(" ", filler))
	}

	// Right zone
	renderSegments(b, r.right)

	// Right cap
	b.WriteString(r.rightCap)
}

// Render returns the fully assembled row as a string.
func (r *Row) Render() string {
	var b strings.Builder
	r.RenderTo(&b)
	return b.String()
}
