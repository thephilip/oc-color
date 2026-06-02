# YAML/JSON Colorization Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Improve YAML colorization by adding a shared scalar value colorizer with IP/timestamp detection, refactoring the YAML highlighter to a state-aware struct, and adding support for block scalars and flow collections.

**Architecture:** A new `output/values.go` provides a shared `colorizeScalarValue` function used by the YAML highlighter (and available to JSON). The YAML highlighter in `output/yaml.go` is rewritten from a stateless line function to a `yamlHL` struct that tracks flow depth and block scalar mode across lines. Public API (`highlightYAML`) is unchanged.

**Tech Stack:** Go stdlib + `github.com/thephilip/oc-color/theme` token system. No new dependencies.

---

### Task 1: Extract shared scalar value colorizer

**Files:**
- Create: `output/values.go`
- Create: `output/values_test.go`
- Modify: `output/yaml.go` (remove `looksNumeric` and `colorizeYAMLValue`, update call site)

- [ ] **Step 1: Write the failing tests**

Create `output/values_test.go`:

```go
package output

import (
	"testing"
)

func TestColorizeScalarValue_Bool(t *testing.T) {
	th := testTheme()
	for _, kw := range []string{"true", "false", "yes", "no", "on", "off"} {
		got := colorizeScalarValue(kw, th)
		want := wrapWithTheme(kw, "info", th)
		if got != want {
			t.Errorf("colorizeScalarValue(%q) = %q, want %q", kw, got, want)
		}
	}
}

func TestColorizeScalarValue_Null(t *testing.T) {
	th := testTheme()
	for _, kw := range []string{"null", "~"} {
		got := colorizeScalarValue(kw, th)
		want := wrapWithTheme(kw, "dim", th)
		if got != want {
			t.Errorf("colorizeScalarValue(%q) = %q, want %q", kw, got, want)
		}
	}
}

func TestColorizeScalarValue_Number(t *testing.T) {
	th := testTheme()
	for _, n := range []string{"42", "3.14", "-7", "0"} {
		got := colorizeScalarValue(n, th)
		want := wrapWithTheme(n, "accent", th)
		if got != want {
			t.Errorf("colorizeScalarValue(%q) = %q, want %q", n, got, want)
		}
	}
}

func TestColorizeScalarValue_PlainUnchanged(t *testing.T) {
	th := testTheme()
	for _, s := range []string{"quay.io/openshift/origin-pod:latest", "my-pod", ""} {
		got := colorizeScalarValue(s, th)
		if got != s {
			t.Errorf("colorizeScalarValue(%q) = %q, want unchanged", s, got)
		}
	}
}
```

- [ ] **Step 2: Run to confirm they fail**

```
go test ./output/... -run TestColorizeScalarValue -v
```

Expected: FAIL — `undefined: colorizeScalarValue`

- [ ] **Step 3: Create `output/values.go`**

```go
package output

import (
	"strings"

	"github.com/thephilip/oc-color/theme"
)

// colorizeScalarValue colorizes an unquoted scalar string using the theme token system.
// Returns s unchanged when no known pattern matches.
func colorizeScalarValue(s string, th theme.Theme) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return s
	}
	switch trimmed {
	case "true", "false", "yes", "no", "on", "off":
		return wrapWithTheme(trimmed, "info", th)
	case "null", "~":
		return wrapWithTheme(trimmed, "dim", th)
	}
	if looksNumeric(trimmed) {
		return wrapWithTheme(trimmed, "accent", th)
	}
	return s
}

func looksNumeric(s string) bool {
	if s == "" {
		return false
	}
	start := 0
	if s[0] == '-' || s[0] == '+' {
		start = 1
		if start >= len(s) {
			return false
		}
	}
	hasDigit := false
	hasDot := false
	for i := start; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			hasDigit = true
		} else if s[i] == '.' && !hasDot {
			hasDot = true
		} else {
			return false
		}
	}
	return hasDigit
}
```

- [ ] **Step 4: Remove `looksNumeric` and `colorizeYAMLValue` from `output/yaml.go`, update `highlightYAMLLine` to call the shared function**

In `output/yaml.go`, delete the `looksNumeric` function entirely (now in `values.go`).

Replace `colorizeYAMLValue` with a call to `colorizeScalarValue`. Find this function:

