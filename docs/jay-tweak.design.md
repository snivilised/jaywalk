# `jay tweak` - Design Document

## Status: Draft

---

## 1. Design Principles

These principles govern every decision in `tweak`. When facing a design
fork, test both options against this list.

1. **True-color is principal** - the user always works in hex. ANSI256
   and ANSI16 are derived automatically and silently.
2. **Promote on open** - any theme opened in `tweak` is silently
   upscaled to full coverage before editing begins. Missing tier values
   are derived from whichever tiers are present. Upscaling is never
   lost, even on undo.
3. **Remove friction** - wherever `tweak` can make a sensible automatic
   decision, it does. The user is only asked to make choices that
   genuinely require human judgement.
4. **Degrade gracefully, visibly** - the user can preview their theme
   at any capability tier (TrueColor, ANSI256, ANSI16) before saving,
   so they can see degradation without leaving `tweak`.
5. **Real code, real content** - the live preview uses the actual
   rendering stack against real filesystem content. No fake renderers
   to maintain.
6. **Coherence by construction** - the gradient workshop ensures
   colours in a theme are visually related by design, not by accident.
7. **Write on exit, not on change** - nothing is written to disk until
   the user explicitly exits or saves. The TUI is a safe exploration
   space.
8. **No implementation concepts in the UX** - internal abstractions
   such as layers, upscaling, and state models must never appear in
   any user-facing string, prompt, label, or error message. User
   language is always concrete and outcome-oriented.

---

## 2. Overview

`tweak` is a new top-level subcommand sitting alongside `walk`, `query`,
and `sprint`. It provides a Bubble Tea TUI for interactive theme
creation and editing. Its output is a YAML theme file written to
`~/.config/jay/themes/`.

```text
jay tweak [--theme <name>]
```

The `--theme` flag is optional. When omitted the home screen offers
theme file management (open / new / save-as) as its first action.

---

## 3. Home Screen

The home screen is a persistent menu offering four entry points plus
theme file management. The interaction style adapts to the task:
wizard-style for the gradient workshop, menu-driven for the palette
editor, picker for bindings, and short wizard for import.

```text
+----------------------------------------------+
|  jay tweak                   [theme: starship]
+----------------------------------------------+
|                                              |
|   1. Gradient Workshop                       |
|      Define seed gradient, harvest steps     |
|      to palette roles                        |
|                                              |
|   2. Palette Editor                          |
|      Edit palette roles directly             |
|                                              |
|   3. Bindings                                |
|      Map component slots to gradients        |
|                                              |
|   4. Import Theme                            |
|      Convert iTerm2 / VS Code / Alacritty /  |
|      Warp themes                             |
|                                              |
|   F: File  (open / new / save / save-as)     |
|   Z: Undo  (reset changes)                   |
|   Q: Quit                                    |
+----------------------------------------------+
```

---

## 4. Theme File Management

Available from the home screen at all times via the `F` key. Presented
as a short menu:

- **Open** - list themes from `~/.config/jay/themes/` (or
  `$JAY_THEMES_DIR`), select one to load. Upscaling runs silently
  before the editor screens are shown.
- **New** - create a new theme by copying the built-in system default.
  Prompts for a name; creates `<name>.yaml` in the themes directory.
- **Save** - write the current state to the open file.
- **Save-as** - write to a new name; leaves the original intact.

---

## 5. Internal State Model

This section describes the internal implementation only. None of these
concepts - layers, upscaling, dirty state - appear anywhere in the
user-facing TUI. User-visible language is defined in section 5.4.

### 5.1 Layer 0 - Raw Loaded State

The `contract.Palette` decoded directly from the theme YAML by
`ThemeLoader.Load`. Stored read-only and never modified. Represents
exactly what was on disk when the session began; may be sparse (e.g.
a role with only `ansi16` set and no `true-color`).

### 5.2 Layer 1 - Upscaled State

