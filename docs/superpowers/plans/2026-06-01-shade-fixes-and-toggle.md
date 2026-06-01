# Shade Fixes and Toggle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the shade-bleed bug and misaligned dry-run sample, then add a `--no-shade` flag and `shade: false` config option so users can disable zebra-striping without switching themes.

**Architecture:** Four tasks in dependency order: fix the bleed in `output.go` (adding `Shade bool` field, temporarily hardcoded true), add `ShadeEnabled()` to config, wire flag+config into `main.go` and fix the sample, then update the README. No new files; all changes are inline.

**Tech Stack:** Go stdlib only. No new dependencies.

---

## File Map

| File | Role |
|------|------|
| `output/output.go` | Add `Shade bool` to `Processor`; fix shade bleed in `processLine` |
| `output/output_test.go` | New — tests for shade bleed fix and Shade=false behavior |
| `config/config.go` | Add `Shade *bool` to `Config`; add `ShadeEnabled()` helper |
| `config/config_test.go` | New — tests for ShadeEnabled |
| `main.go` | Add `--no-shade` flag; resolve shade from flag+config; update `dryRun()` call + sample + help text |
| `README.md` | Add `--no-shade` to flags table; add `shade` key to config example |

---

### Task 1: Fix shade bleed and add Shade field (`output/output.go`)

**Files:**
- Modify: `output/output.go`
- Create: `output/output_test.go`

The shade background isn't reset before each line's trailing `\n`, so it bleeds into the next row. Fix: strip `\n` before applying shade, append `theme.Reset`, restore `\n`. Add `Shade bool` to `Processor` to gate the feature; temporarily wire `Shade: true` in `main.go` so existing behavior is preserved.

- [ ] **Step 1: Write failing tests**

Create `output/output_test.go`:

```go
package output

import (
	"strings"
	"testing"

	"github.com/thephilip/oc-color/theme"
)

func shadeTheme() theme.Theme {
	return theme.Theme{
		Tokens: map[string]theme.TokenStyle{
			"shade": {Background: "#2E3040"},
		},
	}
}

func TestShadeResetAtLineEnd(t *testing.T) {
	proc := Processor{Theme: shadeTheme(), Colour: true, Shade: true}
	// Two data lines: lineNum=1 (no shade), lineNum=2 (shade)
	got := proc.Process("line one\nline two\n")
	lines := strings.Split(got, "\n")
	// lines[1] is the shaded line — must end with theme.Reset
	if !strings.HasSuffix(lines[1], theme.Reset) {
		t.Errorf("shaded line must end with Reset before newline, got: %q", lines[1])
	}
}

func TestShadeDisabledSkipsStripe(t *testing.T) {
	th := shadeTheme()
	proc := Processor{Theme: th, Colour: true, Shade: false}
	got := proc.Process("line one\nline two\n")
	shade := th.Tokens["shade"].BackgroundSequence()
	if strings.Contains(got, shade) {
		t.Errorf("shade must not appear when Processor.Shade is false, got: %q", got)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test -run 'TestShadeResetAtLineEnd|TestShadeDisabledSkipsStripe' -v ./output/...
```

Expected: compile error — `Shade` field not defined on `Processor`.

- [ ] **Step 3: Add `Shade bool` to `Processor` and fix the bleed**

In `output/output.go`, update the `Processor` struct:

```go
type Processor struct {
	Theme   theme.Theme
	Colour  bool
	Shade   bool
	columns []column
	lineNum int
}
```

Replace the shade block in `processLine` (currently lines 167–172):

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

- [ ] **Step 4: Temporarily wire `Shade: true` in `main.go`**

In `main.go`, find the `Processor` construction (two places):

First occurrence — regular processing:
```go
proc := output.Processor{Theme: th, Colour: useColor, Shade: true}
```

Second occurrence — inside `dryRun()`:
```go
proc := output.Processor{Theme: th, Colour: useColor, Shade: true}
```

This preserves existing shade behavior until Task 3 wires up the real config value.

- [ ] **Step 5: Run tests to confirm they pass**

```bash
go test -run 'TestShadeResetAtLineEnd|TestShadeDisabledSkipsStripe' -v ./output/...
```

Expected:
```
--- PASS: TestShadeResetAtLineEnd (0.00s)
--- PASS: TestShadeDisabledSkipsStripe (0.00s)
PASS
```

