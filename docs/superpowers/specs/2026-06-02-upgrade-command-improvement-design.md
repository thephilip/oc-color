# Upgrade Command Improvement Design

**Date:** 2026-06-02
**Status:** Approved

## Goal

Improve `oc color upgrade` so that users without Go installed get a clear, actionable error instead of a confusing `exec` failure.

## Problem

`printUpgrade` unconditionally runs `go install github.com/thephilip/oc-color@latest`. If `go` is not on PATH, the error is opaque and gives no guidance on what to do next.

## Approach

Silent `go` detection with a targeted error message. No noise on the happy path.

1. At the top of `printUpgrade`, call `exec.LookPath("go")`.
2. If `go` is not found, print to stderr and exit 1:

```
error: 'oc color upgrade' requires Go to be installed.
Install Go from https://golang.org/dl, then re-run this command.
If you installed via Krew, upgrade with: kubectl krew upgrade oc-color
```

3. If `go` is found, proceed exactly as today — no extra output.

## Constraints

- No new imports (`exec` is already imported).
- No other files touched.
- Happy path is unchanged: zero extra output.

## Testing

- Unit-testable by injecting a `lookPath` func into `printUpgrade` (or extracting the detection logic).
- Alternatively: test the detection logic as a separate helper that returns an error when `go` is absent.
- The existing upgrade flow (happy path) is not unit-tested today and does not need to be added — it requires a live `go` binary and network.
