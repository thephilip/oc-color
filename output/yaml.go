package output

import (
	"strings"

	"github.com/thephilip/oc-color/theme"
)

func looksLikeYAML(output string) bool {
	trimmed := strings.TrimLeft(output, " \t\n\r")
	return strings.HasPrefix(trimmed, "---")
}

func highlightYAML(input string, th theme.Theme) string {
	lines := strings.SplitAfter(input, "\n")
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(highlightYAMLLine(line, th))
	}
	return b.String()
}

func highlightYAMLLine(line string, th theme.Theme) string {
	trimmed := strings.TrimRight(line, "\n\r")
	suffix := line[len(trimmed):]
	line = trimmed

	if strings.TrimSpace(line) == "" {
		return line + suffix
	}

	var indent string
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		indent += string(line[i])
		i++
	}
	line = line[i:]

	if line == "---" || line == "..." {
		return indent + wrapWithTheme(line, "pink", th) + suffix
	}

	if strings.HasPrefix(line, "#") {
		return indent + wrapWithTheme(line, "dim", th) + suffix
	}

	// List item
	if strings.HasPrefix(line, "- ") || line == "-" {
		rest := ""
		if len(line) > 2 {
			rest = line[2:]
		}
		return indent + wrapWithTheme("-", "pink", th) + " " + colorizeYAMLValue(rest, th) + suffix
	}

	// Key-value
	if idx := strings.Index(line, ":"); idx >= 0 {
		key := line[:idx]
		afterColon := line[idx+1:]

		if key != "" && !strings.Contains(key, " ") {
			colored := indent + wrapWithTheme(key, "key", th) + ":"

			trimmedVal := strings.TrimSpace(afterColon)
			if trimmedVal != "" {
				spaces := afterColon[:len(afterColon)-len(strings.TrimLeft(afterColon, " "))]
				colored += spaces + colorizeYAMLValue(trimmedVal, th)
			} else {
				colored += afterColon
			}
			return colored + suffix
		}
	}

	return indent + line + suffix
}

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