Produced once from layer 0 by `UpscalePalette` immediately after load,
before any editor screen is shown. Every `SemanticColour` is inspected
and missing tier values are derived from whichever tiers are present
(see section 7 for the full rule set). The result is a fully populated
palette where every role carries values for all three tiers. Stored
read-only after construction.

Layer 1 is never discarded. Undo resets the working state to layer 1,
not to layer 0. Upscaled values are always preserved across undo.

### 5.3 Layer 2 - Working State

A mutable copy of layer 1. All creative changes made in the TUI -
palette role assignments, gradient definitions, curve and easing
choices, bindings - are applied here. This is what the live preview
renders against and what is written to disk on save.

```text
Layer 0  Raw loaded      (sparse, read-only, never changes)
   |
   v  UpscalePalette()
Layer 1  Upscaled        (complete, read-only, never changes)
   |
   v  copy
Layer 2  Working         (mutable; what the user edits and
                          what is written to disk on save)
```

### 5.4 User-Visible Language

The three-layer model is invisible to the user. The only user-facing
behaviours that relate to internal state are:

**Undo prompt** - resets working state (layer 2) to upscaled state
(layer 1):

```text
Reset all changes? Any colour values derived from your
existing settings will be kept.
[Y] Reset  [N] Cancel
```

**Exit with creative changes pending**:

```text
You have unsaved changes.
[S] Save  [D] Discard  [C] Cancel
```

**Exit with upscaling only** (no creative changes; layer 2 equals
layer 1 but layer 1 differs from layer 0):

```text
No changes to save, but your theme has been enriched with
additional colour values derived from your existing settings.
Save the enriched version?
[S] Save  [D] Discard  [C] Cancel
```

**Exit with no changes** (layer 2 equals layer 0): exit silently with
no prompt and no write.

### 5.5 Dirty State Tracking

The coordinator tracks two internal flags:

- **Upscale dirty** - layer 1 differs from layer 0 (upscaling derived
  new values).
- **Creative dirty** - layer 2 differs from layer 1 (the user made
  creative changes).

These flags drive the exit flow above. They are never exposed in the
UI directly.

---

## 6. Write Path

The existing config infrastructure (`bedrock/loader.go`,
`bedrock/theme-loader.go`) is read-only. Viper does not participate
in the write path. A new write layer is introduced alongside the
existing loaders.

### 6.1 Theme Writer

`src/app/bedrock/theme-writer.go` serialises the working
`contract.Palette` directly to YAML using `gopkg.in/yaml.v3`. Viper
is not involved.

```go
// ThemeWriter serialises a Palette to a YAML theme file.
type ThemeWriter struct {
    themesDir string
}

// Write encodes palette as a YAML theme file at
// <themesDir>/<name>.yaml. Overwrites any existing file.
func (tw *ThemeWriter) Write(name string, palette contract.Palette) error
```

The output file wraps the palette under the `palette:` top-level key,
matching the structure `ThemeLoader.Load` expects to read back.

### 6.2 Atomic Write

Writes are atomic: write to a `.tmp` file in the same directory, then
`os.Rename` over the target. `os.Rename` is atomic on POSIX systems,
preventing corrupt files on interrupted writes.

### 6.3 View Config Writer

`src/app/bedrock/view-writer.go` serialises updated view config (banner
cascade overrides, `tweak:` preview-path changes) using the same
`gopkg.in/yaml.v3` approach. Changes to `jay.ui.yml` are batched into
the same exit-time write as theme changes.

---

## 7. Upscaling - Full Rule Set

`UpscalePalette` in `src/prism/contract/upscale.go` applies these
rules to every `SemanticColour` in the palette, including gradient
`hi` and `lo` endpoints.

