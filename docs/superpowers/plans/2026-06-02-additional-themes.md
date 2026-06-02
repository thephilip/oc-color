# Additional Built-in Themes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add six new built-in themes (Nord, One Dark, Gruvbox, Catppuccin Mocha, Tokyo Night, Solarized Dark) and migrate Dracula to a separate file so each theme is self-contained and self-registering.

**Architecture:** Each theme lives in its own file in the `theme/` package. An `init()` function in each file registers the theme into the shared `builtins` map. `theme/theme.go`'s `init()` initialises the empty map; all population happens via theme files. No central list to maintain.

**Tech Stack:** Go stdlib only — no new dependencies.

---

## Files

- Modify: `theme/theme.go` — `init()` → empty map, delete `dracula()`
- Create: `theme/dracula.go` — migrated dracula with self-registration
- Create: `theme/nord.go`
- Create: `theme/one-dark.go`
- Create: `theme/gruvbox.go`
- Create: `theme/catppuccin.go`
- Create: `theme/tokyo-night.go`
- Create: `theme/solarized.go`
- Modify: `theme/theme_test.go` — fix custom-theme tests, add new theme tests

---

### Task 1: Write failing tests + fix custom-theme test collision

**Context:** `TestLoadCustomTheme` currently uses "nord" as its example custom theme name. Once we add a builtin nord, `Get("nord")` will hit the builtin first and the test will silently stop testing custom loading. Fix it now to use "mytheme".

**Files:**
- Modify: `theme/theme_test.go`

- [ ] **Step 1: Update `TestLoadCustomTheme` to use "mytheme" instead of "nord"**

In `theme/theme_test.go`, replace `TestLoadCustomTheme` (lines 200–231) with:

```go
func TestLoadCustomTheme(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	themesDir := filepath.Join(home, ".config", "oc-color", "themes")
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `name: mytheme
tokens:
  success: green
  warning: yellow
  error: bold+red
  info: cyan
  accent: purple
  dim: gray
  header: bold+blue+underline
  key: yellow
`
	if err := os.WriteFile(filepath.Join(themesDir, "mytheme.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	th, ok := Get("mytheme")
	if !ok {
		t.Fatal("expected to find mytheme custom theme")
	}
	if th.Name != "mytheme" {
		t.Errorf("expected name 'mytheme', got %q", th.Name)
	}
}
```

- [ ] **Step 2: Update `TestNamesIncludesCustom` to use "mytheme" instead of "nord"**

Replace `TestNamesIncludesCustom` (lines 233–267) with:

```go
func TestNamesIncludesCustom(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	themesDir := filepath.Join(home, ".config", "oc-color", "themes")
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `name: mytheme
tokens:
  success: green
  warning: yellow
  error: bold+red
  info: cyan
  accent: purple
  dim: gray
  header: bold+blue+underline
  key: yellow
`
	if err := os.WriteFile(filepath.Join(themesDir, "mytheme.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	names := Names()
	found := false
	for _, n := range names {
		if n == "mytheme" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'mytheme' in Names() = %v", names)
	}
}
```

- [ ] **Step 3: Append new failing tests for all six themes and the full builtin names list**

Append after `TestNamesIncludesCustom`:

```go
func testBuiltinTheme(t *testing.T, name string) {
	t.Helper()
	th, ok := Get(name)
	if !ok {
		t.Fatalf("%s theme not found", name)
	}
	required := []string{"success", "warning", "error", "info", "accent", "dim", "header", "key"}
	for _, tok := range required {
		if _, exists := th.Tokens[tok]; !exists {
			t.Errorf("%s: missing required token %q", name, tok)
		}
	}
}

func TestNordTheme(t *testing.T)        { testBuiltinTheme(t, "nord") }
func TestOneDarkTheme(t *testing.T)     { testBuiltinTheme(t, "one-dark") }
func TestGruvboxTheme(t *testing.T)     { testBuiltinTheme(t, "gruvbox") }
func TestCatppuccinTheme(t *testing.T)  { testBuiltinTheme(t, "catppuccin") }
func TestTokyoNightTheme(t *testing.T)  { testBuiltinTheme(t, "tokyo-night") }
func TestSolarizedTheme(t *testing.T)   { testBuiltinTheme(t, "solarized") }

func TestAllBuiltinNames(t *testing.T) {
	names := Names()
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}
	expected := []string{"dracula", "nord", "one-dark", "gruvbox", "catppuccin", "tokyo-night", "solarized"}
	for _, want := range expected {
		if !nameSet[want] {
			t.Errorf("Names() missing %q", want)
		}
	}
}
```

- [ ] **Step 4: Run tests — existing tests must pass, new theme tests must fail**

```bash
go test ./theme/... -v 2>&1 | grep -E "^(=== RUN|--- PASS|--- FAIL|FAIL|ok)"
```

Expected: all existing tests PASS, `TestNordTheme` through `TestSolarizedTheme` and `TestAllBuiltinNames` FAIL.

- [ ] **Step 5: Commit**

```bash
git add theme/theme_test.go
git commit -m "test: add failing tests for new built-in themes (red)"
```