```go
func colorizeYAMLValue(val string, th theme.Theme) string {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return val
	}

	if (strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"")) ||
		(strings.HasPrefix(trimmed, "'") && strings.HasSuffix(trimmed, "'")) {
		return wrapWithTheme(trimmed, "success", th)
	}

	switch trimmed {
	case "true", "false", "yes", "no", "on", "off":
		return wrapWithTheme(trimmed, "info", th)
	case "null", "~":
		return wrapWithTheme(trimmed, "dim", th)
	}

	if looksNumeric(trimmed) {
		return wrapWithTheme(trimmed, "accent", th)
	}

	return val
}
```

Replace it with:

```go
func colorizeYAMLValue(val string, th theme.Theme) string {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return val
	}
	if (strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"")) ||
		(strings.HasPrefix(trimmed, "'") && strings.HasSuffix(trimmed, "'")) {
		return wrapWithTheme(trimmed, "success", th)
	}
	return colorizeScalarValue(val, th)
}
```

- [ ] **Step 5: Run tests to confirm they pass**

```
go test ./output/... -run TestColorizeScalarValue -v
```

Expected: PASS for all four tests.

Also run the full suite to confirm nothing regressed:

```
go test ./... -v 2>&1 | tail -20
```

Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add output/values.go output/values_test.go output/yaml.go
git commit -m "refactor: extract shared colorizeScalarValue into output/values.go"
```

---

### Task 2: Add IP address and timestamp detection

**Files:**
- Modify: `output/values.go`
- Modify: `output/values_test.go`

- [ ] **Step 1: Write the failing tests**

Append to `output/values_test.go`:

```go
func TestColorizeScalarValue_IPv4(t *testing.T) {
	th := testTheme()
	for _, ip := range []string{"192.168.1.1", "10.0.0.1", "172.16.0.0/16", "0.0.0.0"} {
		got := colorizeScalarValue(ip, th)
		want := wrapWithTheme(ip, "accent", th)
		if got != want {
			t.Errorf("colorizeScalarValue(%q) = %q, want %q", ip, got, want)
		}
	}
}

func TestColorizeScalarValue_Timestamp(t *testing.T) {
	th := testTheme()
	for _, ts := range []string{
		"2024-06-02",
		"2024-06-02T15:04:05Z",
		"2024-06-02T15:04:05.123456789Z",
		"2024-06-02T15:04:05+05:30",
		"2024-06-02T15:04:05-07:00",
	} {
		got := colorizeScalarValue(ts, th)
		want := wrapWithTheme(ts, "dim", th)
		if got != want {
			t.Errorf("colorizeScalarValue(%q) = %q, want %q", ts, got, want)
		}
	}
}

func TestColorizeScalarValue_NotIPOrTimestamp(t *testing.T) {
	th := testTheme()
	// These look vaguely numeric/date-like but should not match
	for _, s := range []string{"999.999.999.999", "not-a-date", "2024"} {
		got := colorizeScalarValue(s, th)
		if got != s {
			t.Errorf("colorizeScalarValue(%q) = %q, want unchanged", s, got)
		}
	}
}
```

- [ ] **Step 2: Run to confirm they fail**

```
go test ./output/... -run "TestColorizeScalarValue_IPv4|TestColorizeScalarValue_Timestamp|TestColorizeScalarValue_NotIPOrTimestamp" -v
```

Expected: FAIL — IP and timestamp inputs returned unchanged instead of colored.

- [ ] **Step 3: Add regex patterns and detection to `output/values.go`**

Add `regexp` to imports. Add the two compiled regexes at package level (after the `import` block, before `colorizeScalarValue`):

```go
import (
	"regexp"
	"strings"

	"github.com/thephilip/oc-color/theme"
)

var (
	ipv4RE  = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}(/\d{1,2})?$`)
	stampRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}(T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2}))?$`)
)
```

Then in `colorizeScalarValue`, add IP and timestamp checks after the `looksNumeric` check:

```go
func colorizeScalarValue(s string, th theme.Theme) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return s
	}
	switch trimmed {
	case "true", "false", "yes", "no", "on", "off":
		return wrapWithTheme(trimmed, "info", th)
	case "null", "~":
		return wrapWithTheme(trimmed, "dim", th)
	}
	if looksNumeric(trimmed) {
		return wrapWithTheme(trimmed, "accent", th)
	}
	if ipv4RE.MatchString(trimmed) {
		return wrapWithTheme(trimmed, "accent", th)
	}
	if stampRE.MatchString(trimmed) {
		return wrapWithTheme(trimmed, "dim", th)
	}
	return s
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```
go test ./output/... -run TestColorizeScalarValue -v
```