| Present | Absent | Action |
| --- | --- | --- |
| `true-color` | `ansi256`, `ansi16` | Derive both via nearest-colour |
| `true-color`, `ansi256` | `ansi16` | Derive ANSI16 via nearest-colour |
| `true-color`, `ansi16` | `ansi256` | Derive ANSI256 via nearest-colour |
| `ansi16` | `true-color`, `ansi256` | Promote via canonical hex table; derive ANSI256 |
| `ansi256` | `true-color`, `ansi16` | Promote via canonical hex table; derive ANSI16 |
| `ansi16`, `ansi256` | `true-color` | Promote ANSI16 via canonical hex table |
| all three present | - | No action |
| none present | - | No action (role intentionally empty) |

### 7.1 ANSI16 Canonical Hex Table (xterm defaults)

| Name | Hex |
| --- | --- |
| black | #000000 |
| red | #CC0000 |
| green | #4E9A06 |
| yellow | #C4A000 |
| blue | #3465A4 |
| magenta | #75507B |
| cyan | #06989A |
| white | #D3D7CF |
| bright-black | #555753 |
| bright-red | #EF2929 |
| bright-green | #8AE234 |
| bright-yellow | #FCE94F |
| bright-blue | #729FCF |
| bright-magenta | #AD7FA8 |
| bright-cyan | #34E2E2 |
| bright-white | #EEEEEC |

### 7.2 Nearest-Colour Algorithm

Implemented in `src/prism/contract/colour-distance.go`.

- **ANSI256 derivation** - Euclidean distance in RGB space against all
  256 xterm palette entries. Returns the index of the closest match.
- **ANSI16 derivation** - same algorithm restricted to the 16 canonical
  entries above. Returns the name string (e.g. `"cyan"`).

---

## 8. Gradient Model

### 8.1 Shape

Controls how colour steps are distributed across the hi-to-lo range.
Implemented with standard easing functions; no external dependency.

| Curve | Behaviour |
| --- | --- |
| `linear` | Equal spacing (default, current behaviour) |
| `ease-in` | Steps bunch at the hi end |
| `ease-out` | Steps bunch at the lo end |
| `bell` | Steps concentrate in the middle |

### 8.2 Animation Easing

Controls how fast the renderer moves through steps over time during
animation. Backed by Harmonica internally; the user sees only named
presets, never spring parameters.

| Preset | Behaviour |
| --- | --- |
| `linear` | Constant speed (default, current behaviour) |
| `ease-in` | Accelerates into the sweep |
| `ease-out` | Decelerates out of the sweep |
| `bounce` | Overshoots and settles |

### 8.3 Single-Ended Gradients

A gradient with only `hi` or only `lo` defined produces a monochrome
brightness sweep. The missing endpoint is derived automatically
(`deriveBrighter` / `deriveDimmed`, already implemented in `NewTheme`).
Curve type still applies and controls brightness distribution.

### 8.4 Updated YAML Shape

Two new optional fields per gradient definition. Both default to
`linear` so existing themes require no migration.

```yaml
aurora-borealis:
  steps: 64
  curve: ease-in      # optional, default: linear
  easing: bounce      # optional, default: linear
  animate: true
  hi:
    ansi16: cyan
    true-color: "#00E5FF"
  lo:
    ansi16: magenta
    true-color: "#B388FF"
```

### 8.5 Gradient Attribute Cascade

Gradient attributes can be specified at two levels, the lower level
overriding the higher:

```text
Theme gradient definition   (starship.yaml - base)
          |
          v
View-level banner override  (jay.ui.yml - ui.<view>.banner)
```

The banner section in `jay.ui.yml` already carries `steps`. It grows
to support all four gradient attributes as optional overrides:

```yaml
ui:
  highway:
    banner:
      position: bottom
      tick: 100
      steps: 64        # already exists
      curve: ease-out  # new
      easing: bounce   # new
      animate: true    # new
  porthole:
    banner:
      steps: 64
      curve: linear
      easing: ease-out
  linear:
    banner:
      steps: 64
```

`tweak` makes this cascade visible when editing gradient attributes.
If a view-level override is active and winning, the editor shows which
level is in effect and allows editing either level independently.

---

## 9. Colour Tier Model

### 9.1 The Four Tiers

