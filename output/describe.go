package output

import (
	"regexp"
	"strings"
)

var (
	describeSectionHeader = regexp.MustCompile(`^(\s*)([\w][\w\s./_-]*?)(:\s*)$`)
	describeKeyVal        = regexp.MustCompile(`^(\s*)([\w][\w\s./_-]*?)(\s*:\s{2,})(.+)$`)
	describeDash          = regexp.MustCompile(`^(\s*)([-]+\s+)+[-]*$`)
	describeNone          = regexp.MustCompile(`<(none|not set)>`)
	describeEventNormal   = regexp.MustCompile(`^(\s+)(Normal)\b`)
	describeEventWarning  = regexp.MustCompile(`^(\s+)(Warning)\b`)
	describeFalse         = regexp.MustCompile(`\bFalse\b`)
)

func looksLikeDescribe(output string) bool {
	if output == "" {
		return false
	}
	lines := strings.SplitN(output, "\n", 8)
	kvCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if trimmed == "Conditions:" || strings.HasPrefix(trimmed, "Events:") {
			return true
		}
		if describeKeyVal.MatchString(trimmed) {
			kvCount++
			if kvCount >= 2 {
				return true
			}
		}
	}
	return false
}

func (p *Processor) processDescribe(output string) string {
	lines := strings.SplitAfter(output, "\n")
	var buf strings.Builder
	for _, line := range lines {
		buf.WriteString(p.processDescribeLine(line))
	}
	return buf.String()
}

func (p *Processor) processDescribeLine(line string) string {
	trimmed := strings.TrimRight(line, "\n\r")
	suffix := line[len(trimmed):]

	if trimmed == "" {
		return line
	}

	if describeDash.MatchString(trimmed) {
		return wrapWithTheme(trimmed, "dim", p.Theme) + suffix
	}

	if m := describeEventNormal.FindStringSubmatch(trimmed); m != nil {
		indent := m[1]
		etype := m[2]
		rest := trimmed[len(indent)+len(etype):]
		return indent + wrapWithTheme(etype, "info", p.Theme) + p.colorizeStatusWords(rest) + suffix
	}

	if m := describeEventWarning.FindStringSubmatch(trimmed); m != nil {
		indent := m[1]
		etype := m[2]
		rest := trimmed[len(indent)+len(etype):]
		return indent + wrapWithTheme(etype, "warning", p.Theme) + p.colorizeStatusWords(rest) + suffix
	}

	if describeSectionHeader.MatchString(trimmed) {
		m := describeSectionHeader.FindStringSubmatch(trimmed)
		return m[1] + wrapWithTheme(m[2], "header", p.Theme) + wrapWithTheme(m[3], "dim", p.Theme) + suffix
	}

	if m := describeKeyVal.FindStringSubmatch(trimmed); m != nil {
		indent := m[1]
		key := m[2]
		sep := m[3]
		val := m[4]

		if strings.Contains(val, "<none>") || strings.Contains(val, "<not set>") {
			val = describeNone.ReplaceAllStringFunc(val, func(m string) string {
				return wrapWithTheme(m, "dim", p.Theme)
			})
		}
		val = describeFalse.ReplaceAllStringFunc(val, func(m string) string {
			return wrapWithTheme(m, "error", p.Theme)
		})

		return indent + wrapWithTheme(key, "key", p.Theme) + wrapWithTheme(sep, "dim", p.Theme) + p.colorizeStatusWords(val) + suffix
	}

	if strings.Contains(trimmed, "<none>") || strings.Contains(trimmed, "<not set>") {
		result := describeNone.ReplaceAllStringFunc(trimmed, func(m string) string {
			return wrapWithTheme(m, "dim", p.Theme)
		})
		return p.colorizeStatusWords(result) + suffix
	}

	result := p.colorizeStatusWords(trimmed)
	result = describeFalse.ReplaceAllStringFunc(result, func(m string) string {
		return wrapWithTheme(m, "error", p.Theme)
	})
	return result + suffix
}
