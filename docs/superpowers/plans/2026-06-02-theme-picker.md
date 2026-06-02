# Theme Picker Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `oc color themes` — an interactive TUI that lets users preview all themes with live sample output and write the selected theme to config.

**Architecture:** `config.Save()` handles config writing. `picker.go` contains the TUI split into three functions: `runThemePicker` (orchestrates), `runPickerLoop` (raw terminal + input), `drawPicker` (renders). The dry-run sample is extracted to a shared `const drySample` in `main.go`. `"themes"` is pre-scanned in `main()` alongside `completion` and `upgrade`.

**Tech Stack:** `golang.org/x/term` (already a dependency) for raw terminal mode. No new dependencies.

---

## Files

- Modify: `config/config.go` — add `Save(cfg Config) (string, error)`
- Modify: `config/config_test.go` — add Save tests
- Modify: `main.go` — extract `drySample` const, add `"themes"` to pre-scan, update `printHelp()`
- Create: `picker.go` (package main)

---

### Task 1: Add `config.Save()` with TDD

**Files:**
- Modify: `config/config_test.go`
- Modify: `config/config.go`

- [ ] **Step 1: Add Save tests to `config/config_test.go`**

Replace the import block and append three test functions:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShadeEnabledNil(t *testing.T) {
	cfg := Config{}
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

func TestSaveCreatesFileAndParentDirs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := Config{Color: "always", Theme: "nord"}
	path, err := Save(cfg)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	expected := filepath.Join(home, ".config", "oc-color", "config.yaml")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file at %s: %v", path, err)
	}
}

func TestSaveRoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := Config{Color: "always", Theme: "gruvbox"}
	if _, err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after Save() error: %v", err)
	}
	if loaded.Theme != "gruvbox" {
		t.Errorf("Theme = %q, want %q", loaded.Theme, "gruvbox")
	}
	if loaded.Color != "always" {
		t.Errorf("Color = %q, want %q", loaded.Color, "always")
	}
}

func TestSaveReturnsResolvedPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path, err := Save(Config{Theme: "nord"})
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if path == "" {
		t.Error("Save() returned empty path")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("Save() returned non-absolute path: %q", path)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./config/... -run "TestSave" -v
```

Expected: compilation error — `undefined: Save`

- [ ] **Step 3: Add `Save` to `config/config.go`**

Add `"fmt"` to the import block in `config/config.go`, then append `Save` after `Load`:

```go
func Save(cfg Config) (string, error) {
	path, err := configFilePath()
	if err != nil {
		return "", fmt.Errorf("cannot resolve config path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("cannot create config directory: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("cannot marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("cannot write config: %w", err)
	}
	return path, nil
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./config/... -v
```

Expected: all tests PASS including all three new `TestSave*` tests.

- [ ] **Step 5: Commit**

```bash
git add config/config.go config/config_test.go
git commit -m "feat: add config.Save() for writing theme selection"
```

---

### Task 2: Extract `drySample` const in `main.go`

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Extract the sample text to a package-level const**

In `main.go`, add this const immediately before `func main()` (around line 33):

```go
const drySample = `NAMESPACE     NAME                        READY   STATUS              RESTARTS   AGE
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
```

- [ ] **Step 2: Update `dryRun()` to use `drySample`**

Replace the current `dryRun` function:

```go
func dryRun(th theme.Theme, useColor bool, shade bool) {
	sample := `NAMESPACE     NAME ...` // the long literal
	proc := output.Processor{Theme: th, Colour: useColor, Shade: shade}
	fmt.Print(proc.Process(sample))
}
```

With:

```go
func dryRun(th theme.Theme, useColor bool, shade bool) {
	proc := output.Processor{Theme: th, Colour: useColor, Shade: shade}
	fmt.Print(proc.Process(drySample))
}
```

- [ ] **Step 3: Build and run tests to confirm no regression**

```bash
go build ./... && go test ./...
```

Expected: clean build, all tests PASS.

- [ ] **Step 4: Commit**

```bash
git add main.go
git commit -m "refactor: extract drySample const for reuse in theme picker"
```

---

### Task 3: Create `picker.go`

**Files:**
- Create: `picker.go`

- [ ] **Step 1: Create `picker.go`**

```go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/thephilip/oc-color/config"
	"github.com/thephilip/oc-color/output"
	"github.com/thephilip/oc-color/theme"
	"golang.org/x/term"
)

func runThemePicker() {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Fprintln(os.Stderr, "error: theme picker requires an interactive terminal")
		os.Exit(1)
	}

	themes := theme.Names()
	cfg, _ := config.Load()

	cursor := 0
	for i, name := range themes {
		if name == cfg.Theme {
			cursor = i
			break
		}
	}

	selected, ok := runPickerLoop(themes, cursor)
	fmt.Print("\033[2J\033[H")
	if !ok {
		fmt.Println("No changes made.")
		return
	}

	cfg.Theme = selected
	path, err := config.Save(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Theme set: %s\n", selected)
	fmt.Printf("Written to %s\n", path)
}

func runPickerLoop(themes []string, cursor int) (selected string, ok bool) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot enter raw terminal mode: %v\n", err)
		os.Exit(1)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	drawPicker(themes, cursor)

	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf[:1])
		if n == 0 || err != nil {
			return "", false
		}
		switch buf[0] {
		case 'q', 3: // q or Ctrl-C
			return "", false
		case '\r', '\n': // Enter
			return themes[cursor], true
		case '\033': // escape sequence (arrow keys)
			n2, _ := os.Stdin.Read(buf[1:3])
			if n2 == 2 && buf[1] == '[' {
				switch buf[2] {
				case 'A': // up arrow
					cursor = (cursor - 1 + len(themes)) % len(themes)
				case 'B': // down arrow
					cursor = (cursor + 1) % len(themes)
				}
			}
		case 'k': // vim up
			cursor = (cursor - 1 + len(themes)) % len(themes)
		case 'j': // vim down
			cursor = (cursor + 1) % len(themes)
		}
		drawPicker(themes, cursor)
	}
}

func drawPicker(themes []string, cursor int) {
	fmt.Print("\033[2J\033[H")
	fmt.Print("Select a theme  (↑↓ or j/k navigate · Enter select · q quit)\r\n\r\n")
	for i, name := range themes {
		if i == cursor {
			fmt.Printf("▶ %s\r\n", name)
		} else {
			fmt.Printf("  %s\r\n", name)
		}
	}
	fmt.Print("\r\n── Preview ─────────────────────────────────────────────\r\n")
	th, ok := theme.Get(themes[cursor])
	if ok {
		proc := output.Processor{Theme: th, Colour: true, Shade: true}
		preview := strings.ReplaceAll(proc.Process(drySample), "\n", "\r\n")
		fmt.Print(preview)
	}
}
```

- [ ] **Step 2: Verify it builds**

```bash
go build ./...
```

Expected: clean build, no errors.

- [ ] **Step 3: Commit**

```bash
git add picker.go
git commit -m "feat: add theme picker TUI (picker.go)"
```

---

### Task 4: Wire `themes` subcommand + update help

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Add `"themes"` to the pre-scan switch in `main()`**

In `main()`, the current pre-scan switch is:

```go
if len(args) > 0 {
    switch args[0] {
    case "completion":
        ...
    case "upgrade":
        ...
    }
}
```

Add the `"themes"` case:

```go
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
    case "themes":
        runThemePicker()
        return
    }
}
```

- [ ] **Step 2: Add `themes` to `printHelp()`**

In `printHelp()`, find the usage section listing subcommands:

```
Usage:
  oc color [flags] -- <oc-args>
  oc color completion <bash|zsh|fish>
  oc color upgrade
```

Update to:

```
Usage:
  oc color [flags] -- <oc-args>
  oc color completion <bash|zsh|fish>
  oc color upgrade
  oc color themes
```

And add a line in the examples section after `--dry-run`:

```
  oc color themes              # interactive theme picker
```

- [ ] **Step 3: Build and run full test suite**

```bash
go build ./... && go test ./...
```

Expected: clean build, all tests PASS.

- [ ] **Step 4: Commit**

```bash
git add main.go
git commit -m "feat: wire oc color themes subcommand"
```