| Tier | Detection |
| --- | --- |
| TrueColor | `colorprofile.Detect` returns TrueColor |
| ANSI256 | `colorprofile.Detect` returns ANSI256 |
| ANSI16 | `colorprofile.Detect` returns ANSI |
| NoColor | `colorprofile.Detect` returns NoColor |

Detection is already handled by `colorprofile` in `NewTheme`
(`src/prism/contract/theme.go`). No new runtime logic is required.

### 9.2 Automatic Downgrade Derivation

When the user assigns a hex value to a palette role, `tweak` silently
derives ANSI256 and ANSI16 values and populates them in the working
state. Both derived values are shown alongside the hex so the user can
inspect and override if the automatic match is poor.

### 9.3 Preview Tier Switching

The live preview panel exposes a key to simulate a lower-capability
terminal:

```text
  T: toggle tier  [TrueColor | ANSI256 | ANSI16]
```

Switching re-renders the preview using only the fields appropriate to
the selected tier. This is a pure rendering switch; no data changes.

---

## 10. Gradient Visualiser

### 10.1 Interface

The gradient workshop displays a realtime animated visualisation of the
seed gradient. The visualisation is behind a `GradientVisualiser`
interface so it can be swapped without touching any workshop logic.

```go
// GradientVisualiser renders a realtime visual representation of a
// gradient for display in the gradient workshop seed screen.
// Implementations are interchangeable - the workshop calls Render
// on every state change and displays the result without knowing
// which visualiser is active.
type GradientVisualiser interface {
    // Render produces the terminal string for the current gradient.
    // steps is the pre-interpolated colour slice from the working
    // state. animFrame is the current animation tick counter, used
    // by animated visualisers to advance their state.
    Render(steps []contract.Color, curve CurveType,
        easing EasingPreset, animFrame int) string

    // Name returns the display name shown in the visualiser picker.
    Name() string
}
```

### 10.2 Lazy Registration

Visualiser implementations are registered only when `tweak` is
invoked. Bootstrap acts as the launchpad: the `tweak` cobra command's
bootstrap path calls the registration function; the `walk`, `sprint`,
and `query` paths do not. This is consistent with the existing
`prism.RegisterAll()` pattern - explicit registration, no `init()`.

```go
// RegisterVisualisers registers all GradientVisualiser implementations
// for use in the gradient workshop. Called only from the tweak
// bootstrap path.
func RegisterVisualisers() {
    visualiserRegistry.Register(&WaveformVisualiser{})
    visualiserRegistry.Register(&SweepVisualiser{})
    visualiserRegistry.Register(&BloomVisualiser{})
    visualiserRegistry.Register(&BandsVisualiser{})
}
```

### 10.3 User Selection

The user cycles through available visualisers with `V` from within the
gradient workshop seed screen. The active visualiser name is persisted
in `jay.ui.yml` so the preference survives sessions:

```yaml
tweak:
  gradient-visualiser: waveform  # waveform | sweep | bloom | bands
```

### 10.4 Concrete Implementations

| File | Name | Description |
| --- | --- | --- |
| `visualiser-waveform.go` | waveform | Default. A sine wave of half-block characters whose colour follows the gradient sweep. Wave shape reflects the curve type; animation speed reflects the easing preset. A flat sweep bar beneath it provides the accurate colour reference. |
| `visualiser-sweep.go` | sweep | Plain animated horizontal sweep bar. Simple fallback. |
| `visualiser-bloom.go` | bloom | Gradient radiates outward from a central point. Each ring takes one gradient step. Pulses slowly using the easing preset. |
| `visualiser-bands.go` | bands | Flat gradient bar with a braille curve indicator row above it showing the easing curve shape. |

The waveform is the recommended default and the first to be
implemented. The others are registered and available but lower
implementation priority.

---

## 11. Gradient Workshop (Entry Point 1)

The primary creative flow. Wizard-style, one concern per screen. All
changes are applied to the working state only; nothing is written to
disk until the user saves or exits.