---

### Task 2: Migrate Dracula + update theme.go init()

**Files:**
- Create: `theme/dracula.go`
- Modify: `theme/theme.go`

- [ ] **Step 1: Create `theme/dracula.go`**

```go
package theme

func init() {
	builtins["dracula"] = dracula()
}

func dracula() Theme {
	return Theme{
		Name: "dracula",
		Tokens: map[string]TokenStyle{
			"success": {Color: "#50FA7B"},
			"warning": {Color: "#F1FA8C"},
			"error":   {Color: "#FF5555", Bold: true},
			"info":    {Color: "#8BE9FD"},
			"accent":  {Color: "#BD93F9"},
			"pink":    {Color: "#FF79C6"},
			"orange":  {Color: "#FFB86C"},
			"dim":     {Color: "#6272A4"},
			"shade":   {Background: "#2E3040"},
			"header":  {Color: "#BD93F9", Bold: true, Underline: true},
			"key":     {Color: "#F1FA8C"},
			"value":   {Color: "#F8F8F2"},
		},
	}
}
```

- [ ] **Step 2: Update `theme/theme.go` — replace `init()` and delete `dracula()`**

In `theme/theme.go`, replace the current `init()` block:

```go
func init() {
	builtins = map[string]Theme{
		"dracula": dracula(),
	}
}
```

With:

```go
func init() {
	builtins = map[string]Theme{}
}
```

Then delete the entire `dracula()` function from `theme/theme.go` (it now lives in `theme/dracula.go`).

- [ ] **Step 3: Run tests — all existing tests must still pass**

```bash
go test ./theme/... -v 2>&1 | grep -E "^(=== RUN|--- PASS|--- FAIL|FAIL|ok)"
```

Expected: all existing tests PASS (including `TestGetBuiltin` which tests dracula). The six new theme tests still FAIL.

- [ ] **Step 4: Commit**

```bash
git add theme/theme.go theme/dracula.go
git commit -m "refactor: move dracula to theme/dracula.go, self-register via init()"
```

---

### Task 3: Add Nord theme

**Files:**
- Create: `theme/nord.go`

- [ ] **Step 1: Create `theme/nord.go`**

```go
package theme

func init() {
	builtins["nord"] = nord()
}

func nord() Theme {
	return Theme{
		Name: "nord",
		Tokens: map[string]TokenStyle{
			"success": {Color: "#A3BE8C"},
			"warning": {Color: "#EBCB8B"},
			"error":   {Color: "#BF616A", Bold: true},
			"info":    {Color: "#88C0D0"},
			"accent":  {Color: "#B48EAD"},
			"dim":     {Color: "#4C566A"},
			"shade":   {Background: "#3B4252"},
			"header":  {Color: "#81A1C1", Bold: true, Underline: true},
			"key":     {Color: "#EBCB8B"},
			"value":   {Color: "#ECEFF4"},
			"pink":    {Color: "#BF616A"},
			"orange":  {Color: "#D08770"},
		},
	}
}
```

- [ ] **Step 2: Run the Nord test**

```bash
go test ./theme/... -run TestNordTheme -v
```

Expected: `--- PASS: TestNordTheme`

- [ ] **Step 3: Commit**

```bash
git add theme/nord.go
git commit -m "feat: add Nord built-in theme"
```

---

### Task 4: Add One Dark theme

**Files:**
- Create: `theme/one-dark.go`

- [ ] **Step 1: Create `theme/one-dark.go`**

```go
package theme

func init() {
	builtins["one-dark"] = oneDark()
}

func oneDark() Theme {
	return Theme{
		Name: "one-dark",
		Tokens: map[string]TokenStyle{
			"success": {Color: "#98C379"},
			"warning": {Color: "#E5C07B"},
			"error":   {Color: "#E06C75", Bold: true},
			"info":    {Color: "#61AFEF"},
			"accent":  {Color: "#C678DD"},
			"dim":     {Color: "#5C6370"},
			"shade":   {Background: "#2C313C"},
			"header":  {Color: "#61AFEF", Bold: true, Underline: true},
			"key":     {Color: "#E5C07B"},
			"value":   {Color: "#ABB2BF"},
			"pink":    {Color: "#E06C75"},
			"orange":  {Color: "#D19A66"},
		},
	}
}
```

- [ ] **Step 2: Run the One Dark test**

```bash
go test ./theme/... -run TestOneDarkTheme -v
```

Expected: `--- PASS: TestOneDarkTheme`

- [ ] **Step 3: Commit**

```bash
git add theme/one-dark.go
git commit -m "feat: add One Dark built-in theme"
```

---

### Task 5: Add Gruvbox theme

**Files:**
- Create: `theme/gruvbox.go`

- [ ] **Step 1: Create `theme/gruvbox.go`**

