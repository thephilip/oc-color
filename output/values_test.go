package output

import (
	"testing"
)

func TestColorizeScalarValue_Bool(t *testing.T) {
	th := testTheme()
	for _, kw := range []string{"true", "false", "yes", "no", "on", "off"} {
		got := colorizeScalarValue(kw, th)
		want := wrapWithTheme(kw, "info", th)
		if got != want {
			t.Errorf("colorizeScalarValue(%q) = %q, want %q", kw, got, want)
		}
	}
}

func TestColorizeScalarValue_Null(t *testing.T) {
	th := testTheme()
	for _, kw := range []string{"null", "~"} {
		got := colorizeScalarValue(kw, th)
		want := wrapWithTheme(kw, "dim", th)
		if got != want {
			t.Errorf("colorizeScalarValue(%q) = %q, want %q", kw, got, want)
		}
	}
}

func TestColorizeScalarValue_Number(t *testing.T) {
	th := testTheme()
	for _, n := range []string{"42", "3.14", "-7", "0"} {
		got := colorizeScalarValue(n, th)
		want := wrapWithTheme(n, "accent", th)
		if got != want {
			t.Errorf("colorizeScalarValue(%q) = %q, want %q", n, got, want)
		}
	}
}

func TestColorizeScalarValue_PlainUnchanged(t *testing.T) {
	th := testTheme()
	for _, s := range []string{"quay.io/openshift/origin-pod:latest", "my-pod", ""} {
		got := colorizeScalarValue(s, th)
		if got != s {
			t.Errorf("colorizeScalarValue(%q) = %q, want unchanged", s, got)
		}
	}
}
