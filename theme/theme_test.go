package theme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGetBuiltin(t *testing.T) {
	th, ok := Get("dracula")
	if !ok {
		t.Fatal("expected to find builtin dracula theme")
	}
	if th.Name != "dracula" {
		t.Errorf("expected name 'dracula', got %q", th.Name)
	}
	if _, ok := th.Tokens["success"]; !ok {
		t.Error("expected success token in dracula theme")
	}
}

func TestGetBuiltinCaseInsensitive(t *testing.T) {
	_, ok := Get("DRACULA")
	if !ok {
		t.Error("expected case-insensitive lookup to work")
	}
}

func TestGetUnknown(t *testing.T) {
	_, ok := Get("nonexistent")
	if ok {
		t.Error("expected nonexistent theme to not be found")
	}
}

func TestNamesIncludesBuiltin(t *testing.T) {
	names := Names()
	found := false
	for _, n := range names {
		if n == "dracula" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'dracula' in Names()")
	}
}

func TestParseShorthand(t *testing.T) {
	tests := []struct {
		input string
		want  TokenStyle
	}{
		{"green", TokenStyle{Color: "green"}},
		{"bold+red", TokenStyle{Color: "red", Bold: true}},
		{"dim+cyan", TokenStyle{Color: "cyan", Dim: true}},
		{"bold+underline+blue", TokenStyle{Color: "blue", Bold: true, Underline: true}},
		{"italic+yellow", TokenStyle{Color: "yellow", Italic: true}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseShorthand(tt.input)
			if got.Color != tt.want.Color {
				t.Errorf("Color = %q, want %q", got.Color, tt.want.Color)
			}
			if got.Bold != tt.want.Bold {
				t.Errorf("Bold = %v, want %v", got.Bold, tt.want.Bold)
			}
			if got.Dim != tt.want.Dim {
				t.Errorf("Dim = %v, want %v", got.Dim, tt.want.Dim)
			}
			if got.Italic != tt.want.Italic {
				t.Errorf("Italic = %v, want %v", got.Italic, tt.want.Italic)
			}
			if got.Underline != tt.want.Underline {
				t.Errorf("Underline = %v, want %v", got.Underline, tt.want.Underline)
			}
		})
	}
}

func TestSequence(t *testing.T) {
	tests := []struct {
		name  string
		style TokenStyle
		want  string
	}{
		{
			name:  "empty",
			style: TokenStyle{},
			want:  "",
		},
		{
			name:  "named color only",
			style: TokenStyle{Color: "green"},
			want:  "\033[32m",
		},
		{
			name:  "bold red",
			style: TokenStyle{Color: "red", Bold: true},
			want:  "\033[31;1m",
		},
		{
			name:  "dim",
			style: TokenStyle{Color: "cyan", Dim: true},
			want:  "\033[36;2m",
		},
		{
			name:  "bold underline",
			style: TokenStyle{Color: "blue", Bold: true, Underline: true},
			want:  "\033[34;1;4m",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.style.Sequence()
			if got != tt.want {
				t.Errorf("Sequence() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInvalidTheme(t *testing.T) {
	// Test various validation failures
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			"no name",
			`tokens: {success: green}`,
			"theme name is required",
		},
		{
			"no tokens",
			`name: test`,
			"theme has no tokens",
		},
		{
			"empty tokens",
			`name: test
tokens: {}`,
			"theme has no tokens",
		},
		{
			"missing required",
			`name: test
tokens:
  success: green`,
			"missing required token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "test.yaml")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			err := Validate(path)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidateValidTheme(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "valid.yaml")
	content := `name: test
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
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Validate(path); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestLoadCustomTheme(t *testing.T) {
	// Set up custom themes dir
	home := t.TempDir()
	t.Setenv("HOME", home)
	themesDir := filepath.Join(home, ".config", "oc-color", "themes")
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `name: nord
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
	if err := os.WriteFile(filepath.Join(themesDir, "nord.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	th, ok := Get("nord")
	if !ok {
		t.Fatal("expected to find nord theme")
	}
	if th.Name != "nord" {
		t.Errorf("expected name 'nord', got %q", th.Name)
	}
}

func TestNamesIncludesCustom(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	themesDir := filepath.Join(home, ".config", "oc-color", "themes")
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `name: nord
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
	if err := os.WriteFile(filepath.Join(themesDir, "nord.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	names := Names()
	found := false
	for _, n := range names {
		if n == "nord" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'nord' in Names() = %v", names)
	}
}

func TestResetConstant(t *testing.T) {
	if Reset != "\033[0m" {
		t.Errorf("Reset = %q, want %q", Reset, "\033[0m")
	}
}

func TestHexTruecolor(t *testing.T) {
	result := hexTruecolor("#FF5555", true)
	expected := "38;2;255;85;85"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestNamedColor(t *testing.T) {
	if got := namedColor("red", true); got != "31" {
		t.Errorf("got %q, want %q", got, "31")
	}
	if got := namedColor("green", true); got != "32" {
		t.Errorf("got %q, want %q", got, "32")
	}
}

func TestIsNamedColor(t *testing.T) {
	if !isNamedColor("red") {
		t.Error("expected 'red' to be a valid named color")
	}
	if !isNamedColor("purple") {
		t.Error("expected 'purple' to be a valid named color via alias")
	}
	if isNamedColor("neon-chartreuse") {
		t.Error("expected nonsense to be invalid")
	}
}

func TestTokenStyleYAMLString(t *testing.T) {
	var ts TokenStyleYAML
	yamlContent := `"bold+red"`
	if err := yaml.Unmarshal([]byte(yamlContent), &ts); err != nil {
		t.Fatal(err)
	}
	if ts.Color != "red" || !ts.Bold {
		t.Error("expected bold+red")
	}
}

func TestTokenStyleYAMLStruct(t *testing.T) {
	var ts TokenStyleYAML
	yamlContent := `{color: "#50FA7B", bold: true}`
	if err := yaml.Unmarshal([]byte(yamlContent), &ts); err != nil {
		t.Fatal(err)
	}
	if ts.Color != "#50FA7B" || !ts.Bold {
		t.Error("expected color and bold")
	}
}
