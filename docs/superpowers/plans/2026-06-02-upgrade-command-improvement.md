# Upgrade Command Improvement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Go detection to `printUpgrade` so users without Go installed get a clear, actionable error instead of a confusing `exec` failure.

**Architecture:** Extract a `goInstalled(lookPath func(string) (string, error)) bool` helper that accepts a lookPath function for testability. `printUpgrade` calls it with `exec.LookPath`; tests inject a stub. If Go is absent, print three lines to stderr and exit 1. Happy path is unchanged.

**Tech Stack:** Go stdlib only (`os/exec.LookPath` already imported).

---

## Files

- Modify: `main.go` — add `goInstalled` helper, update `printUpgrade`
- Modify: `main_test.go` — add `TestGoInstalled`

---

### Task 1: Write failing test for `goInstalled`

**Files:**
- Modify: `main_test.go`

- [ ] **Step 1: Append `TestGoInstalled` to `main_test.go`**

Add after the last existing test function:

```go
func TestGoInstalled(t *testing.T) {
	found := goInstalled(func(name string) (string, error) {
		return "/usr/local/go/bin/go", nil
	})
	if !found {
		t.Error("goInstalled should return true when lookPath succeeds")
	}

	notFound := goInstalled(func(name string) (string, error) {
		return "", fmt.Errorf("executable file not found in $PATH")
	})
	if notFound {
		t.Error("goInstalled should return false when lookPath fails")
	}
}
```

Also add `"fmt"` to the `main_test.go` import block (it currently only imports `"testing"`):

```go
import (
	"fmt"
	"testing"
)
```

- [ ] **Step 2: Run the test to confirm it fails to compile**

```bash
go test ./... -run TestGoInstalled -v
```

Expected: compilation error — `undefined: goInstalled`

- [ ] **Step 3: Commit the failing test**

```bash
git add main_test.go
git commit -m "test: add goInstalled test (red)"
```

---

### Task 2: Implement `goInstalled` and update `printUpgrade`

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Add the `goInstalled` helper to `main.go`**

Add this function immediately before `printUpgrade` (around line 222):

```go
func goInstalled(lookPath func(string) (string, error)) bool {
	_, err := lookPath("go")
	return err == nil
}
```

- [ ] **Step 2: Update `printUpgrade` to call `goInstalled`**

Replace the current `printUpgrade` function:

```go
func printUpgrade() {
	fmt.Println("Upgrading oc-color to the latest version...")
	cmd := exec.Command("go", "install", "github.com/thephilip/oc-color@latest")
	cmd.Env = append(os.Environ(), "GOPROXY=direct")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: upgrade failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Upgrade complete.")
}
```

With:

```go
func printUpgrade() {
	if !goInstalled(exec.LookPath) {
		fmt.Fprintln(os.Stderr, "error: 'oc color upgrade' requires Go to be installed.")
		fmt.Fprintln(os.Stderr, "Install Go from https://golang.org/dl, then re-run this command.")
		fmt.Fprintln(os.Stderr, "If you installed via Krew, upgrade with: kubectl krew upgrade oc-color")
		os.Exit(1)
	}
	fmt.Println("Upgrading oc-color to the latest version...")
	cmd := exec.Command("go", "install", "github.com/thephilip/oc-color@latest")
	cmd.Env = append(os.Environ(), "GOPROXY=direct")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: upgrade failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Upgrade complete.")
}
```

- [ ] **Step 3: Run all tests**

```bash
go test ./... -v
```

Expected: all tests pass including `TestGoInstalled`.

- [ ] **Step 4: Commit**

```bash
git add main.go
git commit -m "feat: detect Go before upgrade, print actionable error if missing"
```
