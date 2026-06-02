# Theme Picker Design

**Date:** 2026-06-02
**Status:** Approved

## Goal

Add `oc color themes` — an interactive TUI that lets users preview all built-in (and custom) themes with live sample output and write the selected theme to config.

## Invocation

`oc color themes` is added to the subcommand pre-scan in `main()` alongside `completion` and `upgrade`. It calls `runThemePicker()` and exits.

## Code Structure

- **`picker.go`** (package main) — `runThemePicker()` function
- **`config/config.go`** — add `Save(cfg Config) (string, error)`
- **`main.go`** — add `"themes"` case to pre-scan switch; extract dry-run sample into a package-level `const drySample` shared by `dryRun()` and `runThemePicker()`

## TUI Layout and Controls

On each keypress, the screen clears (`\033[2J\033[H`) and redraws:

```
Select a theme  (↑↓  or j/k navigate · Enter select · q quit)

  dracula
▶ catppuccin
  gruvbox
  nord
  one-dark
  solarized
  tokyo-night

── Preview ───────────────────────────────────────────────
NAMESPACE  NAME     READY  STATUS             AGE
default    web-1    1/1    Running            12h
default    web-2    0/1    CrashLoopBackOff   7h
default    db-0     0/1    Pending            5m
```

**Controls:**
- `↑` / `k` — move cursor up (wraps)
- `↓` / `j` — move cursor down (wraps)
- `Enter` — write config and exit
- `q` / `Ctrl-C` — exit without saving

**Implementation:** `runThemePicker` first checks `term.IsTerminal(int(os.Stdout.Fd()))` — if not a terminal (e.g. piped output), prints `error: theme picker requires an interactive terminal` to stderr and exits 1. Otherwise, enters raw terminal via `term.MakeRaw(int(os.Stdin.Fd()))` / `term.Restore` (`golang.org/x/term` already imported). Arrow keys read as 3-byte escape sequences (`\x1b[A` up, `\x1b[B` down). All output to stdout; stderr stays clean.

The theme list comes from `theme.Names()` (already sorted alphabetically).

The preview renders `drySample` through `output.Processor{Theme: th, Colour: true, Shade: true}`.

## Config Save

`config.Save(cfg Config) (string, error)`:
1. Resolves path via `configFilePath()`
2. Creates parent dirs with `os.MkdirAll`
3. Marshals `cfg` with `yaml.Marshal`
4. Writes with `os.WriteFile`
5. Returns resolved path and error

The picker uses it as:
```go
cfg, _ := config.Load()  // preserve existing Color/Shade
cfg.Theme = selected
path, err := config.Save(cfg)
```

On success, prints to stdout:
```
Theme set: catppuccin
Written to /home/user/.config/oc-color/config.yaml
```

On failure, prints error to stderr and exits 1.

## Testing

- `config.Save()` — creates file and parent dirs when absent; writes correct YAML; round-trips cleanly through `Load` → modify theme → `Save` → `Load`
- `config.Save()` — returns resolved path alongside error
- TUI interaction (`runThemePicker`) is not unit-testable; covered by manual verification