### Screen 1 - Define Seed Gradient

The user defines or selects the seed gradient:

- Choose an existing named gradient from the theme as a starting point,
  or start from scratch.
- Assign a name (required; stored as a named entry in
  `highlights.gradients` independently of any palette role
  assignments harvested from it).
- Set hi colour (hex input with live swatch).
- Set lo colour (hex input with live swatch), or leave empty for a
  monochrome sweep.
- Set step count, curve type, and animation easing preset.
- The `GradientVisualiser` renders a realtime animated preview of the
  gradient, updating on every keystroke. The user can cycle
  visualisers with `V`.

### Screen 2 - Colour Grid

The seed gradient is rendered as a navigable grid of N colour cells,
one per step. Each cell shows its step index, a filled block in that
step's true-color hex, and the hex value. The user cursors around the
grid with arrow keys and selects a cell to open the assignment panel.

### Screen 3 - Assignment Panel (inline, alongside grid)

The right half of the screen shows:

- The selected step's hex, derived ANSI256 index, derived ANSI16 name.
- A list of all palette roles with their current assignments.
- Navigation to bind the selected step to a role.

As assignments are made, the live preview panel updates in realtime.

### Screen 4 - Live Preview

Renders the current working-state palette using the real view renderers
against the configured preview content path. The user can switch view
with `V` and switch colour tier with `T`.

---

## 12. Palette Editor (Entry Point 2)

Menu-driven. Lists all palette roles. The user selects a role and edits:

- Current hex value.
- Derived ANSI256 and ANSI16 values (shown; overridable).
- A colour swatch.
- Provenance: which gradient step this colour was harvested from, if
  any (informational only).

All changes apply to the working state. The live preview panel runs
alongside.

**Integration point:** `contract.Palette` and `contract.SemanticColour`
in `src/prism/contract/theme.go`. `NewTheme` re-runs from the edited
working-state palette to produce a new `Theme` for the preview
renderer.

---

## 13. Bindings Editor (Entry Point 3)

Simple two-column picker. Left column: component slot names
(`activity-control`, `banner-control`, `periscope-control`). Right
column: named gradients from the theme. Arrow keys to move; enter to
bind. Changes apply to the working state.

A gradient swatch (animated if `animate: true`) is shown for the
currently highlighted gradient.

**Integration point:** `palette.Highlights.Components` in the theme
YAML, surfaced as `Theme.HighlightsComponents` at runtime.

---

## 14. Import Theme (Entry Point 4)

Short wizard. Converts an external theme file into a first-pass jay
theme loaded directly into the working state for immediate refinement.
The import layer always produces full true-color output so upscaling
is not needed after import.

### 14.1 Supported Source Formats

| Source | File format |
| --- | --- |
| iTerm2 | XML plist (`.itermcolors`) |
| VS Code | JSON (`.json`) |
| Alacritty | TOML (`.toml`) |
| Warp | YAML (`.yaml`) |
| Windows Terminal | JSON (`.json`, low priority) |

### 14.2 Auto-Mapping

Each source format defines colour roles with its own names. The import
layer maps source roles onto jay palette roles using a default mapping
table per format. Unmapped source roles are surfaced to the user in
the review step so they can be routed manually or discarded.

Example iTerm2 mapping:

| iTerm2 role | Jay palette role |
| --- | --- |
| Foreground Color | file |
| Background Color | (not mapped - terminal bg) |
| Bold Color | directory |
| Selection Color | progress |
| Selected Text Color | summary-value |
| Ansi 1 Color (red) | error |
| Ansi 2 Color (green) | progress |
| Ansi 5 Color (magenta) | branch |
| Ansi 6 Color (cyan) | worker |

### 14.3 Wizard Steps

1. Select source format.
2. Point to the source file (file picker).
3. Review auto-mapping - a two-column table: source role with colour
   swatch, jay role with colour swatch. The user can reassign any row.
   Unmapped source roles appear at the bottom for manual routing.