Expected: all PASS.

Full suite:

```
go test ./... 2>&1 | tail -10
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add output/values.go output/values_test.go
git commit -m "feat: add IPv4 and timestamp detection to colorizeScalarValue"
```

---

### Task 3: Refactor yaml.go to state-aware yamlHL struct

This task preserves all existing YAML colorization behavior while switching to a struct-based parser that can carry state across lines (required for Tasks 4 and 5).

**Files:**
- Create: `output/yaml_test.go`
- Modify: `output/yaml.go` (full rewrite of highlighting logic)

- [ ] **Step 1: Write tests for existing YAML behavior**

Create `output/yaml_test.go`:

```go
package output

import (
	"strings"
	"testing"
)

func TestHighlightYAML_DocMarker(t *testing.T) {
	th := testTheme()
	got := highlightYAML("---\n", th)
	want := wrapWithTheme("---", "pink", th) + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHighlightYAML_DocEnd(t *testing.T) {
	th := testTheme()
	got := highlightYAML("...\n", th)
	want := wrapWithTheme("...", "pink", th) + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHighlightYAML_Comment(t *testing.T) {
	th := testTheme()
	got := highlightYAML("# this is a comment\n", th)
	want := wrapWithTheme("# this is a comment", "dim", th) + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHighlightYAML_Key_Colored(t *testing.T) {
	th := testTheme()
	got := highlightYAML("name: my-pod\n", th)
	if !strings.Contains(got, wrapWithTheme("name", "key", th)) {
		t.Errorf("key not colored; got %q", got)
	}
}

func TestHighlightYAML_Value_Bool(t *testing.T) {
	th := testTheme()
	got := highlightYAML("active: true\n", th)
	if !strings.Contains(got, wrapWithTheme("true", "info", th)) {
		t.Errorf("boolean value not colored; got %q", got)
	}
}

func TestHighlightYAML_Value_Null(t *testing.T) {
	th := testTheme()
	got := highlightYAML("deletionTimestamp: null\n", th)
	if !strings.Contains(got, wrapWithTheme("null", "dim", th)) {
		t.Errorf("null value not colored; got %q", got)
	}
}

func TestHighlightYAML_Value_Number(t *testing.T) {
	th := testTheme()
	got := highlightYAML("replicas: 3\n", th)
	if !strings.Contains(got, wrapWithTheme("3", "accent", th)) {
		t.Errorf("numeric value not colored; got %q", got)
	}
}

func TestHighlightYAML_Value_QuotedDouble(t *testing.T) {
	th := testTheme()
	got := highlightYAML("name: \"my-pod\"\n", th)
	if !strings.Contains(got, wrapWithTheme("\"my-pod\"", "success", th)) {
		t.Errorf("double-quoted string not colored; got %q", got)
	}
}

func TestHighlightYAML_Value_QuotedSingle(t *testing.T) {
	th := testTheme()
	got := highlightYAML("name: 'my-pod'\n", th)
	if !strings.Contains(got, wrapWithTheme("'my-pod'", "success", th)) {
		t.Errorf("single-quoted string not colored; got %q", got)
	}
}

func TestHighlightYAML_ListBullet_Colored(t *testing.T) {
	th := testTheme()
	got := highlightYAML("- item\n", th)
	if !strings.Contains(got, wrapWithTheme("-", "pink", th)) {
		t.Errorf("list bullet not colored; got %q", got)
	}
}

func TestHighlightYAML_IndentedKey(t *testing.T) {
	th := testTheme()
	got := highlightYAML("  name: my-pod\n", th)
	if !strings.Contains(got, wrapWithTheme("name", "key", th)) {
		t.Errorf("indented key not colored; got %q", got)
	}
	if !strings.HasPrefix(got, "  ") {
		t.Errorf("indentation not preserved; got %q", got)
	}
}

func TestHighlightYAML_EmptyValue(t *testing.T) {
	th := testTheme()
	got := highlightYAML("labels:\n", th)
	if !strings.Contains(got, wrapWithTheme("labels", "key", th)) {
		t.Errorf("key with empty value not colored; got %q", got)
	}
}

func TestHighlightYAML_PreservesNewlines(t *testing.T) {
	th := testTheme()
	input := "---\nname: pod\n"
	got := highlightYAML(input, th)
	if strings.Count(got, "\n") != strings.Count(input, "\n") {
		t.Errorf("newline count changed: input %d, got %d", strings.Count(input, "\n"), strings.Count(got, "\n"))
	}
}
```

