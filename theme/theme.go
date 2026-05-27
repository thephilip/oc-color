package theme

import (
	"fmt"
	"strconv"
	"strings"
)

type TokenStyle struct {
	Color     string `yaml:"color"`
	Bold      bool   `yaml:"bold"`
	Dim       bool   `yaml:"dim"`
	Italic    bool   `yaml:"italic"`
	Underline bool   `yaml:"underline"`
}

type Theme struct {
	Name   string
	Tokens map[string]TokenStyle
}

var builtins map[string]Theme

func init() {
	builtins = map[string]Theme{
		"dracula": dracula(),
	}
}

func Get(name string) (Theme, bool) {
	t, ok := builtins[strings.ToLower(name)]
	return t, ok
}

func Names() []string {
	names := make([]string, 0, len(builtins))
	for n := range builtins {
		names = append(names, n)
	}
	return names
}

func (ts TokenStyle) Sequence() string {
	var parts []string

	switch {
	case strings.HasPrefix(ts.Color, "#"):
		parts = append(parts, hexTruecolor(ts.Color, true))
	case ts.Color != "":
		parts = append(parts, namedColor(ts.Color, true))
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

func namedColor(name string, fg bool) string {
	named := map[string]string{
		"black":   "30",
		"red":     "31",
		"green":   "32",
		"yellow":  "33",
		"blue":    "34",
		"magenta": "35",
		"cyan":    "36",
		"white":   "37",
	}
	if c, ok := named[strings.ToLower(name)]; ok {
		return c
	}
	return "39"
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
			"header":  {Color: "#BD93F9", Bold: true, Underline: true},
			"key":     {Color: "#F1FA8C"},
			"value":   {Color: "#F8F8F2"},
		},
	}
}
