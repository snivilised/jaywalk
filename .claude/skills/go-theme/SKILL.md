---
name: go-theme
description: When creating or editing UI rendering code, NEVER hardcode lipgloss styles. Always use the prism.Theme system - add semantic colours to Palette, wire them in NewTheme, and reference styles via the Theme struct.
---

# Prism Theme-First Rule

## Purpose

Every visual style in jaywalk **must** be driven by the `prism.Theme` struct, which is constructed from the user-configurable `prism.Palette`. Hardcoded `lipgloss.NewStyle()` calls in rendering code (View functions, renderers, etc.) are forbidden - they bypass the user's colour configuration and create maintenance burden.

## The Three-Step Rule

When you need a new visual style, follow this exact sequence:

### Step 1: Add a `SemanticColour` field to `Palette`

File: `src/prism/palette.go`

Add the field to the `Palette` struct with a `mapstructure` tag matching the YAML key users will write:

```go
// MyElement is the colour of ... shown during traversal.
MyElement SemanticColour `mapstructure:"my-element"`
```

Keep fields grouped by semantic category (traversal nodes, execution, status, summary, concurrent views, highway view, etc.) with section comments.

Then add a sensible ANSI-16 default in `SystemPalette()`:

```go
MyElement: SemanticColour{ANSI16: "bright-magenta"},
```

### Step 2: Add a `lipgloss.Style` field to `Theme`

File: `src/prism/theme.go`

Add the field to the `Theme` struct with a doc comment:

```go
// MyElementStyle is applied to ... in the output.
MyElementStyle lipgloss.Style
```

Then in `NewTheme`:
1. Add a `resolve` call for the palette colour,
2. Construct the style in the returned `Theme` literal:

```go
myElement, err := resolve(palette.MyElement, "my-element")
if err != nil {
    return Theme{}, err
}

// ...

MyElementStyle: lipgloss.NewStyle().
    Foreground(myElement).
    Bold(true),
```

Add the bulk of the style attributes in `NewTheme` - callers should only `.Render()`.

### Step 3: Use the style in rendering code

```go
m.theme.MyElementStyle.Render("some text")
```

Never do this in any View/renderer function:

```go
// WRONG ❌
lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render("text")

// WRONG ❌
myLocal := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))

// RIGHT ✅
m.theme.MyElementStyle.Render("text")
```

## Adding new colours to tests

Every new `Palette` field must be added to the three assertion lists in `src/prism/palette_test.go`:

1. The "all ANSI16 fields resolve without error" list
2. The "no TrueColor or ANSI256 values set" list
3. The "all ANSI16 names are non-empty" map

## Semantic mapping guidelines

When deciding whether to reuse an existing theme style or add a new one, ask: **is this a distinct visual concept that a user would want to configure independently?**

- Same concept, same context → reuse (e.g., "files:" and "dirs:" labels both use `SummaryLabelStyle`)
- Same concept, different context → consider reuse (e.g., error messages in both linear and highway views use `ErrorStyle`)
- Different concept → add new field (e.g., square-bar filled glyphs are visually distinct from progress bars)

## Palette entry naming conventions

| Convention | Example |
|---|---|
| kebab-case in YAML | `box-border`, `landing-strip` |
| PascalCase in Go struct | `BoxBorder`, `LandingStrip` |
| Style field suffix | `BoxBorder` → `BoxStyle.BorderForeground`, `MyElement` → `MyElementStyle` |