- [ ] **Step 2: Run tests to confirm they pass with current code**

```
go test ./output/... -run TestHighlightYAML -v
```

Expected: all PASS (these tests pin the existing behavior before refactoring).

- [ ] **Step 3: Rewrite `output/yaml.go` with the `yamlHL` struct**

Replace the entire file contents:

```go
package output

import (
	"strings"

	"github.com/thephilip/oc-color/theme"
)

func looksLikeYAML(output string) bool {
	trimmed := strings.TrimLeft(output, " \t\n\r")
	return strings.HasPrefix(trimmed, "---")
}

type yamlHL struct {
	theme theme.Theme
}

func highlightYAML(input string, th theme.Theme) string {
	h := &yamlHL{theme: th}
	lines := strings.SplitAfter(input, "\n")
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(h.processLine(line))
	}
	return b.String()
}

func (h *yamlHL) processLine(line string) string {
	trimmed := strings.TrimRight(line, "\n\r")
	suffix := line[len(trimmed):]

	if strings.TrimSpace(trimmed) == "" {
		return line
	}

	iLen := indentLen(trimmed)
	indentStr := trimmed[:iLen]
	content := trimmed[iLen:]

	if content == "---" || content == "..." {
		return indentStr + wrapWithTheme(content, "pink", h.theme) + suffix
	}

	if strings.HasPrefix(content, "#") {
		return indentStr + wrapWithTheme(content, "dim", h.theme) + suffix
	}

	if strings.HasPrefix(content, "- ") || content == "-" {
		rest := ""
		if len(content) > 2 {
			rest = content[2:]
		}
		return indentStr + wrapWithTheme("-", "pink", h.theme) + " " + h.colorizeValue(rest) + suffix
	}

	if idx := strings.Index(content, ":"); idx >= 0 {
		key := content[:idx]
		afterColon := content[idx+1:]

		if key != "" && !strings.Contains(key, " ") {
			colored := indentStr + wrapWithTheme(key, "key", h.theme) + ":"

			valTrimmed := strings.TrimSpace(afterColon)
			if valTrimmed != "" {
				spaces := afterColon[:len(afterColon)-len(strings.TrimLeft(afterColon, " "))]
				colored += spaces + h.colorizeValue(valTrimmed)
			} else {
				colored += afterColon
			}
			return colored + suffix
		}
	}

	return line
}

func (h *yamlHL) colorizeValue(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return s
	}
	if (strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"")) ||
		(strings.HasPrefix(trimmed, "'") && strings.HasSuffix(trimmed, "'")) {
		return wrapWithTheme(trimmed, "success", h.theme)
	}
	return colorizeScalarValue(s, h.theme)
}

func indentLen(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return i
		}
	}
	return len(s)
}
```

- [ ] **Step 4: Run all tests to confirm behavior is preserved**

```
go test ./... -v 2>&1 | tail -30
```

Expected: all PASS. If any YAML test fails, fix `processLine` before continuing.

- [ ] **Step 5: Commit**

```bash
git add output/yaml.go output/yaml_test.go
git commit -m "refactor: rewrite YAML highlighter as state-aware yamlHL struct"
```

---

### Task 4: Add block scalar support

Block scalars (`|` and `>`) mark the start of multi-line string content. Lines after the header are string content until a line with indentation ≤ the key's indentation level is encountered.

**Files:**
- Modify: `output/yaml.go`
- Modify: `output/yaml_test.go`

- [ ] **Step 1: Write the failing tests**

Append to `output/yaml_test.go`:

