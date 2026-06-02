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
	return colorizeScalarValue(trimmed, h.theme)
}

func indentLen(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return i
		}
	}
	return len(s)
}
