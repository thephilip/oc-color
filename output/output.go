package output

import (
	"regexp"
	"strings"

	"github.com/thephilip/oc-color/theme"
)

type Processor struct {
	Theme  theme.Theme
	Colour bool
}

type statusEntry struct {
	pattern *regexp.Regexp
	token   string
}

var statusPatterns []statusEntry

var (
	agePattern = regexp.MustCompile(`\b[1-9]\d*[smhd]\b`)
	headerLine = regexp.MustCompile(`^[A-Z][A-Z\s/]+$`)
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

func (p *Processor) Process(output string) string {
	if !p.Colour {
		return output
	}

	lines := strings.SplitAfter(output, "\n")
	var buf strings.Builder

	for _, line := range lines {
		buf.WriteString(p.processLine(line))
	}

	return buf.String()
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
		return p.applyStyle(line, "header")
	}

	result := p.colorizeStatusWords(line)
	result = agePattern.ReplaceAllStringFunc(result, func(match string) string {
		return p.wrapStyle(match, "dim")
	})
	return result
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
	style, ok := p.Theme.Tokens[token]
	if !ok || style.Sequence() == "" {
		return text
	}
	return style.Sequence() + text + theme.Reset
}