```go
func TestHighlightYAML_BlockScalar_Pipe(t *testing.T) {
	th := testTheme()
	input := "command: |\n  echo hello\n  echo world\n"
	got := highlightYAML(input, th)
	if !strings.Contains(got, wrapWithTheme("|", "dim", th)) {
		t.Errorf("block scalar header '|' not colored dim; got %q", got)
	}
	if !strings.Contains(got, wrapWithTheme("echo hello", "success", th)) {
		t.Errorf("block scalar content not colored success; got %q", got)
	}
	if !strings.Contains(got, wrapWithTheme("echo world", "success", th)) {
		t.Errorf("second block scalar line not colored success; got %q", got)
	}
}

func TestHighlightYAML_BlockScalar_Fold(t *testing.T) {
	th := testTheme()
	input := "message: >\n  long line\n  continuation\n"
	got := highlightYAML(input, th)
	if !strings.Contains(got, wrapWithTheme(">", "dim", th)) {
		t.Errorf("block scalar header '>' not colored dim; got %q", got)
	}
	if !strings.Contains(got, wrapWithTheme("long line", "success", th)) {
		t.Errorf("folded scalar content not colored success; got %q", got)
	}
}

func TestHighlightYAML_BlockScalar_WithChomping(t *testing.T) {
	th := testTheme()
	input := "script: |-\n  set -e\n  echo done\n"
	got := highlightYAML(input, th)
	if !strings.Contains(got, wrapWithTheme("|-", "dim", th)) {
		t.Errorf("block scalar header '|-' not colored dim; got %q", got)
	}
	if !strings.Contains(got, wrapWithTheme("set -e", "success", th)) {
		t.Errorf("block scalar content not colored success; got %q", got)
	}
}

func TestHighlightYAML_BlockScalar_EndedByReducedIndent(t *testing.T) {
	th := testTheme()
	input := "data: |\n  line one\nnextKey: value\n"
	got := highlightYAML(input, th)
	if !strings.Contains(got, wrapWithTheme("nextKey", "key", th)) {
		t.Errorf("key after block scalar not recognized as key; got %q", got)
	}
}

func TestHighlightYAML_BlockScalar_BlankLineInside(t *testing.T) {
	th := testTheme()
	// A blank line inside a block scalar must not end the scalar
	input := "data: |\n  line one\n\n  line two\nnextKey: val\n"
	got := highlightYAML(input, th)
	if !strings.Contains(got, wrapWithTheme("line two", "success", th)) {
		t.Errorf("line after blank inside block scalar not colored success; got %q", got)
	}
	if !strings.Contains(got, wrapWithTheme("nextKey", "key", th)) {
		t.Errorf("key after block scalar not recognized; got %q", got)
	}
}
```

- [ ] **Step 2: Run to confirm they fail**

```
go test ./output/... -run "TestHighlightYAML_BlockScalar" -v
```

Expected: FAIL — block scalar content is processed as key-value lines and produces incorrect output.

- [ ] **Step 3: Add block scalar state to `yamlHL` and update `output/yaml.go`**

Update the `yamlHL` struct, add `isBlockScalarHeader`, and update `processLine`. Replace the entire `output/yaml.go`:

