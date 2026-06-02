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
	return colorizeScalarValue(trimmed, h.theme)
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