```go
package theme

func init() {
	builtins["gruvbox"] = gruvbox()
}

func gruvbox() Theme {
	return Theme{
		Name: "gruvbox",
		Tokens: map[string]TokenStyle{
			"success": {Color: "#B8BB26"},
			"warning": {Color: "#FABD2F"},
			"error":   {Color: "#FB4934", Bold: true},
			"info":    {Color: "#83A598"},
			"accent":  {Color: "#D3869B"},
			"dim":     {Color: "#928374"},
			"shade":   {Background: "#3C3836"},
			"header":  {Color: "#83A598", Bold: true, Underline: true},
			"key":     {Color: "#FABD2F"},
			"value":   {Color: "#EBDBB2"},
			"pink":    {Color: "#D3869B"},
			"orange":  {Color: "#FE8019"},
		},
	}
}
```

- [ ] **Step 2: Run the Gruvbox test**

```bash
go test ./theme/... -run TestGruvboxTheme -v
```

Expected: `--- PASS: TestGruvboxTheme`

- [ ] **Step 3: Commit**

```bash
git add theme/gruvbox.go
git commit -m "feat: add Gruvbox built-in theme"
```

---

### Task 6: Add Catppuccin Mocha theme

**Files:**
- Create: `theme/catppuccin.go`

- [ ] **Step 1: Create `theme/catppuccin.go`**

```go
package theme

func init() {
	builtins["catppuccin"] = catppuccin()
}

func catppuccin() Theme {
	return Theme{
		Name: "catppuccin",
		Tokens: map[string]TokenStyle{
			"success": {Color: "#A6E3A1"},
			"warning": {Color: "#F9E2AF"},
			"error":   {Color: "#F38BA8", Bold: true},
			"info":    {Color: "#89DCEB"},
			"accent":  {Color: "#CBA6F7"},
			"dim":     {Color: "#6C7086"},
			"shade":   {Background: "#181825"},
			"header":  {Color: "#89B4FA", Bold: true, Underline: true},
			"key":     {Color: "#F9E2AF"},
			"value":   {Color: "#CDD6F4"},
			"pink":    {Color: "#F38BA8"},
			"orange":  {Color: "#FAB387"},
		},
	}
}
```

- [ ] **Step 2: Run the Catppuccin test**

```bash
go test ./theme/... -run TestCatppuccinTheme -v
```

Expected: `--- PASS: TestCatppuccinTheme`

- [ ] **Step 3: Commit**

```bash
git add theme/catppuccin.go
git commit -m "feat: add Catppuccin Mocha built-in theme"
```

---

### Task 7: Add Tokyo Night theme

**Files:**
- Create: `theme/tokyo-night.go`

- [ ] **Step 1: Create `theme/tokyo-night.go`**

```go
package theme

func init() {
	builtins["tokyo-night"] = tokyoNight()
}

func tokyoNight() Theme {
	return Theme{
		Name: "tokyo-night",
		Tokens: map[string]TokenStyle{
			"success": {Color: "#9ECE6A"},
			"warning": {Color: "#E0AF68"},
			"error":   {Color: "#F7768E", Bold: true},
			"info":    {Color: "#7DCFFF"},
			"accent":  {Color: "#BB9AF7"},
			"dim":     {Color: "#565F89"},
			"shade":   {Background: "#16161E"},
			"header":  {Color: "#7AA2F7", Bold: true, Underline: true},
			"key":     {Color: "#E0AF68"},
			"value":   {Color: "#C0CAF5"},
			"pink":    {Color: "#F7768E"},
			"orange":  {Color: "#FF9E64"},
		},
	}
}
```

- [ ] **Step 2: Run the Tokyo Night test**

```bash
go test ./theme/... -run TestTokyoNightTheme -v
```

Expected: `--- PASS: TestTokyoNightTheme`

- [ ] **Step 3: Commit**

```bash
git add theme/tokyo-night.go
git commit -m "feat: add Tokyo Night built-in theme"
```

---

### Task 8: Add Solarized Dark theme + verify all tests green

**Files:**
- Create: `theme/solarized.go`

- [ ] **Step 1: Create `theme/solarized.go`**

```go
package theme

func init() {
	builtins["solarized"] = solarized()
}

func solarized() Theme {
	return Theme{
		Name: "solarized",
		Tokens: map[string]TokenStyle{
			"success": {Color: "#859900"},
			"warning": {Color: "#B58900"},
			"error":   {Color: "#DC322F", Bold: true},
			"info":    {Color: "#268BD2"},
			"accent":  {Color: "#6C71C4"},
			"dim":     {Color: "#657B83"},
			"shade":   {Background: "#073642"},
			"header":  {Color: "#268BD2", Bold: true, Underline: true},
			"key":     {Color: "#B58900"},
			"value":   {Color: "#93A1A1"},
			"pink":    {Color: "#D33682"},
			"orange":  {Color: "#CB4B16"},
		},
	}
}
```

- [ ] **Step 2: Run the full test suite**

```bash
go test ./... -v 2>&1 | grep -E "^(=== RUN|--- PASS|--- FAIL|FAIL|ok)"
```

Expected: all tests PASS — including `TestAllBuiltinNames`, all six `Test*Theme` tests, and all pre-existing tests.

- [ ] **Step 3: Commit**

```bash
git add theme/solarized.go
git commit -m "feat: add Solarized Dark built-in theme"
```
