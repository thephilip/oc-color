# Krew Submission Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a GoReleaser release pipeline and submit oc-color to the Krew plugin index.

**Architecture:** GoReleaser cross-compiles five platform binaries on tag push via GitHub Actions, uploads them as a GitHub Release with a `checksums.txt`. The maintainer manually updates `oc-color.yaml` with the new version and SHA256s after each release, then opens a PR to `krew-index` for the initial submission.

**Tech Stack:** GoReleaser v2, GitHub Actions, Krew v1alpha2 plugin manifest format

---

## File Map

| File | Action | Purpose |
|------|--------|---------|
| `.goreleaser.yaml` | Create | Cross-compile config: 5 platforms, archive naming, checksum |
| `.github/workflows/release.yml` | Create | CI: trigger GoReleaser on `v*.*.*` tags |
| `oc-color.yaml` | Modify | Update to v0.8.0 with correct per-platform entries |
| `RELEASING.md` | Create | Step-by-step release + Krew update runbook |

---

### Task 1: Create `.goreleaser.yaml`

**Files:**
- Create: `.goreleaser.yaml`

- [ ] **Step 1: Create the GoReleaser config**

Create `.goreleaser.yaml` at the repo root with this exact content:

```yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    binary: oc-color
    ldflags:
      - -s -w

archives:
  - format: tar.gz
    name_template: "oc-color_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip
    files:
      - none*

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

release:
  github:
    owner: thephilip
    name: oc-color
  draft: false
  prerelease: auto

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
```

- [ ] **Step 2: Verify GoReleaser can parse the config**

Install GoReleaser locally if not present:

```bash
go install github.com/goreleaser/goreleaser/v2@latest
```

Then validate:

```bash
goreleaser check
```

Expected output: `• config is valid`

If GoReleaser is not installed and you don't want to install it locally, skip this step — the GitHub Actions workflow will catch errors on first tag.

- [ ] **Step 3: Commit**

```bash
git add .goreleaser.yaml
git commit -m "ci: add GoReleaser config for multi-platform releases"
```

---

### Task 2: Create GitHub Actions release workflow

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Create the workflows directory and file**

```bash
mkdir -p .github/workflows
```

Create `.github/workflows/release.yml` with this exact content:

```yaml
name: Release

on:
  push:
    tags:
      - "v*.*.*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "ci: add GitHub Actions release workflow"
```

---

### Task 3: Update `oc-color.yaml` plugin manifest

**Files:**
- Modify: `oc-color.yaml`

The current file has a single catch-all platform entry with placeholder SHA256 and stale version. Replace it with five per-platform entries matching the GoReleaser archive naming: `oc-color_{version}_{os}_{arch}.{ext}`.

- [ ] **Step 1: Replace `oc-color.yaml` with the updated manifest**

Overwrite `oc-color.yaml` with this exact content:

```yaml
apiVersion: krew.googlecontainertools.github.com/v1alpha2
kind: Plugin
metadata:
  name: oc-color
spec:
  version: "0.8.0"
  homepage: "https://github.com/thephilip/oc-color"
  shortDescription: "Colorize and syntax-highlight oc command output."
  description: |-
    oc-color is a plugin for the OpenShift CLI (oc) tool that wraps the oc binary,
    intercepting its output and injecting ANSI color codes for improved readability.
    It supports highlighting key status words (e.g., Running, Pending, Error) and can be
    extended to support advanced syntax highlighting.
  platforms:
    - selector:
        matchLabels:
          os: linux
          arch: amd64
      uri: "https://github.com/thephilip/oc-color/releases/download/v0.8.0/oc-color_0.8.0_linux_amd64.tar.gz"
      sha256: "<sha256-linux-amd64>"
      bin: "oc-color"
    - selector:
        matchLabels:
          os: linux
          arch: arm64
      uri: "https://github.com/thephilip/oc-color/releases/download/v0.8.0/oc-color_0.8.0_linux_arm64.tar.gz"
      sha256: "<sha256-linux-arm64>"
      bin: "oc-color"
    - selector:
        matchLabels:
          os: darwin
          arch: amd64
      uri: "https://github.com/thephilip/oc-color/releases/download/v0.8.0/oc-color_0.8.0_darwin_amd64.tar.gz"
      sha256: "<sha256-darwin-amd64>"
      bin: "oc-color"
    - selector:
        matchLabels:
          os: darwin
          arch: arm64
      uri: "https://github.com/thephilip/oc-color/releases/download/v0.8.0/oc-color_0.8.0_darwin_arm64.tar.gz"
      sha256: "<sha256-darwin-arm64>"
      bin: "oc-color"
    - selector:
        matchLabels:
          os: windows
          arch: amd64
      uri: "https://github.com/thephilip/oc-color/releases/download/v0.8.0/oc-color_0.8.0_windows_amd64.zip"
      sha256: "<sha256-windows-amd64>"
      bin: "oc-color.exe"
```

- [ ] **Step 2: Commit**

```bash
git add oc-color.yaml
git commit -m "chore: update plugin manifest to v0.8.0 with per-platform entries"
```

---

### Task 4: Create `RELEASING.md`

**Files:**
- Create: `RELEASING.md`

- [ ] **Step 1: Create RELEASING.md at the repo root**

