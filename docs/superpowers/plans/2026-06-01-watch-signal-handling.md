# Watch Mode Signal Handling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix three correctness issues in `runWatch`: silently swallowed stderr, noisy "interrupted" on Ctrl-C, and stale ANSI codes leaking into the shell prompt on exit.

**Architecture:** All changes are inline fixes to two existing files — `watch.go` and `main.go`. No new files, no new abstractions. Fix 1 removes dead code (stderr buffer + goroutine) and replaces it with a single line. Fix 2 introduces a package-level sentinel error and updates the signal handler and call site. Fix 3 adds one line to an existing defer.

**Tech Stack:** Go stdlib only (`errors`, `os`, `os/exec`). No new dependencies.

---

### Task 1: Fix stderr forwarding (`watch.go`)

**Files:**
- Modify: `watch.go`

The current code captures `oc` stderr into a buffer that is never read. Replace it with `cmd.Stderr = os.Stderr` so errors surface immediately. Remove the now-unused `bytes` and `io` imports.

- [ ] **Step 1: Remove the dead stderr code**

In `watch.go`, delete these lines (currently around lines 48–54):

```go
stderr, _ := cmd.StderrPipe()
var stderrBuf bytes.Buffer
go func() {
	io.Copy(&stderrBuf, stderr)
}()
```

- [ ] **Step 2: Add `cmd.Stderr = os.Stderr` before `cmd.Start()`**

After `cmd.StdoutPipe()` succeeds and before `cmd.Start()`, insert:

```go
cmd.Stderr = os.Stderr
```

The relevant block should look like this after the change:

```go
cmd := exec.Command("oc", args...)
stdout, err := cmd.StdoutPipe()
if err != nil {
	return err
}
cmd.Stderr = os.Stderr
if err := cmd.Start(); err != nil {
	return err
}
```

- [ ] **Step 3: Remove unused imports**

In the `import` block of `watch.go`, remove `"bytes"` and `"io"`. The import block should become:

```go
import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/thephilip/oc-color/output"
	"golang.org/x/term"
)
```

- [ ] **Step 4: Verify it compiles**

```bash
go build ./...
```

Expected: no output, exit 0.

- [ ] **Step 5: Commit**

```bash
git add watch.go
git commit -m "fix: forward oc stderr to os.Stderr in watch mode"
```

---

### Task 2: Silent Ctrl-C (`watch.go`)

**Files:**
- Modify: `watch.go`
- Create: `watch_test.go`

Introduce a package-level `errInterrupted` sentinel. Update the signal handler to call `cmd.Wait()` before returning it (reaps the process). Add `errors` to the import block.

- [ ] **Step 1: Write a failing test for the sentinel**

Create `watch_test.go`:

```go
package main

import (
	"errors"
	"testing"
)

func TestErrInterrupted(t *testing.T) {
	if !errors.Is(errInterrupted, errInterrupted) {
		t.Fatal("errInterrupted sentinel identity broken")
	}
	if errInterrupted.Error() != "interrupted" {
		t.Fatalf("expected message %q, got %q", "interrupted", errInterrupted.Error())
	}
}
```

- [ ] **Step 2: Run the test to confirm it fails**

```bash
go test -run TestErrInterrupted -v .
```

Expected: compile error — `errInterrupted` not defined.

- [ ] **Step 3: Add the sentinel and update the signal handler**

At the top of `watch.go`, after the `import` block, add:

```go
var errInterrupted = errors.New("interrupted")
```

Add `"errors"` to the import block:

```go
import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/thephilip/oc-color/output"
	"golang.org/x/term"
)
```

Update the signal handler case in `runWatch` (currently `cmd.Process.Kill()` + `return fmt.Errorf("interrupted")`):

```go
case <-sigCh:
	cmd.Process.Kill()
	cmd.Wait()
	return errInterrupted
```

- [ ] **Step 4: Run the test to confirm it passes**

```bash
go test -run TestErrInterrupted -v .
```

Expected:
```
--- PASS: TestErrInterrupted (0.00s)
PASS
```

- [ ] **Step 5: Run full test suite**

```bash
go test ./...
```

Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add watch.go watch_test.go
git commit -m "fix: suppress interrupted error on Ctrl-C in watch mode"
```

---

### Task 3: ANSI reset on exit (`watch.go`)

**Files:**
- Modify: `watch.go`

Add `\033[0m` (full ANSI reset) to the terminal-cleanup defer so stale color state doesn't bleed into the shell prompt after watch mode exits.

- [ ] **Step 1: Update the cleanup defer**

Find the defer inside `runWatch` that currently reads:

```go
defer func() {
	os.Stdout.WriteString("\033[?25h")
	os.Stdout.WriteString("\n")
}()
```

Change it to:

```go
defer func() {
	os.Stdout.WriteString("\033[0m")   // reset any stale color/style
	os.Stdout.WriteString("\033[?25h") // show cursor
	os.Stdout.WriteString("\n")
}()
```

- [ ] **Step 2: Build and run tests**

```bash
go build ./... && go test ./...
```

Expected: no errors, all tests pass.

- [ ] **Step 3: Commit**

```bash
git add watch.go
git commit -m "fix: emit ANSI reset before cursor restore on watch exit"
```

---

### Task 4: Suppress `errInterrupted` in `main.go`

**Files:**
- Modify: `main.go`

Update the watch error-check call site so `errInterrupted` exits silently. Add `errors` to the import block.

- [ ] **Step 1: Add `errors` to the import block in `main.go`**

The current import block in `main.go`:

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/thephilip/oc-color/config"
	"github.com/thephilip/oc-color/output"
	"github.com/thephilip/oc-color/theme"
	"golang.org/x/term"
)
```

Add `"errors"`:

```go
import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/thephilip/oc-color/config"
	"github.com/thephilip/oc-color/output"
	"github.com/thephilip/oc-color/theme"
	"golang.org/x/term"
)
```

- [ ] **Step 2: Update the watch error check**

Find the current watch error handling in `main.go`:

```go
err := runWatch(watchArgs, &proc)
if err != nil {
	fmt.Fprintln(os.Stderr, err)
}
```

Change it to:

```go
err := runWatch(watchArgs, &proc)
if err != nil && !errors.Is(err, errInterrupted) {
	fmt.Fprintln(os.Stderr, err)
}
```

- [ ] **Step 3: Build and run full test suite**

```bash
go build ./... && go test ./...
```

Expected: no errors, all tests pass.

- [ ] **Step 4: Commit**

```bash
git add main.go
git commit -m "fix: suppress errInterrupted on Ctrl-C watch exit"
```
