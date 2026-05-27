package output

import (
	"strings"

	"github.com/thephilip/oc-color/theme"
)

func looksLikeJSON(output string) bool {
	trimmed := strings.TrimLeft(output, " \t\n\r")
	return strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")
}

func highlightJSON(input string, th theme.Theme) string {
	h := jsonHL{input: input, theme: th}
	return h.process()
}

type jsonHL struct {
	input string
	pos   int
	theme theme.Theme
}

func (h *jsonHL) process() string {
	var b strings.Builder
	b.Grow(len(h.input) + 256)
	for h.pos < len(h.input) {
		c := h.input[h.pos]
		switch {
		case c == '"':
			b.WriteString(h.readString())
		case c == '{' || c == '}' || c == '[' || c == ']':
			b.WriteString(wrapWithTheme(string(c), "pink", h.theme))
			h.pos++
		case c == ':':
			b.WriteString(wrapWithTheme(":", "dim", h.theme))
			h.pos++
		case c == ',':
			b.WriteByte(',')
			h.pos++
		case c == 't' && h.matchAhead("true"):
			b.WriteString(wrapWithTheme("true", "info", h.theme))
			h.pos += 4
		case c == 'f' && h.matchAhead("false"):
			b.WriteString(wrapWithTheme("false", "info", h.theme))
			h.pos += 5
		case c == 'n' && h.matchAhead("null"):
			b.WriteString(wrapWithTheme("null", "dim", h.theme))
			h.pos += 4
		case c == '-' || (c >= '0' && c <= '9'):
			b.WriteString(h.readNumber())
		case isWhitespace(c):
			b.WriteByte(c)
			h.pos++
		default:
			b.WriteByte(c)
			h.pos++
		}
	}
	return b.String()
}

func (h *jsonHL) readString() string {
	start := h.pos
	h.pos++

	for h.pos < len(h.input) {
		c := h.input[h.pos]
		h.pos++
		if c == '\\' && h.pos < len(h.input) {
			h.pos++
			continue
		}
		if c == '"' {
			break
		}
	}

	raw := h.input[start:h.pos]

	saved := h.pos
	h.skipWS()
	isKey := h.pos < len(h.input) && h.input[h.pos] == ':'
	h.pos = saved

	if isKey {
		return wrapWithTheme(raw, "key", h.theme)
	}
	return wrapWithTheme(raw, "success", h.theme)
}

func (h *jsonHL) readNumber() string {
	start := h.pos
	if h.input[h.pos] == '-' {
		h.pos++
	}
	for h.pos < len(h.input) && h.input[h.pos] >= '0' && h.input[h.pos] <= '9' {
		h.pos++
	}
	if h.pos < len(h.input) && h.input[h.pos] == '.' {
		h.pos++
		for h.pos < len(h.input) && h.input[h.pos] >= '0' && h.input[h.pos] <= '9' {
			h.pos++
		}
	}
	if h.pos < len(h.input) && (h.input[h.pos] == 'e' || h.input[h.pos] == 'E') {
		h.pos++
		if h.pos < len(h.input) && (h.input[h.pos] == '+' || h.input[h.pos] == '-') {
			h.pos++
		}
		for h.pos < len(h.input) && h.input[h.pos] >= '0' && h.input[h.pos] <= '9' {
			h.pos++
		}
	}
	return wrapWithTheme(h.input[start:h.pos], "accent", h.theme)
}

func (h *jsonHL) matchAhead(word string) bool {
	if h.pos+len(word) > len(h.input) {
		return false
	}
	if h.input[h.pos:h.pos+len(word)] != word {
		return false
	}
	if h.pos+len(word) < len(h.input) {
		next := h.input[h.pos+len(word)]
		if isWordChar(next) {
			return false
		}
	}
	return true
}

func (h *jsonHL) skipWS() {
	for h.pos < len(h.input) && isWhitespace(h.input[h.pos]) {
		h.pos++
	}
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}