4. Name the new theme.
5. Confirm - loads the mapped palette into the working state and opens
   the palette editor for further refinement. Does not write to disk
   until the user saves.

Save after import always creates a new file. Save-as is also available.
An imported theme never silently overwrites an existing theme file.

---

## 15. Live Preview - Technical Design

### 15.1 Content Source

The preview path is configured in `jay.ui.yml` under a new `tweak:`
section:

```yaml
tweak:
  preview-path: ""           # default: $HOME
  default-view: highway      # highway | porthole | linear
  gradient-visualiser: waveform
```

### 15.2 Preview Session

The existing session wires up the full lifecycle including resume state
persistence, fault handlers, and save-on-panic behaviour. None of this
is appropriate for `tweak`. A new `PreviewSession` abstraction replaces
the session for the `tweak` context:

- Provides the same interface the coordinator expects.
- Strips out persistence, resume, and fault handling entirely.
- Ctrl-C exits cleanly without saving any navigation state.
- Is additive - the existing session is not modified.

This is a standalone piece of work that must be completed before the
`tweak` coordinator can be implemented (see section 17, issue 4).

### 15.3 View Isolation

The highway, porthole, and linear view models are already largely
message-driven with no direct session reference (confirmed by
inspection of `highway/model.go`: the model receives `OvertureMsg`
and responds to standard message types). The `tweak` coordinator drives
them with the same message types via the preview traversal rather than
a real traversal. A brief audit of each view model is required to
confirm no hidden session coupling before the coordinator is
implemented.

### 15.4 Traversal Mode

A real `agenor` traversal runs against the preview path. The client
callback performs a benign sleep per node:

```go
// The preview callback sleeps briefly per node to simulate
// realistic worker behaviour. No actions or pipelines execute.
func previewCallback(node core.Node) error {
    time.Sleep(previewNodeDelay)
    return nil
}
```

`previewNodeDelay` defaults to `interaction.tui.per-item-delay` from
`jay.yml` (currently `"1s"`).

### 15.5 Auto-Restart Loop

Navigation is perpetual. When the traversal receives the `End`
lifecycle event, `tweak` re-arms a new traversal from the same preview
path rather than tearing down. This loop continues until the user
explicitly exits. No iteration cap.

**Integration point:** a new `tweak-coordinator.go` in
`src/app/controller/` handles the `End` event by re-dispatching rather
than calling `tea.Quit`. The existing `Coordinator` is not modified.

### 15.6 Palette Hot-Swap

When the user changes a palette role or gradient, `tweak` calls
`contract.NewTheme` with the current working-state palette to produce
a new `Theme`, then sends a `ThemeUpdateMsg` to the active view model:

```go
// ThemeUpdateMsg carries a freshly built Theme to the active view
// model. Sent by tweak whenever a working-state palette change
// occurs.
type ThemeUpdateMsg struct {
    Theme Theme
}
```

Each view model's `Update` method gains a `ThemeUpdateMsg` case that
stores the new theme. The next render uses it automatically.

---

## 16. Integration Points Summary

