# YAML/JSON Colorization Improvements

**Date:** 2026-06-02  
**Status:** Approved

## Problem

The current YAML colorizer (`output/yaml.go`) is stateless — it processes lines independently and misses several common patterns in `oc get -o yaml` output:

- Unquoted string values (image names, URIs, arbitrary text) get no color, which is fine, but IP addresses and timestamps — which carry signal — are also left plain.
- Flow sequences (`[app=myapp, tier=frontend]`) and flow mappings are not handled; brackets and their contents are passed through uncolored.
- Block scalar values (`|`, `>`, `|-`, `>-`) are not recognized; content lines after the header are parsed as if they were regular key-value lines, producing incorrect output.

The JSON colorizer (`output/json.go`) is largely correct but shares no value-colorization logic with the YAML colorizer, so any improvements must be duplicated.

## Approach

Improve the existing hand-written parsers. No new dependencies.

## Architecture

Three files change:

| File | Change |
|---|---|
| `output/yaml.go` | Replace stateless line function with a `yamlHL` struct that carries state across lines |
| `output/json.go` | Delegate scalar value colorization to shared function |
| `output/output.go` | Add shared `colorizeScalarValue` function (or extract to `output/values.go` if it grows) |

Public API is unchanged: `highlightYAML(input string, th theme.Theme) string`.

## State Machine

`yamlHL` tracks state across lines:

```go
type yamlHL struct {
    theme       theme.Theme
    flowDepth   int  // >0 inside [...] or {...}
    blockScalar bool // true after a | or > value header
    blockIndent int  // indentation of the key that opened the block
}
```

**Flow depth:** incremented on `[` or `{`, decremented on `]` or `}`. When `flowDepth > 0`, the current line is inside an inline flow collection and is colorized accordingly (brackets pink, commas dim, values via `colorizeScalarValue`).

**Block scalar mode:** set when a YAML value is a block scalar header (`|`, `>`, `|-`, `>-`, `|+`, `>+`, with optional indentation indicator). Cleared when a line's indentation is ≤ `blockIndent`. While active, content lines are colored as `success` (string content) without key-value parsing.

## Token Mapping

| Element | Token | Notes |
|---|---|---|
| Keys | `key` | Unchanged |
| `---` / `...` | `pink` | Unchanged |
| `-` list bullet | `pink` | Unchanged |
| `[` `]` `{` `}` | `pink` | New — matches JSON bracket coloring |
| `,` inside flow collections | `dim` | New — separator, not content |
| Quoted strings | `success` | Unchanged |
| Block scalar content lines | `success` | New |
| Booleans (`true`/`false`/`yes`/`no`/`on`/`off`) | `info` | Unchanged |
| Null (`null`/`~`) | `dim` | Unchanged |
| Numbers | `accent` | Unchanged |
| IP addresses (IPv4) | `accent` | New — same category as numbers |
| Timestamps (RFC3339, date-only) | `dim` | New — de-emphasize metadata noise |
| Comments (`#...`) | `dim` | Unchanged |
| Generic unquoted strings | (none) | Image URIs, arbitrary text left plain |

Colors vary per theme — each theme maps tokens to its own palette. This is intentional; the token assignments define semantic roles, not literal colors.

## Shared Value Colorizer

Extract `colorizeScalarValue(s string, th theme.Theme) string` into `output/output.go` (or `output/values.go`). It receives a trimmed, unquoted scalar string and returns a colorized version, or the original string if no pattern matches.

Matching order:
1. Boolean keywords → `info`
2. Null keywords → `dim`
3. Numeric → `accent` (reuse existing `looksNumeric`)
4. IPv4 address → `accent`
5. RFC3339 / date-only timestamp → `dim`
6. Fallback → return input unchanged

Both `yaml.go` and `json.go` call this function for unquoted scalar values, ensuring consistent colorization across formats.

IPv4 pattern: four octets 0–255 separated by dots, optionally followed by a `/prefix`.  
Timestamp pattern: RFC3339 (`2006-01-02T15:04:05Z`) or date-only (`2006-01-02`), identified by a leading 4-digit year and `-`.

## Flow Collection Colorization

When `flowDepth > 0`, lines are tokenized character by character (similar to the JSON parser) rather than via line-level regex:

- `[` / `{` → emit `pink`, increment depth
- `]` / `}` → emit `pink`, decrement depth
- `,` → emit `dim`
- Quoted strings → emit `success`
- Unquoted tokens → pass through `colorizeScalarValue`
- `:` → emit `dim` (key-value separator within flow mappings)
- Whitespace → pass through

Keys within flow mappings (tokens followed by `:`) are colored `key`.

## Block Scalar Colorization

A line whose trimmed value (after the colon and optional space) matches `^[|>][|\->+]?[0-9]?$` is a block scalar header. On detection:

- Color the block scalar indicator (`|` or `>` and any modifiers) as `dim`
- Set `blockScalar = true`, `blockIndent = indent of the key line`

While `blockScalar` is true, each subsequent line:
- If blank: pass through unchanged (blank lines are valid inside a block scalar)
- If `len(leading spaces) <= blockIndent`: clear `blockScalar`, re-process this line normally
- Otherwise: color the entire line content as `success`

## Error Handling

The colorizer must be a faithful pass-through — if it encounters input it doesn't recognize, it emits the original bytes unchanged. No panics, no truncation. The state machine resets on lines that don't match any pattern.

## Testing

Extend `output/output_test.go` and `output/yaml_test.go` to cover:

- Flow sequence on a single line: `labels: [app=myapp, tier=frontend]`
- Multi-line flow sequence (flowDepth spans lines)
- Block scalar with `|` and `>`
- IP address value
- Timestamp value (RFC3339 and date-only)
- Nested indentation (block scalar ended by reduced indent)
- JSON scalar values (IP, timestamp) via the shared function

Existing tests must continue to pass.
