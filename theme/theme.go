package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type TokenStyle struct {
	Color      string `yaml:"color"`
	Background string `yaml:"background"`
	Bold       bool   `yaml:"bold"`
	Dim        bool   `yaml:"dim"`
	Italic     bool   `yaml:"italic"`
	Underline  bool   `yaml:"underline"`
}

type Theme struct {
	Name   string
	Tokens map[string]TokenStyle
}

type TokenStyleYAML struct {
	TokenStyle
}

func (ts *TokenStyleYAML) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err == nil {
		ts.TokenStyle = parseShorthand(s)
		return nil
	}
	return value.Decode(&ts.TokenStyle)
}

type ThemeFile struct {
	Name   string                    `yaml:"name"`
	Tokens map[string]TokenStyleYAML `yaml:"tokens"`
}

func (tf *ThemeFile) ToTheme() Theme {
	t := Theme{Name: tf.Name, Tokens: make(map[string]TokenStyle, len(tf.Tokens))}
	for k, v := range tf.Tokens {
		t.Tokens[k] = v.TokenStyle
	}
	return t
}

var builtins map[string]Theme

func init() {
	builtins = map[string]Theme{
		"dracula": dracula(),
	}
}

func Get(name string) (Theme, bool) {
	name = strings.ToLower(name)
	if t, ok := builtins[name]; ok {
		return t, true
	}
	return loadCustom(name)
}

func loadCustom(name string) (Theme, bool) {
	dir, err := themesDir()
	if err != nil {
		return Theme{}, false
	}
	for _, ext := range []string{".yaml", ".yml"} {
		path := filepath.Join(dir, name+ext)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var tf ThemeFile
		if err := yaml.Unmarshal(data, &tf); err != nil {
			continue
		}
		if strings.ToLower(tf.Name) != name {
			continue
		}
		return tf.ToTheme(), true
	}
	return Theme{}, false
}

func Names() []string {
	seen := make(map[string]bool)
	for n := range builtins {
		seen[n] = true
	}
	dir, err := themesDir()
	if err == nil {
		entries, err := os.ReadDir(dir)
		if err == nil {
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				name := e.Name()
				if strings.HasSuffix(name, ".yaml") {
					name = strings.TrimSuffix(name, ".yaml")
				} else if strings.HasSuffix(name, ".yml") {
					name = strings.TrimSuffix(name, ".yml")
				} else {
					continue
				}
				seen[strings.ToLower(name)] = true
			}
		}
	}
	result := make([]string, 0, len(seen))
	for n := range seen {
		result = append(result, n)
	}
	sort.Strings(result)
	return result
}

var requiredTokens = []string{"success", "warning", "error", "info", "accent", "dim", "header", "key"}

func Validate(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read theme file: %w", err)
	}
	var tf ThemeFile
	if err := yaml.Unmarshal(data, &tf); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}
	if tf.Name == "" {
		return fmt.Errorf("theme name is required")
	}
	if len(tf.Tokens) == 0 {
		return fmt.Errorf("theme has no tokens defined")
	}
	for _, tok := range requiredTokens {
		if _, ok := tf.Tokens[tok]; !ok {
			return fmt.Errorf("missing required token %q", tok)
		}
	}
	for k, v := range tf.Tokens {
		if v.Color == "" {
			return fmt.Errorf("token %q has no color", k)
		}
		if strings.HasPrefix(v.Color, "#") {
			if len(v.Color) != 7 {
				return fmt.Errorf("token %q: invalid hex color %q (want #RRGGBB)", k, v.Color)
			}
			if _, err := strconv.ParseUint(v.Color[1:], 16, 64); err != nil {
				return fmt.Errorf("token %q: invalid hex color %q: %w", k, v.Color, err)
			}
		} else {
			if !isNamedColor(v.Color) {
				return fmt.Errorf("token %q: unknown color %q", k, v.Color)
			}
		}
	}
	return nil
}

func themesDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "oc-color", "themes"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "oc-color", "themes"), nil
}

