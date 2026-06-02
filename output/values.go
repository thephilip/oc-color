package output

import (
	"regexp"
	"strings"

	"github.com/thephilip/oc-color/theme"
)

var (
	ipv4RE  = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])(/([0-9]|[1-2][0-9]|3[0-2]))?$`)
	stampRE = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])(T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2}))?$`)
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
	if ipv4RE.MatchString(trimmed) {
		return wrapWithTheme(trimmed, "accent", th)
	}
	if stampRE.MatchString(trimmed) {
		return wrapWithTheme(trimmed, "dim", th)
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