```markdown
# Releasing oc-color

## Prerequisites

- Push access to `github.com/thephilip/oc-color`
- GitHub CLI (`gh`) installed and authenticated

---

## Cutting a Release

### 1. Bump the version

In `main.go`, update the version constant (line ~29):

```go
const version = "0.9.0"  // new version
```

### 2. Commit the version bump

```bash
git add main.go
git commit -m "chore: bump version to 0.9.0"
git push origin main
```

### 3. Tag and push

```bash
git tag v0.9.0
git push origin v0.9.0
```

This triggers the GitHub Actions release workflow. Go to
https://github.com/thephilip/oc-color/actions to monitor progress (~3 minutes).

### 4. Download checksums

Once the workflow completes, the release appears at
https://github.com/thephilip/oc-color/releases. Download `checksums.txt`:

```bash
gh release download v0.9.0 -R thephilip/oc-color --pattern checksums.txt
```

### 5. Update `oc-color.yaml`

In `oc-color.yaml`, update:
- `spec.version` to `"0.9.0"`
- All five `uri` fields: replace `0.8.0` with `0.9.0` (in both the tag path and filename)
- All five `sha256` fields: copy values from `checksums.txt` matching each archive name

Example `checksums.txt` lines to look for:

```
abc123...  oc-color_0.9.0_linux_amd64.tar.gz
def456...  oc-color_0.9.0_linux_arm64.tar.gz
ghi789...  oc-color_0.9.0_darwin_amd64.tar.gz
jkl012...  oc-color_0.9.0_darwin_arm64.tar.gz
mno345...  oc-color_0.9.0_windows_amd64.zip
```

### 6. Commit and push `oc-color.yaml`

```bash
git add oc-color.yaml
git commit -m "chore: update plugin manifest to v0.9.0"
git push origin main
```

---

## Initial Krew Index Submission (one-time)

Do this after the first real release (v0.8.0) has SHA256s filled in.

### 1. Fork krew-index

Go to https://github.com/kubernetes-sigs/krew-index and fork it to your GitHub account.

### 2. Clone your fork

```bash
git clone https://github.com/thephilip/krew-index.git
cd krew-index
```

### 3. Add the plugin manifest

```bash
cp /path/to/oc-color/oc-color.yaml plugins/oc-color.yaml
git add plugins/oc-color.yaml
git commit -m "feat: add oc-color plugin"
git push origin main
```

### 4. Open a PR

Go to https://github.com/kubernetes-sigs/krew-index and open a pull request from
your fork. Follow the PR template — reviewers will check that:
- SHA256s are correct
- URIs resolve to real release assets
- The manifest validates with `kubectl krew install --manifest=plugins/oc-color.yaml`

### 5. Future releases

After the initial PR is merged, each new release requires a PR to update
`plugins/oc-color.yaml` in `krew-index` with the new version and SHA256s.
Follow steps 1–6 of "Cutting a Release" above, then open a PR to krew-index
with the updated manifest.
```

- [ ] **Step 2: Commit**

```bash
git add RELEASING.md
git commit -m "docs: add release runbook"
```

---

### Task 5: Push to GitHub and cut v0.8.0

**Files:** None (git operations only)

- [ ] **Step 1: Push all commits**

```bash
git push origin main
```

- [ ] **Step 2: Tag v0.8.0 and push**

```bash
git tag v0.8.0
git push origin v0.8.0
```

Monitor the Actions run at https://github.com/thephilip/oc-color/actions.

- [ ] **Step 3: Download checksums and fill in `oc-color.yaml`**

Wait for the release workflow to complete (~3 minutes), then:

```bash
gh release download v0.8.0 -R thephilip/oc-color --pattern checksums.txt
cat checksums.txt
```

Copy the SHA256 for each of the five archives into `oc-color.yaml`, replacing the `<sha256-*>` placeholders.

- [ ] **Step 4: Commit and push the filled-in manifest**

```bash
git add oc-color.yaml
git commit -m "chore: add v0.8.0 SHA256s to plugin manifest"
git push origin main
```

---

### Task 6: Submit to krew-index

**Files:** None in this repo — work happens in a fork of `kubernetes-sigs/krew-index`

- [ ] **Step 1: Fork and clone krew-index**

Fork https://github.com/kubernetes-sigs/krew-index via the GitHub UI, then:

```bash
git clone https://github.com/thephilip/krew-index.git
cd krew-index
```

- [ ] **Step 2: Copy the manifest**

```bash
cp /home/philip/Downloads/_projects/oc-color/oc-color.yaml plugins/oc-color.yaml
```

- [ ] **Step 3: Validate locally (optional but recommended)**

```bash
kubectl krew install --manifest=plugins/oc-color.yaml
```

Expected: installs successfully. Uninstall after: `kubectl krew uninstall oc-color`

- [ ] **Step 4: Commit and push**

```bash
git add plugins/oc-color.yaml
git commit -m "feat: add oc-color plugin"
git push origin main
```

- [ ] **Step 5: Open the PR**

Go to https://github.com/kubernetes-sigs/krew-index and open a pull request from
your fork. The PR description should include:
- What the plugin does (one sentence)
- Link to the homepage: https://github.com/thephilip/oc-color
- Confirmation that you tested installation with `kubectl krew install --manifest=...`