| Area | File | Change |
| --- | --- | --- |
| Gradient shape + easing | `src/prism/contract/gradients.go` | Add `Curve`, `Easing` to `GradientDef`; update `InterpolateBetween` to accept a curve function |
| Gradient attribute cascade | `src/app/bedrock/view-loader.go` | Extend banner config struct with `Curve`, `Easing`, `Animate` |
| Upscale palette | `src/prism/contract/upscale.go` (new) | `UpscalePalette(p Palette) Palette` |
| Theme hot-swap message | `src/prism/contract/messages.go` (new) | `ThemeUpdateMsg` |
| ANSI16 canonical hex map | `src/prism/contract/ansi16-canonical.go` (new) | Canonical name-to-hex table |
| Nearest-colour algorithms | `src/prism/contract/colour-distance.go` (new) | ANSI256 and ANSI16 nearest-colour matching |
| Theme write path | `src/app/bedrock/theme-writer.go` (new) | `ThemeWriter.Write` |
| View config write path | `src/app/bedrock/view-writer.go` (new) | `ViewConfigWriter.Write` |
| Preview session | `src/app/controller/preview-session.go` (new) | Lightweight session; no persistence, no fault handler |
| Preview content config | `jay.ui.yml` | New `tweak:` section |
| `tweak` command | `src/app/command/tweak-cmd.go` (new) | Thin cobra adapter |
| `tweak` coordinator | `src/app/controller/tweak-coordinator.go` (new) | State model, dirty tracking, auto-restart, exit flow |
| Gradient visualiser interface | `src/prism/tweak/workshop/visualiser.go` (new) | `GradientVisualiser` interface and registry |
| Waveform visualiser | `src/prism/tweak/workshop/visualiser-waveform.go` (new) | Default implementation |
| Sweep visualiser | `src/prism/tweak/workshop/visualiser-sweep.go` (new) | Plain sweep fallback |
| Bloom visualiser | `src/prism/tweak/workshop/visualiser-bloom.go` (new) | Radial bloom |
| Bands visualiser | `src/prism/tweak/workshop/visualiser-bands.go` (new) | Braille curve overlay |
| Gradient workshop TUI | `src/prism/tweak/workshop/` (new) | Seed screen, grid, assignment panel |
| Palette editor TUI | `src/prism/tweak/palette/` (new) | Role list and editor |
| Bindings editor TUI | `src/prism/tweak/bindings/` (new) | Two-column picker |
| Import wizard TUI | `src/prism/tweak/importer/` (new) | Format parsers and mapping tables |

---

## 17. Implementation Issues

Issues are ordered by dependency. Each issue is independently
deliverable with its own Ginkgo/Gomega test suite.

### Issue 1 - Foundation: Data and IO

**Scope:** Pure data and IO. No TUI, no commands.

- `UpscalePalette` with full rule set (section 7).
- `colour-distance.go` - ANSI256 and ANSI16 nearest-colour algorithms.
- `ansi16-canonical.go` - canonical name-to-hex table.
- `ThemeWriter` - atomic YAML write via `gopkg.in/yaml.v3`.
- `ViewConfigWriter` - atomic YAML write for `jay.ui.yml`.

**Acceptance:** all upscale combinations covered by table-driven
Ginkgo specs; round-trip test confirms written YAML is loadable by
`ThemeLoader.Load`.

### Issue 2 - Gradient Model Extension

**Scope:** Extend the existing gradient infrastructure.

- Add `Curve` and `Easing` fields to `GradientDef` in
  `src/prism/contract/`.
- Update `InterpolateBetween` to accept a curve function parameter.
- Implement the four curve functions (linear, ease-in, ease-out, bell).
- Extend banner config structs in `view-loader.go` with `Curve`,
  `Easing`, `Animate`.
- Implement cascade resolution: view-level override wins over theme
  gradient base.

**Acceptance:** existing tests pass unchanged; new specs cover each
curve type and cascade override combination.

### Issue 3 - Gradient Visualiser Interface and Implementations

**Scope:** Standalone visualiser package with no TUI wiring.

- `GradientVisualiser` interface and registry.
- `RegisterVisualisers()` - called only from `tweak` bootstrap path.
- `WaveformVisualiser` - primary implementation.
- `SweepVisualiser` - plain sweep fallback.
- `BloomVisualiser` and `BandsVisualiser` - lower priority, can follow
  in a sub-issue.

**Acceptance:** spy renderer tests confirm each visualiser produces
non-empty output for a range of step counts, curve types, and easing
presets. Registration only occurs when explicitly called.

### Issue 4 - Preview Session and View Isolation Audit

**Scope:** Infrastructure for safe preview traversal.

- Audit highway, porthole, and linear view models for hidden session
  coupling.
