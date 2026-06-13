# Bubbletea View Layouts

This document describes the 2D geometry and layout of the TUI views in Jaywalk.

## Overview

Both the Highway and Porthole views follow a **Vertical Stack** rendering pattern. They construct the final view by concatenating strings (usually via a `strings.Builder`) representing rows of the terminal.

The coordinate system is:

- **X-axis (Columns)**: Managed via `lipgloss` styles and manual padding/justification based on `m.Width`.
- **Y-axis (Rows)**: Managed by the order of calls in the `View()` method and dynamic height calculations for scrollable regions.

---

## Highway View

The highway view provides a high-level, multi-lane traversal visualization.

### Highway Layout Structure

The view is rendered as a vertical stack of components:

1. **Top Banner** (Optional): Rendered outside the main bordered region at the very top.
2. **Main Bordered Region**:
   - **Header**: Top border with the root path and session metadata.
   - **Legend (Top)**: Optional flags row, rendered between the header and the track if `FlagsRowPosition` is `Top`.
   - **Track**: The core visualization containing multiple lanes of traversal.
   - **Legend (Bottom)**: Optional flags row, rendered between the track and the summary if `FlagsRowPosition` is `Bottom` or unset.
   - **Summary**: Final traversal statistics (files, directories, errors, elapsed time).
3. **Bottom Banner** (Optional): Rendered outside the main bordered region at the very bottom, preceded by a separator newline.

### Developer Guide: Adding a New Widget to Highway view

To add a new widget to the Highway view:

1. **State Management**:
   - If the widget is stateless, you can create a transient model in `View()`.
   - If it requires state (like a timer or counter), add it as a field to `Model` in `src/prism/views/highway/model.go`. Update the `Update()` method to handle relevant messages.
2. **Rendering**:
   - Open `src/prism/views/highway/view.go`.
   - Insert a call to your widget's rendering logic into the `View()` method's `strings.Builder` sequence.
3. **Geometry**:
   - **Width**: Pass `m.Width` to the widget. Use `lipgloss` to ensure the widget spans the full width or is correctly aligned.
   - **Height**: Highway view generally allows components to take as many rows as they need (except the track, which is constrained by the terminal height), as it's rendered to an alt-screen.

---

## Porthole View

The porthole view provides a detailed, scrollable stream of traversal events.

### Layout Structure

The view is rendered as a vertical stack of components:

1. **Top Banner** (Optional): Rendered outside the main bordered region at the very top.
2. **Main Bordered Region**:
   - **Header**: Top border with the root path and session metadata.
   - **Legend (Top)**: Optional flags row, rendered between the header and the viewport if `FlagsRowPosition` is `Top`.
   - **Viewport (Body)**: A scrollable region containing the buffered content lines.
   - **Legend (Bottom)**: Optional flags row, rendered between the viewport and the status row if `FlagsRowPosition` is `Bottom` or unset.
   - **Status**: A single line showing real-time traversal statistics.
   - **Bottom Border**: The closing border of the main region.
3. **Footer**:
   - **Exit Prompt**: " • press space to exit" (1 row) displayed after completion.
4. **Bottom Banner** (Optional): Rendered outside the main bordered region at the very bottom, preceded by a separator newline.

### Geometry & Height Budgeting

Unlike the Highway view, the Porthole view has a **fixed-size viewport** that must shrink to accommodate the surrounding chrome.

The `bodyHeight` is the most critical geometric value. It is calculated in `renderBody()` to ensure the bottom border and footer are never pushed off-screen.

**Current Height Budget Formula**:
`bodyHeight = terminalHeight - 6 - legendHeight - bannerOffset`

- **6 Rows (Fixed Chrome)**:
  - Top Border (1)
  - Header Row (1)
  - Header Separator (1)
  - Status Row (1)
  - Bottom Border (1)
  - Exit Prompt (1)
- **Variable Chrome**:
  - `legendHeight`: Height of the flags row (if active).
  - `bannerOffset`: Height of the ANSI banner (if active), plus 1 for the separator newline if positioned at the bottom.

### Developer Guide: Adding a New Widget to Porthole view

To add a new widget to the Porthole view:

1. **State Management**:
   - Add the widget to the `Model` in `src/prism/views/porthole/model.go`.
   - Implement message handling in `Update()`.
2. **Rendering**:
   - Insert the widget's rendering call into the `View()` method in `src/prism/views/porthole/view.go`.
3. **Geometry Adjustments (CRITICAL)**:
   - If your widget adds **new rows** to the vertical stack, you **MUST** update the height budget formula in `renderBody()`.
   - Example: If adding a 2-row footer above the status line, change the constant `6` to `8`.
   - Failure to update this formula will cause the bottom of the view (Bottom Border/Exit Prompt) to be rendered beyond the terminal's visible area.
   - **Width**: Use `m.Width` for full-width widgets, or `bodyWidth` (terminalWidth - gutter - 2) for widgets inside the viewport.
