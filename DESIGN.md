# oc-color — Design & Roadmap

## Vision

`oc-color` is an `oc` CLI plugin that wraps OpenShift commands and colorizes/beautifies
output for improved readability. Think `diff` → `colordiff`, but for `oc`.

---

## Architecture

```
User: oc color get pods

┌─────────────────────┐
│   CLI Arg Parser     │  parse flags (--color, --theme, --no-color),
│   (stdlib flag or    │  strip plugin-specific flags, forward rest to oc
│    github.com/spf13/ │
│    cobra — TBD)      │
└────────┬────────────┘
         │ args (no oc-color flags)
         ▼
┌─────────────────────┐
│   OC Executor        │  exec.Command("oc", args...)
│   (exists in v0.1)   │  captures stdout + stderr separately
└────────┬────────────┘
         │ raw output
         ▼
┌─────────────────────┐
│  Output Processor    │  pipeline of formatters
│  Pipeline            │
│                      │
│  1. TTY Detector     │  — disable colors when piped (auto mode)
│  2. Context Sniffer  │  — detect "get pods" vs "describe" vs "logs" vs -o json/yaml
│  3. Theme Resolver   │  — load Dracula (default) or custom theme from config
│  4. Status Colorizer │  — Running→green, CrashLoopBackOff→bold red, etc.
│  5. Table Styler     │  — headers, dividers, alignment
│  6. Syntax High-     │  — JSON / YAML syntax highlighting via chroma
│     lighter (opt)     │     (only when -o json/yaml detected)
└────────┬────────────┘
         │ colorized output
         ▼
┌─────────────────────┐
│   Terminal Writer    │  write to stdout / stderr with proper
│                      │  reset codes at end
└─────────────────────┘
```

---

## Phased Roadmap

Each phase is self-contained and can be completed in one session. They build on
each other but are ordered so the plugin is usable after every phase.

---

### Phase 1 — Foundation (current state → v0.2)

**Goal:** A working, installable plugin with basic colorization and config.

- [ ] **Initialize Go module correctly**
  - Choose module path (`github.com/<your-org>/oc-color`)
  - Add `go.sum`, pin dependencies

- [ ] **Add YAML config system**
  - Search paths: `$XDG_CONFIG_HOME/oc-color/config.yaml` → `~/.config/oc-color/config.yaml` → built-in defaults
  - Config sections:
    ```yaml
    color: always       # always | never | auto
    theme: dracula      # name or path to custom theme file
    highlight:
      json: true        # enable JSON syntax highlighting on -o json output
      yaml: true        # enable YAML syntax highlighting on -o yaml output
    ```

- [ ] **TTY detection**
  - `--color=auto` (default): colorize only when stdout is a terminal
  - `--color=always`: force colors even when piped
  - `--color=never` / `--no-color`: disable
  - Use `golang.org/x/term` for portable detection

- [ ] **Expanded status map** (covers ~30+ statuses)
  - Pod statuses, OLM statuses, build statuses, deployment statuses
  - All mapped to themed colors

- [ ] **Dracula theme** as the default built-in theme
  - Define colors in a `theme` struct, loadable from config

- [ ] **Self-test mode**
  - `oc color --dry-run` — runs the pipeline against a sample output and prints it
  - Useful for validating themes and spotting regressions

**Deliverable:** `oc color get pods` is noticeably prettier. Config file works. No color bleed when piping.

---

### Phase 2 — Table & Header Styling (v0.3)

**Goal:** `oc get` tabular output looks structured and professional.

- [ ] **Detect tabular output**
  - Heuristic: multi-line output with consistent column spacing, first line is a header
  - Commands: `oc get <resource>`, `oc get <resource> -w`

- [ ] **Header styling**
  - Bold + underline + colored
  - Detect column boundaries

- [ ] **Column styling**
  - Alternate row shading (subtle, optional)
  - Status column gets special color treatment
  - Age/human-readable durations get dimmer color

- [ ] **Divider / separator lines**
  - Lightweight horizontal rules between sections (e.g., after `oc describe` headings)

**Deliverable:** `oc color get pods`, `oc color get nodes`, `oc color get all` look clean.

---

### Phase 3 — JSON / YAML Syntax Highlighting (v0.4)

**Goal:** `oc color get pod -o json` and `-o yaml` output is syntax-highlighted.

- [ ] **JSON tokenizer** (built-in, no external dep)
  - Walk Go's `encoding/json` Decoder tokens
  - Color by token type: keys→yellow, strings→green, numbers→purple, booleans→cyan, null→dim
  - Handle nested objects, arrays, indentation

- [ ] **YAML tokenizer**
  - Lightweight line-by-line tokenization (no full parser needed)
  - Color keys, values, comments, anchors, directives

- [ ] **Respect `--color` setting** in the highlighter

**Deliverable:** `oc color get pod my-pod -o json` and `-o yaml` are syntax highlighted.

---

