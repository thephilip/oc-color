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
	return colorizeScalarValue(trimmed, h.theme)
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