- `PreviewSession` - lightweight session abstraction. No persistence,
  no resume, no fault handler. Ctrl-C exits cleanly.
- Document any view model changes required by the audit.

**Acceptance:** `PreviewSession` drives a real agenor traversal of a
test directory and produces the expected lifecycle messages without
touching any persistence path.

### Issue 5 - `tweak` Command and Coordinator Skeleton

**Scope:** Command wiring and state model. No editor screens yet.

- `tweak-cmd.go` - thin cobra adapter; calls `RegisterVisualisers()`
  in its bootstrap path.
- `tweak-coordinator.go` - implements the three-layer state model,
  dirty tracking, undo, exit flow with correct user-facing prompts,
  and the perpetual auto-restart loop.
- Home screen TUI with all four entry points visible but not yet
  functional.
- `ThemeUpdateMsg` and the `Update` case in each view model.

**Acceptance:** `jay tweak` launches, shows the home screen, handles
undo and exit prompts correctly, and drives the live preview with
auto-restart.

### Issue 6 - Gradient Workshop

**Scope:** Full gradient workshop TUI.

- Seed screen with gradient definition and live visualiser.
- Colour grid navigation.
- Assignment panel with palette role binding.
- Named gradient saved to `highlights.gradients` in working state.

**Acceptance:** a gradient can be defined, stepped through, and its
steps assigned to palette roles; the live preview updates in realtime.

### Issue 7 - Palette Editor

**Scope:** Direct palette role editing.

- Role list with current values.
- Hex editor with live swatch.
- Derived ANSI256 and ANSI16 display with manual override.
- Provenance display for gradient-harvested colours.

**Acceptance:** editing a role updates the live preview and the working
state; derived values update automatically on hex change.

### Issue 8 - Bindings Editor

**Scope:** Component-to-gradient binding.

- Two-column picker.
- Animated gradient swatch for highlighted gradient.

**Acceptance:** a binding change is reflected in the working state and
survives a save/reload cycle.

### Issue 9 - Import Theme

**Scope:** External theme conversion.

- iTerm2 parser (XML plist).
- VS Code parser (JSON).
- Alacritty parser (TOML).
- Warp parser (YAML).
- Auto-mapping tables per format.
- Review and reassignment UI.
- Save / save-as after import (always creates a new file).

**Acceptance:** a real iTerm2 `.itermcolors` file is imported, mapped,
and written as a valid jay theme loadable by `ThemeLoader.Load`.

---

## 18. New Config Fields

### 18.1 `jay.ui.yml` - tweak section (new)

```yaml
tweak:
  preview-path: ""                # directory used for preview traversal
  default-view: highway           # highway | porthole | linear
  gradient-visualiser: waveform   # waveform | sweep | bloom | bands
```

### 18.2 `jay.ui.yml` - banner sections (extended)

```yaml
ui:
  highway:
    banner:
      steps: 64
      curve: linear    # new
      easing: linear   # new
      animate: true    # new
  porthole:
    banner:
      steps: 64
      curve: linear    # new
      easing: linear   # new
      animate: true    # new
  linear:
    banner:
      steps: 64
      curve: linear    # new
      easing: linear   # new
```

### 18.3 Theme YAML - gradient entries (extended)

```yaml
palette:
  highlights:
    gradients:
      aurora-borealis:
        steps: 64
        curve: ease-in    # new
        easing: bounce    # new
        animate: true
        hi:
          ansi16: cyan
          true-color: "#00E5FF"
        lo:
          ansi16: magenta
          true-color: "#B388FF"
```

---

## 19. Out of Scope (First Iteration)

- `jay.yml` main config editing.
- `--theme` persistent flag propagation to subcommands (deferred).
- NerdFont glyph editing (planned separately).
- Multi-stop gradients (hi-mid-lo or arbitrary colour stops).
- Runtime flag override of gradient attributes (third cascade level).
- Per-action undo stack (undo is a full reset to upscaled state).
- Windows Terminal import (low priority; format overlaps VS Code JSON).