- [ ] **Step 6: Run full suite**

```bash
go test ./...
```

Expected: all pass.

- [ ] **Step 7: Commit**

```bash
git add output/output.go output/output_test.go main.go
git commit -m "fix: correct shade bleed and add Shade field to Processor"
```

---

### Task 2: Add shade config (`config/config.go`)

**Files:**
- Modify: `config/config.go`
- Create: `config/config_test.go`

Add `Shade *bool` to `Config` and a `ShadeEnabled()` helper. Using a pointer allows distinguishing "not set in file" (nil → true) from "explicitly false".

- [ ] **Step 1: Write failing tests**

Create `config/config_test.go`:

```go
package config

import (
	"testing"
)

func TestShadeEnabledNil(t *testing.T) {
	cfg := Config{} // Shade not set → nil
	if !cfg.ShadeEnabled() {
		t.Error("nil Shade should default to enabled")
	}
}

func TestShadeEnabledFalse(t *testing.T) {
	f := false
	cfg := Config{Shade: &f}
	if cfg.ShadeEnabled() {
		t.Error("Shade=false should return disabled")
	}
}

func TestShadeEnabledTrue(t *testing.T) {
	tr := true
	cfg := Config{Shade: &tr}
	if !cfg.ShadeEnabled() {
		t.Error("Shade=true should return enabled")
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test -run 'TestShadeEnabled' -v ./config/...
```

Expected: compile error — `ShadeEnabled` not defined, `Shade` field not in `Config`.

- [ ] **Step 3: Update `config/config.go`**

Replace the file contents with:

```go
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Color string `yaml:"color"`
	Theme string `yaml:"theme"`
	Shade *bool  `yaml:"shade,omitempty"`
}

func (c Config) ShadeEnabled() bool {
	if c.Shade == nil {
		return true
	}
	return *c.Shade
}

func Default() Config {
	return Config{
		Color: "auto",
		Theme: "dracula",
	}
}

func Load() (Config, error) {
	cfg := Default()

	path, err := configFilePath()
	if err != nil {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func configFilePath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "oc-color", "config.yaml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "oc-color", "config.yaml"), nil
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test -run 'TestShadeEnabled' -v ./config/...
```

Expected:
```
--- PASS: TestShadeEnabledNil (0.00s)
--- PASS: TestShadeEnabledFalse (0.00s)
--- PASS: TestShadeEnabledTrue (0.00s)
PASS
```

- [ ] **Step 5: Run full suite**

```bash
go test ./...
```

Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add config/config.go config/config_test.go
git commit -m "feat: add Shade config field and ShadeEnabled helper"
```

---

### Task 3: Wire `--no-shade`, fix dry-run sample, update help (`main.go`)

**Files:**
- Modify: `main.go`

Add `--no-shade` flag. Resolve shade from flag + config (flag wins). Replace the hardcoded `Shade: true` from Task 1 with the resolved value. Rewrite the `dryRun()` sample with correct column alignment. Add `--no-shade` to the help text.

- [ ] **Step 1: Add `noShade bool` to the `flags` struct**

In `main.go`, update the `flags` struct:

```go
type flags struct {
	colorMode       string
	themeName       string
	dryRun          bool
	showVer         bool
	showHelp        bool
	listThemes      bool
	validateTheme   string
	completionShell string
	showUpgrade     bool
	watchMode       bool
	noShade         bool
}
```

- [ ] **Step 2: Parse `--no-shade` in `parseFlags`**

In the `parseFlags` switch, add a case after `--no-color`:

```go
case arg == "--no-shade":
	f.noShade = true
