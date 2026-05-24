# Gradients Configuration Guide

## Overview

The highway view animation system supports optional colour gradient overlays that apply smooth colour transitions across animation frames. This creates visually engaging effects for the spinner/film strip animations in the highway view.

Gradients are configured via the theme's `highlights.gradients` section and applied to specific components via `highlights.components`.

## Configuration Location

Configure gradients in your theme file (`palette.jay.yml` or `jay.ui.yml`):

```yaml
highlights:
  gradients:
    highway-animation:
      hi: red
      lo: blue
  
    sunset:
      hi: orange
      lo: yellow
      steps: 16  # default is 8

  components:
    highway-animation: highway-animation  # map component to gradient
```

## Gradient Definition Syntax

Each gradient definition has the following fields:

| Field | Type | Required? | Description |
|-------|------|-----------|-------------|
| `hi` | colour name or hex | No | High-end colour (brighter endpoint) |
| `lo` | colour name or hex | No | Low-end colour (darker endpoint) |
| `steps` | integer | No | Number of interpolated colours (default: 8) |

### Colour Values

Accepted colour formats:

- **ANSI-16 names**: `"red"`, `"cyan"`, `"bright-blue"`
- **ANSI-256 numbers**: `"123"`  
- **TrueColor hex**: `"#FF5733"`, `"#00AAFF"`
- **Raw colour functions**: Use lipgloss-compatible format

### Single Endpoint Gradients

You can define a gradient with only one endpoint - the system will derive the other:

```yaml
highlights:
  gradients:
    sunset:
      hi: orange  # lo will be auto-derived (darker)
```

When only `hi` is specified, `lo` is derived by dimming. When only `lo` is specified, `hi` is derived by brightening.

## Available Components

Currently supported components for gradient overlays:

- `highway-animation` - applied to frame animations (spinners)

Additional components can be added in future releases.

## Example Themes

### Simple Two-Colour Gradient

```yaml
highlights:
  gradients:
    rainbow:
      hi: red
      lo: purple
  
  components:
    highway-animation: rainbow
```

### Multi-Stop Custom Gradient

```yaml
highlights:
  gradients:
    custom-gradient:
      hi: "255,100,50"  # RGB with alpha
      lo: "100,200,255"
      steps: 12

  components:
    highway-animation: custom-gradient
```

## Technical Details

- **Interpolation Method**: Linear interpolation across all RGB channels independently
- **Sliding Window Effect**: Gradient positions slide across frames each tick, reversing direction at boundaries (no wrap-around)
- **State Management**: All state lives in memory during one CLI invocation - no JSON persistence needed
- **Performance**: Pre-computed colour steps cached in theme; on-demand computation supported

## Backward Compatibility

Gradient overlays are optional. Lanes without configured gradients render using their default static styles with no change to existing behavior.

## See Also

- [Prism Theme Schema](./theme.schema.md)
- [Highway View Configuration](./highway-view.md)
- [Colour Tier Detection](./colour-profiles.md)