func parseShorthand(s string) TokenStyle {
	parts := strings.Split(s, "+")
	ts := TokenStyle{}
	for _, p := range parts {
		switch strings.ToLower(p) {
		case "bold":
			ts.Bold = true
		case "dim":
			ts.Dim = true
		case "italic":
			ts.Italic = true
		case "underline":
			ts.Underline = true
		default:
			if ts.Color == "" {
				ts.Color = p
			}
		}
	}
	return ts
}

func (ts TokenStyle) Sequence() string {
	var parts []string

	if ts.Color != "" {
		switch ColorCapability {
		case CapTruecolor, Cap256:
			if strings.HasPrefix(ts.Color, "#") {
				parts = append(parts, hexTruecolor(ts.Color, true))
			} else {
				parts = append(parts, namedColor(ts.Color, true))
			}
		default:
			if strings.HasPrefix(ts.Color, "#") {
				parts = append(parts, hexToNamed(ts.Color, true))
			} else {
				parts = append(parts, namedColor(ts.Color, true))
			}
		}
	}

	if ts.Bold {
		parts = append(parts, "1")
	}
	if ts.Dim {
		parts = append(parts, "2")
	}
	if ts.Italic {
		parts = append(parts, "3")
	}
	if ts.Underline {
		parts = append(parts, "4")
	}

	if len(parts) == 0 {
		return ""
	}
	return "\033[" + strings.Join(parts, ";") + "m"
}

func (ts TokenStyle) BackgroundSequence() string {
	if ts.Background == "" {
		return ""
	}
	switch ColorCapability {
	case CapTruecolor, Cap256:
		if strings.HasPrefix(ts.Background, "#") {
			return "\033[" + hexTruecolor(ts.Background, false) + "m"
		}
		return "\033[" + namedColor(ts.Background, false) + "m"
	default:
		if strings.HasPrefix(ts.Background, "#") {
			return "\033[" + hexToNamed(ts.Background, false) + "m"
		}
		return "\033[" + namedColor(ts.Background, false) + "m"
	}
}

const Reset = "\033[0m"

func hexTruecolor(hex string, fg bool) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return ""
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	if fg {
		return fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
	}
	return fmt.Sprintf("48;2;%d;%d;%d", r, g, b)
}

func hexToNamed(hex string, fg bool) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return "39"
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return namedColor(closestNamedName(int(r), int(g), int(b)), fg)
}

var namedColorNames = []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}
var namedColorValues = [][3]int{
	{0, 0, 0},
	{170, 0, 0},
	{0, 170, 0},
	{170, 85, 0},
	{0, 0, 170},
	{170, 0, 170},
	{0, 170, 170},
	{170, 170, 170},
}

var namedAliases = map[string]string{
	"purple":  "magenta",
	"violet":  "magenta",
	"grey":    "white",
	"gray":    "white",
	"default": "39",
}

func closestNamedName(r, g, b int) string {
	best := "white"
	bestDist := 3 * 256 * 256
	for i, vals := range namedColorValues {
		dr := r - vals[0]
		dg := g - vals[1]
		db := b - vals[2]
		dist := dr*dr + dg*dg + db*db
		if dist < bestDist {
			bestDist = dist
			best = namedColorNames[i]
		}
	}
	return best
}

var namedColorMap = map[string]string{
	"black":   "30",
	"red":     "31",
	"green":   "32",
	"yellow":  "33",
	"blue":    "34",
	"magenta": "35",
	"purple":  "35",
	"cyan":    "36",
	"white":   "37",
}

func resolveNamed(name string) string {
	n := strings.ToLower(name)
	if c, ok := namedColorMap[n]; ok {
		return c
	}
	if alias, ok := namedAliases[n]; ok {
		if c, ok := namedColorMap[alias]; ok {
			return c
		}
	}
	return "39"
}

func isNamedColor(name string) bool {
	n := strings.ToLower(name)
	if _, ok := namedColorMap[n]; ok {
		return true
	}
	if _, ok := namedAliases[n]; ok {
		return true
	}
	return false
}

func namedColor(name string, fg bool) string {
	c := resolveNamed(name)
	if !fg {
		if c == "39" {
			return "49"
		}
		if len(c) == 2 && c[0] == '3' {
			return "4" + string(c[1])
		}
	}
	return c
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
