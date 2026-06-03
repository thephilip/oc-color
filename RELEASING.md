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

### Troubleshooting

**Tagged the wrong commit or need to redo a release?**

```bash
git tag -d v0.9.0              # delete local tag
git push origin --delete v0.9.0  # delete remote tag
# Fix the issue, then re-run from step 1
```

Note: GoReleaser will also have created a GitHub release — delete it at https://github.com/thephilip/oc-color/releases before re-tagging.

### 4. Download checksums

Once the workflow completes, the release appears at
https://github.com/thephilip/oc-color/releases. Download `checksums.txt`:

```bash
gh release download v0.9.0 -R thephilip/oc-color --pattern checksums.txt
```

**Note:** If the command fails because assets aren't ready yet, wait a moment and retry — GoReleaser can occasionally take longer than 3 minutes.

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

Copy the exact hex hash from each line — do not type these manually.

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
- The manifest can be validated locally with `kubectl krew install --manifest=plugins/oc-color.yaml` (requires krew installed)

### 5. Future releases

After the initial PR is merged, each new release requires a PR to update
`plugins/oc-color.yaml` in `krew-index` with the new version and SHA256s.
Follow steps 1–6 of "Cutting a Release" above, then open a PR to krew-index
with the updated manifest.
