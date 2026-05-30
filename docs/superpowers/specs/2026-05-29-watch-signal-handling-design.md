# Watch Mode Signal Handling — Design Spec

**Date:** 2026-05-29
**Scope:** `watch.go`, `main.go`
**Version target:** v0.8.x patch

---

## Problem

Three correctness issues in `runWatch`:

1. **Stderr swallowed** — `cmd.StderrPipe()` captures `oc` stderr into a buffer that is never read or printed. If `oc` errors during a watch session (auth expired, resource not found, etc.), the user sees nothing.
2. **Noisy Ctrl-C** — The signal handler returns `fmt.Errorf("interrupted")`, which `main.go` prints to stderr. A normal user action produces an unexpected error message.
3. **Stale ANSI on exit** — The cleanup defer restores the cursor but does not emit a full ANSI reset (`\033[0m`). If a colorized line was partially written when the signal fires, stale color or style codes can bleed into the next shell prompt.

---

## Approach

Minimal inline fixes. No new abstractions, no restructuring. Each change is independent and isolated to a small number of lines.

---

## Changes

### Fix 1 — Stderr forwarding (`watch.go`)

**Remove** the `StderrPipe` + dead-buffer goroutine (5 lines):

```go
// REMOVE:
stderr, _ := cmd.StderrPipe()
var stderrBuf bytes.Buffer
go func() {
    io.Copy(&stderrBuf, stderr)
}()
```

**Replace** with a single assignment before `cmd.Start()`:

```go
cmd.Stderr = os.Stderr
```

`bytes` and `io` imports can be removed if no longer used elsewhere in the file.

**Behavioral note:** If `oc` writes to stderr while a terminal redraw is in progress, the error text may interleave with ANSI codes. This is an acceptable edge case — a visible error is strictly better than a silently discarded one.

---

### Fix 2 — Silent Ctrl-C (`watch.go` + `main.go`)

**In `watch.go`**, add a package-level sentinel and update the signal handler:

```go
var errInterrupted = errors.New("interrupted")
```

```go
case <-sigCh:
    cmd.Process.Kill()
    cmd.Wait() // reap the process
    return errInterrupted
```

**In `main.go`**, suppress the sentinel at the call site:

```go
err := runWatch(watchArgs, &proc)
if err != nil && !errors.Is(err, errInterrupted) {
    fmt.Fprintln(os.Stderr, err)
}
```

`errors` import is already present in the stdlib; no new dependency.

---

### Fix 3 — ANSI reset on exit (`watch.go`)

Update the terminal-cleanup defer to emit a full reset before restoring the cursor:

```go
defer func() {
    os.Stdout.WriteString("\033[0m")   // reset any stale color/style
    os.Stdout.WriteString("\033[?25h") // show cursor
    os.Stdout.WriteString("\n")
}()
```

The reset must come first — writing the cursor-show sequence while in a colored state has no effect on color bleed.

---

## Files Changed

| File | Changes |
|------|---------|
| `watch.go` | Remove StderrPipe + goroutine; add `cmd.Stderr = os.Stderr`; add `errInterrupted` sentinel; call `cmd.Wait()` in signal handler; add `\033[0m` to cleanup defer |
| `main.go` | Suppress `errInterrupted` in watch error check |

---

## Out of Scope

- Stderr interleaving with redraws (acceptable edge case, not worth the complexity of a mutex)
- Switching from `SIGKILL` to `SIGTERM` + drain (the `oc` process is a passthrough; abrupt kill is fine)
- Context-based cancellation refactor (no benefit at this scale)
