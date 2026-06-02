# Additional Built-in Themes Design

**Date:** 2026-06-02
**Status:** Approved

## Goal

Add six new built-in themes (Nord, One Dark, Gruvbox, Catppuccin Mocha, Tokyo Night, Solarized Dark) and refactor the existing Dracula theme to follow a consistent file-per-theme pattern.

## File Organization

Each theme lives in its own file in the `theme/` package. The file self-registers via `init()`. No central registration needed — adding a new theme is just adding a new file.

**New files:**
- `theme/dracula.go` — moved from `theme/theme.go`
- `theme/nord.go`
- `theme/one-dark.go`
- `theme/gruvbox.go`
- `theme/catppuccin.go`
- `theme/tokyo-night.go`
- `theme/solarized.go`

**Modified:** `theme/theme.go` — `init()` becomes `builtins = map[string]Theme{}`; `dracula()` function removed.

## Self-Registration Pattern

Every theme file follows this structure:

```go
package theme

func init() {
    builtins["<name>"] = <name>()
}

func <name>() Theme {
    return Theme{
        Name: "<name>",
        Tokens: map[string]TokenStyle{ ... },
    }
}
```

## Token Colors

All themes define these tokens: `success`, `warning`, `error` (bold), `info`, `accent`, `dim`, `header` (bold+underline), `key`, `value`, `pink`, `orange`, `shade` (background-only — sets `Background` field, no `Color` field, matching the dracula shade pattern).

| Token | Nord | One Dark | Gruvbox | Catppuccin | Tokyo Night | Solarized |
|---|---|---|---|---|---|---|
| success | `#A3BE8C` | `#98C379` | `#B8BB26` | `#A6E3A1` | `#9ECE6A` | `#859900` |
| warning | `#EBCB8B` | `#E5C07B` | `#FABD2F` | `#F9E2AF` | `#E0AF68` | `#B58900` |
| error | `#BF616A` bold | `#E06C75` bold | `#FB4934` bold | `#F38BA8` bold | `#F7768E` bold | `#DC322F` bold |
| info | `#88C0D0` | `#61AFEF` | `#83A598` | `#89DCEB` | `#7DCFFF` | `#268BD2` |
| accent | `#B48EAD` | `#C678DD` | `#D3869B` | `#CBA6F7` | `#BB9AF7` | `#6C71C4` |
| dim | `#4C566A` | `#5C6370` | `#928374` | `#6C7086` | `#565F89` | `#657B83` |
| header | `#81A1C1` bold+ul | `#61AFEF` bold+ul | `#83A598` bold+ul | `#89B4FA` bold+ul | `#7AA2F7` bold+ul | `#268BD2` bold+ul |
| key | `#EBCB8B` | `#E5C07B` | `#FABD2F` | `#F9E2AF` | `#E0AF68` | `#B58900` |
| value | `#ECEFF4` | `#ABB2BF` | `#EBDBB2` | `#CDD6F4` | `#C0CAF5` | `#93A1A1` |
| pink | `#BF616A` | `#E06C75` | `#D3869B` | `#F38BA8` | `#F7768E` | `#D33682` |
| orange | `#D08770` | `#D19A66` | `#FE8019` | `#FAB387` | `#FF9E64` | `#CB4B16` |
| shade (bg only) | `#3B4252` | `#2C313C` | `#3C3836` | `#181825` | `#16161E` | `#073642` |

## Testing

Add to `theme/theme_test.go`:

1. One test per new theme: `Get("<name>")` returns `ok=true` and all 8 required tokens (`success`, `warning`, `error`, `info`, `accent`, `dim`, `header`, `key`) are present in the token map.
2. One test confirming `Names()` includes all 7 built-in theme names.

The `dracula.go` migration requires no new test — existing dracula tests continue to pass unchanged.

## Dracula Migration

`dracula()` moves from `theme/theme.go` to `theme/dracula.go` and gains its own `init()`. The function body is unchanged. The old `init()` in `theme.go` is replaced with `builtins = map[string]Theme{}`.