```

- [ ] **Step 3: Resolve shade after loading config**

After the theme lookup succeeds (after the `if !ok { ... os.Exit(1) }` block), add:

```go
shadeEnabled := cfg.ShadeEnabled()
if flags.noShade {
	shadeEnabled = false
}
```

- [ ] **Step 4: Replace hardcoded `Shade: true` with resolved value**

Update the `Processor` construction for regular processing:

```go
proc := output.Processor{Theme: th, Colour: useColor, Shade: shadeEnabled}
```

Update the `dryRun` call and function signature. Change the call site:

```go
if flags.dryRun {
	dryRun(th, useColor, shadeEnabled)
	return
}
```

Change the `dryRun` function signature and its internal `Processor`:

```go
func dryRun(th theme.Theme, useColor bool, shade bool) {
	sample := `NAMESPACE     NAME                        READY   STATUS              RESTARTS   AGE
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
`
	proc := output.Processor{Theme: th, Colour: useColor, Shade: shade}
	fmt.Print(proc.Process(sample))
}
```

The column positions in this sample (all rows):

| Column | Start | Width |
|--------|-------|-------|
| NAMESPACE | 0 | 14 |
| NAME | 14 | 28 |
| READY | 42 | 8 |
| STATUS | 50 | 20 |
| RESTARTS | 70 | 11 |
| AGE | 81 | — |

- [ ] **Step 5: Add `--no-shade` to `printHelp()`**

In `printHelp()`, add after the `--no-color` line:

```
  --no-shade           Disable zebra-stripe row shading
```

The updated flags section of the help string:

```go
fmt.Print(`oc-color — colorize oc command output

Usage:
  oc color [flags] -- <oc-args>
  oc color completion <bash|zsh|fish>
  oc color upgrade

Flags:
  --color <mode>       Color mode: always, never, auto (default: auto)
  --no-color           Shorthand for --color=never
  --no-shade           Disable zebra-stripe row shading
  --theme <name>       Theme name (default: dracula)
  --list-themes        List available themes
  --validate-theme <path>  Validate a theme YAML file
  --watch              Watch mode (equivalent to oc -w). Clean in-place redraw.
  --dry-run            Process sample output to preview colors
  --version            Print version
  --help, -h           Show this help

Examples:
  oc color get pods
  oc color get pods -w
  oc color --watch get pods
  oc color --color=always get pods | less -R
  oc color --theme dracula get pods -o json
  oc color --theme nord get pods
  oc color --no-shade get pods
  oc color --list-themes
  oc color --validate-theme ~/.config/oc-color/themes/nord.yaml
  oc color --dry-run

  # Generate shell completion scripts:
  oc color completion bash > /etc/bash_completion.d/oc-color
  oc color completion zsh  > /usr/share/zsh/site-functions/_oc-color
  oc color completion fish > ~/.config/fish/completions/oc-color.fish

Config: ~/.config/oc-color/config.yaml
Themes:  ~/.config/oc-color/themes/*.yaml
`)
```

- [ ] **Step 6: Build and run full test suite**

```bash
go build ./... && go test ./...
```

Expected: clean build, all tests pass.

- [ ] **Step 7: Smoke test**

```bash
./oc-color --dry-run
./oc-color --no-shade --dry-run
```

First command: zebra-striped output with no color bleed between rows.
Second command: same output with no shading at all.

- [ ] **Step 8: Commit**

```bash
git add main.go
git commit -m "feat: add --no-shade flag, wire shade config, fix dry-run sample alignment"
```

---

### Task 4: Update README (`README.md`)

**Files:**
- Modify: `README.md`

Add `--no-shade` to the flags table and `shade` to the configuration section.

- [ ] **Step 1: Add `--no-shade` to the Flags table**

In `README.md`, find the Flags table and add a row after `--no-color`:

```markdown
| `--no-shade` | Disable zebra-stripe row shading |
```

The updated table section:

```markdown
| Flag | Description |
|------|-------------|
| `--color <mode>` | Color mode: `always`, `never`, `auto` (default: `auto`) |
| `--no-color` | Shorthand for `--color=never` |
| `--no-shade` | Disable zebra-stripe row shading |
| `--theme <name>` | Theme name (default: `dracula`) |
| `--list-themes` | List available themes |
| `--validate-theme <path>` | Validate a theme YAML file |
| `--dry-run` | Process sample output to preview colors |
| `--version` | Print version |
| `--help`, `-h` | Show help |
| `completion <shell>` | Generate shell completion script (`bash`, `zsh`, `fish`) |
```

- [ ] **Step 2: Add `shade` to the Configuration section**

Find the config YAML example in README.md:

```yaml
color: auto      # auto, always, never
theme: dracula   # theme name
```

Replace with:

```yaml
color: auto      # auto, always, never
theme: dracula   # theme name
shade: true      # set to false to disable zebra-stripe row shading
```

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: add --no-shade flag and shade config key to README"
```
