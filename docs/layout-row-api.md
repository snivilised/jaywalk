# Prism Layout Row API

The `layout.Row` API provides a fluent interface for partitioning a fixed terminal width into a set of left-aligned, right-aligned, and flexible segments. It is primarily used to render consistent horizontal rows with borders (caps) and precise spacing.

## Core Concept

A `Row` decomposes a line of text into three main zones:

1. **Left Zone**: Segments added via `Content()` or `Fixed()`. These are rendered in the order they are added.
2. **Flex Zone** (Optional): A single flexible segment that occupies all remaining space after fixed and right-aligned segments are accounted for.
3. **Right Zone**: Segments added via `RightContent()` or `RightFixed()`. These are pushed to the right edge of the row as a group.

### Visual Layout

```txt
┌─ left1 [gap] left2 ... [gap] flex [gap] left3 ... [gap] filler [gap] right1 [gap] right2 ─┐
  ^                                                                                  ^
  leftCap                                                                        rightCap
```

---

## API Reference

### Construction

`NewRow(width int) *Row`
Creates a new row with the specified inner width (excluding caps).

### Borders

`Caps(left, right string) *Row`
Sets the strings to be rendered at the absolute start and end of the row. Commonly used for border characters like `│`.

### The Left Zone

Segments added here are rendered from left to right.

- `Content(s string) *Row`: Adds a segment whose width is determined by the content (`lipgloss.Width`).
- `Fixed(w int, s string) *Row`: Adds a segment with an explicit width `w`. If the content is shorter than `w`, it is padded with trailing spaces.

### The Flex Zone

The flex zone allows a segment to dynamically expand to fill available space.

- `Flex(truncate bool) *Row`: Enables the flexible segment.
  - Call this *after* any left-aligned segments that should appear to its left.
  - Any `Content()` or `Fixed()` calls *after* `Flex()` will be placed to the right of the flex content, but still left-aligned.
  - If `truncate` is true, content exceeding the allocated width is truncated with an ellipsis (`…`).
- `FlexWidth() int`: Returns the computed width of the flex segment. **Crucial**: Call this before `SetFlexContent` if you need to render the content with a specific width constraint.
- `SetFlexContent(s string) *Row`: Sets the content for the flexible segment.

### The Right Zone

Segments added here are right-aligned as a collective group.

- `RightContent(s string) *Row`: Adds a content-sized segment to the right zone.
- `RightFixed(w int, s string) *Row`: Adds a fixed-width segment to the right zone.

### Spacing

- `Gap(n int) *Row`: Adds `n` spaces after the most recently added segment. This works for left, flex, and right segments.

### Rendering

- `Render() string`: Assembles the row and returns the final string.
- `RenderTo(b *strings.Builder)`: Writes the assembled row directly into a builder.

---

## Usage Examples

### Basic Bordered Row

A simple row with a label and a right-aligned value.

```go
row := layout.NewRow(m.Width - 2).
    Caps("│ ", " │").
    Content("Files visited:").
    Gap(2).
    RightContent("1,234")
```

### Complex Row with Flex Content

A row where a path is truncated in the middle, with a fixed-width status on the left and a label on the right.

```go
row := layout.NewRow(m.Width - 4).
    Caps("│ ", " │").
    Fixed(10, "Status: OK").
    Gap(2).
    Flex(true). // Enable truncation
    RightContent("v1.0.0")

// Use FlexWidth to style the content if needed
path := "very/long/path/to/some/deeply/nested/file.txt"
row.SetFlexContent(path)
```

## Implementation Details

1. **Sizing**: The layout algorithm first sums the widths of all fixed and content-sized segments.
2. **Flex Calculation**: `flexWidth = totalWidth - (sum of all left and right segments) - flexGap`.
3. **Filler**: If the total width of all segments (including the rendered flex content) is less than the `Row` width, a "filler" of spaces is inserted between the left zone and the right zone to ensure right-aligned segments are pushed to the edge.
4. **Unicode Handling**: Truncation is performed on runes to correctly handle multi-byte characters and emojis.
