# Shade Fixes and Toggle — Design Spec

**Date:** 2026-06-01
**Scope:** `output/output.go`, `config/config.go`, `main.go`
**Version target:** v0.9.0

---

## Problem

Three related issues with the zebra-stripe (shade) feature:

1. **Shade bleeding** — The background ANSI sequence set on even rows is never reset before the trailing `\n`. The terminal carries that background into the next (unshaded) row, making ALL rows after the first shaded row appear shaded.

2. **Dry-run sample misaligned** — The hand-crafted sample in `dryRun()` has inconsistent column padding. READY lands at byte position 42, 43, or 44 depending on the row (should always be 42). This causes the column parser to slice at wrong positions and produces visually crooked output.

3. **No shade toggle** — Users who dislike zebra-striping have no way to disable it without switching to a custom theme that omits the `shade` token.

---

## Design

### Fix 1 — Shade bleeding (`output/output.go`)

In `processLine`, the shade application block currently reads:

```go
if p.lineNum%2 == 0 {
    if shade := p.Theme.Tokens["shade"].BackgroundSequence(); shade != "" {
        result = strings.ReplaceAll(result, theme.Reset, theme.Reset+shade)
        result = shade + result
    }
}
```

Replace with logic that strips the trailing `\n` before applying shade, then appends `theme.Reset` and restores `\n`:

```go
if p.Shade && p.lineNum%2 == 0 {
    if shade := p.Theme.Tokens["shade"].BackgroundSequence(); shade != "" {
        nl := ""
        if strings.HasSuffix(result, "\n") {
            result = result[:len(result)-1]
            nl = "\n"
        }
        result = strings.ReplaceAll(result, theme.Reset, theme.Reset+shade)
        result = shade + result + theme.Reset + nl
    }
}
```

This ensures every shaded line ends with a full ANSI reset before the newline, preventing bleed into the next row.

---

### Fix 2 — Dry-run sample alignment (`main.go`)

Rewrite the `sample` string literal in `dryRun()` with exact column widths:

| Column   | Start | Width |
|----------|-------|-------|
| NAMESPACE | 0    | 14    |
| NAME      | 14   | 28    |
| READY     | 42   | 8     |
| STATUS    | 50   | 20    |
| RESTARTS  | 70   | 11    |
| AGE       | 81   | —     |

Every row padded to these boundaries:

```
NAMESPACE     NAME                        READY   STATUS              RESTARTS   AGE
default       web-1                       1/1     Running             0          12h
default       web-2                       0/1     CrashLoopBackOff    7          12h
default       db-0                        0/1     Pending             0          5m
default       cache-6b8d4                 0/1     ContainerCreating   0          30s
default       old-job-x7f2                0/1     Evicted             0          24h
kube-system   coredns-5d4b                1/1     Running             0          30d
kube-system   metrics-server              0/1     ImagePullBackOff    3          2h
default       batch-processor             0/1     Error               1          10m
default       init-container-pod          0/1     Init:0/1            0          1m
default       long-running                1/1     Running             0          7d
default       failed-build-1              0/1     Failed              0          1h
default       node-affinity-pod           0/1     NodeAffinity        0          15m
default       big-data                    1/1     Running             0          3d
default       pending-pod                 0/1     Unknown             0          5m
default       OOM-killed-app              0/1     OOMKilled           0          1m
default       terminated-job              0/1     Completed           0          6h
```

No logic changes — string literal only.

---

### Feature — Shade toggle

#### `config/config.go`

Add `Shade` field to `Config`, defaulting to `true`:

```go
type Config struct {
    Color string `yaml:"color"`
    Theme string `yaml:"theme"`
    Shade *bool  `yaml:"shade,omitempty"`
}

func Default() Config {
    t := true
    return Config{
        Color: "auto",
        Theme: "dracula",
        Shade: &t,
    }
}
```

Using `*bool` (pointer) allows distinguishing "user explicitly set false" from "not set" (which defaults to true). A helper resolves it:

```go
func (c Config) ShadeEnabled() bool {
    if c.Shade == nil {
        return true
    }
    return *c.Shade
}
```

#### `main.go`

Add `noShade bool` to the `flags` struct. Parse `--no-shade` in `parseFlags`. Resolve shade after loading config (flag overrides config):

```go
shadeEnabled := cfg.ShadeEnabled()
if flags.noShade {
    shadeEnabled = false
}
```

Pass into the processor:

```go
proc := output.Processor{Theme: th, Colour: useColor, Shade: shadeEnabled}
```

Update `--help` text to include `--no-shade`.

#### `output/output.go`

Add `Shade bool` to `Processor`:

```go
type Processor struct {
    Theme   theme.Theme
    Colour  bool
    Shade   bool
    columns []column
    lineNum int
}
```

The shade block in `processLine` is already gated on `p.Shade` (see Fix 1 above).

#### README

Add `--no-shade` to the Flags table. Add `shade: false` example to the Configuration section.

---

## Files Changed

| File | Changes |
|------|---------|
| `output/output.go` | Add `Shade bool` to `Processor`; fix shade bleed in `processLine` |
| `config/config.go` | Add `Shade *bool` to `Config`; add `ShadeEnabled()` helper; update `Default()` |
| `main.go` | Add `--no-shade` flag; resolve shade from flag+config; pass to `Processor`; update `dryRun()` sample; update `--help` |
| `README.md` | Add `--no-shade` to flags table; add `shade` to config example |

---

## Out of Scope

- `--shade=always|never` enum style (no meaningful "auto" mode exists)
- Per-theme shade defaults (theme controls color; config/flag controls on/off)
- Shade on odd rows instead of even (current behavior is correct, staggering bug is fixed by Fix 1)
