package output

import (
	"strings"
	"testing"

	"github.com/thephilip/oc-color/theme"
)

func shadeTheme() theme.Theme {
	return theme.Theme{
		Tokens: map[string]theme.TokenStyle{
			"shade": {Background: "#2E3040"},
		},
	}
}

func TestShadeResetAtLineEnd(t *testing.T) {
	proc := Processor{Theme: shadeTheme(), Colour: true, Shade: true}
	// Two data lines: lineNum=1 (no shade), lineNum=2 (shade)
	got := proc.Process("line one\nline two\n")
	lines := strings.Split(got, "\n")
	// lines[1] is the shaded line — must end with theme.Reset
	if !strings.HasSuffix(lines[1], theme.Reset) {
		t.Errorf("shaded line must end with Reset before newline, got: %q", lines[1])
	}
}

func TestShadeDisabledSkipsStripe(t *testing.T) {
	th := shadeTheme()
	proc := Processor{Theme: th, Colour: true, Shade: false}
	got := proc.Process("line one\nline two\n")
	shade := th.Tokens["shade"].BackgroundSequence()
	if strings.Contains(got, shade) {
		t.Errorf("shade must not appear when Processor.Shade is false, got: %q", got)
	}
}

func TestShadeNotAppliedToHeader(t *testing.T) {
	proc := Processor{Theme: shadeTheme(), Colour: true, Shade: true}
	// Header resets lineNum to 0; row1=lineNum1 (no shade), row2=lineNum2 (shade)
	got := proc.Process("NAME   STATUS\nfoo    Running\nbar    Pending\n")
	shade := shadeTheme().Tokens["shade"].BackgroundSequence()
	lines := strings.Split(got, "\n")
	// Header (lines[0]) must not contain shade background
	if strings.Contains(lines[0], shade) {
		t.Errorf("header must not be shaded, got: %q", lines[0])
	}
	// Row 2 (lines[2], lineNum=2) must contain shade
	if !strings.Contains(lines[2], shade) {
		t.Errorf("even data row must be shaded, got: %q", lines[2])
	}
}