```go
package output

import (
	"strings"

	"github.com/thephilip/oc-color/theme"
)

func looksLikeYAML(output string) bool {
	trimmed := strings.TrimLeft(output, " \t\n\r")
	return strings.HasPrefix(trimmed, "---")
}

type yamlHL struct {
	theme       theme.Theme
	blockScalar bool
	blockIndent int
}

func highlightYAML(input string, th theme.Theme) string {
	h := &yamlHL{theme: th}
	lines := strings.SplitAfter(input, "\n")
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(h.processLine(line))
	}
	return b.String()
}

func (h *yamlHL) processLine(line string) string {
	trimmed := strings.TrimRight(line, "\n\r")
	suffix := line[len(trimmed):]

	if strings.TrimSpace(trimmed) == "" {
		return line
	}

	iLen := indentLen(trimmed)
	indentStr := trimmed[:iLen]
	content := trimmed[iLen:]

	// Block scalar content: color as string until indentation returns to key level
	if h.blockScalar {
		if iLen <= h.blockIndent {
			h.blockScalar = false
			// fall through to normal processing
		} else {
			return indentStr + wrapWithTheme(content, "success", h.theme) + suffix
		}
	}

	if content == "---" || content == "..." {
		return indentStr + wrapWithTheme(content, "pink", h.theme) + suffix
	}

	if strings.HasPrefix(content, "#") {
		return indentStr + wrapWithTheme(content, "dim", h.theme) + suffix
	}

	if strings.HasPrefix(content, "- ") || content == "-" {
		rest := ""
		if len(content) > 2 {
			rest = content[2:]
		}
		return indentStr + wrapWithTheme("-", "pink", h.theme) + " " + h.colorizeValue(rest) + suffix
	}

	if idx := strings.Index(content, ":"); idx >= 0 {
		key := content[:idx]
		afterColon := content[idx+1:]

		if key != "" && !strings.Contains(key, " ") {
			colored := indentStr + wrapWithTheme(key, "key", h.theme) + ":"

			valTrimmed := strings.TrimSpace(afterColon)
			if valTrimmed != "" {
				spaces := afterColon[:len(afterColon)-len(strings.TrimLeft(afterColon, " "))]
				if isBlockScalarHeader(valTrimmed) {
					h.blockScalar = true
					h.blockIndent = iLen
					colored += spaces + wrapWithTheme(valTrimmed, "dim", h.theme)
				} else {
					colored += spaces + h.colorizeValue(valTrimmed)
				}
			} else {
				colored += afterColon
			}
			return colored + suffix
		}
	}

	return line
}

func (h *yamlHL) colorizeValue(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return s
	}
	if (strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"")) ||
		(strings.HasPrefix(trimmed, "'") && strings.HasSuffix(trimmed, "'")) {
		return wrapWithTheme(trimmed, "success", h.theme)
	}
	return colorizeScalarValue(s, h.theme)
}

// isBlockScalarHeader reports whether s is a YAML block scalar indicator:
// | or > optionally followed by a chomping indicator (- or +) and/or
// an indentation indicator (1-9) in either order.
func isBlockScalarHeader(s string) bool {
	if len(s) == 0 || (s[0] != '|' && s[0] != '>') {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if c != '-' && c != '+' && (c < '1' || c > '9') {
			return false
		}
	}
	return true
}

func indentLen(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return i
		}
	}
	return len(s)
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```
go test ./output/... -run "TestHighlightYAML" -v
```

Expected: all PASS.

Full suite:

```
go test ./... 2>&1 | tail -10
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add output/yaml.go output/yaml_test.go
git commit -m "feat: add block scalar support to YAML highlighter"
```

---

### Task 5: Add flow collection support

Flow sequences (`[a, b, c]`) and flow mappings (`{key: val}`) tokenize inline content. Brackets are colored `pink`, commas `dim`, keys `key`, scalar values via `colorizeScalarValue`. Multi-line flow collections are tracked via `flowDepth`.

**Files:**
- Modify: `output/yaml.go`
- Modify: `output/yaml_test.go`

- [ ] **Step 1: Write the failing tests**

Append to `output/yaml_test.go`:

```go
func TestHighlightYAML_FlowSequence_Brackets(t *testing.T) {
	th := testTheme()
	got := highlightYAML("labels: [app=myapp, tier=frontend]\n", th)
	if !strings.Contains(got, wrapWithTheme("[", "pink", th)) {
		t.Errorf("opening bracket not colored pink; got %q", got)
	}
	if !strings.Contains(got, wrapWithTheme("]", "pink", th)) {
		t.Errorf("closing bracket not colored pink; got %q", got)
	}
}

func TestHighlightYAML_FlowSequence_Comma(t *testing.T) {
	th := testTheme()
	got := highlightYAML("labels: [app=myapp, tier=frontend]\n", th)
	if !strings.Contains(got, wrapWithTheme(",", "dim", th)) {
		t.Errorf("comma not colored dim; got %q", got)
	}
}

func TestHighlightYAML_FlowMapping_Braces(t *testing.T) {
	th := testTheme()
	got := highlightYAML("options: {key: val}\n", th)
	if !strings.Contains(got, wrapWithTheme("{", "pink", th)) {
		t.Errorf("opening brace not colored pink; got %q", got)
	}
	if !strings.Contains(got, wrapWithTheme("}", "pink", th)) {
		t.Errorf("closing brace not colored pink; got %q", got)
	}
}

func TestHighlightYAML_FlowMapping_Key(t *testing.T) {
	th := testTheme()
	got := highlightYAML("options: {timeout: 30}\n", th)
	if !strings.Contains(got, wrapWithTheme("timeout", "key", th)) {
		t.Errorf("key inside flow mapping not colored; got %q", got)
	}
}

