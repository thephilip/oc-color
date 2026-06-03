# Krew Submission Design

**Date:** 2026-06-03  
**Status:** Approved

## Overview

Add a GoReleaser-based release pipeline to `oc-color` and submit the plugin to the Krew index. Future releases will be semi-automated: tagging triggers CI to build and publish binaries; the maintainer manually updates `oc-color.yaml` with new SHA256s before opening a PR to `krew-index`.

## Components

### 1. GoReleaser config (`.goreleaser.yaml`)

Builds five targets in a single CI run:

| OS      | Arch  | Archive format |
|---------|-------|----------------|
| linux   | amd64 | `.tar.gz`      |
| linux   | arm64 | `.tar.gz`      |
| darwin  | amd64 | `.tar.gz`      |
| darwin  | arm64 | `.tar.gz`      |
| windows | amd64 | `.zip`         |

Archive naming: `oc-color_{version}_{os}_{arch}.{ext}` — matches the URI pattern already in `oc-color.yaml`.

Binary inside each archive: `oc-color` (or `oc-color.exe` on Windows).

GoReleaser also generates `checksums.txt` (SHA256 for all archives) and attaches everything to the GitHub Release automatically.

No snapshot builds, no homebrew tap, no Docker image — binaries only.

### 2. GitHub Actions release workflow (`.github/workflows/release.yml`)

- **Trigger:** `push` to tags matching `v*.*.*`
- **Steps:** `actions/checkout` (fetch-depth: 0 for GoReleaser changelog), setup Go (version from `go.mod`), run `goreleaser release --clean`
- **Secret:** `GITHUB_TOKEN` — automatically available in Actions, no manual setup
- Tests and lint are out of scope for this workflow; add a separate CI workflow later if desired

### 3. Plugin manifest (`oc-color.yaml`)

Updated to v0.8.0. One `platforms` entry per target (5 total). Each entry specifies:

- `uri`: `https://github.com/thephilip/oc-color/releases/download/v{version}/oc-color_{version}_{os}_{arch}.{ext}`
- `sha256`: filled in from `checksums.txt` after each release (placeholder until first real release)
- `bin`: `oc-color` (or `oc-color.exe` on Windows)
- `selector`: `matchLabels` for `os` and `arch`

Existing short description and full description are retained as-is.

### 4. Release runbook (`RELEASING.md`)

Committed to repo root. Covers the full release checklist:

1. Bump version constant in `main.go`
2. Commit and push to `main`
3. Tag and push: `git tag v0.x.0 && git push origin v0.x.0`
4. Wait for GitHub Actions to complete
5. Download `checksums.txt` from the new GitHub Release
6. Update `oc-color.yaml`: bump version string, paste SHA256s
7. Commit and push `oc-color.yaml`

**Initial Krew submission (one-time):**

8. Fork `kubernetes-sigs/krew-index`
9. Copy `oc-color.yaml` to `plugins/oc-color.yaml` in the fork
10. Open a PR against `krew-index` main — follow their PR template
11. After merge, future releases update `plugins/oc-color.yaml` via PR

## Files Changed

| File | Action |
|------|--------|
| `.goreleaser.yaml` | Create |
| `.github/workflows/release.yml` | Create |
| `oc-color.yaml` | Update (version, URIs, SHA256 placeholders) |
| `RELEASING.md` | Create |

## Out of Scope

- Automated Krew manifest PRs on every release (GoReleaser krew publisher)
- Homebrew tap
- Docker image
- Separate lint/test CI workflow
- Windows arm64
