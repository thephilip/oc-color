# Flag Parsing Refactor Design

**Date:** 2026-06-02
**Status:** Approved

## Goal

Replace the manual flag loop in `main.go` with stdlib `flag.FlagSet`. Eliminate the bespoke `flagValue` helper, reduce code, and gain correct `--help`/`-h` handling for free.

## Approach

Option A: `flag.FlagSet` with subcommand pre-scan.

- Pre-scan `os.Args[1:]` in `main()` for `completion` and `upgrade` subcommands before flag parsing.
- Use `flag.NewFlagSet` with `flag.ContinueOnError` for all remaining flag parsing.
- `fs.Args()` replaces the manual `remaining` accumulator.
- No new dependencies.

## `flags` Struct Changes

Fields removed:
- `showHelp` — handled automatically by `flag.ErrHelp` / `fs.Usage`
- `completionShell` — moved to pre-scan in `main()`
- `showUpgrade` — moved to pre-scan in `main()`

Field added:
- `noColor bool` — required so `--no-color` can be registered as a real flag; resolved to `colorMode = "never"` after `fs.Parse`

## `parseFlags` Changes

```go
func parseFlags(args []string) (flags, []string) {
    f := flags{themeName: "dracula"}
    fs := flag.NewFlagSet("oc-color", flag.ContinueOnError)
    fs.StringVar(&f.colorMode,     "color",          "",       "Color mode: always, never, auto")
    fs.BoolVar(&f.noColor,         "no-color",       false,    "Shorthand for --color=never")
    fs.BoolVar(&f.noShade,         "no-shade",       false,    "Disable zebra-stripe row shading")
    fs.StringVar(&f.themeName,     "theme",          "dracula","Theme name")
    fs.BoolVar(&f.dryRun,          "dry-run",        false,    "Process sample output to preview colors")
    fs.BoolVar(&f.showVer,         "version",        false,    "Print version")
    fs.BoolVar(&f.listThemes,      "list-themes",    false,    "List available themes")
    fs.StringVar(&f.validateTheme, "validate-theme", "",       "Validate a theme YAML file")
    fs.BoolVar(&f.watchMode,       "watch",          false,    "Watch mode")
    fs.Usage = printHelp
    _ = fs.Parse(args) // ErrHelp causes os.Exit(0) via fs.Usage; other errors are logged by flag itself
    if f.noColor {
        f.colorMode = "never"
    }
    return f, fs.Args()
}
```

`flagValue` helper is deleted.

## `main()` Changes

Subcommand pre-scan added before `parseFlags`:

```go
args := os.Args[1:]
if len(args) > 0 {
    switch args[0] {
    case "completion":
        shell := "bash"
        if len(args) > 1 {
            shell = args[1]
        }
        printCompletion(shell)
        return
    case "upgrade":
        printUpgrade()
        return
    }
}
```

Checks removed from `main()`:
- `if flags.showHelp { printHelp(); return }`
- `if flags.completionShell != "" { ... }`
- `if flags.showUpgrade { ... }`

## Behavioral Change

**Flags must precede oc args.** `flag.FlagSet` stops at the first non-flag argument, so `oc color get pods --theme dracula` would pass `--theme dracula` through to `oc`. The correct form is `oc color --theme dracula get pods`. All documented README examples already follow this order; no docs update needed beyond the commit message.

## Help Handling

`fs.Usage = printHelp` connects our existing help printer to the flag set. When `-h` or `--help` is passed, `flag.FlagSet` calls `printHelp()` then returns `flag.ErrHelp`. With `ContinueOnError`, parse returns early with zero-valued flags, so `main()` would fall through to normal execution unless we explicitly handle it:

```go
if err := fs.Parse(args); errors.Is(err, flag.ErrHelp) {
    os.Exit(0)
}
```

## Testing

`TestParseFlagsNoShade` in `main_test.go` is unchanged (same signature). New cases to add:
- `--no-color` sets `colorMode = "never"`
- `--color=always` sets `colorMode = "always"`
- `--watch get pods -w` → `watchMode=true`, remaining=`["get","pods","-w"]`

Subcommand pre-scan lives in `main()` and is covered by the existing manual test path rather than unit tests.