### Phase 4 — `oc describe` Beautification (v0.5)

**Goal:** `oc describe` output is structured and scannable.

- [ ] **Section header detection** (lines like `=== ...` or `Conditions:`)
  - Bold + accent color for section titles
  - Collapsible thinking (future stretch)

- [ ] **Key-value pair formatting**
  - Align values at the same column within a section
  - Color keys consistently

- [ ] **Event/Warning highlighting**
  - Warnings and error events in `oc describe pod` get standout colors

**Deliverable:** `oc color describe pod my-pod` is visually scannable.

---

### Phase 5 — Theme System & Customization (v0.6)

**Goal:** Users can define and share themes.

- [ ] **Theme file format**
  ```yaml
  name: dracula
  base:
    background: default
    foreground: "#F8F8F2"
  ansi:
    black: "#21222C"
    red: "#FF5555"
    green: "#50FA7B"
    yellow: "#F1FA8C"
    blue: "#BD93F9"
    magenta: "#FF79C6"
    cyan: "#8BE9FD"
    white: "#F8F8F2"
  tokens:
    running: green
    pending: yellow
    error: red
    crash_loop_back_off: bold+red
    header: bold+cyan+underline
    json_key: yellow
    json_string: green
    json_number: purple
    json_bool: cyan
    json_null: dim
  ```

- [ ] **Theme discovery**
  - Built-in: `dracula` (default)
  - Custom: `~/.config/oc-color/themes/*.yaml`
  - `oc color --list-themes` to show available

- [ ] **Truecolor / 256-color support**
  - Detect terminal color capabilities
  - Fall back gracefully: truecolor → 256 → 16 → no color

- [ ] **Theme validation**
  - `oc color --validate-theme path/to/theme.yaml`
  - Reports missing keys, invalid colors, etc.

**Deliverable:** `oc color --theme nord get pods`, custom themes, `--list-themes`.

---

### Phase 6 — Completion & Polish (v0.7+)

- [ ] **Performance optimization** — memoize regex compilation, buffer management
- [ ] **Shell completion** — `oc color completion [bash|zsh|fish]`
- [ ] **`--watch` support** — clean redraw on `oc get pods -w` (no scroll junk)
- [ ] **`oc color --version`** — prints version + git commit
- [ ] **`oc color --help`** — well-structured help output
- [ ] **`oc color --dry-run` with sample data** — show all features without a real cluster

---

## Security Considerations

| Concern | Mitigation |
|---------|-----------|
| **Supply chain** | Minimize external dependencies. Pin versions in `go.mod` + `go.sum`. Use well-audited libs only (`golang.org/x/term`, `gopkg.in/yaml.v3`). |
| **No network calls** | Plugin never reaches out. No telemetry, no update checks. Zero network in the codebase. |
| **No code execution** | Plugin does not eval or exec anything user-provided. The only `exec` is the wrapped `oc` binary. |
| **Config injection** | YAML config is strictly validated. Invalid fields are ignored with a warning. No `exec` or `include` directives in config. |
| **Secret exposure** | Plugin does not read, log, or transmit secrets. If `oc` output contains secrets, they pass through unchanged (no accidental highlighting that could expose them). |
| **File write scope** | Config writes are limited to `$XDG_CONFIG_HOME/oc-color/` and `~/.config/oc-color/`. No writes outside these paths. |

---

## Dependencies (intentionally minimal)

| Package | Why | Risk |
|---------|-----|------|
| `golang.org/x/term` | TTY detection, terminal width | Official Go team, zero network, minimal |
| `gopkg.in/yaml.v3` | Config/theme file parsing | De facto standard, widely audited |
| stdlib `encoding/json` | JSON tokenizer for highlighting | Built-in, zero risk |
| stdlib `os/exec` | Running `oc` | Built-in |

No HTTP clients, no template engines, no ORMs. The entire dependency tree fits on one screen.

---

## Quick Start for Development

```bash
# Clone / enter the directory
cd oc-color

# Build
go build -o oc-color .

# Run as plugin (binary must be named oc-color on PATH)
cp oc-color ~/.local/bin/oc-color
oc color get pods

# Or run standalone for testing
./oc-color get pods

# Run tests
go test ./...

# Lint (if golangci-lint is installed)
golangci-lint run
```

---

## Agent Dispatch Guidance (for OpenCode)

When working on this project via OpenCode agents, use these agents:

| Area | Agent | Reason |
|------|-------|--------|
| Writing Go code | `backend-dev` | Go backend logic |
| Config file parsing | `backend-dev` | File I/O, YAML parsing |
| Color/Terminal logic | `backend-dev` | ANSI codes, terminal detection |
| Security review | `security-auditor` | Before shipping — review config parsing, exec, I/O |
| Code review | `code-reviewer` | After each phase — correctness, idiomatic Go |
| Documentation | (manual) | DESIGN.md and README updates |

---

*Last updated: 2026-05-27*