func TestHighlightYAML_FlowSequence_IPv4Items(t *testing.T) {
	th := testTheme()
	got := highlightYAML("ips: [192.168.1.1, 10.0.0.1]\n", th)
	if !strings.Contains(got, wrapWithTheme("192.168.1.1", "accent", th)) {
		t.Errorf("IPv4 inside flow sequence not colored accent; got %q", got)
	}
}

func TestHighlightYAML_FlowSequence_BoolItems(t *testing.T) {
	th := testTheme()
	got := highlightYAML("flags: [true, false]\n", th)
	if !strings.Contains(got, wrapWithTheme("true", "info", th)) {
		t.Errorf("bool inside flow sequence not colored info; got %q", got)
	}
}

func TestHighlightYAML_FlowSequence_QuotedItems(t *testing.T) {
	th := testTheme()
	got := highlightYAML(`tags: ["prod", "web"]` + "\n", th)
	if !strings.Contains(got, wrapWithTheme(`"prod"`, "success", th)) {
		t.Errorf("quoted string inside flow sequence not colored success; got %q", got)
	}
}
```

- [ ] **Step 2: Run to confirm they fail**

```
go test ./output/... -run "TestHighlightYAML_Flow" -v
```

Expected: FAIL — brackets and their contents are passed through uncolored.

- [ ] **Step 3: Add `flowDepth` to `yamlHL` and `colorizeFlow` method; update `processLine`**

Replace the entire `output/yaml.go`:

```go
package output

import (
	"strings"

	"github.com/thephilip/oc-color/theme"
)

func looksLikeYAML(output string) bool {
	trimmed := strings.TrimLeft(output, " \t\n\r")
	return strings.HasPrefix(trimmed, "---")
}

type yamlHL struct {
	theme       theme.Theme
	blockScalar bool
	blockIndent int
	flowDepth   int
}

func highlightYAML(input string, th theme.Theme) string {
	h := &yamlHL{theme: th}
	lines := strings.SplitAfter(input, "\n")
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(h.processLine(line))
	}
	return b.String()
}

func (h *yamlHL) processLine(line string) string {
	trimmed := strings.TrimRight(line, "\n\r")
	suffix := line[len(trimmed):]

	if strings.TrimSpace(trimmed) == "" {
		return line
	}

	iLen := indentLen(trimmed)
	indentStr := trimmed[:iLen]
	content := trimmed[iLen:]

	// Continuation of a multi-line flow collection
	if h.flowDepth > 0 {
		return indentStr + h.colorizeFlow(content) + suffix
	}

	// Block scalar content: color as string until indentation returns to key level
	if h.blockScalar {
		if iLen <= h.blockIndent {
			h.blockScalar = false
			// fall through to normal processing
		} else {
			return indentStr + wrapWithTheme(content, "success", h.theme) + suffix
		}
	}

	if content == "---" || content == "..." {
		return indentStr + wrapWithTheme(content, "pink", h.theme) + suffix
	}

	if strings.HasPrefix(content, "#") {
		return indentStr + wrapWithTheme(content, "dim", h.theme) + suffix
	}

	if strings.HasPrefix(content, "- ") || content == "-" {
		rest := ""
		if len(content) > 2 {
			rest = content[2:]
		}
		return indentStr + wrapWithTheme("-", "pink", h.theme) + " " + h.colorizeValue(rest) + suffix
	}

	if idx := strings.Index(content, ":"); idx >= 0 {
		key := content[:idx]
		afterColon := content[idx+1:]

		if key != "" && !strings.Contains(key, " ") {
			colored := indentStr + wrapWithTheme(key, "key", h.theme) + ":"

			valTrimmed := strings.TrimSpace(afterColon)
			if valTrimmed != "" {
				spaces := afterColon[:len(afterColon)-len(strings.TrimLeft(afterColon, " "))]
				if isBlockScalarHeader(valTrimmed) {
					h.blockScalar = true
					h.blockIndent = iLen
					colored += spaces + wrapWithTheme(valTrimmed, "dim", h.theme)
				} else if valTrimmed[0] == '[' || valTrimmed[0] == '{' {
					colored += spaces + h.colorizeFlow(valTrimmed)
				} else {
					colored += spaces + h.colorizeValue(valTrimmed)
				}
			} else {
				colored += afterColon
			}
			return colored + suffix
		}
	}

	return line
}

