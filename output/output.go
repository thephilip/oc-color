package output

import (
	"regexp"
	"strings"

	"github.com/thephilip/oc-color/theme"
)

type column struct {
	name  string
	start int
}

type Processor struct {
	Theme   theme.Theme
	Colour  bool
	columns []column
}

type statusEntry struct {
	pattern *regexp.Regexp
	token   string
}

var statusPatterns []statusEntry

var (
	agePattern   = regexp.MustCompile(`\b[1-9]\d*[smhd]\b`)
	headerLine   = regexp.MustCompile(`^[A-Z][A-Z\s/]+$`)
	readyPattern = regexp.MustCompile(`(\d+)/(\d+)`)
	colSplit     = regexp.MustCompile(`  +`)
)

func init() {
	raw := map[string]string{
		`\bCrashLoopBackOff\b`:           "error",
		`\bImagePullBackOff\b`:           "error",
		`\bErrImagePull\b`:               "error",
		`\bErrImageNeverPull\b`:          "error",
		`\bOOMKill(ed)?\b`:               "error",
		`\bEvicted\b`:                    "error",
		`\bRunContainerError\b`:          "error",
		`\bCreateContainerConfigError\b`: "error",
		`\bInvalidImageName\b`:           "error",
		`\bNodeLost\b`:                   "error",
		`\bError\b`:                      "error",
		`\bFailed\b`:                     "error",
		`\bCancelled\b`:                  "error",
		`\bFailedBuild\b`:                "error",
		`\bReplicaFailure\b`:             "error",

		`\bPending\b`:           "warning",
		`\bContainerCreating\b`: "warning",
		`\bPodInitializing\b`:   "warning",
		`\bInit:\S+\b`:          "warning",
		`\bTerminating\b`:       "warning",
		`\bPreempting\b`:        "warning",
		`\bUnschedulable\b`:     "warning",
		`\bUnknown\b`:           "warning",
		`\bWarning\b`:           "warning",
		`\bNew\b`:               "warning",
		`\bNodeAffinity\b`:      "warning",

		`\bRunning\b`:     "success",
		`\bActive\b`:      "success",
		`\bSucceeded\b`:   "success",
		`\bComplete\b`:    "success",
		`\bCompleted\b`:   "success",
		`\bReady\b`:       "success",
		`\bTrue\b`:        "success",
		`\bAvailable\b`:   "success",
		`\bBound\b`:       "success",
		`\bCreated\b`:     "success",
		`\bEstablished\b`: "success",
	}
	for pat, token := range raw {
		statusPatterns = append(statusPatterns, statusEntry{
			pattern: regexp.MustCompile(pat),
			token:   token,
		})
	}
}

func (p *Processor) parseColumns(header string) {
	parts := colSplit.Split(header, -1)
	p.columns = p.columns[:0]
	searchPos := 0
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		idx := strings.Index(header[searchPos:], name)
		if idx < 0 {
			continue
		}
		start := searchPos + idx
		p.columns = append(p.columns, column{name: name, start: start})
		searchPos = start + len(name)
	}
}

func (p *Processor) Process(output string) string {
	if !p.Colour {
		return output
	}

	trimmed := strings.TrimSpace(output)
	if looksLikeJSON(trimmed) {
		return highlightJSON(output, p.Theme)
	}
	if looksLikeYAML(trimmed) {
		return highlightYAML(output, p.Theme)
	}
	if looksLikeDescribe(trimmed) {
		return p.processDescribe(output)
	}

	lines := strings.SplitAfter(output, "\n")
	var buf strings.Builder

	for _, line := range lines {
		buf.WriteString(p.processLine(line))
	}

	return buf.String()
}

func wrapWithTheme(text, token string, th theme.Theme) string {
	style, ok := th.Tokens[token]
	if !ok || style.Sequence() == "" {
		return text
	}
	return style.Sequence() + text + theme.Reset
}

func (p *Processor) processLine(line string) string {
	if line == "" {
		return line
	}

	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return line
	}

	if headerLine.MatchString(trimmed) {
		p.parseColumns(line)
		return p.applyStyle(line, "header")
	}

	if len(p.columns) >= 2 {
		return p.processTabularLine(line)
	}

	result := p.colorizeStatusWords(line)
	result = agePattern.ReplaceAllStringFunc(result, func(match string) string {
		return p.wrapStyle(match, "dim")
	})
	return result
}

func (p *Processor) processTabularLine(line string) string {
	if line == "" || len(p.columns) < 2 {
		return line
	}

	statusIdx := -1
	ageIdx := -1
	readyIdx := -1
	for i, col := range p.columns {
		switch strings.ToLower(col.name) {
		case "status":
			statusIdx = i
		case "age":
			ageIdx = i
		case "ready":
			readyIdx = i
		}
	}

	var result strings.Builder
	result.Grow(len(line) + 256)

	for i, col := range p.columns {
		if col.start >= len(line) {
			break
		}
		var colEnd int
		if i+1 < len(p.columns) {
			colEnd = p.columns[i+1].start
		} else {
			colEnd = len(line)
		}
		if colEnd > len(line) {
			colEnd = len(line)
		}
		if col.start >= colEnd {
			continue
		}

		raw := line[col.start:colEnd]
		trimmed := strings.TrimSpace(raw)

		leadingWS := len(raw) - len(strings.TrimLeft(raw, " "))
		prefix := raw[:leadingWS]
		body := raw[leadingWS:]

		var styled string
		if i == statusIdx && trimmed != "" {
			styled = strings.TrimRightFunc(p.colorizeStatusWords(body), isSpace)
		} else if i == ageIdx && trimmed != "" {
			styled = agePattern.ReplaceAllStringFunc(body, func(m string) string {
				return p.wrapStyle(m, "dim")
			})
		} else if i == readyIdx && trimmed != "" {
			styled = p.colorizeReady(body)
		} else {
			styled = body
		}

		result.WriteString(prefix)
		result.WriteString(styled)
	}

	return result.String()
}

func (p *Processor) colorizeReady(text string) string {
	return readyPattern.ReplaceAllStringFunc(text, func(match string) string {
		m := readyPattern.FindStringSubmatch(match)
		if len(m) != 3 {
			return match
		}
		if m[1] == "0" {
			return p.wrapStyle(match, "warning")
		}
		if m[1] == m[2] {
			return p.wrapStyle(match, "success")
		}
		return p.wrapStyle(match, "warning")
	})
}

func (p *Processor) colorizeStatusWords(text string) string {
	result := text
	for _, entry := range statusPatterns {
		result = entry.pattern.ReplaceAllStringFunc(result, func(match string) string {
			return p.wrapStyle(match, entry.token)
		})
	}
	return result
}

func (p *Processor) applyStyle(line, token string) string {
	style, ok := p.Theme.Tokens[token]
	if !ok || style.Sequence() == "" {
		return line
	}
	return style.Sequence() + line + theme.Reset
}

func (p *Processor) wrapStyle(text, token string) string {
	return wrapWithTheme(text, token, p.Theme)
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}
