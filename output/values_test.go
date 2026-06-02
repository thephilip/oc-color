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

func TestColorizeScalarValue_IPv4(t *testing.T) {
	th := testTheme()
	for _, ip := range []string{"192.168.1.1", "10.0.0.1", "172.16.0.0/16", "0.0.0.0"} {
		got := colorizeScalarValue(ip, th)
		want := wrapWithTheme(ip, "accent", th)
		if got != want {
			t.Errorf("colorizeScalarValue(%q) = %q, want %q", ip, got, want)
		}
	}
}

func TestColorizeScalarValue_Timestamp(t *testing.T) {
	th := testTheme()
	for _, ts := range []string{
		"2024-06-02",
		"2024-06-02T15:04:05Z",
		"2024-06-02T15:04:05.123456789Z",
		"2024-06-02T15:04:05+05:30",
		"2024-06-02T15:04:05-07:00",
	} {
		got := colorizeScalarValue(ts, th)
		want := wrapWithTheme(ts, "dim", th)
		if got != want {
			t.Errorf("colorizeScalarValue(%q) = %q, want %q", ts, got, want)
		}
	}
}

func TestColorizeScalarValue_NotIPOrTimestamp(t *testing.T) {
	th := testTheme()
	// These look vaguely numeric/date-like but should not match
	for _, s := range []string{"999.999.999.999", "not-a-date", "2024-13-01"} {
		got := colorizeScalarValue(s, th)
		if got != s {
			t.Errorf("colorizeScalarValue(%q) = %q, want unchanged", s, got)
		}
	}
}