func (h *yamlHL) colorizeValue(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return s
	}
	if (strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"")) ||
		(strings.HasPrefix(trimmed, "'") && strings.HasSuffix(trimmed, "'")) {
		return wrapWithTheme(trimmed, "success", h.theme)
	}
	return colorizeScalarValue(s, h.theme)
}

// colorizeFlow tokenizes a flow collection string character by character,
// coloring brackets, commas, keys, and scalar values.
func (h *yamlHL) colorizeFlow(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		c := s[i]
		switch {
		case c == '[' || c == '{':
			b.WriteString(wrapWithTheme(string(c), "pink", h.theme))
			h.flowDepth++
			i++
		case c == ']' || c == '}':
			if h.flowDepth > 0 {
				h.flowDepth--
			}
			b.WriteString(wrapWithTheme(string(c), "pink", h.theme))
			i++
		case c == ',':
			b.WriteString(wrapWithTheme(",", "dim", h.theme))
			i++
		case c == ':':
			b.WriteString(wrapWithTheme(":", "dim", h.theme))
			i++
		case c == ' ' || c == '\t':
			b.WriteByte(c)
			i++
		case c == '"' || c == '\'':
			quote := c
			start := i
			i++
			for i < len(s) {
				if s[i] == '\\' {
					i += 2
					continue
				}
				if s[i] == quote {
					i++
					break
				}
				i++
			}
			b.WriteString(wrapWithTheme(s[start:i], "success", h.theme))
		default:
			// read an unquoted token
			start := i
			for i < len(s) && s[i] != ',' && s[i] != ']' && s[i] != '}' && s[i] != ':' && s[i] != ' ' && s[i] != '\t' {
				i++
			}
			token := s[start:i]
			// peek past whitespace to detect key (token followed by ':')
			j := i
			for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
				j++
			}
			if j < len(s) && s[j] == ':' {
				b.WriteString(wrapWithTheme(token, "key", h.theme))
			} else {
				b.WriteString(colorizeScalarValue(token, h.theme))
			}
		}
	}
	return b.String()
}

// isBlockScalarHeader reports whether s is a YAML block scalar indicator:
// | or > optionally followed by a chomping indicator (- or +) and/or
// an indentation indicator (1-9) in either order.
func isBlockScalarHeader(s string) bool {
	if len(s) == 0 || (s[0] != '|' && s[0] != '>') {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if c != '-' && c != '+' && (c < '1' || c > '9') {
			return false
		}
	}
	return true
}

func indentLen(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return i
		}
	}
	return len(s)
}
```

- [ ] **Step 4: Run all tests**

```
go test ./output/... -run "TestHighlightYAML" -v
```

Expected: all PASS.

Full suite:

```
go test ./... 2>&1 | tail -10
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add output/yaml.go output/yaml_test.go
git commit -m "feat: add flow collection support to YAML highlighter"
```

---

## Self-Review

**Spec coverage:**
- Shared `colorizeScalarValue` function → Task 1 ✓
- IP address detection (`accent`) → Task 2 ✓
- Timestamp detection (`dim`) → Task 2 ✓
- `yamlHL` struct with `flowDepth`, `blockScalar`, `blockIndent` → Tasks 3–5 ✓
- Block scalar header detection and content coloring → Task 4 ✓
- Flow collection brackets (`pink`), commas (`dim`), keys (`key`), scalars → Task 5 ✓
- JSON: no change needed — character-based parser already handles all types; IPs/timestamps appear as quoted strings and are colored `success` ✓
- Error handling (pass-through on unrecognized input): covered by the `return line` fallback in `processLine` ✓
- Existing tests preserved: Task 3 pins existing behavior before refactoring ✓

**Placeholder scan:** No TBDs. All code blocks are complete.

**Type consistency:**
- `yamlHL` struct fields (`blockScalar bool`, `blockIndent int`, `flowDepth int`) are defined once in Task 3 and carried through Tasks 4 and 5 — each task shows the full struct, preventing drift.
- `colorizeScalarValue(s string, th theme.Theme) string` — consistent across Tasks 1, 2, and all callers.
- `isBlockScalarHeader(s string) bool` — defined in Task 4, referenced only within `processLine` in the same task.
- `indentLen(s string) int` — defined in Task 3, present in all subsequent full-file replacements.
