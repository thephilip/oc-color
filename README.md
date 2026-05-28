# oc-color

Colorize and syntax-highlight `oc` command output. Think `diff` → `colordiff`, but for `oc`.

[![Go Version](https://img.shields.io/badge/Go-1.26.3-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](/LICENSE)

## Installation

### Krew plugin (recommended)

Requires [krew](https://krew.sigs.k8s.io/) plugin manager:

```bash
kubectl krew install oc-color
```

### Go install

```bash
go install github.com/thephilip/oc-color@latest
```

### Build from source

```bash
git clone https://github.com/thephilip/oc-color.git
cd oc-color
go build -o oc-color .
cp oc-color ~/.local/bin/
```

## Quick start

```bash
oc color get pods
oc color get pods -o json
oc color describe pod my-pod
oc color --theme dracula get pods
oc color --dry-run
```

## Features

| Feature | Description |
|---------|-------------|
| **Status colorization** | 30+ pod/build/deployment statuses colored by severity — `Running` in green, `CrashLoopBackOff` in bold red, `Pending` in yellow, etc. |
| **Table header styling** | Bold + underline + accent color on column headers |
| **Age/duration dimming** | Values like `12h`, `5m` rendered in a dim theme color |
| **JSON highlighting** | Built-in tokenizer (no dependencies) — keys, strings, numbers, booleans, null all colorized |
| **YAML highlighting** | Line-by-line tokenizer — document delimiters, keys, list markers, comments, and values highlighted |
| **Describe beautification** | Section headers, key-value pairs, event types (`Normal`/`Warning`), `<none>` dimming, `False` conditions highlighted |
| **Theme system** | Built-in Dracula theme. Custom YAML themes with `--theme`, `--list-themes`, `--validate-theme` |
| **TTY detection** | Auto-disable colors when piping. `--color=always\|never\|auto` flag |
| **Dry-run mode** | `--dry-run` processes sample output to preview colors without a real cluster |

## Flags

| Flag | Description |
|------|-------------|
| `--color <mode>` | Color mode: `always`, `never`, `auto` (default: `auto`) |
| `--no-color` | Shorthand for `--color=never` |
| `--theme <name>` | Theme name (default: `dracula`) |
| `--list-themes` | List available themes |
| `--validate-theme <path>` | Validate a theme YAML file |
| `--dry-run` | Process sample output to preview colors |
| `--version` | Print version |
| `--help`, `-h` | Show help |

## Configuration

Config file at `~/.config/oc-color/config.yaml` (or `$XDG_CONFIG_HOME/oc-color/config.yaml`):

```yaml
color: auto      # auto, always, never
theme: dracula   # theme name
```

## Themes

Built-in theme: **dracula**.

Custom themes go in `~/.config/oc-color/themes/<name>.yaml` (or `$XDG_CONFIG_HOME/oc-color/themes/<name>.yaml`).

### Theme file format

Supports string shorthand and structured YAML:

```yaml
name: nord
tokens:
  success: green
  warning: yellow
  error: bold+red
  info: cyan
  accent: "#5E81AC"
  dim: dim+white
  header: bold+underline+"#8FBCBB"
  key: "#88C0D0"
```

Required tokens: `success`, `warning`, `error`, `info`, `accent`, `dim`, `header`, `key`.

Validate a theme with:

```bash
oc color --validate-theme ~/.config/oc-color/themes/nord.yaml
```

## Examples

```bash
# Basic pod listing with colorized statuses
oc color get pods

# Force colors even when piping to less
oc color --color=always get pods | less -R

# JSON output with syntax highlighting
oc color get pod my-pod -o json

# YAML output with syntax highlighting
oc color get pod my-pod -o yaml

# Describe output beautification
oc color describe pod my-pod

# Use a custom theme
oc color --theme nord get pods

# List available themes
oc color --list-themes

# Preview color output without a cluster
oc color --dry-run
```

## Development

```bash
go build -o oc-color .
go test ./...
```
